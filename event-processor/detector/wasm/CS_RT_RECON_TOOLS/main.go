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
	ID          = "CS_RT_RECON_TOOLS"
	Name        = "Use of network reconnaissance tools"
	Description = "The detector detects attempts to start applications that use the libpcap library. Such attempts may indicate use of network reconnaissance tools."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeNoArgs  = "Detected that the `%s` process was started and a library used for network reconnaissance was downloaded"
	KprobeDefault = "Detected that the `%s` process was started using the `%s` arguments and a library used for network reconnaissance was downloaded"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_file_permission", "security_mmap_file"},

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
			Id: "TA0007",
			Techniques: []string{
				"T1046",
			},
		},
	}
)

const (
	// File access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/linux/fs.h#L101
	MAY_READ = 4

	// Memory page access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L10
	PROT_READ = 1
)

var (
	libPath = glob.MustCompile("*/libpcap.so*")
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

		switch function {
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_READ == 0 {
				return resp, nil
			}

		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_READ {
				return resp, nil
			}

		default:
			return resp, nil
		}

		path = args[0].GetFileArg().GetPath()

		if libPath.Match(path) {
			resp.TacticsCovered = mitreTactics
			switch {
			case (function == "security_file_permission") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeNoArgs, binary)
			case (function == "security_file_permission") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDefault, binary, binaryArgs)
			case (function == "security_mmap_file") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeNoArgs, binary)
			case (function == "security_mmap_file") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDefault, binary, binaryArgs)
			}
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

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
