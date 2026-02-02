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
	ID          = "CS_RT_RECON_TOOLS"
	Name        = "Use of network reconnaissance tools"
	Description = "The detector detects attempts to start applications that use the libpcap library. Such attempts may indicate use of network reconnaissance tools."
	Version     = 1
	Author      = "Runtime Radar Team"

	License = "Apache License 2.0"
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
	libPath = glob.MustCompile("*/libpcap.so*")
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
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_READ == 0 {
				return resp, nil
			}

		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_READ {
				return resp, nil
			}

		default:
			return resp, nil
		}

		path = args[0].GetFileArg().GetPath()

		if libPath.Match(path) {
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM4OTAzNTE5ODYwMDYwNTQ6MzM2NjAyNA==",
            "pid": 3366024,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/nmap",
            "arguments": "localhost",
            "flags": "execve rootcwd clone",
            "start_time": "2024-07-01T15:51:27.632215549Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-raw",
                "container": {
                    "id": "cri-o://cf7b0a3fa193adee581364d4f46c6665284aa7785641c09ea2580fbfee8ac4f8",
                    "name": "test-pod-debian-raw",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-07-01T15:31:30Z",
                    "pid": 1697,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "cf7b0a3fa193adee581364d4f46c666",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM4OTAzMDQyMTA3ODExMTc6MzM2NTM2OQ==",
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
                    "CAP_NET_RAW"
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
                    "CAP_NET_RAW"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026532263,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533285,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533567,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533568,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533568,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533286,
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
                    "inum": 4026533569,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3366024,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM4OTAzMDQyMTA3ODExMTc6MzM2NTM2OQ==",
            "pid": 3365369,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-07-01T15:50:39.856990647Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-raw",
                "container": {
                    "id": "cri-o://cf7b0a3fa193adee581364d4f46c6665284aa7785641c09ea2580fbfee8ac4f8",
                    "name": "test-pod-debian-raw",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-07-01T15:31:30Z",
                    "pid": 1460,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "cf7b0a3fa193adee581364d4f46c666",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjM4OTAzMDQyMDk1MzM0MDA6MzM2NTM2OQ==",
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
                    "CAP_NET_RAW"
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
                    "CAP_NET_RAW"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026532263,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533285,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533567,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533568,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533568,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533286,
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
                    "inum": 4026533569,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3365369,
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
        "function_name": "security_mmap_file",
        "args": [
            {
                "file_arg": {
                    "mount": "",
                    "path": "/usr/lib/x86_64-linux-gnu/libpcap.so.1.10.3",
                    "flags": "",
                    "permission": "-rw-r--r--"
                },
                "label": ""
            },
            {
                "uint_arg": 3,
                "label": ""
            },
            {
                "uint_arg": 2066,
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
    "node_name": "experts-k8s-cs",
    "time": "2024-07-01T15:51:27.632603799Z",
    "aggregation_info": null
}
*/
