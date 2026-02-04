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
	ID          = "CS_RT_SSH_TUNNEL_CREATE"
	Name        = "SSH tunnel creation"
	Description = "The detector detects if a local or remote network tunnel was created using the SSH service."
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
	sshBin = glob.MustCompile("*/ssh")

	localForwardRegex   = regexp.MustCompile(`\s*-[^-\s]*L\s+`)
	remoteForwardRegex  = regexp.MustCompile(`\s*-[^-\s]*R\s+`)
	dynamicForwardRegex = regexp.MustCompile(`\s*-[^-\s]*D\s+`)
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
		process := ev.ProcessExec
		binary := process.GetProcess().GetBinary()
		args := process.GetProcess().GetArguments()

		if sshBin.Match(binary) &&
			(localForwardRegex.MatchString(args) || remoteForwardRegex.MatchString(args) || dynamicForwardRegex.MatchString(args)) {
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

			return resp, nil
		}

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

LOCAL FORWARDING

{
    "process_exec": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2NDQxODg3MTU0Mzg6MTIzMTI0OA==",
            "pid": 1231248,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/ssh",
            "arguments": "-L 127.0.0.1:8080:10.0.2.5:6443 ubuntu@10.0.2.5",
            "flags": "execve rootcwd clone",
            "start_time": "2024-04-05T14:47:50.982165630Z",
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
                    "pid": 71358,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MDI5NjI4MDU2NjY6MTIzMDg1OQ==",
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
            "tid": 1231248,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MDI5NjI4MDU2NjY6MTIzMDg1OQ==",
            "pid": 1230859,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-04-05T14:47:09.756256165Z",
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
                    "pid": 71351,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MDI5NjEzMDM0NzU6MTIzMDg1OQ==",
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
            "tid": 1230859,
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
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-04-05T14:47:50.982165341Z",
    "aggregation_info": null
}

REMOTE FORWARDING
{
    "process_exec": {
        "process": {
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MTE0ODQ4Mjg5MzE6MTIzMDk2MA==",
            "pid": 1230960,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/ssh",
            "arguments": "-R 127.0.0.1:8080:10.0.2.5:6443 ubuntu@10.0.2.5",
            "flags": "execve rootcwd clone",
            "start_time": "2024-04-05T14:47:18.278278973Z",
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
                    "pid": 71357,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MDI5NjI4MDU2NjY6MTIzMDg1OQ==",
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
            "tid": 1230960,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MDI5NjI4MDU2NjY6MTIzMDg1OQ==",
            "pid": 1230859,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-04-05T14:47:09.756256165Z",
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
                    "pid": 71351,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjQ1MDU2MDI5NjEzMDM0NzU6MTIzMDg1OQ==",
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
            "tid": 1230859,
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
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-04-05T14:47:18.278279002Z",
    "aggregation_info": null
}

DYNAMIC FORWARDING
{
    "process_exec": {
        "process": {
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZToxOTg5NjIwODczNjA5MDQ4NzoxODk3NzYx",
            "pid": 1897761,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/ssh",
            "arguments": "-D 8888 -N admin@10.125.11.114",
            "flags": "execve rootcwd clone",
            "start_time": "2025-04-22T22:12:35.848223325Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://657178534beb38ff65dc8f120862c823de19056e47a31a16caae849b2ce90efa",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-03-04T11:57:50Z",
                    "pid": 4025,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "657178534beb38ff65dc8f120862c82",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZToxOTg5NjAxMDY5OTg1MTE4NzoxODkxMTMz",
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
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
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
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
                    "CAP_SETFCAP"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026533358,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533363,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533365,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533366,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533366,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533192,
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
                    "inum": 4026533367,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1897761,
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
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZToxOTg5NjAxMDY5OTg1MTE4NzoxODkxMTMz",
            "pid": 1891133,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2025-04-22T22:09:17.811981949Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://657178534beb38ff65dc8f120862c823de19056e47a31a16caae849b2ce90efa",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-03-04T11:57:50Z",
                    "pid": 3855,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "657178534beb38ff65dc8f120862c82",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZToxOTg5NjAxMDY1NzcxMDc1NjoxODkxMTMz",
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
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
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
                    "CAP_NET_RAW",
                    "CAP_SYS_CHROOT",
                    "CAP_MKNOD",
                    "CAP_AUDIT_WRITE",
                    "CAP_SETFCAP"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026533358,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533363,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533365,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533366,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533366,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533192,
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
                    "inum": 4026533367,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1891133,
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
        "ancestors": []
    },
    "node_name": "cs-master-node",
    "time": "2025-04-22T22:12:35.848220758Z",
    "aggregation_info": null
}
*/
