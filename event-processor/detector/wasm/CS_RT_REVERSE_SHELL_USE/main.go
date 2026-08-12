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
	ID          = "CS_RT_REVERSE_SHELL_USE"
	Name        = "Reverse shell use"
	Description = "The detector detects if data was sent via a network connection by utilities that are not designed for this. This behavior may indicate use of a reverse shell."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeUtilsNoArgs   = "Detected that data was sent to the host with IP address `%s` and port `%d` by the `%s` utility"
	KprobeUtilsDefault  = "Detected that data was sent to the host with IP address `%s` and port `%d` by the `%s` utility, which was started using the `%s` arguments"
	KprobeShellsNoArgs  = "Detected that data was sent to the host with IP address `%s` and port `%d` by the `%s` command shell"
	KprobeShellsDefault = "Detected that data was sent to the host with IP address `%s` and port `%d` by the `%s` command shell, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"tcp_sendmsg"},

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
				"T1059.004",
				"T1059.006",
			},
		},
	}
)

var (
	commonLinuxTools = []glob.Glob{
		glob.MustCompile("*/ls"),
		glob.MustCompile("*/cp"),
		glob.MustCompile("*/mv"),
		glob.MustCompile("*/chmod"),
		glob.MustCompile("*/chown"),
		glob.MustCompile("*/rmdir"),
		glob.MustCompile("*/ln"),
		glob.MustCompile("*/df"),
		glob.MustCompile("*/du"),
		glob.MustCompile("*/cat"),
		glob.MustCompile("*/[a-z]grep"),
		glob.MustCompile("*/id"),
		glob.MustCompile("*/whoami"),
		glob.MustCompile("*/uname"),
		glob.MustCompile("*/ps"),
		glob.MustCompile("*/top"),
		glob.MustCompile("*/make"),
		glob.MustCompile("*/gcc"),
		glob.MustCompile("*/g++"),
	}

	shells = []glob.Glob{
		glob.MustCompile("*/ash"),
		glob.MustCompile("*/bash"),
		glob.MustCompile("*/csh"),
		glob.MustCompile("*/dash"),
		glob.MustCompile("*/ksh"),
		glob.MustCompile("*/sh"),
		glob.MustCompile("*/tcsh"),
		glob.MustCompile("*/zsh"),
		glob.MustCompile("*/pwsh"),
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
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		if function != "tcp_sendmsg" {
			return resp, nil
		}

		dstIp := args[0].GetSockArg().GetDaddr()
		dstPort := args[0].GetSockArg().GetDport()

		for _, tool := range commonLinuxTools {
			if tool.Match(binary) {
				resp.TacticsCovered = mitreTactics
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeUtilsNoArgs, dstIp, dstPort, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeUtilsDefault, dstIp, dstPort, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_HIGH

				return resp, nil // <-- threat detected
			}
		}

		for _, shell := range shells {
			if shell.Match(binary) {
				resp.TacticsCovered = mitreTactics
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeShellsNoArgs, dstIp, dstPort, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeShellsDefault, dstIp, dstPort, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_HIGH

				return resp, nil // <-- threat detected
			}
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
