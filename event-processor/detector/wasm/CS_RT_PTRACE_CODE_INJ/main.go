//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_PTRACE_CODE_INJ"
	Name        = "Code injection into executable process through ptrace"
	Description = "The detector detects code injection into an executable process through the ptrace() system call, which may indicate attempts to masquerade as a legitimate process, access the process memory and system or network resources, and escalate privileges."
	Version     = 1
	Author      = "Runtime Radar Team"

	License = "Apache License 2.0"
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

var (
	// Taken from here (link from ESC): https://github.com/torvalds/linux/blob/master/include/uapi/linux/ptrace.h
	ptraceMaliciousRequests = []int32{
		4,     // PTRACE_POKETEXT (Copy word *data* to the address *addr*)
		5,     // PTRACE_POKEDATA (Copy word *data* to the address *addr*)
		6,     // PTRACE_POKEUSR (Copy the word *data* to offset *addr* in the traced process' USER area)
		13,    // PTRACE_SETREGS (Set all general purpose registers used by a processes)
		15,    // PTRACE_SETFPREGS (Set all floating point registers used by a processes)
		16,    // PTRACE_ATTACH (Attach to a process with SIGSTOP sending)
		17,    // PTRACE_DETACH (Detach from a traced process)
		19,    // PTRACE_SETFPXREGS (Set all extended floating point registers used by a processes)
		16899, // PTRACE_SETSIGINFO (Set signal information for a traced process)
		16901, // PTRACE_SETREGSET (Modify a traced process registers)
		16902, // PTRACE_SEIZE (Attach to a process without SIGSTOP sending)
		16903, // PTRACE_INTERRUPT (Try to stop traced process)
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
		args := kprobe.GetArgs()

		if strings.HasSuffix(functionName, "sys_ptrace") {
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >=1", len(args))
			}
			req := args[0].GetIntArg()

			for _, mal := range ptraceMaliciousRequests {
				if req == mal {
					resp.Severity = api.DetectResp_HIGH // <-- threat detected

					return resp, nil
				}
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
	  "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMxOTEwMjI4NzY4MTM4NTc6NzUwNzA1",
	  "pid": 750705,
	  "uid": 0,
	  "cwd": "/root/ptrace",
	  "binary": "/root/ptrace/exploit",
	  "arguments": "914",
	  "flags": "execve clone",
	  "start_time": "2024-03-21T09:37:29.670264484Z",
	  "auid": 4294967295,
	  "pod": {
		"namespace": "default",
		"name": "test-pod-ubuntu-ptrace",
		"labels": [],
		"container": {
		  "id": "cri-o://d401cf5283b757aa15aabd23e793c42840d658f1b5457250f32a9aa101371046",
		  "name": "test-pod-ubuntu-ptrace",
		  "image": {
			"id": "docker.io/library/ubuntu@sha256:a4fab1802f08df089c4b2e0a1c8f1a06f573bd1775687d07fef4076d3a2e4900",
			"name": "docker.io/library/ubuntu:focal"
		  },
		  "start_time": "2024-03-21T09:31:26Z",
		  "pid": 916,
		  "maybe_exec_probe": false
		},
		"pod_labels": {},
		"workload": "test-pod-ubuntu-ptrace",
		"workload_kind": "Pod"
	  },
	  "docker": "d401cf5283b757aa15aabd23e793c42",
	  "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMxOTA2Njk4Mjk2NTkxODA6NzQ2Mjg4",
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
		  "CAP_SYS_PTRACE"
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
		  "CAP_SYS_PTRACE"
		],
		"inheritable": []
	  },
	  "ns": {
		"uts": {
		  "inum": 4026533756,
		  "is_host": false
		},
		"ipc": {
		  "inum": 4026533757,
		  "is_host": false
		},
		"mnt": {
		  "inum": 4026535184,
		  "is_host": false
		},
		"pid": {
		  "inum": 4026535185,
		  "is_host": false
		},
		"pid_for_children": {
		  "inum": 4026535185,
		  "is_host": false
		},
		"net": {
		  "inum": 4026533758,
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
		  "inum": 4026535186,
		  "is_host": false
		},
		"user": {
		  "inum": 4026531837,
		  "is_host": true
		}
	  },
	  "tid": 750705,
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
	  "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMxOTA2Njk4Mjk2NTkxODA6NzQ2Mjg4",
	  "pid": 746288,
	  "uid": 0,
	  "cwd": "/",
	  "binary": "/usr/bin/bash",
	  "arguments": "",
	  "flags": "execve rootcwd",
	  "start_time": "2024-03-21T09:31:36.623109545Z",
	  "auid": 4294967295,
	  "pod": {
		"namespace": "default",
		"name": "test-pod-ubuntu-ptrace",
		"labels": [],
		"container": {
		  "id": "cri-o://d401cf5283b757aa15aabd23e793c42840d658f1b5457250f32a9aa101371046",
		  "name": "test-pod-ubuntu-ptrace",
		  "image": {
			"id": "docker.io/library/ubuntu@sha256:a4fab1802f08df089c4b2e0a1c8f1a06f573bd1775687d07fef4076d3a2e4900",
			"name": "docker.io/library/ubuntu:focal"
		  },
		  "start_time": "2024-03-21T09:31:26Z",
		  "pid": 7,
		  "maybe_exec_probe": false
		},
		"pod_labels": {},
		"workload": "test-pod-ubuntu-ptrace",
		"workload_kind": "Pod"
	  },
	  "docker": "d401cf5283b757aa15aabd23e793c42",
	  "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMxOTA2Njk4MjgzMDIyMzg6NzQ2Mjg4",
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
		  "CAP_SYS_PTRACE"
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
		  "CAP_SYS_PTRACE"
		],
		"inheritable": []
	  },
	  "ns": {
		"uts": {
		  "inum": 4026533756,
		  "is_host": false
		},
		"ipc": {
		  "inum": 4026533757,
		  "is_host": false
		},
		"mnt": {
		  "inum": 4026535184,
		  "is_host": false
		},
		"pid": {
		  "inum": 4026535185,
		  "is_host": false
		},
		"pid_for_children": {
		  "inum": 4026535185,
		  "is_host": false
		},
		"net": {
		  "inum": 4026533758,
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
		  "inum": 4026535186,
		  "is_host": false
		},
		"user": {
		  "inum": 4026531837,
		  "is_host": true
		}
	  },
	  "tid": 746288,
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
	"function_name": "__x64_sys_ptrace",
	"args": [
	  {
		"int_arg": 4,
		"label": ""
	  },
	  {
		"int_arg": 914,
		"label": ""
	  }
	],
	"return": {
	  "int_arg": 0,
	  "label": ""
	},
	"action": "KPROBE_ACTION_POST",
	"stack_trace": [],
	"policy_name": "ptrace-monitoring"
  },
  "node_name": "experts-k8s-cs",
  "time": "2024-03-21T09:37:29.671244063Z",
  "aggregation_info": null
}

*/
