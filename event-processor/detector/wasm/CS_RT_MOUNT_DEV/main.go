//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_MOUNT_DEV"
	Name        = "Mounting of devices from /dev"
	Description = "The detector detects the mount() system calls to mount devices from the /dev directory, which may indicate an attacker's attempt to gain access to the file system of the container host. The detector assigns low severity to unsuccessful attempts and high severity to successful ones."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeSuccessNoArgs  = "Detected that the `%s` device was successfully mounted on the mount point `%s` by the `%s` process"
	KprobeSuccessDefault = "Detected that the `%s` device was successfully mounted on the mount point `%s` by the `%s` process, which was started using the `%s` arguments"
	KprobeTryNoArgs      = "Detected a failed attempt to mount the `%s` device on the mount point `%s` made by the `%s` process"
	KprobeTryDefault     = "Detected a failed attempt to mount the `%s` device on the mount point `%s` made by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"sys_mount"},

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
				"T1611",
			},
		},
	}
)

var (
	devPattern = glob.MustCompile("/dev/*")
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
		functionName := kprobe.GetFunctionName()

		if strings.HasSuffix(functionName, "sys_mount") {
			ret := kprobe.GetReturn()
			r := ret.GetIntArg()

			args := kprobe.GetArgs()
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			a := args[0].GetStringArg()
			if !devPattern.Match(a) {
				return resp, nil
			}

			dev := args[0].GetStringArg()
			mountPoint := args[1].GetStringArg()

			if r == 0 {
				// According to mount() documentation 0 return value means success.
				resp.TacticsCovered = mitreTactics
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeSuccessNoArgs, dev, mountPoint, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeSuccessDefault, dev, mountPoint, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_HIGH // <-- threat detected
			} else {
				// Else, we assume that it was just an attempt, so mark it with low severity, but do not ignore.
				// This can be removed later if there will be a lot of false positives.
				resp.TacticsCovered = mitreTactics
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeTryNoArgs, dev, mountPoint, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeTryDefault, dev, mountPoint, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_LOW // <-- threat detected
			}

			return resp, nil
		}

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
