//go:build tinygo.wasm

package main

import (
	"context"
	"strings"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_REV_ENG"
	Name        = "Debugging and reverse engineering"
	Description = "The detector detects the ptrace system calls used by special software for debugging and reverse engineering, which may indicate an attacker's activity in the target system."
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
		"PROCESS_KPROBE": {"sys_ptrace"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)

// init is required for TinyGo to compile to Wasm.
func init() {
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
		functionName := kprobe.GetFunctionName()

		if strings.HasSuffix(functionName, "sys_ptrace") {
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		}

	case *tetragon.GetEventsResponse_ProcessTracepoint:
		// Nothing here
	}

	return resp, nil
}

/* Example event (JSON):

{
    "process_kprobe": {
        "process": {
            "exec_id": "a2luZC1jb250cm9sLXBsYW5lOjQ0MDEzNTk4NTk4NzE0NDozNzQzNzc0",
            "pid": 3743774,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/strace",
            "arguments": "ls",
            "flags": "execve rootcwd clone",
            "start_time": "2023-11-16T16:39:52.483787451Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://e44de197bbd76a379eaa7863361455ac188af62ccd7b284a5c944183d9dc1d30",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:b55e2651b71408015f8068dd74e1d04404a8fa607dd2cfe284b4824c11f4d9bd",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2023-11-08T14:45:39Z",
                    "pid": 48403
                },
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "e44de197bbd76a379eaa7863361455a",
            "parent_exec_id": "a2luZC1jb250cm9sLXBsYW5lOjQzOTM5NTM3OTkxODgxOTozNzM3NDg3",
            "refcnt": 2,
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
                ]
            },
            "ns": {
                "uts": {
                    "inum": 4026532967
                },
                "ipc": {
                    "inum": 4026532972
                },
                "mnt": {
                    "inum": 4026533013
                },
                "pid": {
                    "inum": 4026533014
                },
                "pid_for_children": {
                    "inum": 4026533014
                },
                "net": {
                    "inum": 4026532532
                },
                "cgroup": {
                    "inum": 4026532330,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3743774,
            "process_credentials": {
                "uid": 0,
                "gid": 0,
                "euid": 0,
                "egid": 0,
                "suid": 0,
                "sgid": 0,
                "fsuid": 0,
                "fsgid": 0
            }
        },
        "parent": {
            "exec_id": "a2luZC1jb250cm9sLXBsYW5lOjQzOTM5NTM3OTkxODgxOTozNzM3NDg3",
            "pid": 3737487,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "flags": "execve rootcwd clone",
            "start_time": "2023-11-16T16:27:31.877718626Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://e44de197bbd76a379eaa7863361455ac188af62ccd7b284a5c944183d9dc1d30",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:b55e2651b71408015f8068dd74e1d04404a8fa607dd2cfe284b4824c11f4d9bd",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2023-11-08T14:45:39Z",
                    "pid": 48324
                },
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "e44de197bbd76a379eaa7863361455a",
            "parent_exec_id": "a2luZC1jb250cm9sLXBsYW5lOjQzOTM5NTM1MzY2OTg0MTozNzM3NDc3",
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
                ]
            },
            "ns": {
                "uts": {
                    "inum": 4026532967
                },
                "ipc": {
                    "inum": 4026532972
                },
                "mnt": {
                    "inum": 4026533013
                },
                "pid": {
                    "inum": 4026533014
                },
                "pid_for_children": {
                    "inum": 4026533014
                },
                "net": {
                    "inum": 4026532532
                },
                "cgroup": {
                    "inum": 4026532330,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3737487,
            "process_credentials": {
                "uid": 0,
                "gid": 0,
                "euid": 0,
                "egid": 0,
                "suid": 0,
                "sgid": 0,
                "fsuid": 0,
                "fsgid": 0
            }
        },
        "function_name": "__x64_sys_ptrace",
        "args": [
            {
                "int_arg": 0
            }
        ],
        "action": "KPROBE_ACTION_POST",
        "policy_name": "ptrace"
    },
    "node_name": "kind-control-plane",
    "time": "2023-11-16T16:39:52.484759149Z"
}

*/
