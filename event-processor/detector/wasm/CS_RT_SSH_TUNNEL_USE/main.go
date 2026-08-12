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
	ID          = "CS_RT_SSH_TUNNEL_USE"
	Name        = "Use of SSH tunnel"
	Description = "The detector detects network activity related to network tunneling of a local SSH service. The activity may indicate that an attacker tried to achieve persistence on the system, bypass existing network restrictions, or hide their activity."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeLocal  = "Detected a connection to a host with IP address `%s` and port `%d` initiated by the SSH server `%s`, which was started using the `%s` arguments. This may indicate that a local tunnel was used."
	KprobeRemote = "Detected that port `%d` was listened to by the SSH server `%s`, which was started using the `%s` arguments. This may indicate that a remote tunnel was used."
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"tcp_connect", "inet_csk_listen_start"},

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

const (
	sshdPort = 22
)

var (
	sshdBin = glob.MustCompile("*/sshd")
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

		if !sshdBin.Match(binary) {
			return resp, nil
		}

		switch function {
		// If sshd process is trying to connect to some host, than someone is using a ssh-tunnel created via local forwarding
		case "tcp_connect":
			resp.TacticsCovered = mitreTactics
			args := kprobe.GetArgs()
			dstIp := args[0].GetSockArg().GetDaddr()
			dstPort := args[0].GetSockArg().GetDport()
			resp.Reason = fmt.Sprintf(KprobeLocal, dstIp, dstPort, binary, binaryArgs)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil

		// If sshd process is trying to open not a tcp/22 port, than someone is opening a ssh-tunnel via remote forwarding
		case "inet_csk_listen_start":
			args := kprobe.GetArgs()

			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			socket := args[0].GetSockArg()
			sport := socket.GetSport()

			if sport != sshdPort {
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(KprobeRemote, sport, binary, binaryArgs)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

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
