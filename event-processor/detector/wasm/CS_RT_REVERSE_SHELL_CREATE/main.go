//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_REVERSE_SHELL_CREATE"
	Name        = "Reverse shell creation"
	Description = "The detector detects signs that a reverse shell was created."
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
		"PROCESS_KPROBE": {"do_dup2"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
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
		functionName := kprobe.GetFunctionName()

		if functionName == "do_dup2" {
			args := kprobe.GetArgs()

			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			}

			fileArg := args[0].GetFileArg()
			filePerm := fileArg.GetPermission()

			// Trigger when stdin file descriptor (0) is duplicated into a socket.
			if filePerm[0] == 's' {
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			}
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNjgzOTcxNTEzODQxMTU6NzU3MzQ0",
            "pid": 757344,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve",
            "start_time": "2024-08-08T14:25:32.797594902Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "cri-o://426ccd7bdd6e9565a3f2767765b0c1fc160c8132c331884a6000759307b4fae2",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "31d5e503c34f4496a263fb3557575cf53e6a40add4c459370120c7454985f7b7",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-18T19:36:11Z",
                    "pid": 4194,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "426ccd7bdd6e9565a3f2767765b0c1f",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNjgxODA2NTc2NDI1NDE6NzU1MDgw",
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
                    "inum": 4026532264,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532265,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026532792,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532266,
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
                    "inum": 4026532794,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 757344,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNjgxODA2NTc2NDI1NDE6NzU1MDgw",
            "pid": 755080,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-08-08T14:21:56.303852687Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "cri-o://426ccd7bdd6e9565a3f2767765b0c1fc160c8132c331884a6000759307b4fae2",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "31d5e503c34f4496a263fb3557575cf53e6a40add4c459370120c7454985f7b7",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-18T19:36:11Z",
                    "pid": 4194,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "426ccd7bdd6e9565a3f2767765b0c1f",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNjgxODA2NTY3MDkzMjY6NzU1MDgw",
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
                    "inum": 4026532264,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532265,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026532792,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532266,
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
                    "inum": 4026532794,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 755080,
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
        "function_name": "do_dup2",
        "args": [
            {
                "file_arg": {
                    "mount": "",
                    "path": "",
                    "flags": "",
                    "permission": "srwxrwxrwx"
                },
                "label": ""
            },
            {
                "int_arg": 0,
                "label": "fd"
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "dup",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-08-08T14:25:32.798469136Z",
    "aggregation_info": null
}

*/
