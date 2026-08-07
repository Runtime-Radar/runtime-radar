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
	ID          = "CS_RT_RAW_SOCKETS"
	Name        = "Creation of raw network socket"
	Description = "The detector detects if a raw network socket was created, which may indicate network traffic interception or reconnaissance."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeNoArgs  = "Detected that the special `%s` socket was created by the `%s` process"
	KprobeDefault = "Detected that the special `%s` socket was created by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_socket_create"},

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
			Id: "TA0003",
			Techniques: []string{
				"T1205.002",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1205.002",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1040",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1046",
				"T1040",
			},
		},
		{
			Id: "TA0011",
			Techniques: []string{
				"T1205.002",
			},
		},
	}
)

const (
	// Socket Address Family: https://elixir.bootlin.com/linux/latest/source/include/linux/socket.h#L188
	AF_INET  = 2
	AF_INET6 = 10

	// Socket Type: https://elixir.bootlin.com/linux/latest/source/include/linux/net.h#L64
	SOCK_RAW    = 3
	SOCK_PACKET = 10
)

var (
	binaryWhitelist = []*regexp.Regexp{
		regexp.MustCompile(`/NetworkManager$`),
		regexp.MustCompile(`/dhclient$`),
		regexp.MustCompile(`/f?ping6?$`),
		regexp.MustCompile(`/traceroute6?$`),
		regexp.MustCompile(`/(?:ip|x)tables.*$`),
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

		if function != "security_socket_create" {
			return resp, nil
		}

		// Exclude whitelisted binary from detect logic
		for _, bin := range binaryWhitelist {
			if bin.MatchString(binary) {
				return resp, nil
			}
		}

		args := kprobe.GetArgs()

		if len(args) < 2 {
			return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
		}

		socketFamily := args[0].GetIntArg()
		socketType := args[1].GetIntArg()

		// Regular RAW sockets
		if (socketFamily == AF_INET || socketFamily == AF_INET6) && (socketType == SOCK_RAW || socketType == SOCK_PACKET) {
			resp.TacticsCovered = mitreTactics
			socketTypeSymbolic := ""
			switch socketType {
			case 3:
				socketTypeSymbolic = "SOCK_RAW"
			case 10:
				socketTypeSymbolic = "SOCK_PACKET"
			}
			if binaryArgs == "" {
				resp.Reason = fmt.Sprintf(KprobeNoArgs, socketTypeSymbolic, binary)
			} else {
				resp.Reason = fmt.Sprintf(KprobeDefault, socketTypeSymbolic, binary, binaryArgs)
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
