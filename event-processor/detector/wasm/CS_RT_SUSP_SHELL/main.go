//go:build tinygo.wasm

package main

import (
	"context"

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

// main is required for TinyGo to compile to Wasm.
func main() {
	api.RegisterDetector(Detector{})
}

type Detector struct{}

func (d Detector) Info(ctx context.Context, req *api.InfoReq) (*api.InfoResp, error) {
	return &api.InfoResp{
		Id:          ID,
		Name:        Name,
		Description: Description,
		Version:     Version,
		Author:      Author,
		License:     License,
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

/* Example event (JSON):

{
    "process_exec": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODkzODA1NzY1NDI3Mjc6MTczODQ2NA==",
            "pid": 1738464,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd clone",
            "start_time": "2024-03-22T12:56:47.369992957Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "labels": [],
                "container": {
                    "id": "cri-o://332cff0fb99f03f8b4fb9633f245ba21ad43339eeafd3b0ac3e5d9a38371262c",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-02-27T13:26:02Z",
                    "pid": 62632,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODkzNzUyOTI1MDExMDY6MTczNzA5OA==",
            "refcnt": 0,
            "cap": {
                "permitted": [
                    "CAP_CHOWN",
                    "DAC_OVERRIDE",
                    "CAP_FOWNER",
                    "CAP_FSETID",
                    "CAP_KILL",
                    "CAP_SETGID",
                    "CAP_SETUID",
                    "CAP_SETPCAP",
                    "CAP_NET_BIND_SERVICE"
                ],
                "effective": [
                    "CAP_CHOWN",
                    "DAC_OVERRIDE",
                    "CAP_FOWNER",
                    "CAP_FSETID",
                    "CAP_KILL",
                    "CAP_SETGID",
                    "CAP_SETUID",
                    "CAP_SETPCAP",
                    "CAP_NET_BIND_SERVICE"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026534978,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534979,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026535073,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026535074,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026535074,
                    "is_host": false
                },
                "net": {
                    "inum": 4026534980,
                    "is_host": false
                },
                "time": {
                    "inum": 4026531834,
                    "is_host": true
                },
                "time_for_children": {
                    "inum": 4026531834,
                    "is_host": true
                },
                "cgroup": {
                    "inum": 4026535075,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1738464,
            "process_credentials": {
                "uid": 0,
                "gid": 0,
                "euid": 0,
                "egid": 0,
                "suid": 0,
                "sgid": 0,
                "fsuid": 0,
                "fsgid": 0,
                "securebits": [],
                "caps": null,
                "user_ns": null
            },
            "binary_properties": null
        },
        "parent": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODkzNzUyOTI1MDExMDY6MTczNzA5OA==",
            "pid": 1737098,
            "uid": 0,
            "cwd": "/",
            "binary": "/tmp/linpeas.sh",
            "arguments": "/tmp/linpeas.sh",
            "flags": "execve rootcwd clone",
            "start_time": "2024-03-22T12:56:42.085951573Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "labels": [],
                "container": {
                    "id": "cri-o://332cff0fb99f03f8b4fb9633f245ba21ad43339eeafd3b0ac3e5d9a38371262c",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-02-27T13:26:02Z",
                    "pid": 61338,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODY1NjA1MDI5Mzg4NDQ6MTcwODI4Mg==",
            "refcnt": 0,
            "cap": {
                "permitted": [
                    "CAP_CHOWN",
                    "DAC_OVERRIDE",
                    "CAP_FOWNER",
                    "CAP_FSETID",
                    "CAP_KILL",
                    "CAP_SETGID",
                    "CAP_SETUID",
                    "CAP_SETPCAP",
                    "CAP_NET_BIND_SERVICE"
                ],
                "effective": [
                    "CAP_CHOWN",
                    "DAC_OVERRIDE",
                    "CAP_FOWNER",
                    "CAP_FSETID",
                    "CAP_KILL",
                    "CAP_SETGID",
                    "CAP_SETUID",
                    "CAP_SETPCAP",
                    "CAP_NET_BIND_SERVICE"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026534978,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534979,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026535073,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026535074,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026535074,
                    "is_host": false
                },
                "net": {
                    "inum": 4026534980,
                    "is_host": false
                },
                "time": {
                    "inum": 4026531834,
                    "is_host": true
                },
                "time_for_children": {
                    "inum": 4026531834,
                    "is_host": true
                },
                "cgroup": {
                    "inum": 4026535075,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1737098,
            "process_credentials": {
                "uid": 0,
                "gid": 0,
                "euid": 0,
                "egid": 0,
                "suid": 0,
                "sgid": 0,
                "fsuid": 0,
                "fsgid": 0,
                "securebits": [],
                "caps": null,
                "user_ns": null
            },
            "binary_properties": null
        },
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-03-22T12:56:47.369992646Z",
    "aggregation_info": null
}

*/
