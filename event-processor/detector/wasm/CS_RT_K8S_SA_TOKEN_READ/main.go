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
	ID          = "CS_RT_K8S_SA_TOKEN_READ"
	Name        = "Reading of Kubernetes authentication token"
	Description = "The detector detects reading of the Kubernetes authentication token, which may indicate that the Kubernetes account was compromised."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeReadNoArgs  = "Detected that the Kubernetes token `%s` was read by the `%s` process"
	KprobeReadDefault = "Detected that the Kubernetes token `%s` was read by the `%s` process, which was started using the `%s` arguments"
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
			Id: "TA0001",
			Techniques: []string{
				"T1078.001",
				"T1078.003",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1078.001",
				"T1078.003",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1078.001",
				"T1078.003",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1550.001",
				"T1078.001",
				"T1078.003",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1528",
				"T1552.001",
			},
		},
		{
			Id: "TA0008",
			Techniques: []string{
				"T1550.001",
			},
		},
	}
)

var (
	// Following code is commented and will probably be removed later on, because listed tools run on host itself, out of container context,
	// and will be filtered out by monitoring agent most of the time because of runtime filters.
	// The names of tools coded here can also be used by an attacker to avoid detection.
	//
	// legitUtilPatterns contains list of utils which are allowed to read ServiceAccount token.
	// legitUtilPatterns = []glob.Glob{
	// 	glob.MustCompile("*/flanneld"),
	// 	glob.MustCompile("*/kube-proxy"),
	// 	glob.MustCompile("*/etcd"),
	// 	glob.MustCompile("*/kube-apiserver"),
	// 	glob.MustCompile("*/coredns"),
	// 	glob.MustCompile("*/kube-controller"),
	// 	glob.MustCompile("*/kubectl"),
	// 	glob.MustCompile("*/storage-provisioner"),
	// }

	// fileRegex is a regex for path to a file where ServiceAccount token is stored.
	fileRegex = regexp.MustCompile("secrets/kubernetes.io/serviceaccount.+token$")
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

		containerIDPart := kprobe.GetProcess().GetDocker()

		// Process events only from runtime env
		if containerIDPart == "" {
			return resp, nil
		}

		switch function {
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != 4 { // need MAY_READ
				return resp, nil
			}
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&1 != 1 { // need PROT_READ
				return resp, nil
			}
		default:
			return resp, nil
		}

		// This code commented intentionally, see the comment above.
		// for _, p := range legitUtilPatterns {
		// 	if p.Match(binary) {
		// 		return resp, nil
		// 	}
		// }

		path := args[0].GetFileArg().GetPath()

		if fileRegex.MatchString(path) {
			resp.TacticsCovered = mitreTactics
			switch {
			case binaryArgs == "":
				resp.Reason = fmt.Sprintf(KprobeReadNoArgs, path, binary)
			case binaryArgs != "":
				resp.Reason = fmt.Sprintf(KprobeReadDefault, path, binary, binaryArgs)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected
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
