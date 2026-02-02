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
	ID          = "CS_RT_CORE_PATTERN_MODIFY"
	Name        = "Change of generation settings for process dump file"
	Description = "The detector detects if the core_pattern file was edited in the procfs file system, which may indicate an attacker's attempt to elevate their privileges and leave the isolated container environment."
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
		"PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
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
	corePattern = glob.MustCompile("*/core_pattern")
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
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()
		path := ""

		switch function {
		// Trigger when security function check for file write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_WRITE {
				return resp, nil
			}

			path = args[0].GetFileArg().GetPath()

		// Trigger when security function check for memory page write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_WRITE == 0 {
				return resp, nil
			}

			path = args[0].GetFileArg().GetPath()

		// Trigger when security function check if truncating a file is allowed.
		// https://elixir.bootlin.com/linux/v6.10.6/source/security/security.c#L1923
		case "security_path_truncate":
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			path = args[0].GetPathArg().GetPath()

		default:
			return resp, nil
		}

		if corePattern.Match(path) {
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

			return resp, nil
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
            "exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDo2NDI2MTQxMjE4NTI4MTU5OjIwNTg0ODI=",
            "pid": 2058482,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "-c \"echo test > /sysproc/core_pattern\"",
            "flags": "execve rootcwd clone",
            "start_time": "2025-07-28T13:34:05.420097350Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-umh",
                "container": {
                    "id": "containerd://6d8981b036e9fcc730a8746235e218f4d85341b1d9b526fd49a773c0917d3230",
                    "name": "test-pod-debian-umh",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-07-25T14:20:32Z",
                    "pid": 1571,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-umh",
                "workload_kind": "Pod"
            },
            "docker": "6d8981b036e9fcc730a8746235e218f",
            "parent_exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDo2NDI2MTI1MDgwOTg3MjczOjIwNTgyNTQ=",
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
                    "CAP_NET_BIND_SERVICE",
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
                    "CAP_SETFCAP"
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
                    "CAP_NET_BIND_SERVICE",
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
                    "CAP_SETFCAP"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026534486,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534487,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534489,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534490,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534490,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533923,
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
                    "inum": 4026534491,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2058482,
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
            "binary_properties": null,
            "user": null
        },
        "parent": {
            "exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDo2NDI2MTI1MDgwOTg3MjczOjIwNTgyNTQ=",
            "pid": 2058254,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2025-07-28T13:33:49.282556644Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-umh",
                "container": {
                    "id": "containerd://6d8981b036e9fcc730a8746235e218f4d85341b1d9b526fd49a773c0917d3230",
                    "name": "test-pod-debian-umh",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-07-25T14:20:32Z",
                    "pid": 1565,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-umh",
                "workload_kind": "Pod"
            },
            "docker": "6d8981b036e9fcc730a8746235e218f",
            "parent_exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDo2NDI2MTI1MDc5MzAwMzI5OjIwNTgyNTQ=",
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
                    "CAP_NET_BIND_SERVICE",
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
                    "CAP_SETFCAP"
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
                    "CAP_NET_BIND_SERVICE",
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
                    "CAP_SETFCAP"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026534486,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534487,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534489,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534490,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534490,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533923,
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
                    "inum": 4026534491,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2058254,
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
            "binary_properties": null,
            "user": null
        },
        "function_name": "security_file_permission",
        "args": [
            {
                "file_arg": {
                    "mount": "",
                    "path": "/sysproc/core_pattern",
                    "flags": "",
                    "permission": "-rw-r--r--"
                },
                "label": ""
            },
            {
                "int_arg": 2,
                "label": ""
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "file-monitoring",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "szy-k8s.appsec-stand.cloud.local",
    "time": "2025-07-28T13:34:05.422534198Z",
    "aggregation_info": null
}

*/
