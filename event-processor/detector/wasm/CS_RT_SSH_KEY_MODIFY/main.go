//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_SSH_KEY_MODIFY"
	Name        = "Suspicious change of SSH keys"
	Description = "The detector detects suspicious changes in files that contain SSH keys, which may indicate that an existing account was compromised or that an attacker attempted to achieve persistence on the system."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
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

var (
	sshKeysPath    = regexp.MustCompile(`^(/root/|/home/.*/)\.ssh/`)
	filesToExclude = regexp.MustCompile(`\.ssh2?/(config|known_hosts)$`)
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

		if sshKeysPath.MatchString(path) && !filesToExclude.MatchString(path) {
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

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
      "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjE4NDE3OTE1MDY5OTEzMDQ6NDA4NzQxOA==",
      "pid": 4087418,
      "uid": 0,
      "cwd": "/",
      "binary": "/usr/bin/vim",
      "arguments": "/root/.ssh/authorized_keys",
      "flags": "execve rootcwd clone",
      "start_time": "2024-03-05T18:50:18.300441972Z",
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
          "pid": 1968,
          "maybe_exec_probe": false
        },
        "pod_labels": {},
        "workload": "test-pod-debian",
        "workload_kind": "Pod"
      },
      "docker": "332cff0fb99f03f8b4fb9633f245ba2",
      "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjE4NDE2ODY2ODU1Njg4NzI6NDA4NjQxMA==",
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
      "tid": 4087418,
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
      "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjE4NDE2ODY2ODU1Njg4NzI6NDA4NjQxMA==",
      "pid": 4086410,
      "uid": 0,
      "cwd": "/",
      "binary": "/usr/bin/bash",
      "arguments": "",
      "flags": "execve rootcwd",
      "start_time": "2024-03-05T18:48:33.479019614Z",
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
          "pid": 1959,
          "maybe_exec_probe": false
        },
        "pod_labels": {},
        "workload": "test-pod-debian",
        "workload_kind": "Pod"
      },
      "docker": "332cff0fb99f03f8b4fb9633f245ba2",
      "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjE4NDE2ODY2ODE1NjIwOTQ6NDA4NjQxMA==",
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
      "tid": 4086410,
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
          "path": "/root/.ssh/authorized_keys",
          "flags": ""
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
    "stack_trace": [],
    "policy_name": "file-monitoring"
  },
  "node_name": "experts-k8s-cs",
  "time": "2024-03-05T18:50:25.450598445Z",
  "aggregation_info": null
}


*/
