//go:build tinygo.wasm

package main

import (
	"context"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_CRYPTOMINER"
	Name        = "Cryptominer execution"
	Description = "The detector detects if any known cryptominers were started or stopped. The latter may indicate an attacker's attempt to get rid of competing loads."
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
		"PROCESS_EXEC": {},

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
	minerTools = []*regexp.Regexp{
		regexp.MustCompile(`/[^/]*stratum[^/]*$`),
		regexp.MustCompile(`/[^/]*minerd[^/]*$`),
		regexp.MustCompile(`/[^/]*xmr[^/]*$`),
		regexp.MustCompile(`/[^/]*cryptonight[^/]*$`),
		regexp.MustCompile(`/[^/]*haiduc[^/]*$`),
		regexp.MustCompile(`/sparky\.sh$`),
		regexp.MustCompile(`/2238Xae$`),
		regexp.MustCompile(`/[^/]*yastrm[^/]*$`),
		regexp.MustCompile(`/[^/]*biden[1l][^/]*$`),
		regexp.MustCompile(`/xrx$`),
		regexp.MustCompile(`/zzh$`),
		regexp.MustCompile(`/arx645$`),
		regexp.MustCompile(`/[^/]+kthread[^/]*$`),
		regexp.MustCompile(`/[^/]*kdevtmpfsi[^/]*$`),
		regexp.MustCompile(`/pppsssdm$`),
		regexp.MustCompile(`/[^/]*kernelx[^/]*$`),
		regexp.MustCompile(`/[^/]*pg_mem[^/]*$`),
	}

	minerArgs = []*regexp.Regexp{
		regexp.MustCompile(`\-\-cpu-priority`),
		regexp.MustCompile(`\-\-donate-level`),
		regexp.MustCompile(`\-\-nicehash`),
		regexp.MustCompile(`\-\-algo`),
		regexp.MustCompile(`stratum2?\+`),
	}

	killingTools = regexp.MustCompile(`/p?kill(?:all)?$`)

	killArgs = []glob.Glob{
		glob.MustCompile("*stratum*"),
		glob.MustCompile("*minerd*"),
		glob.MustCompile("*xmr*"),
		glob.MustCompile("*cryptonight*"),
		glob.MustCompile("*haiduc*"),
		glob.MustCompile("*sparky.sh*"),
		glob.MustCompile("*2238Xae*"),
		glob.MustCompile("*yastrm*"),
		glob.MustCompile("*biden*"),
		glob.MustCompile("*xrx*"),
		glob.MustCompile("*zzh*"),
		glob.MustCompile("*arx645*"),
		glob.MustCompile("*kthread*"),
		glob.MustCompile("*kdevtmpfsi*"),
		glob.MustCompile("*pppsssdm*"),
		glob.MustCompile("*kernelx*"),
		glob.MustCompile("*pg_mem*"),
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
		exec := ev.ProcessExec
		binary := exec.GetProcess().GetBinary()
		args := exec.GetProcess().GetArguments()

		for _, mt := range minerTools {
			// Trigger on explicit miner utility name.
			if mt.MatchString(binary) {
				resp.Severity = api.DetectResp_HIGH // <-- threat detected
				return resp, nil
			}
		}

		// Trigger on explicit miner utility args.
		for _, ma := range minerArgs {
			if ma.MatchString(args) {
				resp.Severity = api.DetectResp_HIGH // <-- threat detected
				return resp, nil
			}
		}

		// Trigger on attempts of killing a miner process.
		if killingTools.MatchString(binary) {
			for _, ka := range killArgs {
				if ka.Match(args) {
					resp.Severity = api.DetectResp_HIGH // <-- threat detected
					return resp, nil
				}
			}
		}

		return resp, nil

	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessTracepoint:
		// Nothing here
	}

	return resp, nil
}

/* Example event (JSON):
{
    "process_exit": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjc3NjE3NjE1NzgxMjI1OTk6MjIxNzU1MQ==",
            "pid": 2217551,
            "uid": 0,
            "cwd": "/opt/xmrig/xmrig-6.22.0",
            "binary": "/usr/bin/pkill",
            "arguments": "-f xmrig",
            "flags": "execve clone",
            "start_time": "2024-08-15T11:14:57.224332234Z",
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
                    "pid": 4043,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-privileged",
                "workload_kind": "Pod"
            },
            "docker": "497809a2cbb0695fbd5383072cfb59e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjc3NjE3MzY1MzMzODc2NTQ6MjIxNzI4Ng==",
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
            "tid": 2217551,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjc3NjE3MzY1MzMzODc2NTQ6MjIxNzI4Ng==",
            "pid": 2217286,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-08-15T11:14:32.179597219Z",
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
                    "pid": 4012,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-privileged",
                "workload_kind": "Pod"
            },
            "docker": "497809a2cbb0695fbd5383072cfb59e",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjc3NjE3MzY1MzE4NzUxMTU6MjIxNzI4Ng==",
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
            "tid": 2217286,
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
        "signal": "",
        "status": 0,
        "time": "2024-08-15T11:14:57.231061804Z"
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-08-15T11:14:57.231060954Z",
    "aggregation_info": null
}


*/
