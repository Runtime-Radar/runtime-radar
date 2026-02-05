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
	ID          = "CS_RT_CVE_2022_0492"
	Name        = "Container isolation vulnerability exploitation (CVE-2022-0492)"
	Description = "The detector detects attempts to modify the notify_on_release and release_agent files, which may indicate an attempt to exploit vulnerability CVE-2022-0492. An attacker can exploit this vulnerability to escalate privileges and break out from the isolated container environment."
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

var (
	fileRegex = regexp.MustCompile(`(?:release_agent|notify_on_release)$`)
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

		// The first 15 digits of the container ID
		containerIDPart := kprobe.GetProcess().GetDocker()
		functionName := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		// Empty containerIDPart means that we are not in container environment (vanilla Docker or k8s)
		if containerIDPart == "" {
			return resp, nil
		}

		path := ""

		switch functionName {
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != 2 { // need MAY_WRITE
				return resp, nil
			}

			path = args[0].GetFileArg().GetPath()
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&2 != 2 { // need PROT_WRITE
				return resp, nil
			}

			path = args[0].GetFileArg().GetPath()
		case "security_path_truncate":
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			path = args[0].GetPathArg().GetPath()
		default:
			return resp, nil
		}

		if fileRegex.MatchString(path) {
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

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
      "exec_id": "OjkxMjM1NTUzMTgyNTQxOjY2NjA0",
      "pid": 66604,
      "uid": 0,
      "cwd": "/",
      "binary": "/usr/bin/bash",
      "flags": "execve rootcwd clone",
      "start_time": "2024-03-22T08:02:04.345014271Z",
      "auid": 4294967295,
      "docker": "b2ce6ed48a40531cbbcdbf025b47467",
      "parent_exec_id": "OjkxMjM1MTcxNDIwNjE0OjY2NTY3",
      "refcnt": 1,
      "cap": {
        "permitted": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ],
        "effective": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ]
      },
      "tid": 66604,
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
      "exec_id": "OjkxMjM1MTcxNDIwNjE0OjY2NTY3",
      "pid": 66567,
      "uid": 0,
      "cwd": "/run/containerd/io.containerd.runtime.v2.task/moby/b2ce6ed48a40531cbbcdbf025b4746725ce07ddcb4d5b96715d80a37967218d5",
      "binary": "/usr/bin/containerd-shim-runc-v2",
      "arguments": "-namespace moby -id b2ce6ed48a40531cbbcdbf025b4746725ce07ddcb4d5b96715d80a37967218d5 -address /run/containerd/containerd.sock",
      "flags": "execve clone",
      "start_time": "2024-03-22T08:02:03.963251745Z",
      "auid": 4294967295,
      "parent_exec_id": "OjkxMjM1MTUzNjE5ODY3OjY2NTU5",
      "cap": {
        "permitted": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ],
        "effective": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ]
      },
      "tid": 66567,
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
    "function_name": "security_file_permission",
    "args": [
      {
        "file_arg": {
          "path": "/tmp/cgrp/x/notify_on_release"
        }
      },
      {
        "int_arg": 2
      }
    ],
    "return": {
      "int_arg": 0
    },
    "action": "KPROBE_ACTION_POST",
    "policy_name": "file-monitoring"
  },
  "time": "2024-03-22T08:06:38.859706015Z"
}

{
  "process_kprobe": {
    "process": {
      "exec_id": "OjkxMjM1NTUzMTgyNTQxOjY2NjA0",
      "pid": 66604,
      "uid": 0,
      "cwd": "/",
      "binary": "/usr/bin/bash",
      "flags": "execve rootcwd clone",
      "start_time": "2024-03-22T08:02:04.345014271Z",
      "auid": 4294967295,
      "docker": "b2ce6ed48a40531cbbcdbf025b47467",
      "parent_exec_id": "OjkxMjM1MTcxNDIwNjE0OjY2NTY3",
      "refcnt": 1,
      "cap": {
        "permitted": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ],
        "effective": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ]
      },
      "tid": 66604,
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
      "exec_id": "OjkxMjM1MTcxNDIwNjE0OjY2NTY3",
      "pid": 66567,
      "uid": 0,
      "cwd": "/run/containerd/io.containerd.runtime.v2.task/moby/b2ce6ed48a40531cbbcdbf025b4746725ce07ddcb4d5b96715d80a37967218d5",
      "binary": "/usr/bin/containerd-shim-runc-v2",
      "arguments": "-namespace moby -id b2ce6ed48a40531cbbcdbf025b4746725ce07ddcb4d5b96715d80a37967218d5 -address /run/containerd/containerd.sock",
      "flags": "execve clone",
      "start_time": "2024-03-22T08:02:03.963251745Z",
      "auid": 4294967295,
      "parent_exec_id": "OjkxMjM1MTUzNjE5ODY3OjY2NTU5",
      "cap": {
        "permitted": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ],
        "effective": [
          "CAP_CHOWN",
          "DAC_OVERRIDE",
          "CAP_DAC_READ_SEARCH",
          "CAP_FOWNER",
          "CAP_FSETID",
          "CAP_KILL",
          "CAP_SETGID",
          "CAP_SETUID",
          "CAP_SETPCAP",
          "CAP_LINUX_IMMUTABLE",
          "CAP_NET_BIND_SERVICE",
          "CAP_NET_BROADCAST",
          "CAP_NET_ADMIN",
          "CAP_NET_RAW",
          "CAP_IPC_LOCK",
          "CAP_IPC_OWNER",
          "CAP_SYS_MODULE",
          "CAP_SYS_RAWIO",
          "CAP_SYS_CHROOT",
          "CAP_SYS_PTRACE",
          "CAP_SYS_PACCT",
          "CAP_SYS_ADMIN",
          "CAP_SYS_BOOT",
          "CAP_SYS_NICE",
          "CAP_SYS_RESOURCE",
          "CAP_SYS_TIME",
          "CAP_SYS_TTY_CONFIG",
          "CAP_MKNOD",
          "CAP_LEASE",
          "CAP_AUDIT_WRITE",
          "CAP_AUDIT_CONTROL",
          "CAP_SETFCAP",
          "CAP_MAC_OVERRIDE",
          "CAP_MAC_ADMIN",
          "CAP_SYSLOG",
          "CAP_WAKE_ALARM",
          "CAP_BLOCK_SUSPEND",
          "CAP_AUDIT_READ"
        ]
      },
      "tid": 66567,
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
    "function_name": "security_file_permission",
    "args": [
      {
        "file_arg": {
          "path": "/tmp/cgrp/release_agent"
        }
      },
      {
        "int_arg": 2
      }
    ],
    "return": {
      "int_arg": 0
    },
    "action": "KPROBE_ACTION_POST",
    "policy_name": "file-monitoring"
  },
  "time": "2024-03-22T08:11:14.343791166Z"
}

*/
