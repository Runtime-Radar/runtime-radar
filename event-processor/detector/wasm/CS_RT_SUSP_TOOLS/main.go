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
	ID          = "CS_RT_SUSP_TOOLS"
	Name        = "Start of suspicious utilities and executable files"
	Description = "The detector detects if suspicious system utilities were started and files were executed from potentially harmful directories (for example, /tmp, /sys, or /boot). Such activity may indicate an attempt to carry out an LOLBins attack as well as execution of malicious code loaded by an attacker."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecPrivilegesNoArgs  = "Detected a suspicious start of the privilege management utility `%s`"
	ExecPrivilegesDefault = "Detected a suspicious start of the privilege management utility `%s` with the `%s` arguments"
	ExecNetworkNoArgs     = "Detected a suspicious start of the network utility `%s`"
	ExecNetworkDefault    = "Detected a suspicious start of the network utility `%s` with the `%s` arguments"
	ExecSoftwareNoArgs    = "Detected a suspicious start of the package and application installation utility `%s`"
	ExecSoftwareDefault   = "Detected a suspicious start of the package and application installation utility `%s` with the `%s` arguments"
	ExecDownloadNoArgs    = "Detected a suspicious start of the file downloading utility `%s`"
	ExecDownloadDefault   = "Detected a suspicious start of the file downloading utility `%s` with the `%s` arguments"
	ExecContainersNoArgs  = "Detected a suspicious start of the utility `%s` designed for exploiting the container infrastructure"
	ExecContainersDefault = "Detected a suspicious start of the utility `%s` designed for exploiting the container infrastructure with the `%s` arguments"
	ExecEnvironNoArgs     = "Detected a suspicious start of the utility `%s` designed for obtaining environment variables"
	ExecEnvironDefault    = "Detected a suspicious start of the utility `%s` designed for obtaining environment variables with the `%s` arguments"
	ExecNohupNoArgs       = "Detected a suspicious start of the utility `%s` designed for ignoring process termination signals"
	ExecNohupDefault      = "Detected a suspicious start of the utility `%s` designed for ignoring process termination signals with the `%s` arguments"
	ExecMiscNoArgs        = "Detected a suspicious start of the `%s` utility"
	ExecMiscDefault       = "Detected a suspicious start of the `%s` utility with the `%s` arguments"
	ExecSuspDirsNoArgs    = "Detected a suspicious start of the executable file `%s` from the `%s` directory, which does not contain executable files"
	ExecSuspDirsDefault   = "Detected a suspicious start of the executable file `%s` with the `%s` arguments from the `%s` directory, which does not contain executable files"
	ExecNohupChildNoArgs  = "Detected a suspicious start of the `%s` process by the nohup parent process to ignore process termination signals"
	ExecNohupChildDefault = "Detected a suspicious start of the `%s` process with the `%s` arguments by the nohup parent process to ignore process termination signals"
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
				"T1609",
				"T1053.007",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1098.006",
				"T1543.005",
				"T1053.007",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1098.006",
				"T1543.005",
				"T1611",
				"T1068",
				"T1053.007",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1564.011",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1552.007",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1613",
				"T1016.001",
				"T1049",
				"T1033",
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

	mitreTacticsPrivileges = []*api.MitreTactic{
		{
			Id: "TA0007",
			Techniques: []string{
				"T1033",
			},
		},
	}

	mitreTacticsDownload = []*api.MitreTactic{
		{
			Id: "TA0007",
			Techniques: []string{
				"T1016.001",
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

	mitreTacticsNetwork = []*api.MitreTactic{
		{
			Id: "TA0007",
			Techniques: []string{
				"T1016.001",
				"T1049",
			},
		},
	}

	mitreTacticsContainers = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1609",
				"T1053.007",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1098.006",
				"T1543.005",
				"T1053.007",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1098.006",
				"T1543.005",
				"T1611",
				"T1068",
				"T1053.007",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1552.007",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1613",
			},
		},
	}

	mitreTacticsNohup = []*api.MitreTactic{
		{
			Id: "TA0005",
			Techniques: []string{
				"T1564.011",
			},
		},
	}
)

type suspiciousTool struct {
	pattern glob.Glob
	class   string
}

var (
	suspiciousTools = []suspiciousTool{
		// sudo and su a rarely used in runtime, containers are either rooted or rootless with no need to escalate privileges for a normal user.
		{pattern: glob.MustCompile("*/sudo"), class: "Privileges"},
		{pattern: glob.MustCompile("*/su"), class: "Privileges"},
		{pattern: glob.MustCompile("*/id"), class: "Privileges"},
		{pattern: glob.MustCompile("*/whoami"), class: "Privileges"},

		// Network tools rarely used in containers
		{pattern: glob.MustCompile("*/arp"), class: "Network"},
		{pattern: glob.MustCompile("*/ifconfig"), class: "Network"},
		{pattern: glob.MustCompile("*/ip"), class: "Network"},
		{pattern: glob.MustCompile("*/netstat"), class: "Network"},
		{pattern: glob.MustCompile("*/ss"), class: "Network"},

		// Usually required packets and applications installed during image build stage, otherwise suspicious
		{pattern: glob.MustCompile("*/apk"), class: "Software"},
		{pattern: glob.MustCompile("*/apt"), class: "Software"},
		{pattern: glob.MustCompile("*/apt-get"), class: "Software"},
		{pattern: glob.MustCompile("*/dnf"), class: "Software"},
		{pattern: glob.MustCompile("*/dpkg"), class: "Software"},
		{pattern: glob.MustCompile("*/git"), class: "Software"},
		{pattern: glob.MustCompile("*/git-remote-*"), class: "Software"},
		{pattern: glob.MustCompile("*/rpm"), class: "Software"},
		{pattern: glob.MustCompile("*/svn"), class: "Software"},
		{pattern: glob.MustCompile("*/yum"), class: "Software"},

		// Download tools
		{pattern: glob.MustCompile("*/curl"), class: "Download"},
		{pattern: glob.MustCompile("*/ftp"), class: "Download"},
		{pattern: glob.MustCompile("*/sftp"), class: "Download"},
		{pattern: glob.MustCompile("*/sftp-server"), class: "Download"},
		{pattern: glob.MustCompile("*/wget"), class: "Download"},

		// Containers and namespace tools
		{pattern: glob.MustCompile("*/crictl"), class: "Containers"},
		{pattern: glob.MustCompile("*/docker"), class: "Containers"},
		{pattern: glob.MustCompile("*/kubeadm"), class: "Containers"},
		{pattern: glob.MustCompile("*/kubectl"), class: "Containers"},
		{pattern: glob.MustCompile("*/nsenter"), class: "Containers"},
		{pattern: glob.MustCompile("*/podman"), class: "Containers"},
		{pattern: glob.MustCompile("*/unshare"), class: "Containers"},

		// Environment variables
		{pattern: glob.MustCompile("*/env"), class: "Environ"},
		{pattern: glob.MustCompile("*/printenv"), class: "Environ"},

		// Nohup
		{pattern: glob.MustCompile("*/nohup"), class: "Nohup"},

		// Miscellaneous
		{pattern: glob.MustCompile("*/dd"), class: "Misc"},
	}

	// Directories that normally do not contain binaries.
	suspDirs = []glob.Glob{
		glob.MustCompile("/boot/*"),
		glob.MustCompile("/dev/*"),
		glob.MustCompile("/home/*"),
		glob.MustCompile("/media/*"),
		glob.MustCompile("/mnt/*"),
		glob.MustCompile("/run/*"),
		glob.MustCompile("/srv/*"),
		glob.MustCompile("/sys/*"),
		glob.MustCompile("/tmp/*"),
		glob.MustCompile("/var/*"),
	}

	nohupBin = glob.MustCompile(`*/nohup`)
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

		// Trigger on binary if its parent binary is nohup
		if nohupBin.Match(parentBinary) && (args == "") {
			resp.TacticsCovered = mitreTacticsNohup
			resp.Reason = fmt.Sprintf(ExecNohupChildNoArgs, binary)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		} else if nohupBin.Match(parentBinary) {
			resp.TacticsCovered = mitreTacticsNohup
			resp.Reason = fmt.Sprintf(ExecNohupChildDefault, binary, args)
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		}

		// Trigger on process from suspicious directories
		for _, dir := range suspDirs {
			binDir := binary[0 : strings.LastIndex(binary, "/")+1]
			if dir.Match(binary) && (args == "") {
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSuspDirsNoArgs, binary, binDir)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			} else if dir.Match(binary) {
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSuspDirsDefault, binary, args, binDir)
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

		for _, tool := range suspiciousTools {
			if !tool.pattern.Match(binary) {
				continue
			}
			switch {
			case (tool.class == "Privileges") && (args == ""):
				resp.TacticsCovered = mitreTacticsPrivileges
				resp.Reason = fmt.Sprintf(ExecPrivilegesNoArgs, binary)
			case (tool.class == "Privileges") && (args != ""):
				resp.TacticsCovered = mitreTacticsPrivileges
				resp.Reason = fmt.Sprintf(ExecPrivilegesDefault, binary, args)
			case (tool.class == "Network") && (args == ""):
				resp.TacticsCovered = mitreTacticsNetwork
				resp.Reason = fmt.Sprintf(ExecNetworkNoArgs, binary)
			case (tool.class == "Network") && (args != ""):
				resp.TacticsCovered = mitreTacticsNetwork
				resp.Reason = fmt.Sprintf(ExecNetworkDefault, binary, args)
			case (tool.class == "Software") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSoftwareNoArgs, binary)
			case (tool.class == "Software") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSoftwareDefault, binary, args)
			case (tool.class == "Download") && (args == ""):
				resp.TacticsCovered = mitreTacticsDownload
				resp.Reason = fmt.Sprintf(ExecDownloadNoArgs, binary)
			case (tool.class == "Download") && (args != ""):
				resp.TacticsCovered = mitreTacticsDownload
				resp.Reason = fmt.Sprintf(ExecDownloadDefault, binary, args)
			case (tool.class == "Containers") && (args == ""):
				resp.TacticsCovered = mitreTacticsContainers
				resp.Reason = fmt.Sprintf(ExecContainersNoArgs, binary)
			case (tool.class == "Containers") && (args != ""):
				resp.TacticsCovered = mitreTacticsContainers
				resp.Reason = fmt.Sprintf(ExecContainersDefault, binary, args)
			case (tool.class == "Environ") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecEnvironNoArgs, binary)
			case (tool.class == "Environ") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecEnvironDefault, binary, args)
			case (tool.class == "Nohup") && (args == ""):
				resp.TacticsCovered = mitreTacticsNohup
				resp.Reason = fmt.Sprintf(ExecNohupNoArgs, binary)
			case (tool.class == "Nohup") && (args != ""):
				resp.TacticsCovered = mitreTacticsNohup
				resp.Reason = fmt.Sprintf(ExecNohupDefault, binary, args)
			case (tool.class == "Misc") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecMiscNoArgs, binary)
			case (tool.class == "Misc") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecMiscDefault, binary, args)
			}
			resp.Severity = api.DetectResp_LOW // <-- threat detected

			return resp, nil
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

