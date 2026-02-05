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
	ID          = "CS_RT_SUSP_FILE_READ"
	Name        = "Suspicious reading of sensitive system files"
	Description = "The detector detects suspicious reading of system files with utilities from uncommon directories, such as /tmp and /home, which may indicate that system configuration data is being collected."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	// File access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/linux/fs.h#L101
	MAY_READ = 4

	// Memory page access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L10
	PROT_READ = 1
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
	sensitiveFiles = []glob.Glob{
		glob.MustCompile("/etc/shadow"),                  // users shadow info
		glob.MustCompile("/etc/sudoers*"),                // superuser security policies
		glob.MustCompile("/etc/pam.*"),                   // authentication module settings
		glob.MustCompile("/etc/security/pwquality.conf"), // password policy (pam)
		glob.MustCompile("/etc/*-release"),               // distribution info
		glob.MustCompile("/usr/lib/os-release"),          // distribution info (for the most distributions /etc/os-release is a symlink to /usr/lib/os-release)
		glob.MustCompile("/proc/*/environ"),              // environment variables
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
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

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

		sensitiveFile := false

		for _, file := range sensitiveFiles {
			if file.Match(path) {
				sensitiveFile = true
				break
			}
		}

		if !sensitiveFile {
			return resp, nil
		}

		// find out sensitive file reading method
		// trigger on specific utils reading sensitive files
		for _, util := range readUtils {
			if util.Match(binary) {
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			}
		}

		// trigger on utils from suspicious directories
		for _, dir := range suspDirs {
			if dir.Match(binary) {
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

/* Example event (JSON):

{
    "process_kprobe": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM0NDI1Nzk5OTE2NjMzNDQ6MzI1OTM0MA==",
            "pid": 3259340,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/head",
            "arguments": "/etc/shadow",
            "flags": "execve rootcwd clone",
            "start_time": "2024-03-24T07:30:06.785114023Z",
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
                    "pid": 70734,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM0NDE2OTY2NzQwOTM2OTM6MzI1MDU5MQ==",
            "refcnt": 1,
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
            "tid": 3259340,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM0NDE2OTY2NzQwOTM2OTM6MzI1MDU5MQ==",
            "pid": 3250591,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-03-24T07:15:23.467544178Z",
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
                    "pid": 70721,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM0NDE2OTY2NzI2NjUwMDI6MzI1MDU5MQ==",
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
            "tid": 3250591,
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
        "function_name": "security_file_permission",
        "args": [
            {
                "file_arg": {
                    "mount": "",
                    "path": "/etc/shadow",
                    "flags": ""
                },
                "label": ""
            },
            {
                "int_arg": 4,
                "label": ""
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "stack_trace": [],
        "policy_name": "file-monitoring"
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-03-24T07:30:06.785822196Z",
    "aggregation_info": null
}

*/
