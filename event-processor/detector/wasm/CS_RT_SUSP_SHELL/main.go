//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_SUSP_SHELL"
	Name        = "Suspicious startup of command shell"
	Description = "The detector detects suspicious startups of the command shell. They may indicate attempts to exploit RCE vulnerabilities, open a remote communication channel, or use the GTFOBins utilities for privilege escalation."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecNoArgs  = "Detected a suspicious startup of the `%s` command shell whose parent process was created by the executable file `%s` from the `%s` directory"
	ExecDefault = "Detected a suspicious startup of the `%s` command shell with the `%s` arguments. The parent process of the shell was created by the executable file `%s` from the `%s` directory."
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
				"T1190",
			},
		},
		{
			Id: "TA0002",
			Techniques: []string{
				"T1059.004",
				"T1059.006",
				"T1059.011",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1543",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1543",
				"T1611",
				"T1068",
			},
		},
	}
)

var (
	shells = []glob.Glob{
		glob.MustCompile("*/ash"),
		glob.MustCompile("*/bash"),
		glob.MustCompile("*/csh"),
		glob.MustCompile("*/dash"),
		glob.MustCompile("*/ksh"),
		glob.MustCompile("*/sh"),
		glob.MustCompile("*/tcsh"),
		glob.MustCompile("*/zsh"),
	}

	suspParents = []glob.Glob{
		glob.MustCompile("*/awk"),
		glob.MustCompile("*/busybox"),
		glob.MustCompile("*/capsh"),
		glob.MustCompile("*/certbot"),
		glob.MustCompile("*/choom"),
		glob.MustCompile("*/cowsay"),
		glob.MustCompile("*/cpan"),
		glob.MustCompile("*/cpio"),
		glob.MustCompile("*/cpulimit"),
		glob.MustCompile("*/ed"),
		glob.MustCompile("*/emacs"),
		glob.MustCompile("*/env"),
		glob.MustCompile("*/ex"),
		glob.MustCompile("*/expect"),
		glob.MustCompile("*/facter"),
		glob.MustCompile("*/find"),
		glob.MustCompile("*/flock"),
		glob.MustCompile("*/ftp"),
		glob.MustCompile("*/gcc"),
		glob.MustCompile("*/gdb"),
		glob.MustCompile("*/gem"),
		glob.MustCompile("*/ghc"),
		glob.MustCompile("*/gtester"),
		glob.MustCompile("*/hping3"),
		glob.MustCompile("*/ionice"),
		glob.MustCompile("*/irb"),
		glob.MustCompile("*/java"),
		glob.MustCompile("*/jjs"),
		glob.MustCompile("*/jrunscript"),
		glob.MustCompile("*/knife"),
		glob.MustCompile("*/latex"),
		glob.MustCompile("*/latexmk"),
		glob.MustCompile("*/less"),
		glob.MustCompile("*/lftp"),
		glob.MustCompile("*/logsave"),
		glob.MustCompile("*/ltrace"),
		glob.MustCompile("*/lua"),
		glob.MustCompile("*/lualatex"),
		glob.MustCompile("*/luatex"),
		glob.MustCompile("*/more"),
		glob.MustCompile("*/msgfilter"),
		glob.MustCompile("*/mysql"),
		glob.MustCompile("*/nano"),
		glob.MustCompile("*/neofetch"),
		glob.MustCompile("*/nice"),
		glob.MustCompile("*/nmap"),
		glob.MustCompile("*/node"),
		glob.MustCompile("*/nohup"),
		glob.MustCompile("*/npm"),
		glob.MustCompile("*/nsenter"),
		glob.MustCompile("*/octave-cli"),
		glob.MustCompile("*/openvpn"),
		glob.MustCompile("*/pager"),
		glob.MustCompile("*/pdflatex"),
		glob.MustCompile("*/pdftex"),
		glob.MustCompile("*/perf"),
		glob.MustCompile("*/perl"),
		glob.MustCompile("*/php"),
		glob.MustCompile("*/pic"),
		glob.MustCompile("*/pip"),
		glob.MustCompile("*/pip3"),
		glob.MustCompile("*/postgres"),
		glob.MustCompile("*/pry"),
		glob.MustCompile("*/psftp"),
		glob.MustCompile("*/puppet"),
		glob.MustCompile("*/python"),
		glob.MustCompile("*/python3"),
		glob.MustCompile("*/rake"),
		glob.MustCompile("*/rlwrap"),
		glob.MustCompile("*/rpm"),
		glob.MustCompile("*/rpmdb"),
		glob.MustCompile("*/rpmverify"),
		glob.MustCompile("*/rsync"),
		glob.MustCompile("*/ruby"),
		glob.MustCompile("*/scanmem"),
		glob.MustCompile("*/scp"),
		glob.MustCompile("*/screen"),
		glob.MustCompile("*/script"),
		glob.MustCompile("*/sed"),
		glob.MustCompile("*/setarch"),
		glob.MustCompile("*/sftp"),
		glob.MustCompile("*/sg"),
		glob.MustCompile("*/socat"),
		glob.MustCompile("*/socket"),
		glob.MustCompile("*/sqlite3"),
		glob.MustCompile("*/ssh"),
		glob.MustCompile("*/stdbuf"),
		glob.MustCompile("*/strace"),
		glob.MustCompile("*/tar"),
		glob.MustCompile("*/task"),
		glob.MustCompile("*/taskset"),
		glob.MustCompile("*/tasksh"),
		glob.MustCompile("*/tclsh"),
		glob.MustCompile("*/tex"),
		glob.MustCompile("*/time"),
		glob.MustCompile("*/timeout"),
		glob.MustCompile("*/tmate"),
		glob.MustCompile("*/tshark"),
		glob.MustCompile("*/unshare"),
		glob.MustCompile("*/vi"),
		glob.MustCompile("*/view"),
		glob.MustCompile("*/vim"),
		glob.MustCompile("*/vimdiff"),
		glob.MustCompile("*/watch"),
		glob.MustCompile("*/wish"),
		glob.MustCompile("*/xargs"),
		glob.MustCompile("*/xdotool"),
		glob.MustCompile("*/xelatex"),
		glob.MustCompile("*/xetex"),
		glob.MustCompile("*/zip"),
		glob.MustCompile("*/ash"),
		glob.MustCompile("*/bash"),
		glob.MustCompile("*/csh"),
		glob.MustCompile("*/ksh"),
		glob.MustCompile("*/sh"),
		glob.MustCompile("*/tcsh"),
		glob.MustCompile("*/zsh"),
		glob.MustCompile("*/dash"),
	}

	suspDirs = []glob.Glob{
		glob.MustCompile("/home/*"),
		glob.MustCompile("/tmp/*"),
		glob.MustCompile("/var/*"),
		glob.MustCompile("/boot/*"),
		glob.MustCompile("/media/*"),
		glob.MustCompile("/mnt/*"),
		glob.MustCompile("/srv/*"),
		glob.MustCompile("/sys/*"),
		glob.MustCompile("/dev/*"),
		glob.MustCompile("/run/*"),
		glob.MustCompile("/sys/*"),
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
		parentBinary := exec.GetParent().GetBinary()

		shell := false
		suspParent := false
		suspDir := false

		for _, sh := range shells {
			if sh.Match(binary) {
				shell = true
				break
			}
		}

		// the executed process is not a shell
		if !shell {
			return resp, nil
		}

		for _, sp := range suspParents {
			if sp.Match(parentBinary) {
				suspParent = true
				break
			}
		}

		// trigger on parent running in suspicious directory
		for _, dir := range suspDirs {
			if dir.Match(parentBinary) {
				suspDir = true
				break
			}
		}

		if suspParent || suspDir {
			resp.TacticsCovered = mitreTactics
			parentBinaryDir := parentBinary[0 : strings.LastIndex(parentBinary, "/")+1]
			if args == "" {
				resp.Reason = fmt.Sprintf(ExecNoArgs, binary, parentBinary, parentBinaryDir)
			} else {
				resp.Reason = fmt.Sprintf(ExecDefault, binary, args, parentBinary, parentBinaryDir)
			}
			resp.Severity = api.DetectResp_HIGH
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
