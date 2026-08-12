//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_PROCESS_CAP_RAISE"
	Name        = "Privilege escalation: capabilities"
	Description = "This detector detects processes running with excessive capabilities, which may indicate an attacker's attempt to escalate their privileges in the system."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeEffectiveNoArgs  = "Detected that the `%s` process was started with the following excessive effective capabilities: `%s`"
	KprobeEffectiveDefault = "Detected that the `%s` process was started using the `%s` arguments and with the following excessive effective capabilities: `%s`"
	KprobePermittedNoArgs  = "Detected that the `%s` process was started with the following excessive permitted capabilities: `%s`"
	KprobePermittedDefault = "Detected that the `%s` process was started using the `%s` arguments and with the following excessive permitted capabilities: `%s`"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
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
				"T1543",
			},
		},
	}
)

var (
	excessiveCapabilities = []tetragon.CapabilitiesType{
		tetragon.CapabilitiesType_CAP_DAC_READ_SEARCH,
		tetragon.CapabilitiesType_CAP_SYS_ADMIN,
		tetragon.CapabilitiesType_CAP_SYS_MODULE,
		tetragon.CapabilitiesType_CAP_SYS_RAWIO,
		tetragon.CapabilitiesType_CAP_NET_ADMIN,
		tetragon.CapabilitiesType_CAP_SYS_CHROOT,
		tetragon.CapabilitiesType_CAP_SYS_PTRACE,
		tetragon.CapabilitiesType_CAP_NET_RAW,
		tetragon.CapabilitiesType_CAP_SYS_BOOT,
		tetragon.CapabilitiesType_CAP_SYSLOG,
	}
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
		proc := kprobe.GetProcess()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		if function != "commit_creds" {
			return resp, nil
		}

		if len(args) < 1 {
			return nil, fmt.Errorf("unexpected args len. got %d, expected >= 1", len(args))
		}

		committedCreds := args[0].GetProcessCredentialsArg()

		procEffectiveCaps := proc.GetCap().GetEffective()
		procPermittedCaps := proc.GetCap().GetPermitted()

		effectiveCapsMatch := false
		effectiveCapsList := ""

		for _, capability := range committedCreds.GetCaps().GetEffective() {
			if slices.Contains(excessiveCapabilities, capability) && !slices.Contains(procEffectiveCaps, capability) && !effectiveCapsMatch {
				effectiveCapsList += tetragon.CapabilitiesType_name[int32(capability)]
				effectiveCapsMatch = true
			} else if slices.Contains(excessiveCapabilities, capability) && !slices.Contains(procEffectiveCaps, capability) && effectiveCapsMatch {
				effectiveCapsList += ", " + tetragon.CapabilitiesType_name[int32(capability)]
			}
		}

		if effectiveCapsMatch {
			resp.TacticsCovered = mitreTactics
			if binaryArgs == "" {
				resp.Reason = fmt.Sprintf(KprobeEffectiveNoArgs, binary, effectiveCapsList)
			} else {
				resp.Reason = fmt.Sprintf(KprobeEffectiveDefault, binary, binaryArgs, effectiveCapsList)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

			return resp, nil
		}

		permittedCapsMatch := false
		permittedCapsList := ""

		for _, capability := range committedCreds.GetCaps().GetPermitted() {
			if slices.Contains(excessiveCapabilities, capability) && !slices.Contains(procPermittedCaps, capability) && !permittedCapsMatch {
				permittedCapsList += tetragon.CapabilitiesType_name[int32(capability)]
				permittedCapsMatch = true
			} else if slices.Contains(excessiveCapabilities, capability) && !slices.Contains(procPermittedCaps, capability) && permittedCapsMatch {
				permittedCapsList += ", " + tetragon.CapabilitiesType_name[int32(capability)]
			}
		}

		if permittedCapsMatch {
			resp.TacticsCovered = mitreTactics
			if binaryArgs == "" {
				resp.Reason = fmt.Sprintf(KprobePermittedNoArgs, binary, permittedCapsList)
			} else {
				resp.Reason = fmt.Sprintf(KprobePermittedDefault, binary, binaryArgs, permittedCapsList)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

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
