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
	ID          = "CS_RT_OPENSSL_USE"
	Name        = "OpenSSL use for malicious purposes"
	Description = "The detector detects if the OpenSSL utility was used to perform an attacker's scenarios: malicious code loading, reading or writing of arbitrary files, or data exchange via a TLS or SSL server or client."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecCustomLibLoad      = "Detected that arbitrary code was loaded by the `%s` process, which was started using the `%s` arguments"
	ExecFileRead           = "Detected a file read attempt made by the `%s` process, which was started using the `%s` arguments"
	ExecFileWrite          = "Detected a file write attempt made by the `%s` process, which was started using the `%s` arguments"
	ExecDataExchangeClient = "Detected that a TLS or SSL client for exchanging data was started by the `%s` process, which was started using the `%s` arguments"
	ExecDataExchangeServer = "Detected that a TLS or SSL server for exchanging data was started by the `%s` process, which was started using the `%s` arguments"
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
			Id: "TA0002",
			Techniques: []string{
				"T1129",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1140",
				"T1027.013",
			},
		},
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
			},
		},
		{
			Id: "TA0011",
			Techniques: []string{
				"T1105",
			},
		},
		{
			Id: "TA0040",
			Techniques: []string{
				"T1486",
			},
		},
	}

	mitreTacticsCustomLibLoad = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1129",
			},
		},
	}

	mitreTacticsFileAccess = []*api.MitreTactic{
		{
			Id: "TA0005",
			Techniques: []string{
				"T1140",
				"T1027.013",
			},
		},
		{
			Id: "TA0010",
			Techniques: []string{
				"T1048.001",
			},
		},
		{
			Id: "TA0040",
			Techniques: []string{
				"T1486",
			},
		},
	}

	mitreTacticsDataExchange = []*api.MitreTactic{
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

type opensslArgPattern struct {
	pattern *regexp.Regexp
	class   string
}

var (
	opensslArgPatterns = []opensslArgPattern{
		// Command for loading custom user library as crypto engine.
		{pattern: regexp.MustCompile(`^engine\s(?:dynamic\s)?.*-pre SO_PATH:.*-pre LOAD`), class: "CustomLibLoad"},
		// Load custom user library using -engine option as part of other openssl commands.
		{pattern: regexp.MustCompile(`^\w+\s-engine\s`), class: "CustomLibLoad"},
		// Read arbitrary files
		{pattern: regexp.MustCompile(`(?:^enc$)|(?:^enc.*-in)`), class: "FileRead"},
		// Write arbitrary files
		{pattern: regexp.MustCompile(`(?:^enc$)|(?:^enc.*-out)`), class: "FileWrite"},
		// Run a HTTPS client for data exchange
		{pattern: regexp.MustCompile(`^s_client.*`), class: "DataExchangeClient"},
		// Run a HTTPS server for data exchange
		{pattern: regexp.MustCompile(`^s_server.*`), class: "DataExchangeServer"},
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
		exec := ev.ProcessExec
		binary := exec.GetProcess().GetBinary()
		args := exec.GetProcess().GetArguments()

		for _, opensslArgs := range opensslArgPatterns {
			switch {
			// Custom library load case
			case opensslArgs.pattern.MatchString(args) && (opensslArgs.class == "CustomLibLoad"):
				resp.TacticsCovered = mitreTacticsCustomLibLoad
				resp.Reason = fmt.Sprintf(ExecCustomLibLoad, binary, args)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			// File read case
			case opensslArgs.pattern.MatchString(args) && (opensslArgs.class == "FileRead"):
				resp.TacticsCovered = mitreTacticsFileAccess
				resp.Reason = fmt.Sprintf(ExecFileRead, binary, args)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			// File write case
			case opensslArgs.pattern.MatchString(args) && (opensslArgs.class == "FileWrite"):
				resp.TacticsCovered = mitreTacticsFileAccess
				resp.Reason = fmt.Sprintf(ExecFileWrite, binary, args)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			// Data exchange client case
			case opensslArgs.pattern.MatchString(args) && (opensslArgs.class == "DataExchangeClient"):
				resp.TacticsCovered = mitreTacticsDataExchange
				resp.Reason = fmt.Sprintf(ExecDataExchangeClient, binary, args)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			// Data exchange server case
			case opensslArgs.pattern.MatchString(args) && (opensslArgs.class == "DataExchangeServer"):
				resp.TacticsCovered = mitreTacticsDataExchange
				resp.Reason = fmt.Sprintf(ExecDataExchangeServer, binary, args)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

		return resp, nil

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

