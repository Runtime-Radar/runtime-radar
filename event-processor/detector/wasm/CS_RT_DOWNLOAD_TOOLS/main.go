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
	ID          = "CS_RT_DOWNLOAD_TOOLS"
	Name        = "Network activity: file downloading"
	Description = "The detector detects network activity of utilities used for file transferring and downloading. Downloaded files can be malicious tools, payloads, or data used by an adversary to prepare for further actions."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	KprobeNoArgs  = "Detected that data was downloaded from a node available at IP address `%s` and port `%d` by the `%s` process"
	KprobeDefault = "Detected that data was downloaded from a node available at IP address `%s` and port `%d` by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"tcp_connect"},

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
			Id: "TA0008",
			Techniques: []string{
				"T1570",
			},
		},
		{
			Id: "TA0010",
			Techniques: []string{
				"T1048.001",
				"T1048.002",
				"T1048.003",
			},
		},
		{
			Id: "TA0011",
			Techniques: []string{
				"T1105",
			},
		},
	}
)

// downloadTool represents a combination of binary's and its arguments' patterns.
type downloadTool struct {
	binaryPattern glob.Glob
	argsPatterns  []*regexp.Regexp
}

var downloadTools = []downloadTool{
	// tools whose arguments we don't check
	{binaryPattern: glob.MustCompile("*/wget")},
	{binaryPattern: glob.MustCompile("*/ftp")},
	{binaryPattern: glob.MustCompile("*/svn")},
	{binaryPattern: glob.MustCompile("*/git-remote-*")},
	{binaryPattern: glob.MustCompile("*/sftp")},
	{binaryPattern: glob.MustCompile("*/sftp-server")},

	// tools whose arguments we check
	{
		binaryPattern: glob.MustCompile("*/curl"),
		argsPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(^-|\s+-)[^-\s]*(o|O)\S*\s*\S+`),
			regexp.MustCompile(`--((output)|(remote-name))\s*\S+`),
		},
	},
	{
		binaryPattern: glob.MustCompile("*/ssh"),
		argsPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\s*-[^-\s]*s\s+`),
		},
	},
}

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
		args := kprobe.GetProcess().GetArguments()
		function := kprobe.GetFunctionName()
		functionArgs := kprobe.GetArgs()

		if function != "tcp_connect" {
			return resp, nil
		}

		match := false

		for _, dt := range downloadTools {
			if !dt.binaryPattern.Match(binary) {
				continue // no need to check args as binary doesn't match
			}

			// in case of some tools args are not checked
			if len(dt.argsPatterns) == 0 {
				match = true
				break
			}

			for _, p := range dt.argsPatterns {
				if p.MatchString(args) {
					match = true
					break
				}
			}
		}

		if match {
			resp.TacticsCovered = mitreTactics
			dstIp := functionArgs[0].GetSockArg().GetDaddr()
			dstPort := functionArgs[0].GetSockArg().GetDport()
			if args == "" {
				resp.Reason = fmt.Sprintf(KprobeNoArgs, dstIp, dstPort, binary)
			} else {
				resp.Reason = fmt.Sprintf(KprobeDefault, dstIp, dstPort, binary, args)
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
