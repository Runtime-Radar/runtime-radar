//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_FILELESS_EXEC"
	Name        = "Fileless process execution"
	Description = "The detector detects if a process was executed without an executable file in the file system, which may indicate an attacker's attempt to conceal their activity."
	Version     = 3
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
		inode := exec.GetProcess().GetBinaryProperties().GetFile().GetInode()

		if inode != nil {
			binLinks := exec.GetProcess().GetBinaryProperties().GetFile().GetInode().GetLinks().GetValue()
			if binLinks == 0 {
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

				return resp, nil
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
    "process_exec": {
        "process": {
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZToxMDI5MDcyMTQxNTA0NTUwOjEyODczODg=",
            "pid": 1287388,
            "uid": 0,
            "cwd": "/root/fileless",
            "binary": "/proc/self/fd/4",
            "arguments": "",
            "flags": "execve",
            "start_time": "2025-07-07T12:14:01.263533986Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://7bcd23290beff948c1f9516218d1e8ae7c21a754ca6c00ac7e261a14e12e989d",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-06-25T14:24:35Z",
                    "pid": 5592,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "7bcd23290beff948c1f9516218d1e8a",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZToxMDI5MDcyMTM3NzQ5ODI1OjEyODczODg=",
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
                    "inum": 4026533911,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533912,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533914,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533915,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533915,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533829,
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
                    "inum": 4026533916,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1287388,
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
            "binary_properties": {
                "setuid": null,
                "setgid": null,
                "privileges_changed": [],
                "file": {
                    "inode": {
                        "number": "202749",
                        "links": 0
                    },
                    "path": ""
                }
            },
            "user": null
        },
        "parent": {
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZToxMDI5MDcyMTM3NzQ5ODI1OjEyODczODg=",
            "pid": 1287388,
            "uid": 0,
            "cwd": "/root/fileless",
            "binary": "/root/fileless/fileless",
            "arguments": "memfd",
            "flags": "execve clone",
            "start_time": "2025-07-07T12:14:01.259780503Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://7bcd23290beff948c1f9516218d1e8ae7c21a754ca6c00ac7e261a14e12e989d",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-06-25T14:24:35Z",
                    "pid": 5592,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "7bcd23290beff948c1f9516218d1e8a",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZToxMDI4MjY0NDU5ODM2NDkzOjEyNjIxODQ=",
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
                    "inum": 4026533911,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533912,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533914,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533915,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533915,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533829,
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
                    "inum": 4026533916,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1287388,
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
    "time": "2025-07-07T12:14:01.263533522Z",
    "aggregation_info": null
}

*/
