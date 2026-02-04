//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_REVERSE_SHELL_USE"
	Name        = "Reverse shell use"
	Description = "The detector detects signs that a reverse shell is used."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"tcp_sendmsg"},

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
	commonLinuxTools = []glob.Glob{
		glob.MustCompile("*/ls"),
		glob.MustCompile("*/cp"),
		glob.MustCompile("*/mv"),
		glob.MustCompile("*/chmod"),
		glob.MustCompile("*/chown"),
		glob.MustCompile("*/rmdir"),
		glob.MustCompile("*/ln"),
		glob.MustCompile("*/df"),
		glob.MustCompile("*/du"),
		glob.MustCompile("*/cat"),
		glob.MustCompile("*/[a-z]grep"),
		glob.MustCompile("*/id"),
		glob.MustCompile("*/whoami"),
		glob.MustCompile("*/uname"),
		glob.MustCompile("*/ps"),
		glob.MustCompile("*/top"),
		glob.MustCompile("*/make"),
		glob.MustCompile("*/gcc"),
		glob.MustCompile("*/g++"),
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

		if function != "tcp_sendmsg" {
			return resp, nil
		}

		for _, tool := range commonLinuxTools {
			if tool.Match(binary) {
				resp.Severity = api.DetectResp_HIGH
				return resp, nil // <-- threat detected
			}
		}

		for _, shell := range shells {
			if shell.Match(binary) {
				resp.Severity = api.DetectResp_HIGH
				return resp, nil // <-- threat detected
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
            "exec_id": "YWtvbWlzb3YtOC04LWV4cGVyaW1lbnRzLWFrb21pc292LWdyb3VwLTE6Mjc4MjAwMzM4MjU0NjM0OjM2NzQ1ODA=",
            "pid": 3674580,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/ls",
            "arguments": "-la",
            "flags": "execve rootcwd clone",
            "start_time": "2024-05-22T18:26:04.097059929Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "cri-o://9f1e6eb99ef555cc81c4b805f5e34c7270980bb48682331a8c95367b7702f6bd",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-19T17:38:40Z",
                    "pid": 803,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "9f1e6eb99ef555cc81c4b805f5e34c7",
            "parent_exec_id": "YWtvbWlzb3YtOC04LWV4cGVyaW1lbnRzLWFrb21pc292LWdyb3VwLTE6Mjc3NDc5MDY3MjE0MzQyOjM2NjUwOTM=",
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
                    "inum": 4026532835,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532836,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533197,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533198,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533198,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532837,
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
                    "inum": 4026531835,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3674580,
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
            "exec_id": "YWtvbWlzb3YtOC04LWV4cGVyaW1lbnRzLWFrb21pc292LWdyb3VwLTE6Mjc3NDc5MDY3MjE0MzQyOjM2NjUwOTM=",
            "pid": 3665093,
            "uid": 0,
            "cwd": "/",
            "binary": "/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-05-22T18:14:02.826032778Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "cri-o://9f1e6eb99ef555cc81c4b805f5e34c7270980bb48682331a8c95367b7702f6bd",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-19T17:38:40Z",
                    "pid": 783,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "9f1e6eb99ef555cc81c4b805f5e34c7",
            "parent_exec_id": "YWtvbWlzb3YtOC04LWV4cGVyaW1lbnRzLWFrb21pc292LWdyb3VwLTE6Mjc3NDc5MDU3MTU4MDYzOjM2NjUwOTM=",
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
                    "inum": 4026532835,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532836,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533197,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533198,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533198,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532837,
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
                    "inum": 4026531835,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3665093,
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
        "function_name": "tcp_sendmsg",
        "args": [
            {
                "sock_arg": {
                    "family": "AF_INET",
                    "type": "SOCK_STREAM",
                    "protocol": "IPPROTO_TCP",
                    "mark": 0,
                    "priority": 0,
                    "saddr": "10.100.132.73",
                    "daddr": "10.0.2.142",
                    "sport": 39424,
                    "dport": 8888,
                    "cookie": "18384485878554729216",
                    "state": "TCP_ESTABLISHED"
                },
                "label": ""
            },
            {
                "int_arg": 1174,
                "label": ""
            }
        ],
        "return": null,
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "connect",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "experiments",
    "time": "2024-05-22T18:26:04.098669109Z",
    "aggregation_info": null
}

*/
