//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_CVE_2026_31431"
	Name        = "Copy Fail vulnerability"
	Description = "The detector detects if the AF_ALG network socket was created, which may indicate an attempt to exploit vulnerability CVE-2026-31431."
	Version     = 1
	Author      = "Runtime Radar Team"

	License = "Apache License 2.0"
)

const (
	// Socket Address Family: https://elixir.bootlin.com/linux/latest/source/include/linux/socket.h#L188
	AF_ALG = 38
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_socket_create"},

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

		if function != "security_socket_create" {
			return resp, nil
		}

		args := kprobe.GetArgs()

		if len(args) < 2 {
			return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
		}

		socketFamily := args[0].GetIntArg()

		// Trigger on AF_ALG socket creation
		if socketFamily == AF_ALG {
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
            "exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDozMDE0NDI1MTg3NDc3NDc0OjE3ODA4MzI=",
            "pid": 1780832,
            "uid": 0,
            "cwd": "/",
            "binary": "/bin/copyfail",
            "arguments": "",
            "flags": "execve rootcwd clone inInitTree",
            "start_time": "2026-05-06T12:48:22.187586619Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "copy-fail-poc-7b69b8cb65-kbqmh",
                "container": {
                    "id": "containerd://eff8f2497170de2bbfd2a4c9c01c9223d5b3c8b9157462c4822e51a7760bc084",
                    "name": "copy-fail-poc",
                    "image": {
                        "id": "ghcr.io/percivalll/copy-fail-cve-2026-31431-kubernetes-poc@sha256:b28d157b1034f7412611606852b62bd15c108d49b86f12fc143ae3edc9403358",
                        "name": "ghcr.io/percivalll/copy-fail-cve-2026-31431-kubernetes-poc:latest"
                    },
                    "start_time": "2026-05-06T12:48:22Z",
                    "pid": 7,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app.kubernetes.io/name": "copy-fail-poc",
                    "app.kubernetes.io/part-of": "cve-2026-31431",
                    "pod-template-hash": "7b69b8cb65"
                },
                "workload": "copy-fail-poc",
                "workload_kind": "Deployment"
            },
            "docker": "eff8f2497170de2bbfd2a4c9c01c922",
            "parent_exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDozMDE0NDI1MTg2MzM2MzU1OjE3ODA4MTk=",
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
                    "inum": 4026533896,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533897,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533899,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533900,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533900,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533795,
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
                    "inum": 4026533901,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1780832,
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
            "in_init_tree": true
        },
        "parent": {
            "exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDozMDE0NDI1MTg2MzM2MzU1OjE3ODA4MTk=",
            "pid": 1780819,
            "uid": 0,
            "cwd": "/",
            "binary": "/bin/sh",
            "arguments": "-c \"/bin/copyfail && sleep 365d\"",
            "flags": "execve rootcwd clone inInitTree",
            "start_time": "2026-05-06T12:48:22.186445771Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "copy-fail-poc-7b69b8cb65-kbqmh",
                "container": {
                    "id": "containerd://eff8f2497170de2bbfd2a4c9c01c9223d5b3c8b9157462c4822e51a7760bc084",
                    "name": "copy-fail-poc",
                    "image": {
                        "id": "ghcr.io/percivalll/copy-fail-cve-2026-31431-kubernetes-poc@sha256:b28d157b1034f7412611606852b62bd15c108d49b86f12fc143ae3edc9403358",
                        "name": "ghcr.io/percivalll/copy-fail-cve-2026-31431-kubernetes-poc:latest"
                    },
                    "start_time": "2026-05-06T12:48:22Z",
                    "pid": 1,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app.kubernetes.io/name": "copy-fail-poc",
                    "app.kubernetes.io/part-of": "cve-2026-31431",
                    "pod-template-hash": "7b69b8cb65"
                },
                "workload": "copy-fail-poc",
                "workload_kind": "Deployment"
            },
            "docker": "eff8f2497170de2bbfd2a4c9c01c922",
            "parent_exec_id": "c3p5LWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDozMDE0NDI0MDY3NTA0Njk4OjE3ODA3Mzc=",
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
                    "inum": 4026533896,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533897,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533899,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533900,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533900,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533795,
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
                    "inum": 4026533901,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1780819,
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
            "in_init_tree": true
        },
        "function_name": "security_socket_create",
        "args": [
            {
                "int_arg": 38,
                "label": "Family"
            },
            {
                "int_arg": 5,
                "label": "Type"
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "listen-socket",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "cs-master-node",
    "time": "2026-05-06T12:48:31.181791290Z",
    "aggregation_info": null,
    "cluster_name": ""
}
*/
