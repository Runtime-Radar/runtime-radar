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
	ID          = "CS_RT_SSH_TUNNEL_USE"
	Name        = "Use of SSH tunnel"
	Description = "The detector detects network activity related to network tunneling of a local SSH service."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	sshdPort = 22
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"tcp_connect", "inet_csk_listen_start"},

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
	sshdBin = glob.MustCompile("*/sshd")
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

		if !sshdBin.Match(binary) {
			return resp, nil
		}

		switch function {
		// If sshd process is trying to connect to some host, than someone is using a ssh-tunnel created via local forwarding
		case "tcp_connect":
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil

		// If sshd process is trying to open not a tcp/22 port, than someone is opening a ssh-tunnel via remote forwarding
		case "inet_csk_listen_start":
			args := kprobe.GetArgs()

			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			socket := args[0].GetSockArg()
			sport := socket.GetSport()

			if sport != sshdPort {
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
			}
		}

		// Nothing here
	case *tetragon.GetEventsResponse_ProcessTracepoint:
		// Nothing here
	}

	return resp, nil
}

/* Example event (JSON):

LOCAL FORWARDING

{
    "process_kprobe": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDQyMDU5MjE0Njg1ODM6MTU2ODY2Ng==",
            "pid": 1568666,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/sbin/sshd",
            "arguments": "-D -e -f /etc/ssh/sshd_config -R",
            "flags": "execve rootcwd clone",
            "start_time": "2024-07-04T14:22:21.567678504Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "openssh-server",
                "container": {
                    "id": "cri-o://d665767d27511008a0c12fae494594ef5daf1d01af7ee3a213c5ebc14348fce8",
                    "name": "openssh-server",
                    "image": {
                        "id": "docker.io/panubo/sshd@sha256:6c2864b883d9e78565a5b0ed127a0a0bd8289fa046b76f55d12d1bd544a4e0b8",
                        "name": "docker.io/panubo/sshd:latest"
                    },
                    "start_time": "2024-07-04T14:12:58Z",
                    "pid": 68,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app": "sshd"
                },
                "workload": "openssh-server",
                "workload_kind": "Pod"
            },
            "docker": "d665767d27511008a0c12fae494594e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDM2NDg0Nzc0MDgxNTI6MTU2MDQ1MQ==",
            "refcnt": 2,
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
                    "inum": 4026533570,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533571,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533964,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533572,
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
            "tid": 1568666,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDM2NDg0Nzc0MDgxNTI6MTU2MDQ1MQ==",
            "pid": 1560451,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/sbin/sshd",
            "arguments": "-D -e -f /etc/ssh/sshd_config",
            "flags": "execve rootcwd clone",
            "start_time": "2024-07-04T14:13:04.123618324Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "openssh-server",
                "container": {
                    "id": "cri-o://d665767d27511008a0c12fae494594ef5daf1d01af7ee3a213c5ebc14348fce8",
                    "name": "openssh-server",
                    "image": {
                        "id": "docker.io/panubo/sshd@sha256:6c2864b883d9e78565a5b0ed127a0a0bd8289fa046b76f55d12d1bd544a4e0b8",
                        "name": "docker.io/panubo/sshd:latest"
                    },
                    "start_time": "2024-07-04T14:12:58Z",
                    "pid": 42,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app": "sshd"
                },
                "workload": "openssh-server",
                "workload_kind": "Pod"
            },
            "docker": "d665767d27511008a0c12fae494594e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDM2NDI2NjM0OTk3MDk6MTU2MDMxMA==",
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
                    "inum": 4026533570,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533571,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533964,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533572,
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
            "tid": 1560451,
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
        "function_name": "tcp_connect",
        "args": [
            {
                "sock_arg": {
                    "family": "AF_INET",
                    "type": "SOCK_STREAM",
                    "protocol": "IPPROTO_TCP",
                    "mark": 0,
                    "priority": 0,
                    "saddr": "10.244.0.245",
                    "daddr": "10.244.0.232",
                    "sport": 38854,
                    "dport": 8080,
                    "cookie": "18446618302131282688",
                    "state": "TCP_SYN_SENT"
                },
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
    "node_name": "experts-k8s-cs",
    "time": "2024-07-04T14:24:52.461134895Z",
    "aggregation_info": null
}

REMOTE FORWARDING
{
    "process_kprobe": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDQ2MzcxOTI0NzY0MDM6MTU3MjkzOA==",
            "pid": 1572938,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/sbin/sshd",
            "arguments": "-D -e -f /etc/ssh/sshd_config -R",
            "flags": "execve rootcwd clone",
            "start_time": "2024-07-04T14:29:32.838686118Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "openssh-server",
                "container": {
                    "id": "cri-o://d665767d27511008a0c12fae494594ef5daf1d01af7ee3a213c5ebc14348fce8",
                    "name": "openssh-server",
                    "image": {
                        "id": "docker.io/panubo/sshd@sha256:6c2864b883d9e78565a5b0ed127a0a0bd8289fa046b76f55d12d1bd544a4e0b8",
                        "name": "docker.io/panubo/sshd:latest"
                    },
                    "start_time": "2024-07-04T14:12:58Z",
                    "pid": 80,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app": "sshd"
                },
                "workload": "openssh-server",
                "workload_kind": "Pod"
            },
            "docker": "d665767d27511008a0c12fae494594e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDM2NDg0Nzc0MDgxNTI6MTU2MDQ1MQ==",
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
                    "inum": 4026533570,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533571,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533964,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533572,
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
            "tid": 1572938,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDM2NDg0Nzc0MDgxNTI6MTU2MDQ1MQ==",
            "pid": 1560451,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/sbin/sshd",
            "arguments": "-D -e -f /etc/ssh/sshd_config",
            "flags": "execve rootcwd clone",
            "start_time": "2024-07-04T14:13:04.123618324Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "openssh-server",
                "container": {
                    "id": "cri-o://d665767d27511008a0c12fae494594ef5daf1d01af7ee3a213c5ebc14348fce8",
                    "name": "openssh-server",
                    "image": {
                        "id": "docker.io/panubo/sshd@sha256:6c2864b883d9e78565a5b0ed127a0a0bd8289fa046b76f55d12d1bd544a4e0b8",
                        "name": "docker.io/panubo/sshd:latest"
                    },
                    "start_time": "2024-07-04T14:12:58Z",
                    "pid": 42,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app": "sshd"
                },
                "workload": "openssh-server",
                "workload_kind": "Pod"
            },
            "docker": "d665767d27511008a0c12fae494594e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQxNDM2NDI2NjM0OTk3MDk6MTU2MDMxMA==",
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
                    "inum": 4026533570,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533571,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533964,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533965,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533572,
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
            "tid": 1560451,
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
        "function_name": "inet_csk_listen_start",
        "args": [
            {
                "sock_arg": {
                    "family": "AF_INET",
                    "type": "SOCK_STREAM",
                    "protocol": "IPPROTO_TCP",
                    "mark": 0,
                    "priority": 0,
                    "saddr": "0.0.0.0",
                    "daddr": "0.0.0.0",
                    "sport": 58000,
                    "dport": 0,
                    "cookie": "18446618298011991360",
                    "state": "TCP_CLOSE"
                },
                "label": ""
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "network-socket",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-07-04T14:29:35.351080886Z",
    "aggregation_info": null
}
*/
