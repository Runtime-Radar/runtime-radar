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
	ID          = "CS_RT_FIFO_FILE_CREATE"
	Name        = "Creation of named pipe file"
	Description = "The detector detects signs that a named pipe file was created. Applications can use such a file to bypass existing audit policies when exchanging data; for example, to create a reverse shell."
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
		"PROCESS_KPROBE": {"security_path_mknod"},

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
	suspDirs = []glob.Glob{
		glob.MustCompile("/boot/*"),
		glob.MustCompile("/dev/*"),
		glob.MustCompile("/home/*"),
		glob.MustCompile("/media/*"),
		glob.MustCompile("/mnt/*"),
		glob.MustCompile("/run/*"),
		glob.MustCompile("/srv/*"),
		glob.MustCompile("/tmp/*"),
		glob.MustCompile("/var/*"),
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
		args := kprobe.GetArgs()
		path := ""

		switch function {
		case "security_path_mknod":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			}

			path = args[0].GetPathArg().GetPath()
		default:
			return resp, nil
		}

		for _, dir := range suspDirs {
			if dir.Match(path) {
				resp.Severity = api.DetectResp_HIGH // <-- threat detected

				return resp, nil
			}
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
            "exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODA3MTQ5NjA3OTAyMjM5MDoxOTE5OTAw",
            "pid": 1919900,
            "uid": 0,
            "cwd": "/tmp",
            "binary": "/usr/bin/mkfifo",
            "arguments": "fifo-test",
            "flags": "execve clone",
            "start_time": "2025-12-10T08:23:14.409290982Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://35ae39802175a6d328218c2d35d878bf0402b37a229ee28b2960161c567eca5d",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:7e5bc0e499a8d50cb1e32287944a90b9ec8fd7d500673e75daff3f52882f5798",
                        "name": "docker.io/library/debian:12"
                    },
                    "start_time": "2025-12-10T07:50:30Z",
                    "pid": 54,
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
            "docker": "35ae39802175a6d328218c2d35d878b",
            "parent_exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODA3MTQ2MjM2MjUzMTYzMToxOTE5MjAw",
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
                    "inum": 4026533802,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533803,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533805,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533806,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533806,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533163,
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
                    "inum": 4026533807,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1919900,
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
            "exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODA3MTQ2MjM2MjUzMTYzMToxOTE5MjAw",
            "pid": 1919200,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2025-12-10T08:22:40.692800397Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://35ae39802175a6d328218c2d35d878bf0402b37a229ee28b2960161c567eca5d",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:7e5bc0e499a8d50cb1e32287944a90b9ec8fd7d500673e75daff3f52882f5798",
                        "name": "docker.io/library/debian:12"
                    },
                    "start_time": "2025-12-10T07:50:30Z",
                    "pid": 45,
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
            "docker": "35ae39802175a6d328218c2d35d878b",
            "parent_exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODA3MTQ2MjM2MTU0OTAwNjoxOTE5MjAw",
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
                    "inum": 4026533802,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533803,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533805,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533806,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533806,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533163,
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
                    "inum": 4026533807,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1919200,
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
        "function_name": "security_path_mknod",
        "args": [
            {
                "path_arg": {
                    "mount": "",
                    "path": "/tmp/fifo-test",
                    "flags": "",
                    "permission": "---------"
                },
                "label": "new file path"
            },
            {
                "uint_arg": 4534,
                "label": "new file mode"
            }
        ],
        "return": null,
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "io-streams",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": [],
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2025-12-10T08:23:14.410150687Z",
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
