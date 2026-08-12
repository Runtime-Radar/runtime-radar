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
	ID          = "CS_RT_SUSP_FILE_WRITE"
	Name        = "Suspicious changes in sensitive system files"
	Description = "The detector detects suspicious changes in system files, such as /etc/passwd and /etc/shadow, which may indicate an attacker's attempt to alter the system configuration or gain privileged access."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (

	// Accounts
	KprobeUtilsWriteAccountsNoArgs      = "Detected that the `%s` file with accounts was edited by the `%s` utility"
	KprobeUtilsWriteAccountsDefault     = "Detected that the `%s` file with accounts was edited by the `%s` utility, which was started using the `%s` arguments"
	KprobeUtilsMmapAccountsNoArgs       = "Detected that the `%s` file with accounts was memory-mapped by the `%s` utility"
	KprobeUtilsMmapAccountsDefault      = "Detected that the `%s` file with accounts was memory-mapped by the `%s` utility, which was started using the `%s` arguments"
	KprobeUtilsRenameAccountsNoArgs     = "Detected that the `%s` file with accounts was replaced with the `%s` file by the `%s` utility"
	KprobeUtilsRenameAccountsDefault    = "Detected that the `%s` file with accounts was replaced with the `%s` file by the `%s` utility, which was started using the `%s` arguments"
	KprobeShellsWriteAccountsNoArgs     = "Detected that the `%s` file with accounts was edited by the `%s` command shell"
	KprobeShellsWriteAccountsDefault    = "Detected that the `%s` file with accounts was edited by the `%s` command shell, which was started using the `%s` arguments"
	KprobeShellsMmapAccountsNoArgs      = "Detected that the `%s` file with accounts was memory-mapped by the `%s` command shell"
	KprobeShellsMmapAccountsDefault     = "Detected that the `%s` file with accounts was memory-mapped by the `%s` command shell, which was started using the `%s` arguments"
	KprobeShellsRenameAccountsNoArgs    = "Detected that the `%s` file with accounts was replaced with the `%s` file by the `%s` command shell"
	KprobeShellsRenameAccountsDefault   = "Detected that the `%s` file with accounts was replaced with the `%s` file by the `%s` command shell, which was started using the `%s` arguments"
	KprobeDirsWriteAccountsNoArgs       = "Detected that the `%s` file with accounts was edited by the executable file `%s` from the `%s` directory"
	KprobeDirsWriteAccountsDefault      = "Detected that the `%s` file with accounts was edited by the executable file `%s` from the `%s` directory, which was started using the `%s` arguments"
	KprobeDirsMmapAccountsNoArgs        = "Detected that the `%s` file with accounts was memory-mapped by the executable file `%s` from the `%s` directory"
	KprobeDirsMmapAccountsDefault       = "Detected that the `%s` file with accounts was memory-mapped by the executable file `%s` from the `%s` directory, which was started using the `%s` arguments"
	KprobeDirsRenameAccountsNoArgs      = "Detected that the `%s` file with accounts was replaced with the `%s` file by the executable file `%s` from the `%s` directory"
	KprobeDirsRenameAccountsDefault     = "Detected that the `%s` file with accounts was replaced with the `%s` file by the executable file `%s` from the `%s` directory, which was started using the `%s` arguments"

	// PAM
	KprobeUtilsWritePamNoArgs           = "Detected that the `%s` file with the system authentication and authorization settings was edited by the `%s` utility"
	KprobeUtilsWritePamDefault          = "Detected that the `%s` file with the system authentication and authorization settings was edited by the `%s` utility, which was started using the `%s` arguments"
	KprobeUtilsMmapPamNoArgs            = "Detected that the `%s` file with the system authentication and authorization settings was memory-mapped by the `%s` utility"
	KprobeUtilsMmapPamDefault           = "Detected that the `%s` file with the system authentication and authorization settings was memory-mapped by the `%s` utility, which was started using the `%s` arguments"
	KprobeUtilsRenamePamNoArgs          = "Detected that the `%s` file with the system authentication and authorization settings was replaced with the `%s` file by the `%s` utility"
	KprobeUtilsRenamePamDefault         = "Detected that the `%s` file with the system authentication and authorization settings was replaced with the `%s` file by the `%s` utility, which was started using the `%s` arguments"
	KprobeShellsWritePamNoArgs          = "Detected that the `%s` file with the system authentication and authorization settings was edited by the `%s` command shell"
	KprobeShellsWritePamDefault         = "Detected that the `%s` file with the system authentication and authorization settings was edited by the `%s` command shell, which was started using the `%s` arguments"
	KprobeShellsMmapPamNoArgs           = "Detected that the `%s` file with the system authentication and authorization settings was memory-mapped by the `%s` command shell"
	KprobeShellsMmapPamDefault          = "Detected that the `%s` file with the system authentication and authorization settings was memory-mapped by the `%s` command shell, which was started using the `%s` arguments"
	KprobeShellsRenamePamNoArgs         = "Detected that the `%s` file with the system authentication and authorization settings was replaced with the `%s` file by the `%s` command shell"
	KprobeShellsRenamePamDefault        = "Detected that the `%s` file with the system authentication and authorization settings was replaced with the `%s` file by the `%s` command shell, which was started using the `%s` arguments"
	KprobeDirsWritePamNoArgs            = "Detected that the `%s` file with the system authentication and authorization settings was edited by the executable file `%s` from the `%s` directory"
	KprobeDirsWritePamDefault           = "Detected that the `%s` file with the system authentication and authorization settings was edited by the executable file `%s` from the `%s` directory, which was started using the `%s` arguments"
	KprobeDirsMmapPamNoArgs             = "Detected that the `%s` file with the system authentication and authorization settings was memory-mapped by the executable file `%s` from the `%s` directory"
	KprobeDirsMmapPamDefault            = "Detected that the `%s` file with the system authentication and authorization settings was memory-mapped by the executable file `%s` from the `%s` directory, which was started using the `%s` arguments"
	KprobeDirsRenamePamNoArgs           = "Detected that the `%s` file with the system authentication and authorization settings was replaced with the `%s` file by the executable file `%s` from the `%s` directory"
	KprobeDirsRenamePamDefault          = "Detected that the `%s` file with the system authentication and authorization settings was replaced with the `%s` file by the executable file `%s` from the `%s` directory, which was started using the `%s` arguments"

	// KernelOptions
	KprobeUtilsWriteKernelOptsNoArgs    = "Detected that the `%s` file with the kernel settings was edited by the `%s` utility"
	KprobeUtilsWriteKernelOptsDefault   = "Detected that the `%s` file with the kernel settings was edited by the `%s` utility started using the `%s` arguments"
	KprobeUtilsMmapKernelOptsNoArgs     = "Detected that the `%s` file with the kernel settings was memory-mapped by the `%s` utility"
	KprobeUtilsMmapKernelOptsDefault    = "Detected that the `%s` file with the kernel settings was memory-mapped by the `%s` utility started using the `%s` arguments"
	KprobeUtilsRenameKernelOptsNoArgs   = "Detected that the `%s` file with the kernel settings was replaced with the `%s` file by the `%s` utility"
	KprobeUtilsRenameKernelOptsDefault  = "Detected that the `%s` file with the kernel settings was replaced with the `%s` file by the `%s` utility started using the `%s` arguments"
	KprobeShellsWriteKernelOptsNoArgs   = "Detected that the `%s` file with the kernel settings was edited by the `%s` command shell"
	KprobeShellsWriteKernelOptsDefault  = "Detected that the `%s` file with the kernel settings was edited by the `%s` command shell started using the `%s` arguments"
	KprobeShellsMmapKernelOptsNoArgs    = "Detected that the `%s` file with the kernel settings was memory-mapped by the `%s` command shell"
	KprobeShellsMmapKernelOptsDefault   = "Detected that the `%s` file with the kernel settings was memory-mapped by the `%s` command shell started using the `%s` arguments"
	KprobeShellsRenameKernelOptsNoArgs  = "Detected that the `%s` file with the kernel settings was replaced with the `%s` file by the `%s` command shell"
	KprobeShellsRenameKernelOptsDefault = "Detected that the `%s` file with the kernel settings was replaced with the `%s` file by the `%s` command shell started using the `%s` arguments"
	KprobeDirsWriteKernelOptsNoArgs     = "Detected that the `%s` file with the kernel settings was edited by the executable file `%s` from the `%s` directory"
	KprobeDirsWriteKernelOptsDefault    = "Detected that the `%s` file with the kernel settings was edited by the executable file `%s` from the `%s` directory started using the `%s` arguments"
	KprobeDirsMmapKernelOptsNoArgs      = "Detected that the `%s` file with the kernel settings was memory-mapped by the executable file `%s` from the `%s` directory"
	KprobeDirsMmapKernelOptsDefault     = "Detected that the `%s` file with the kernel settings was memory-mapped by the executable file `%s` from the `%s` directory started using the `%s` arguments"
	KprobeDirsRenameKernelOptsNoArgs    = "Detected that the `%s` file with the kernel settings was replaced with the `%s` file by the executable file `%s` from the `%s` directory"
	KprobeDirsRenameKernelOptsDefault   = "Detected that the `%s` file with the kernel settings was replaced with the `%s` file by the executable file `%s` from the `%s` directory started using the `%s` arguments"
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

var (
	mitreTactics = []*api.MitreTactic{
		{
			Id: "TA0003",
			Techniques: []string{
				"T1098",
				"T1136.001",
				"T1556.003",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1098",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1556.003",
				"T1562.001",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1556.003",
			},
		},
		{
			Id: "TA0040",
			Techniques: []string{
				"T1531",
			},
		},
	}

	mitreTacticsPam = []*api.MitreTactic{
		{
			Id: "TA0003",
			Techniques: []string{
				"T1556.003",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1556.003",
				"T1562.001",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1556.003",
			},
		},
	}

	mitreTacticsAccounts = []*api.MitreTactic{
		{
			Id: "TA0003",
			Techniques: []string{
				"T1098",
				"T1136.001",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1098",
			},
		},
		{
			Id: "TA0040",
			Techniques: []string{
				"T1531",
			},
		},
	}

	mitreTacticsKernel = []*api.MitreTactic{
		{
			Id: "TA0005",
			Techniques: []string{
				"T1562",
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

type sensitiveFile struct {
	pattern glob.Glob
	class   string
}

var (
	sensitiveFiles = []sensitiveFile{
		// Accounts info
		{pattern: glob.MustCompile("/etc/passwd"), class: "Accounts"},
		// Accounts password info
		{pattern: glob.MustCompile("/etc/shadow"), class: "Accounts"},
		// Groups info
		{pattern: glob.MustCompile("/etc/group"), class: "Accounts"},
		// Groups password info
		{pattern: glob.MustCompile("/etc/gshadow"), class: "Accounts"},
		// Superuser security policies
		{pattern: glob.MustCompile("/etc/sudoers*"), class: "Accounts"},
		// Shadow utils settings
		{pattern: glob.MustCompile("/etc/login.defs"), class: "PAM"},
		// Authentication module settings
		{pattern: glob.MustCompile("/etc/pam.*"), class: "PAM"},
		// Password policy (PAM)
		{pattern: glob.MustCompile("/etc/security/pwquality.conf"), class: "PAM"},
		// Password history (PAM)
		{pattern: glob.MustCompile("/etc/security/opasswd"), class: "PAM"},
		// Kernel parameters
		{pattern: glob.MustCompile("/etc/sysctl.*"), class: "KernelOptions"},
	}

	writeUtils = []glob.Glob{
		glob.MustCompile("*/awk"),
		glob.MustCompile("*/cp"),
		glob.MustCompile("*/dd"),
		glob.MustCompile("*/diff"),
		glob.MustCompile("*/emacs"),
		glob.MustCompile("*/gawk"),
		glob.MustCompile("*/java"),
		glob.MustCompile("*/mc"),
		glob.MustCompile("*/mcdiff"),
		glob.MustCompile("*/mcedit"),
		glob.MustCompile("*/mcview"),
		glob.MustCompile("*/mv"),
		glob.MustCompile("*/nano"),
		glob.MustCompile("*/perl"),
		glob.MustCompile("*/php"),
		glob.MustCompile("*/python*"),
		glob.MustCompile("*/ruby"),
		glob.MustCompile("*/sed"),
		glob.MustCompile("*/vi"),
		glob.MustCompile("*/vim"),
	}

	shells = []glob.Glob{
		glob.MustCompile("*/ash"),
		glob.MustCompile("*/bash"),
		glob.MustCompile("*/csh"),
		glob.MustCompile("*/dash"),
		glob.MustCompile("*/ksh"),
		glob.MustCompile("*/sh"),
		glob.MustCompile("*/tcsh"),
		glob.MustCompile("*/zsh"),
		glob.MustCompile("*/pwsh"),
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
		path := ""
		newFile := ""
		action := ""

		switch function {
		// Trigger when security function check for file write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_WRITE {
				return resp, nil
			}

			action = "write"
			path = args[0].GetFileArg().GetPath()

		// Trigger when security function check for memory page write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_WRITE == 0 {
				return resp, nil
			}

			action = "mmap"
			path = args[0].GetFileArg().GetPath()

		// Trigger when security function check if truncating a file is allowed.
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
			case file.pattern.Match(path) && (file.class == "KernelOptions"):
				resp.TacticsCovered = mitreTacticsKernel
				affectedFile = file
				sensitiveFile = true
			case file.pattern.Match(path):
				resp.TacticsCovered = mitreTactics
				affectedFile = file
				sensitiveFile = true
			}
		}

		if !sensitiveFile {
			return resp, nil
		}

		// Find out sensitive file reading method.
		// Trigger on utils for implicit writing
		for _, util := range writeUtils {
			if !util.Match(binary) {
				continue
			}
			switch {
			case (action == "write") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsWriteAccountsNoArgs, path, binary)
			case (action == "write") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsWriteAccountsDefault, path, binary, binaryArgs)
			case (action == "write") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsWritePamNoArgs, path, binary)
			case (action == "write") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsWritePamDefault, path, binary, binaryArgs)
			case (action == "write") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsWriteKernelOptsNoArgs, path, binary)
			case (action == "write") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsWriteKernelOptsDefault, path, binary, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsMmapAccountsNoArgs, path, binary)
			case (action == "mmap") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsMmapAccountsDefault, path, binary, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsMmapPamNoArgs, path, binary)
			case (action == "mmap") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsMmapPamDefault, path, binary, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsMmapKernelOptsNoArgs, path, binary)
			case (action == "mmap") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsMmapKernelOptsDefault, path, binary, binaryArgs)
			case (action == "rename") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsRenameAccountsNoArgs, path, newFile, binary)
			case (action == "rename") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsRenameAccountsDefault, path, newFile, binary, binaryArgs)
			case (action == "rename") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsRenamePamNoArgs, path, newFile, binary)
			case (action == "rename") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsRenamePamDefault, path, newFile, binary, binaryArgs)
			case (action == "rename") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsRenameKernelOptsNoArgs, path, newFile, binary)
			case (action == "rename") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeUtilsRenameKernelOptsDefault, path, newFile, binary, binaryArgs)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

			return resp, nil
		}

		// Trigger on write using shell
		for _, shell := range shells {
			if !shell.Match(binary) {
				continue
			}
			switch {
			case (action == "write") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsWriteAccountsNoArgs, path, binary)
			case (action == "write") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsWriteAccountsDefault, path, binary, binaryArgs)
			case (action == "write") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsWritePamNoArgs, path, binary)
			case (action == "write") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsWritePamDefault, path, binary, binaryArgs)
			case (action == "write") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsWriteKernelOptsNoArgs, path, binary)
			case (action == "write") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsWriteKernelOptsDefault, path, binary, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsMmapAccountsNoArgs, path, binary)
			case (action == "mmap") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsMmapAccountsDefault, path, binary, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsMmapPamNoArgs, path, binary)
			case (action == "mmap") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsMmapPamDefault, path, binary, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsMmapKernelOptsNoArgs, path, binary)
			case (action == "mmap") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsMmapKernelOptsDefault, path, binary, binaryArgs)
			case (action == "rename") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsRenameAccountsNoArgs, path, newFile, binary)
			case (action == "rename") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsRenameAccountsDefault, path, newFile, binary, binaryArgs)
			case (action == "rename") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsRenamePamNoArgs, path, newFile, binary)
			case (action == "rename") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsRenamePamDefault, path, newFile, binary, binaryArgs)
			case (action == "rename") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeShellsRenameKernelOptsNoArgs, path, newFile, binary)
			case (action == "rename") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeShellsRenameKernelOptsDefault, path, newFile, binary, binaryArgs)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

			return resp, nil
		}

		// Trigger on utils from suspicious directories
		for _, dir := range suspDirs {
			if !dir.Match(binary) {
				continue
			}
			binDir := binary[0 : strings.LastIndex(binary, "/") + 1]
			switch {
			case (action == "write") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsWriteAccountsNoArgs, path, binary, binDir)
			case (action == "write") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsWriteAccountsDefault, path, binary, binDir, binaryArgs)
			case (action == "write") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsWritePamNoArgs, path, binary, binDir)
			case (action == "write") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsWritePamDefault, path, binary, binDir, binaryArgs)
			case (action == "write") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsWriteKernelOptsNoArgs, path, binary, binDir)
			case (action == "write") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsWriteKernelOptsDefault, path, binary, binDir, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsMmapAccountsNoArgs, path, binary, binDir)
			case (action == "mmap") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsMmapAccountsDefault, path, binary, binDir, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsMmapPamNoArgs, path, binary, binDir)
			case (action == "mmap") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsMmapPamDefault, path, binary, binDir, binaryArgs)
			case (action == "mmap") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsMmapKernelOptsNoArgs, path, binary, binDir)
			case (action == "mmap") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsMmapKernelOptsDefault, path, binary, binDir, binaryArgs)
			case (action == "rename") && (affectedFile.class == "Accounts") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsRenameAccountsNoArgs, path, newFile, binary, binDir)
			case (action == "rename") && (affectedFile.class == "Accounts") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsRenameAccountsDefault, path, newFile, binary, binDir, binaryArgs)
			case (action == "rename") && (affectedFile.class == "PAM") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsRenamePamNoArgs, path, newFile, binary, binDir)
			case (action == "rename") && (affectedFile.class == "PAM") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsRenamePamDefault, path, newFile, binary, binDir, binaryArgs)
			case (action == "rename") && (affectedFile.class == "KernelOptions") && (binaryArgs == ""):
				resp.Reason = fmt.Sprintf(KprobeDirsRenameKernelOptsNoArgs, path, newFile, binary, binDir)
			case (action == "rename") && (affectedFile.class == "KernelOptions") && (binaryArgs != ""):
				resp.Reason = fmt.Sprintf(KprobeDirsRenameKernelOptsDefault, path, newFile, binary, binDir, binaryArgs)
			}
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

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
