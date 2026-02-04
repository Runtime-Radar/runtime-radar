//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_MOUNT_DEV"
	Name        = "Device mounting from /dev directory"
	Description = "The detector detects the mount() system calls to mount devices from the /dev directory, which may indicate an attacker's attempt to gain access to the file system of the container host. The detector assigns low severity to unsuccessful attempts and high severity to successful ones."
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
		"PROCESS_KPROBE": {"sys_mount"},

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
	devPattern = glob.MustCompile("/dev/*")
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

		if strings.HasSuffix(functionName, "sys_mount") {
			ret := kprobe.GetReturn()
			r := ret.GetIntArg()

			args := kprobe.GetArgs()
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			a := args[0].GetStringArg()
			if !devPattern.Match(a) {
				return resp, nil
			}

			if r == 0 {
				// According to mount() documentation 0 return value means success.
				resp.Severity = api.DetectResp_HIGH // <-- threat detected
			} else {
				// Else, we assume that it was just an attempt, so mark it with low severity, but do not ignore.
				// This can be removed later if there will be a lot of false positives.
				resp.Severity = api.DetectResp_LOW // <-- threat detected
			}

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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcwNjgyNzc1NzI1NzU1NjQ6NDAwODMyNw==",
            "pid": 4008327,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/mount",
            "arguments": "/dev/vda1 /mnt/hola",
            "flags": "execve rootcwd clone",
            "start_time": "2024-08-07T10:36:53.218786013Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-privileged",
                "container": {
                    "id": "cri-o://497809a2cbb0695fbd5383072cfb59e5c3b6905ab1ac336147cccbcf4a2b224e",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "31d5e503c34f4496a263fb3557575cf53e6a40add4c459370120c7454985f7b7",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-18T19:36:18Z",
                    "pid": 3497,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-privileged",
                "workload_kind": "Pod"
            },
            "docker": "497809a2cbb0695fbd5383072cfb59e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcwNjgxOTk2NzczNDg0OTE6NDAwNzU3Ng==",
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
                    "CAP_AUDIT_READ",
                    "CAP_PERFMON",
                    "CAP_BPF",
                    "CAP_CHECKPOINT_RESTORE"
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
                    "CAP_AUDIT_READ",
                    "CAP_PERFMON",
                    "CAP_BPF",
                    "CAP_CHECKPOINT_RESTORE"
                ],
                "inheritable": [
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
                    "CAP_AUDIT_READ",
                    "CAP_PERFMON",
                    "CAP_BPF",
                    "CAP_CHECKPOINT_RESTORE"
                ]
            },
            "ns": {
                "uts": {
                    "inum": 4026532795,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532796,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533178,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533284,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533284,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532797,
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
            "tid": 4008327,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcwNjgxOTk2NzczNDg0OTE6NDAwNzU3Ng==",
            "pid": 4007576,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-08-07T10:35:35.323559007Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-privileged",
                "container": {
                    "id": "cri-o://497809a2cbb0695fbd5383072cfb59e5c3b6905ab1ac336147cccbcf4a2b224e",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "31d5e503c34f4496a263fb3557575cf53e6a40add4c459370120c7454985f7b7",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-18T19:36:18Z",
                    "pid": 3490,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-privileged",
                "workload_kind": "Pod"
            },
            "docker": "497809a2cbb0695fbd5383072cfb59e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcwNjgxOTk2NzU0NDcwMTg6NDAwNzU3Ng==",
            "refcnt": 0,
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
                    "CAP_AUDIT_READ",
                    "CAP_PERFMON",
                    "CAP_BPF",
                    "CAP_CHECKPOINT_RESTORE"
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
                    "CAP_AUDIT_READ",
                    "CAP_PERFMON",
                    "CAP_BPF",
                    "CAP_CHECKPOINT_RESTORE"
                ],
                "inheritable": [
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
                    "CAP_AUDIT_READ",
                    "CAP_PERFMON",
                    "CAP_BPF",
                    "CAP_CHECKPOINT_RESTORE"
                ]
            },
            "ns": {
                "uts": {
                    "inum": 4026532795,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532796,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533178,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533284,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533284,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532797,
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
            "tid": 4007576,
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
        "function_name": "__x64_sys_mount",
        "args": [
            {
                "string_arg": "/dev/vda1",
                "label": ""
            },
            {
                "string_arg": "/mnt/hola",
                "label": ""
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "mount",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-08-07T10:36:53.223566947Z",
    "aggregation_info": null
}

*/
