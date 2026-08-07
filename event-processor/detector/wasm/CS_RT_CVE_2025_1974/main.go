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
	ID          = "CS_RT_CVE_2025_1974"
	Name        = "Ingress Nightmare vulnerability (malicious code execution)"
	Description = "The detector detects if a shared library was loaded from the directory /tmp/nginx/client-body/, which may indicate that malicious code was loaded and executed to exploit the Ingress Nightmare vulnerability (CVE-2025-1974)."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeDetect = "Detected that arbitrary code was executed by the `%s` process, which was started using the `%s` arguments through the Ingress Nightmare vulnerability exploitation"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_mmap_file"},

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
				"T1190",
			},
		},
		{
			Id: "TA0002",
			Techniques: []string{
				"T1129",
			},
		},
		{
			Id: "TA0008",
			Techniques: []string{
				"T1210",
			},
		},
	}
)

const (
	// Memory page access permissions.
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L10
	PROT_READ = 1
)

var (
	// NGINX binary.
	nginxBinary = glob.MustCompile("*/nginx")

	// Path to request body buffered by NGINX.
	libPath = glob.MustCompile("/tmp/nginx/client-body/*")
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
		// Trigger when security function check for memory page write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_READ == 0 {
				return resp, nil
			}

		default:
			return resp, nil
		}

		path = args[0].GetFileArg().GetPath()

		if nginxBinary.Match(binary) && libPath.Match(path) {
			resp.TacticsCovered = mitreTactics
			resp.Reason = fmt.Sprintf(KprobeDetect, path, binary, binaryArgs)
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
