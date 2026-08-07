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
	ID          = "CS_RT_CRYPTOMINER"
	Name        = "Cryptominer startup"
	Description = "The detector detects if any known cryptominers were started or stopped, which may indicate an attacker's attempt to get rid of competing loads."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecMinersNoArgs  = "Detected that the `%s` cryptominer was started"
	ExecMinersDefault = "Detected that the `%s` cryptominer was started using the `%s` arguments"
	ExecMinerArgs     = "Detected a suspicious start of the `%s` process with the `%s` arguments typical of cryptominers"
	ExecKillArgs      = "Detected an attempt to finish the `%s` process using the `%s` utility started with the `%s` arguments typical of cryptominer startup scripts"
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
				"T1204.003",
			},
		},
		{
			Id: "TA0040",
			Techniques: []string{
				"T1496",
			},
		},
	}
)

var (
	minerTools = []*regexp.Regexp{
		regexp.MustCompile(`/[^/]*stratum[^/]*$`),
		regexp.MustCompile(`/[^/]*minerd[^/]*$`),
		regexp.MustCompile(`/[^/]*xmr[^/]*$`),
		regexp.MustCompile(`/[^/]*cryptonight[^/]*$`),
		regexp.MustCompile(`/[^/]*haiduc[^/]*$`),
		regexp.MustCompile(`/sparky\.sh$`),
		regexp.MustCompile(`/2238Xae$`),
		regexp.MustCompile(`/[^/]*yastrm[^/]*$`),
		regexp.MustCompile(`/[^/]*biden[1l][^/]*$`),
		regexp.MustCompile(`/xrx$`),
		regexp.MustCompile(`/zzh$`),
		regexp.MustCompile(`/arx645$`),
		regexp.MustCompile(`/[^/]+kthread[^/]*$`),
		regexp.MustCompile(`/[^/]*kdevtmpfsi[^/]*$`),
		regexp.MustCompile(`/pppsssdm$`),
		regexp.MustCompile(`/[^/]*kernelx[^/]*$`),
		regexp.MustCompile(`/[^/]*pg_mem[^/]*$`),
	}

	minerArgs = []*regexp.Regexp{
		regexp.MustCompile(`\-\-cpu-priority`),
		regexp.MustCompile(`\-\-donate-level`),
		regexp.MustCompile(`\-\-nicehash`),
		regexp.MustCompile(`\-\-algo`),
		regexp.MustCompile(`stratum2?\+`),
	}

	killingTools = regexp.MustCompile(`/p?kill(?:all)?$`)

	killArgs = []glob.Glob{
		glob.MustCompile("*stratum*"),
		glob.MustCompile("*minerd*"),
		glob.MustCompile("*xmr*"),
		glob.MustCompile("*cryptonight*"),
		glob.MustCompile("*haiduc*"),
		glob.MustCompile("*sparky.sh*"),
		glob.MustCompile("*2238Xae*"),
		glob.MustCompile("*yastrm*"),
		glob.MustCompile("*biden*"),
		glob.MustCompile("*xrx*"),
		glob.MustCompile("*zzh*"),
		glob.MustCompile("*arx645*"),
		glob.MustCompile("*kthread*"),
		glob.MustCompile("*kdevtmpfsi*"),
		glob.MustCompile("*pppsssdm*"),
		glob.MustCompile("*kernelx*"),
		glob.MustCompile("*pg_mem*"),
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

		for _, mt := range minerTools {
			// Trigger on explicit miner utility name.
			if mt.MatchString(binary) {
				resp.TacticsCovered = mitreTactics
				if args == "" {
					resp.Reason = fmt.Sprintf(ExecMinersNoArgs, binary)
				} else {
					resp.Reason = fmt.Sprintf(ExecMinersDefault, binary, args)
				}
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			}
		}

		// Trigger on explicit miner utility args.
		for _, ma := range minerArgs {
			if ma.MatchString(args) {
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecMinerArgs, binary, args)
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			}
		}

		// Trigger on attempts of killing a miner process.
		if killingTools.MatchString(binary) {
			for _, ka := range killArgs {
				if ka.Match(args) {
					resp.TacticsCovered = mitreTactics
					resp.Reason = fmt.Sprintf(ExecKillArgs, binary, args)
					resp.Severity = api.DetectResp_HIGH // <-- threat detected

					return resp, nil
				}
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
