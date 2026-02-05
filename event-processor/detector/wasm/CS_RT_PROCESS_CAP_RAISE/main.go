//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_PROCESS_CAP_RAISE"
	Name        = "Privilege escalation: capabilities"
	Description = "This detector detects processes running with excessive capabilities, which may indicate an attacker's attempt to escalate their privileges in the system."
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
		"PROCESS_KPROBE": {"commit_creds"},

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
	excessiveCapabilities = []tetragon.CapabilitiesType{
		tetragon.CapabilitiesType_CAP_DAC_READ_SEARCH,
		tetragon.CapabilitiesType_CAP_SYS_ADMIN,
		tetragon.CapabilitiesType_CAP_SYS_MODULE,
		tetragon.CapabilitiesType_CAP_SYS_RAWIO,
		tetragon.CapabilitiesType_CAP_NET_ADMIN,
		tetragon.CapabilitiesType_CAP_SYS_CHROOT,
		tetragon.CapabilitiesType_CAP_SYS_PTRACE,
		tetragon.CapabilitiesType_CAP_NET_RAW,
		tetragon.CapabilitiesType_CAP_SYS_BOOT,
		tetragon.CapabilitiesType_CAP_SYSLOG,
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

		proc := kprobe.GetProcess()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		if function != "commit_creds" {
			return resp, nil
		}

		if len(args) < 1 {
			return nil, fmt.Errorf("unexpected args len. got %d, expected >= 1", len(args))
		}

		committedCreds := args[0].GetProcessCredentialsArg()

		procEffectiveCaps := proc.GetCap().GetEffective()
		procPermittedCaps := proc.GetCap().GetPermitted()

		for _, cap := range committedCreds.GetCaps().GetEffective() {
			if slices.Contains(excessiveCapabilities, cap) && !slices.Contains(procEffectiveCaps, cap) {
				resp.Severity = api.DetectResp_HIGH
				return resp, nil
			}
		}

		for _, cap := range committedCreds.GetCaps().GetPermitted() {
			if slices.Contains(excessiveCapabilities, cap) && !slices.Contains(procPermittedCaps, cap) {
				resp.Severity = api.DetectResp_HIGH
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQyNTA2NTUxMjU5NjQ4NDU6MjkwODcxMg==",
            "pid": 2908712,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/unshare",
            "arguments": "-UrCpf bash",
            "flags": "execve",
            "start_time": "2024-04-02T15:58:01.919414806Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-priv",
                "labels": [],
                "container": {
                    "id": "cri-o://53e518ba04e25dd3c786818f8b7681a8e37254baa57930b06a2bd4cea74e93d0",
                    "name": "test-pod-debian-priv",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-03-21T13:35:21Z",
                    "pid": 446,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-priv",
                "workload_kind": "Pod"
            },
            "docker": "53e518ba04e25dd3c786818f8b7681a",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQyNTA2NTUxMjUwOTM0OTQ6MjkwODcxMQ==",
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
                    "CAP_SYS_CHROOT",
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
                    "CAP_SYS_CHROOT",
                    "CAP_SETFCAP"
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
            "tid": 2908712,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQyNTA2NTUxMjUwOTM0OTQ6MjkwODcxMQ==",
            "pid": 2908711,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/unshare",
            "arguments": "-UrCpf bash",
            "flags": "execve rootcwd clone",
            "start_time": "2024-04-02T15:58:01.918543722Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-priv",
                "labels": [],
                "container": {
                    "id": "cri-o://53e518ba04e25dd3c786818f8b7681a8e37254baa57930b06a2bd4cea74e93d0",
                    "name": "test-pod-debian-priv",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-03-21T13:35:21Z",
                    "pid": 446,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-priv",
                "workload_kind": "Pod"
            },
            "docker": "53e518ba04e25dd3c786818f8b7681a",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQyNTA2MzU0NjYzMzUyMjQ6MjkwODQ5MQ==",
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
                    "CAP_SYS_CHROOT",
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
                    "CAP_SYS_CHROOT",
                    "CAP_SETFCAP"
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
            "tid": 2908711,
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
        "function_name": "commit_creds",
        "args": [
            {
                "process_credentials_arg": {
                    "uid": 0,
                    "gid": 0,
                    "euid": 0,
                    "egid": 0,
                    "suid": 0,
                    "sgid": 0,
                    "fsuid": 0,
                    "fsgid": 0,
                    "securebits": [],
                    "caps": {
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
                        "inheritable": []
                    },
                    "user_ns": {
                        "level": 1,
                        "uid": 0,
                        "gid": 0,
                        "ns": {
                            "inum": 4026533174,
                            "is_host": false
                        }
                    }
                },
                "label": ""
            }
        ],
        "return": null,
        "action": "KPROBE_ACTION_POST",
        "stack_trace": [],
        "policy_name": "privilege-raise-monitoring"
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-04-02T15:58:01.919951802Z",
    "aggregation_info": null
}
*/
