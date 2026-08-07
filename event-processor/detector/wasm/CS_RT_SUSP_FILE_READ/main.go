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
	ID          = "CS_RT_SUSP_FILE_READ"
	Name        = "Suspicious reading of sensitive system files"
	Description = "The detector detects suspicious reading of system files with utilities from uncommon directories, such as /tmp and /home, which may indicate that system configuration data is being collected."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (

	// Account
	KprobeUtilsReadAccountsNoArgs  = "Detected that the `%s` file with accounts was read by the `%s` utility"
	KprobeUtilsReadAccountsDefault = "Detected that the `%s` file with accounts was read by the `%s` utility, which was started using the `%s` arguments"
	KprobeDirsReadAccountsNoArgs   = "Detected that the `%s` file with accounts was read by the executable file `%s` from the `%s` directory"
	KprobeDirsReadAccountsDefault  = "Detected that the `%s` file with accounts was read by the executable file `%s` from the `%s` directory started using the `%s` arguments"

	// PAM
	KprobeUtilsReadPamNoArgs  = "Detected that the `%s` file with the system authentication and authorization settings was read by the `%s` utility"
	KprobeUtilsReadPamDefault = "Detected that the `%s` file with the system authentication and authorization settings was read by the `%s` utility started using the `%s` arguments"
	KprobeDirsReadPamNoArgs   = "Detected that the `%s` file with the system authentication and authorization settings was read by the executable file `%s` from the `%s` directory"
	KprobeDirsReadPamDefault  = "Detected that the `%s` file with the system authentication and authorization settings was read by the executable file `%s` from the `%s` directory (started using the `%s` arguments)"

	// SystemInfo
	KprobeUtilsReadSysInfoNoArgs  = "Detected that the `%s` file with system information was read by the `%s` utility"
	KprobeUtilsReadSysInfoDefault = "Detected that the `%s` file with system information was read by the `%s` utility started using the `%s` arguments"
	KprobeDirsReadSysInfoNoArgs   = "Detected that the `%s` file with system information was read by the executable file `%s` from the `%s` directory"
	KprobeDirsReadSysInfoDefault  = "Detected that the `%s` file with system information was read by the executable file `%s` from the `%s` directory (started using the `%s` arguments)"

	// EnvironmentVariables
	KprobeUtilsReadEnvVarNoArgs  = "Detected that the `%s` file with environment variables was read by the `%s` utility"
	KprobeUtilsReadEnvVarDefault = "Detected that the `%s` file with environment variables was read by the `%s` utility started using the `%s` arguments"
	KprobeDirsReadEnvVarNoArgs   = "Detected that the `%s` file with environment variables was read by the executable file `%s` from the `%s` directory"
	KprobeDirsReadEnvVarDefault  = "Detected that the `%s` file with environment variables was read by the executable file `%s` from the `%s` directory (started using the `%s` arguments)"
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
				"T1078.003",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1003.008",
				"T1552.001",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1087.001",
				"T1083",
				"T1201",
				"T1069.001",
			},
		},
		{
			Id: "TA0009",
			Techniques: []string{
				"T1005",
			},
		},
	}

	mitreTacticsPam = []*api.MitreTactic{
		{
			Id: "TA0007",
			Techniques: []string{
				"T1083",
				"T1201",
			},
		},
		{
			Id: "TA0009",
			Techniques: []string{
				"T1005",
			},
		},
	}

	mitreTacticsAccounts = []*api.MitreTactic{
		{
			Id: "TA0001",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1078.003",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1003.008",
				"T1552.001",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1087.001",
				"T1083",
				"T1069.001",
			},
		},
		{
			Id: "TA0009",
			Techniques: []string{
				"T1005",
			},
		},
	}

	mitreTacticsOther = []*api.MitreTactic{
		{
			Id: "TA0007",
			Techniques: []string{
				"T1083",
			},
		},
		{
			Id: "TA0009",
			Techniques: []string{
				"T1005",
			},
		},
	}
)

const (
	// File access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/linux/fs.h#L101
	MAY_READ = 4

	// Memory page access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L10
	PROT_READ = 1
)

type sensitiveFile struct {
	pattern glob.Glob
	class   string
}

var (
	sensitiveFiles = []sensitiveFile{
		// Accounts password info
		{pattern: glob.MustCompile("/etc/shadow"), class: "Accounts"},
		// Superuser security policies
		{pattern: glob.MustCompile("/etc/sudoers*"), class: "Accounts"},
		// Authentication module settings
		{pattern: glob.MustCompile("/etc/pam.*"), class: "PAM"},
		// Password policy (PAM)
		{pattern: glob.MustCompile("/etc/security/pwquality.conf"), class: "PAM"},
		// Distribution info
		{pattern: glob.MustCompile("/etc/*-release"), class: "SystemInfo"},
		// Distribution info (for the most distributions /etc/os-release is a symlink to /usr/lib/os-release)
		{pattern: glob.MustCompile("/usr/lib/os-release"), class: "SystemInfo"},
		// Environment variables
		{pattern: glob.MustCompile("/proc/*/environ"), class: "EnvironmentVariables"},
	}

	readUtils = []glob.Glob{
		glob.MustCompile("*/awk"),
		glob.MustCompile("*/cat"),
		glob.MustCompile("*/cp"),
		glob.MustCompile("*/dd"),
		glob.MustCompile("*/diff"),
		glob.MustCompile("*/egrep"),
		glob.MustCompile("*/emacs"),
		glob.MustCompile("*/gawk"),
		glob.MustCompile("*/grep"),
		glob.MustCompile("*/head"),
		glob.MustCompile("*/java"),
		glob.MustCompile("*/less"),
		glob.MustCompile("*/mc"),
		glob.MustCompile("*/mcdiff"),
		glob.MustCompile("*/mcedit"),
		glob.MustCompile("*/mcview"),
		glob.MustCompile("*/more"),
		glob.MustCompile("*/nano"),
		glob.MustCompile("*/perl"),
		glob.MustCompile("*/php"),
		glob.MustCompile("*/python*"),
		glob.MustCompile("*/ruby"),
		glob.MustCompile("*/sed"),
		glob.MustCompile("*/sort"),
		glob.MustCompile("*/tac"),
		glob.MustCompile("*/tail"),
		glob.MustCompile("*/uniq"),
		glob.MustCompile("*/vi"),
		glob.MustCompile("*/vim"),
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
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		binaryArgs := kprobe.GetProcess().GetArguments()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()
		var affectedFile sensitiveFile
		sensitiveFile := false

		switch function {
		// trigger when security function check for file read access
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_READ {
				return resp, nil
			}
		// trigger when security function check for memory page read access
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&1 != PROT_READ {
				return resp, nil
			}
		default:
			return resp, nil
		}

		path := args[0].GetFileArg().GetPath()

		for _, file := range sensitiveFiles {
			switch {
			case file.pattern.Match(path) && (file.class == "Accounts"):
				resp.TacticsCovered = mitreTacticsAccounts
				affectedFile = file
				sensitiveFile = true
			case file.pattern.Match(path) && (file.class == "PAM"):
				resp.TacticsCovered = mitreTacticsPam
				affectedFile = file
				sensitiveFile = true
			case file.pattern.Match(path):
				resp.TacticsCovered = mitreTacticsOther
				affectedFile = file
				sensitiveFile = true
			}
		}

		if !sensitiveFile {
			return resp, nil
		}

		// find out sensitive file reading method
		// trigger on specific utils reading sensitive files
		for _, util := range readUtils {
			if !util.Match(binary) {
				continue
			}
			switch {
			case (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadAccountsNoArgs, path, binary)
			case (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadAccountsDefault, path, binary, binaryArgs)
			case (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadPamNoArgs, path, binary)
			case (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadPamDefault, path, binary, binaryArgs)
			case (affectedFile.class == "SystemInfo") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadSysInfoNoArgs, path, binary)
			case (affectedFile.class == "SystemInfo") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadSysInfoDefault, path, binary, binaryArgs)
			case (affectedFile.class == "EnvironmentVariables") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadEnvVarNoArgs, path, binary)
			case (affectedFile.class == "EnvironmentVariables") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsReadEnvVarDefault, path, binary, binaryArgs)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

			return resp, nil
		}

		// trigger on utils from suspicious directories
		for _, dir := range suspDirs {
			if !dir.Match(binary) {
				continue
			}
			binDir := binary[0 : strings.LastIndex(binary, "/")+1]
			switch {
			case (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadAccountsNoArgs, path, binary, binDir)
			case (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadAccountsDefault, path, binary, binDir, binaryArgs)
			case (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadPamNoArgs, path, binary, binDir)
			case (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadPamDefault, path, binary, binDir, binaryArgs)
			case (affectedFile.class == "SystemInfo") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadSysInfoNoArgs, path, binary, binDir)
			case (affectedFile.class == "SystemInfo") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadSysInfoDefault, path, binary, binDir, binaryArgs)
			case (affectedFile.class == "EnvironmentVariables") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadEnvVarNoArgs, path, binary, binDir)
			case (affectedFile.class == "EnvironmentVariables") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsReadEnvVarDefault, path, binary, binDir, binaryArgs)
			}
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
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
