//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_HARDLINK_CREATE"
	Name        = "Creation of hard links to sensitive system files"
	Description = "The detector detects if hard links to sensitive system files were created, which may indicate an attacker's attempt to bypass the existing file monitoring rules."
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
		"PROCESS_KPROBE": {"security_path_link"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
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
		function := kprobe.GetFunctionName()

		switch function {
		case "security_path_link":
			resp.Severity = api.DetectResp_HIGH // <-- threat detected

			return resp, nil
		default:
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
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZTo3NzM0MzkzMDM4MDk5MzMyOjUyMDAzOA==",
            "pid": 520038,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/ln",
            "arguments": "/etc/shadow /tmp/token",
            "flags": "execve rootcwd clone",
            "start_time": "2025-12-10T07:55:01.513635226Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://30af0bcafcdec812b8ed60f86b3ccdbd54da68e48152bedd79e4aa1751acdcb4",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-09-11T19:30:01Z",
                    "pid": 144,
                    "maybe_exec_probe": false,
                    "security_context": {
                        "privileged": false
                    }
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod",
                "pod_annotations": {}
            },
            "docker": "30af0bcafcdec812b8ed60f86b3ccdb",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZTo3NzMwOTU4MTU3OTcwMTY4OjMzMDIyNw==",
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
                    "inum": 4026534078,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534079,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534081,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534082,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534082,
                    "is_host": false
                },
                "net": {
                    "inum": 4026534018,
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
                    "inum": 4026534083,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 520038,
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
            "user": null,
            "in_init_tree": false
        },
        "parent": {
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZTo3NzMwOTU4MTU3OTcwMTY4OjMzMDIyNw==",
            "pid": 330227,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2025-12-10T06:57:46.633506777Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://30af0bcafcdec812b8ed60f86b3ccdbd54da68e48152bedd79e4aa1751acdcb4",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-09-11T19:30:01Z",
                    "pid": 104,
                    "maybe_exec_probe": false,
                    "security_context": {
                        "privileged": false
                    }
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod",
                "pod_annotations": {}
            },
            "docker": "30af0bcafcdec812b8ed60f86b3ccdb",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZTo3NzMwOTU4MTUzMTU4MzQ1OjMzMDIyNw==",
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
                    "inum": 4026534078,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534079,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534081,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534082,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534082,
                    "is_host": false
                },
                "net": {
                    "inum": 4026534018,
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
                    "inum": 4026534083,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 330227,
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
            "user": null,
            "in_init_tree": false
        },
        "function_name": "security_path_link",
        "args": [
            {
                "path_arg": {
                    "mount": "",
                    "path": "/etc/shadow",
                    "flags": "",
                    "permission": "-rw-r-----"
                },
                "label": "existing file"
            },
            {
                "path_arg": {
                    "mount": "",
                    "path": "/tmp/token",
                    "flags": "",
                    "permission": "---------"
                },
                "label": "new hard link"
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
        "user_stack_trace": [],
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2025-12-10T07:55:01.526266305Z",
    "aggregation_info": null,
    "cluster_name": "",
    "node_labels": {
        "beta.kubernetes.io/arch": "amd64",
        "beta.kubernetes.io/os": "linux",
        "kubernetes.io/arch": "amd64",
        "kubernetes.io/hostname": "experts-k8s-cs",
        "kubernetes.io/os": "linux",
        "node-role.kubernetes.io/control-plane": "",
        "node.kubernetes.io/exclude-from-external-load-balancers": ""
    }
}
*/
