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
	ID          = "CS_RT_CVE_2022_0492"
	Name        = "Container isolation vulnerability exploitation (CVE-2022-0492)"
	Description = "The detector detects attempts to modify the notify_on_release and release_agent files, which may indicate an attempt to exploit vulnerability CVE-2022-0492. An attacker can exploit this vulnerability to escalate privileges and break out from the isolated container environment."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeWriteNoArgs   = "Detected that the `%s` file was edited by the `%s` process, which may indicate an attempt to exploit vulnerability CVE-2022-0492"
	KprobeWriteDefault  = "Detected that the `%s` file was edited by the `%s` process (started using the `%s` arguments), which may indicate an attempt to exploit vulnerability CVE-2022-0492"
	KprobeMmapNoArgs    = "Detected that the `%s` file was memory-mapped by the `%s` process, which may indicate an attempt to exploit vulnerability CVE-2022-0492"
	KprobeMmapDefault   = "Detected that the `%s` file was memory-mapped by the `%s` process (started using the `%s` arguments), which may indicate an attempt to exploit vulnerability CVE-2022-0492"
	KprobeRenameNoArgs  = "Detected that the `%s` file was replaced with the `%s` file by the `%s` process, which may indicate an attempt to exploit vulnerability CVE-2022-0492"
	KprobeRenameDefault = "Detected that the `%s` file was replaced with the `%s` file by the `%s` process (started using the `%s` arguments), which may indicate an attempt to exploit vulnerability CVE-2022-0492"
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
			Id: "TA0004",
			Techniques: []string{
				"T1611",
			},
		},
	}
)

var (
	fileRegex = regexp.MustCompile(`(?:release_agent|notify_on_release)$`)
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

		// The first 15 digits of the container ID
		containerIDPart := kprobe.GetProcess().GetDocker()

		// Empty containerIDPart means that we are not in container environment (vanilla Docker or k8s)
		if containerIDPart == "" {
			return resp, nil
		}

		switch function {
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != 2 { // need MAY_WRITE
				return resp, nil
			}

			action = "write"
			path = args[0].GetFileArg().GetPath()

		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&2 != 2 { // need PROT_WRITE
				return resp, nil
			}

			action = "mmap"
			path = args[0].GetFileArg().GetPath()

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

		if fileRegex.MatchString(path) {
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
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

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
