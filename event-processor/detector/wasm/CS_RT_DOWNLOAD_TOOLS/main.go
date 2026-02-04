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
	ID          = "CS_RT_DOWNLOAD_TOOLS"
	Name        = "Network activity: file downloading"
	Description = "The detector detects suspicious network activity of utilities used for file transferring and downloading."
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
		"PROCESS_KPROBE": {"tcp_connect"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)

// downloadTool represents a combination of binary's and its arguments' patterns.
type downloadTool struct {
	binaryPattern glob.Glob
	argsPatterns  []*regexp.Regexp
}

var downloadTools = []downloadTool{
	// tools whose arguments we don't check
	{binaryPattern: glob.MustCompile("*/wget")},
	{binaryPattern: glob.MustCompile("*/ftp")},
	{binaryPattern: glob.MustCompile("*/svn")},
	{binaryPattern: glob.MustCompile("*/git-remote-*")},
	{binaryPattern: glob.MustCompile("*/sftp")},
	{binaryPattern: glob.MustCompile("*/sftp-server")},

	// tools whose arguments we check
	{
		binaryPattern: glob.MustCompile("*/curl"),
		argsPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(^-|\s+-)[^-\s]*(o|O)\S*\s*\S+`),
			regexp.MustCompile(`--((output)|(remote-name))\s*\S+`),
		},
	},
	{
		binaryPattern: glob.MustCompile("*/ssh"),
		argsPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\s*-[^-\s]*s\s+`),
		},
	},
}

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
		args := kprobe.GetProcess().GetArguments()
		function := kprobe.GetFunctionName()

		if function != "tcp_connect" {
			return resp, nil
		}

		match := false

		for _, dt := range downloadTools {
			if !dt.binaryPattern.Match(binary) {
				continue // no need to check args as binary doesn't match
			}

			// in case of some tools args are not checked
			if len(dt.argsPatterns) == 0 {
				match = true
				break
			}

			for _, p := range dt.argsPatterns {
				if p.MatchString(args) {
					match = true
					break
				}
			}
		}

		if match {
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODUwNzQ3NzcyMjkxMTg6MTY5MzczMg==",
            "pid": 1693732,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/curl",
            "arguments": "-LO https://github.com/carlospolop/PEASS-ng/releases/latest/download/linpeas.sh",
            "flags": "execve rootcwd clone",
            "start_time": "2024-03-22T11:45:01.570679989Z",
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
                    "pid": 60210,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODMxMzQwNzgzMDQwMjY6MTY3NDQ4NQ==",
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
            "tid": 1693732,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODMxMzQwNzgzMDQwMjY6MTY3NDQ4NQ==",
            "pid": 1674485,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-03-22T11:12:40.871754656Z",
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
                    "pid": 60201,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODMxMzQwNzQ2MjE2NDY6MTY3NDQ4NQ==",
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
            "tid": 1674485,
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
        "function_name": "tcp_connect",
        "args": [
            {
                "sock_arg": {
                    "family": "AF_INET",
                    "type": "SOCK_STREAM",
                    "protocol": "IPPROTO_TCP",
                    "mark": 0,
                    "priority": 0,
                    "saddr": "10.244.0.97",
                    "daddr": "140.82.121.4",
                    "sport": 55106,
                    "dport": 443,
                    "cookie": "18446635049337231936",
                    "state": "TCP_SYN_SENT"
                },
                "label": ""
            }
        ],
        "return": null,
        "action": "KPROBE_ACTION_POST",
        "stack_trace": [],
        "policy_name": "tcp-out-monitoring"
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-03-22T11:45:01.589721382Z",
    "aggregation_info": null
}

*/
