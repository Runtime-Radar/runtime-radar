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
	ID          = "CS_RT_SCHED_TASK_MOD"
	Name        = "Suspicious changes in task scheduler configuration files"
	Description = "The detector detects changes in configuration files of a task scheduler (for example, cron), which may result in an attacker achieving persistence or planned malicious actions."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (

	// Anacron
	KprobeWriteAnacronNoArgs   = "Detected that the configuration file `%s` of the anacron task scheduler was edited by the `%s` process"
	KprobeWriteAnacronDefault  = "Detected that the configuration file `%s` of the anacron task scheduler was edited by the %s process, which was started using the `%s` arguments"
	KprobeMmapAnacronNoArgs    = "Detected that the anacron task scheduler file `%s` was memory-mapped by the `%s` process"
	KprobeMmapAnacronDefault   = "Detected that the anacron task scheduler file `%s` was memory-mapped by the `%s` process, which was started using the `%s` arguments"
	KprobeRenameAnacronNoArgs  = "Detected that the configuration file `%s` of the anacron task scheduler was replaced with the `%s` file by the `%s` process"
	KprobeRenameAnacronDefault = "Detected that the configuration file `%s` of the anacron task scheduler was replaced with the `%s` file by the `%s` process, which was started using the `%s` arguments"

	// At
	KprobeWriteAtNoArgs   = "Detected that the configuration file `%s` of the at task scheduler was edited by the `%s` process"
	KprobeWriteAtDefault  = "Detected that the configuration file `%s` of the at task scheduler was edited by the %s process, which was started using the `%s` arguments"
	KprobeMmapAtNoArgs    = "Detected that the at task scheduler file `%s` was memory-mapped by the `%s` process"
	KprobeMmapAtDefault   = "Detected that the at task scheduler file `%s` was memory-mapped by the `%s` process, which was started using the `%s` arguments"
	KprobeRenameAtNoArgs  = "Detected that the configuration file `%s` of the at task scheduler was replaced with the `%s` file by the `%s` process"
	KprobeRenameAtDefault = "Detected that the configuration file `%s` of the at task scheduler was replaced with the `%s` file by the `%s` process, which was started using the `%s` arguments"

	// Cron
	KprobeWriteCronNoArgs   = "Detected that the configuration file `%s` of the cron task scheduler was edited by the `%s` process"
	KprobeWriteCronDefault  = "Detected that the configuration file `%s` of the cron task scheduler was edited by the %s process, which was started using the `%s` arguments"
	KprobeMmapCronNoArgs    = "Detected that the cron task scheduler file `%s` was memory-mapped by the `%s` process"
	KprobeMmapCronDefault   = "Detected that the cron task scheduler file `%s` was memory-mapped by the `%s` process, which was started using the `%s` arguments"
	KprobeRenameCronNoArgs  = "Detected that the configuration file `%s` of the cron task scheduler was replaced with the `%s` file by the `%s` process"
	KprobeRenameCronDefault = "Detected that the configuration file `%s` of the cron task scheduler was replaced with the `%s` file by the `%s` process, which was started using the `%s` arguments"

	// Systemd Timers
	KprobeWriteSystemdTimersNoArgs   = "Detected that the `%s` file of the systemd subsystem timer was edited by the `%s` process"
	KprobeWriteSystemdTimersDefault  = "Detected that the `%s` file of the systemd subsystem timer was edited by the `%s` process, which was started using the `%s` arguments"
	KprobeMmapSystemdTimersNoArgs    = "Detected that the `%s` file of the systemd subsystem timer was memory-mapped by the `%s` process"
	KprobeMmapSystemdTimersDefault   = "Detected that the `%s` file of the systemd subsystem timer was memory-mapped by the `%s` process, which was started using the `%s` arguments"
	KprobeRenameSystemdTimersNoArgs  = "Detected that the `%s` file of the systemd subsystem timer was replaced with the `%s` file by the `%s` process"
	KprobeRenameSystemdTimersDefault = "Detected that the `%s` file of the systemd subsystem timer was replaced with the `%s` file by the `%s` process, which was started using the `%s` arguments"
)

var (
	mitreTactics = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1053.002",
				"T1053.003",
				"T1053.006",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1053.002",
				"T1053.003",
				"T1053.006",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1053.002",
				"T1053.003",
				"T1053.006",
			},
		},
	}

	mitreTacticsAt = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1053.002",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1053.002",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1053.002",
			},
		},
	}

	mitreTacticsCron = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1053.003",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1053.003",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1053.003",
			},
		},
	}

	mitreTacticsSystemdTimers = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1053.006",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1053.006",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1053.006",
			},
		},
	}

	mitreTacticsAnacron = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1053",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1053",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1053",
			},
		},
	}
)

const (
	// File access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/linux/fs.h#L100
	MAY_WRITE = 2

	// Memory page access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L11
	PROT_WRITE = 2
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate", "security_path_rename"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)

type schedulerFile struct {
	pattern   glob.Glob
	scheduler string
}

var (
	schedulerFiles = []schedulerFile{
		// system task scheduler
		{pattern: glob.MustCompile("/etc/crontab"), scheduler: "Cron"},
		{pattern: glob.MustCompile("/etc/anacrontab"), scheduler: "Anacron"},
		{pattern: glob.MustCompile("/etc/cron.d/*"), scheduler: "Cron"},
		// tasks with predefined hourly interval
		{pattern: glob.MustCompile("/etc/cron.hourly/*"), scheduler: "Cron"},
		// tasks with predefined daily interval
		{pattern: glob.MustCompile("/etc/cron.daily/*"), scheduler: "Cron"},
		// tasks with predefined weekly interval
		{pattern: glob.MustCompile("/etc/cron.weekly/*"), scheduler: "Cron"},
		// tasks with predefined monthly interval
		{pattern: glob.MustCompile("/etc/cron.monthly/*"), scheduler: "Cron"},
		// user task scheduler
		{pattern: glob.MustCompile("/var/spool/cron/*"), scheduler: "Cron"},
		{pattern: glob.MustCompile("/var/spool/anacron/*"), scheduler: "Anacron"},
		// user access list for scheduler
		{pattern: glob.MustCompile("/etc/cron.deny"), scheduler: "Cron"},
		{pattern: glob.MustCompile("/etc/cron.allow"), scheduler: "Cron"},
		// user task scheduler
		{pattern: glob.MustCompile("/var/spool/at/*"), scheduler: "At"},
		// user access list for scheduler
		{pattern: glob.MustCompile("/etc/at.deny"), scheduler: "At"},
		{pattern: glob.MustCompile("/etc/at.allow"), scheduler: "At"},
		// systemd task scheduler
		{pattern: glob.MustCompile("/etc/systemd/system/*.timer"), scheduler: "SystemdTimers"},
		{pattern: glob.MustCompile("/usr/local/lib/systemd/system/*.timer"), scheduler: "SystemdTimers"},
		{pattern: glob.MustCompile("/lib/systemd/system/*.timer"), scheduler: "SystemdTimers"},
		{pattern: glob.MustCompile("/usr/lib/systemd/system/*.timer"), scheduler: "SystemdTimers"},
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
		path := ""
		newFile := ""
		action := ""

		switch function {
		// trigger when security function check for file write access
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_WRITE {
				return resp, nil
			}

			action = "write"
			path = args[0].GetFileArg().GetPath()

		// trigger when security function check for memory page write access
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_WRITE == 0 {
				return resp, nil
			}

			action = "mmap"
			path = args[0].GetFileArg().GetPath()

		// trigger when security function check if truncating a file is allowed
		// https://elixir.bootlin.com/linux/v6.10.6/source/security/security.c#L1923
		case "security_path_truncate":
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			action = "write"
			path = args[0].GetPathArg().GetPath()

		// Trigger when security function check if renaming a file is allowed.
		// https://elixir.bootlin.com/linux/v6.15.7/source/security/security.c#L2005
		case "security_path_rename":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			}

			action = "rename"
			path = args[1].GetPathArg().GetPath()
			newFile = args[0].GetPathArg().GetPath()

		default:
			return resp, nil
		}

		for _, file := range schedulerFiles {
			if file.pattern.Match(path) {
				switch {
				case file.scheduler == "Anacron":
					resp.TacticsCovered = mitreTacticsAnacron
				case file.scheduler == "At":
					resp.TacticsCovered = mitreTacticsAt
				case file.scheduler == "Cron":
					resp.TacticsCovered = mitreTacticsCron
				case file.scheduler == "SystemdTimers":
					resp.TacticsCovered = mitreTacticsSystemdTimers
				default:
					resp.TacticsCovered = mitreTactics
				}
				switch {
				case (file.scheduler == "Anacron") && (action == "write") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeWriteAnacronNoArgs, path, binary)
				case (file.scheduler == "Anacron") && (action == "write") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeWriteAnacronDefault, path, binary, binaryArgs)
				case (file.scheduler == "At") && (action == "write") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeWriteAtNoArgs, path, binary)
				case (file.scheduler == "At") && (action == "write") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeWriteAtDefault, path, binary, binaryArgs)
				case (file.scheduler == "Cron") && (action == "write") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeWriteCronNoArgs, path, binary)
				case (file.scheduler == "Cron") && (action == "write") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeWriteCronDefault, path, binary, binaryArgs)
				case (file.scheduler == "SystemdTimers") && (action == "write") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeWriteSystemdTimersNoArgs, path, binary)
				case (file.scheduler == "SystemdTimers") && (action == "write") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeWriteSystemdTimersDefault, path, binary, binaryArgs)
				case (file.scheduler == "Anacron") && (action == "mmap") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeMmapAnacronNoArgs, path, binary)
				case (file.scheduler == "Anacron") && (action == "mmap") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeMmapAnacronDefault, path, binary, binaryArgs)
				case (file.scheduler == "At") && (action == "mmap") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeMmapAtNoArgs, path, binary)
				case (file.scheduler == "At") && (action == "mmap") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeMmapAtDefault, path, binary, binaryArgs)
				case (file.scheduler == "Cron") && (action == "mmap") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeMmapCronNoArgs, path, binary)
				case (file.scheduler == "Cron") && (action == "mmap") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeMmapCronDefault, path, binary, binaryArgs)
				case (file.scheduler == "SystemdTimers") && (action == "mmap") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeMmapSystemdTimersNoArgs, path, binary)
				case (file.scheduler == "SystemdTimers") && (action == "mmap") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeMmapSystemdTimersDefault, path, binary, binaryArgs)
				case (file.scheduler == "Anacron") && (action == "rename") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeRenameAnacronNoArgs, path, newFile, binary)
				case (file.scheduler == "Anacron") && (action == "rename") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeRenameAnacronDefault, path, newFile, binary, binaryArgs)
				case (file.scheduler == "At") && (action == "rename") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeRenameAtNoArgs, path, newFile, binary)
				case (file.scheduler == "At") && (action == "rename") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeRenameAtDefault, path, newFile, binary, binaryArgs)
				case (file.scheduler == "Cron") && (action == "rename") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeRenameCronNoArgs, path, newFile, binary)
				case (file.scheduler == "Cron") && (action == "rename") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeRenameCronDefault, path, newFile, binary, binaryArgs)
				case (file.scheduler == "SystemdTimers") && (action == "rename") && (binaryArgs == ""):
					resp.Reason = fmt.Sprintf(KprobeRenameSystemdTimersNoArgs, path, newFile, binary)
				case (file.scheduler == "SystemdTimers") && (action == "rename") && (binaryArgs != ""):
					resp.Reason = fmt.Sprintf(KprobeRenameSystemdTimersDefault, path, newFile, binary, binaryArgs)
				}
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

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
