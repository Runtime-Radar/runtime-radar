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
	ID          = "CS_RT_BIN_PERM_RAISE"
	Name        = "Change of file access permissions"
	Description = "The detector detects if execution permissions for files in the boot, dev, home, media, mnt, run, sys, tmp, and var directories were granted. Existence of such permissions may be a sign that an attacker tried to execute an unauthorized and malicious code inside a container."
	Version     = 3
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecChmod     = "Detected that execution permissions were granted by the `%s` process, which was started using the `%s` arguments"
	KprobeNoArgs  = "Detected that execution permissions were granted to the `%s` file by the `%s` process"
	KprobeDefault = "Detected that execution permissions were granted to the `%s` file by the `%s` process, which was started using the `%s` arguments"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_EXEC":   {},
		"PROCESS_KPROBE": {"security_path_chmod"},

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
			Id: "TA0004",
			Techniques: []string{
				"T1543",
				"T1068",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1222.002",
			},
		},
	}
)

var (
	chmod       = glob.MustCompile(`*/chmod`)
	execPerm    = regexp.MustCompile(`^(?:[\-Rcfv\s]+)?(?:(?:[0-7]?[0-7][1357][0-7]\s|[0-7]?[1357][0-7][0-7]\s|[0-7]?[0-7][0-7][1357]\s)|(?:[ugoa]*[+=]x\s))`)
	suspDirArgs = regexp.MustCompile(`^(?:[\-Rcfv\s]+)?(?:(?:[ugoa]*[-=+][rwxX])|(?:[0-7]{3,4}))(?:.*)\/?(?:boot|dev|home|media|mnt|run|sys|tmp|var)\/`)
	relPath     = regexp.MustCompile(`^(?:[\-Rcfv\s]+)?(?:(?:[ugoa]*[-=+][rwxX])|(?:[0-7]{3,4}))(?:.*)\s(?:[^\/])`)

	suspDirCwd = []glob.Glob{
		glob.MustCompile("/boot*"),
		glob.MustCompile("/dev*"),
		glob.MustCompile("/home*"),
		glob.MustCompile("/media*"),
		glob.MustCompile("/mnt*"),
		glob.MustCompile("/run*"),
		glob.MustCompile("/srv*"),
		glob.MustCompile("/sys*"),
		glob.MustCompile("/tmp*"),
		glob.MustCompile("/var*"),
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
		cwd := exec.GetProcess().GetCwd()

		// Check binary.
		if !chmod.Match(binary) {
			return resp, nil
		}

		// Check if execute access permissions are in args.
		execPermSet := false
		if execPerm.MatchString(args) {
			execPermSet = true
		} else {
			return resp, nil
		}

		// Check if suspicious directory is in args.
		suspiciousDir := false
		if suspDirArgs.MatchString(args) {
			suspiciousDir = true
		}

		// Next sequence is for cases like: cd /tmp && chmod +x ./xmrig).
		if !suspiciousDir {
			// Check if current directory is suspicious.
			suspiciousDirInCwd := false
			for _, d := range suspDirCwd {
				if d.Match(cwd) {
					suspiciousDirInCwd = true
				}
			}

			// Check if path in args is relative.
			if relPath.MatchString(args) && suspiciousDirInCwd {
				suspiciousDir = true
			}
		}

		// Trigger only on executable bits in arguments and suspicious directories.
		if execPermSet && suspiciousDir {
			resp.TacticsCovered = mitreTactics
			resp.Reason = fmt.Sprintf(ExecChmod, binary, args)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		}

		return resp, nil

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
		case "security_path_chmod":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			}

			path = args[0].GetPathArg().GetPath()
		default:
			return resp, nil
		}

		for _, dir := range suspDirCwd {
			if dir.Match(path) {
				resp.TacticsCovered = mitreTactics
				if binaryArgs == "" {
					resp.Reason = fmt.Sprintf(KprobeNoArgs, path, binary)
				} else {
					resp.Reason = fmt.Sprintf(KprobeDefault, path, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
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
