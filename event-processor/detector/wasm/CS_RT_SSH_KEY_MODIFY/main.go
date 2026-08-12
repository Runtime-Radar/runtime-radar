//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_SSH_KEY_MODIFY"
	Name        = "Suspicious change of SSH keys"
	Description = "The detector detects suspicious changes in files that contain SSH keys, which may indicate that an existing account was compromised or that an attacker attempted to achieve persistence on the system."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeWriteNoArgs   = "Detected that the `%s` file with SSH keys was edited by the `%s` process"
	KprobeWriteDefault  = "Detected that the `%s` file with SSH keys was edited by the `%s` process, which was started using the `%s` arguments"
	KprobeMmapNoArgs    = "Detected that the `%s` file with SSH keys was memory-mapped by the `%s` process"
	KprobeMmapDefault   = "Detected that the `%s` file with SSH keys was memory-mapped by the `%s` process, which was started using the `%s` arguments"
	KprobeRenameNoArgs  = "Detected that the `%s` file with SSH keys was replaced with the `%s` file by the `%s` process"
	KprobeRenameDefault = "Detected that the `%s` file with SSH keys was replaced with the `%s` file by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate", "security_path_rename"},

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
			Id: "TA0001",
			Techniques: []string{
				"T1133",
				"T1078.003",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1133",
				"T1078.003",
				"T1098.004",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1078.003",
				"T1098.004",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1078.003",
			},
		},
	}
)

const (
	// File access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/linux/fs.h#L100
	MAY_WRITE = 2

	// Memory page access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L11
	PROT_WRITE = 2
)

var (
	sshKeysPath    = regexp.MustCompile(`^(/root/|/home/.*/)\.ssh/`)
	filesToExclude = regexp.MustCompile(`\.ssh2?/(config|known_hosts)$`)
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
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		binaryArgs := kprobe.GetProcess().GetArguments()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()
		path := ""
		newFile := ""
		action := ""

		switch function {
		// Trigger when security function check for file write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_WRITE {
				return resp, nil
			}

			action = "write"
			path = args[0].GetFileArg().GetPath()

		// Trigger when security function check for memory page write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_WRITE == 0 {
				return resp, nil
			}

			action = "mmap"
			path = args[0].GetFileArg().GetPath()

		// Trigger when security function check if truncating a file is allowed.
		// https://elixir.bootlin.com/linux/v6.10.6/source/security/security.c#L1923
		case "security_path_truncate":
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			action = "write"
			path = args[0].GetPathArg().GetPath()

		// Trigger when security function check if renaming a file is allowed.
		// https://elixir.bootlin.com/linux/v6.15.7/source/security/security.c#L2005
		case "security_path_rename":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			}

			action = "rename"
			path = args[1].GetPathArg().GetPath()
			newFile = args[0].GetPathArg().GetPath()

		default:
			return resp, nil
		}

		if sshKeysPath.MatchString(path) && !filesToExclude.MatchString(path) {
			resp.TacticsCovered = mitreTactics
			switch {
			case (action == "write") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeWriteNoArgs, path, binary)
			case (action == "write") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeWriteDefault, path, binary, binaryArgs)
			case (action == "mmap") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeMmapNoArgs, path, binary)
			case (action == "mmap") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeMmapDefault, path, binary, binaryArgs)
			case (action == "rename") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeRenameNoArgs, path, newFile, binary)
			case (action == "rename") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeRenameDefault, path, newFile, binary, binaryArgs)
			}
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
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
