//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_SSH_TUNNEL_CREATE"
	Name        = "SSH tunnel creation"
	Description = "The detector detects attempts to create a local, remote, or dynamic network tunnel using the SSH service, which may indicate that an attacker tried to achieve persistence on the system, bypass existing network restrictions, or hide their activity."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecLocal   = "Detected that a local SSH tunnel was created by the `%s` process, which was started using the `%s` arguments"
	ExecRemote  = "Detected that a remote SSH tunnel was created by the `%s` process, which was started using the `%s` arguments"
	ExecDynamic = "Detected that a dynamic SSH tunnel was created by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_EXEC": {},

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
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1133",
			},
		},
		{
			Id: "TA0008",
			Techniques: []string{
				"T1021.004",
			},
		},
		{
			Id: "TA0011",
			Techniques: []string{
				"T1572",
			},
		},
	}
)

var (
	sshBin = glob.MustCompile("*/ssh")

	localForwardRegex   = regexp.MustCompile(`\s*-[^-\s]*L\s+`)
	remoteForwardRegex  = regexp.MustCompile(`\s*-[^-\s]*R\s+`)
	dynamicForwardRegex = regexp.MustCompile(`\s*-[^-\s]*D\s+`)
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
		process := ev.ProcessExec
		binary := process.GetProcess().GetBinary()
		args := process.GetProcess().GetArguments()

		switch {
		case sshBin.Match(binary) && localForwardRegex.MatchString(args):
			resp.TacticsCovered = mitreTactics
			resp.Reason = fmt.Sprintf(ExecLocal, binary, args)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		case sshBin.Match(binary) && remoteForwardRegex.MatchString(args):
			resp.TacticsCovered = mitreTactics
			resp.Reason = fmt.Sprintf(ExecRemote, binary, args)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		case sshBin.Match(binary) && dynamicForwardRegex.MatchString(args):
			resp.TacticsCovered = mitreTactics
			resp.Reason = fmt.Sprintf(ExecDynamic, binary, args)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		default:
			return resp, nil
		}

	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		// Nothing here
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
