//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_PRIV_ESC"
	Name        = "Privilege escalation"
	Description = "The detector detects if a root user started a process or if a process called the commit_creds function with the UID/EUID == 0 or GID/EGID == 0 parameter to gain root privileges, which may indicate an attacker's attempt to escalate their privileges."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecNoArgs    = "Detected that the `%s` process escalated privileges for the root user"
	ExecDefault   = "Detected privilege escalation for the root user by the `%s` process, which was started using the `%s` arguments"
	KprobeNoArgs  = "Detected that the `%s` process escalated privileges to the root level"
	KprobeDefault = "Detected privilege escalation to the root level by the `%s` process, which was started using the `%s` arguments"
)

const (
	// UID of superuser
	rootUID = 0
	// GID of superuser
	rootGID = 0
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_EXEC":   {},
		"PROCESS_KPROBE": {"commit_creds"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)

var (
	mitreTactics = []*api.MitreTactic{
		{
			Id: "TA0004",
			Techniques: []string{
				"T1548.001",
				"T1543",
				"T1068",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1548.001",
			},
		},
	}

	mitreTacticsSuidSgid = []*api.MitreTactic{
		{
			Id: "TA0004",
			Techniques: []string{
				"T1548.001",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1548.001",
			},
		},
	}

	mitreTacticsPrivEsc = []*api.MitreTactic{
		{
			Id: "TA0004",
			Techniques: []string{
				"T1543",
				"T1068",
			},
		},
	}
)

var (
	runcInitBinPattern = glob.MustCompile("/dev/fd/*")
)

// init is required for TinyGo to compile to Wasm.
func init() {
	api.RegisterDetector(Detector{})
}

type Detector struct{}

func (d Detector) Info(ctx context.Context, req *api.InfoReq) (*api.InfoResp, error) {
	return &api.InfoResp{
		Id:             ID,
		Name:           Name,
		Description:    Description,
		Version:        Version,
		Author:         Author,
		License:        License,
		TacticsCovered: mitreTactics,
	}, nil
}

func (d Detector) TriggerCriteria(ctx context.Context, req *api.TriggerCriteriaReq) (*api.TriggerCriteriaResp, error) {
	resp := &api.TriggerCriteriaResp{
		Criteria: make(map[string]*api.TriggerCriteriaResp_FuncNames, len(triggerCriteria)),
	}

	for k, v := range triggerCriteria {
		resp.Criteria[k] = &api.TriggerCriteriaResp_FuncNames{FuncNames: v}
	}

	return resp, nil
}

func (d Detector) Detect(ctx context.Context, req *api.DetectReq) (*api.DetectResp, error) {
	// Detector info added to DetectResp because detector info is always correlated to response, thus
	// to avoid +1 Wasm call on detect.
	resp := &api.DetectResp{
		// Default response indicates that nothing detected (this is redundant and put here just for reference,
		// as Severity == api.DetectResp_NONE == 0 when omitted (default zero value)).
		Severity: api.DetectResp_NONE,
	}

	event := req.GetEvent().GetEvent()

	switch ev := event.(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		exec := ev.ProcessExec
		binary := exec.GetProcess().GetBinary()
		args := exec.GetProcess().GetArguments()
		parentBinary := exec.GetParent().GetBinary()
		parentArgs := exec.GetParent().GetArguments()

		// exclude runc init container
		if runcInitBinPattern.Match(binary) && (args == "init") && (parentBinary == "/proc/self/exe") && (parentArgs == "init") {
			return resp, nil
		}

		// trigger on binary with setuid bit
		if suid := exec.GetProcess().GetBinaryProperties().GetSetuid(); suid != nil {
			if suid.Value == rootUID {
				resp.TacticsCovered = mitreTacticsSuidSgid
				resp.Reason = fmt.Sprintf(ExecNoArgs, binary)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

		// trigger on binary with setgid bit
		if sgid := exec.GetProcess().GetBinaryProperties().GetSetgid(); sgid != nil {
			if sgid.Value == rootGID {
				resp.TacticsCovered = mitreTacticsSuidSgid
				resp.Reason = fmt.Sprintf(ExecDefault, binary)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		binaryArgs := kprobe.GetProcess().GetArguments()
		functionName := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		if functionName != "commit_creds" {
			return resp, nil // <-- return
		}

		committedCreds := args[0].GetProcessCredentialsArg()

		if committedCreds == nil {
			return resp, nil // <-- return
		}

		// We are taking into account only processes which are running from regular user
		if kprobe.GetProcess().GetUid().GetValue() != rootUID {
			// Checking EUID will also cover cases with setuid binary
			if (committedCreds.GetUid().GetValue() == rootUID) || (committedCreds.GetEuid().GetValue() == rootUID) {
				resp.TacticsCovered = mitreTacticsPrivEsc
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeNoArgs, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeDefault, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_CRITICAL // <-- threat detected

				return resp, nil // <-- return
			}
		}

		processCreds := kprobe.GetProcess().GetProcessCredentials()

		if processCreds == nil {
			return resp, nil // <-- return
		}

		if processCreds.GetGid().GetValue() != rootGID {
			// Checking EGID will also cover cases with setgid binary
			if (committedCreds.GetGid().GetValue() == rootGID) || (committedCreds.GetEgid().GetValue() == rootGID) {
				resp.TacticsCovered = mitreTacticsPrivEsc
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeNoArgs, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeDefault, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_CRITICAL // <-- threat detected

				return resp, nil // <-- return
			}
		}

		return resp, nil

	case *tetragon.GetEventsResponse_ProcessTracepoint:
		// Nothing here
	}

	return resp, nil
}

// tacticTechniques is a constructor for *api.MitreTactic which makes its initialization less verbose.
func tacticTechniques(tacticID string, techniqueIDs ...string) *api.MitreTactic {
	return &api.MitreTactic{
		Id:         tacticID,
		Techniques: techniqueIDs,
	}
}
