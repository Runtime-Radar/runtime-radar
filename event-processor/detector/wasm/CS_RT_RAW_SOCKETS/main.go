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
	ID          = "CS_RT_RAW_SOCKETS"
	Name        = "Creation of raw network socket"
	Description = "The detector detects if a raw network socket was created, which may indicate network traffic interception or reconnaissance."
	Version     = 2
	Author      = "Runtime Radar Team"

	License = "Apache License 2.0"
)

const (
	// Socket Address Family: https://elixir.bootlin.com/linux/latest/source/include/linux/socket.h#L188
	AF_INET  = 2
	AF_INET6 = 10

	// Socket Type: https://elixir.bootlin.com/linux/latest/source/include/linux/net.h#L64
	SOCK_RAW    = 3
	SOCK_PACKET = 10
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

var (
	binaryWhitelist = []*regexp.Regexp{
		regexp.MustCompile(`/NetworkManager$`),
		regexp.MustCompile(`/dhclient$`),
		regexp.MustCompile(`/f?ping6?$`),
		regexp.MustCompile(`/traceroute6?$`),
		regexp.MustCompile(`/(?:ip|x)tables.*$`),
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
		binary := kprobe.GetProcess().GetBinary()

		if function != "security_socket_create" {
			return resp, nil
		}

		// Exclude whitelisted binary from detect logic
		for _, bin := range binaryWhitelist {
			if bin.MatchString(binary) {
				return resp, nil
			}
		}

		args := kprobe.GetArgs()

		if len(args) < 2 {
			return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
		}

		socketFamily := args[0].GetIntArg()
		socketType := args[1].GetIntArg()

		// Regular RAW sockets
		if (socketFamily == AF_INET || socketFamily == AF_INET6) && (socketType == SOCK_RAW || socketType == SOCK_PACKET) {
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjY5MTYxNzYzNjg0ODk0MTQ6NDA3MjM4Nw==",
            "pid": 4072387,
            "uid": 0,
            "cwd": "/root/raw_socket",
            "binary": "/root/raw_socket/a.out",
            "arguments": "",
            "flags": "execve clone",
            "start_time": "2024-05-03T12:23:23.161940060Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-raw",
                "labels": [],
                "container": {
                    "id": "cri-o://c50e7f66b70c88b0321e18c42edd8077623ac7b1563cb13400815158ea3b987b",
                    "name": "test-pod-debian-raw",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-03T12:14:26Z",
                    "pid": 1499,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "c50e7f66b70c88b0321e18c42edd807",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjY5MTU2NDU0MDMzMjY2NzA6NDA2NTU5OQ==",
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
                    "CAP_NET_RAW"
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
                    "CAP_NET_RAW"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026533977,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533978,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534087,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534088,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534088,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533979,
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
                    "inum": 4026534089,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 4072387,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjY5MTU2NDU0MDMzMjY2NzA6NDA2NTU5OQ==",
            "pid": 4065599,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-05-03T12:14:32.196777366Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-raw",
                "labels": [],
                "container": {
                    "id": "cri-o://c50e7f66b70c88b0321e18c42edd8077623ac7b1563cb13400815158ea3b987b",
                    "name": "test-pod-debian-raw",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-03T12:14:26Z",
                    "pid": 7,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "c50e7f66b70c88b0321e18c42edd807",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjY5MTU2NDU0MDA1MjI1NjQ6NDA2NTU5OQ==",
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
                    "CAP_NET_RAW"
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
                    "CAP_NET_RAW"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026533977,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533978,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534087,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534088,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534088,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533979,
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
                    "inum": 4026534089,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 4065599,
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
        "function_name": "security_socket_create",
        "args": [
            {
                "int_arg": 2,
                "label": "Family"
            },
            {
                "int_arg": 3,
                "label": "Type"
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "stack_trace": [],
        "policy_name": "raw-socket-monitoring"
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-05-03T12:23:23.162428452Z",
    "aggregation_info": null
}

*/
