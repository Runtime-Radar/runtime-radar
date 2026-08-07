//go:build tinygo.wasm

package main

import (
	"context"
	"regexp"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_OPENSSL_CUSTOM_LIB_LOAD"
	Name        = "OpenSSL: custom library loading"
	Description = "The detector detects if the OpenSSL utility was used to load custom libraries, which may indicate an attacker's attempt to run malicious code."
	Version     = 1
	Author      = "Runtime Radar Team"

	License = "Apache License 2.0"
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
	// Command for loading custom user library as crypto engine.
	libraryLoadCmd = regexp.MustCompile(`^engine\s(?:dynamic\s)?.*-pre SO_PATH:.*-pre LOAD`)
	// Load custom user library using -engine option as part of other openssl commands.
	libraryLoadOption = regexp.MustCompile(`^\w+\s-engine\s`)
)

// init is required for TinyGo to compile to Wasm.
func init() {
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
		args := exec.GetProcess().GetArguments()

		if libraryLoadCmd.MatchString(args) || libraryLoadOption.MatchString(args) {
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected
			return resp, nil
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
            "exec_id": "cHRjcy1kZWJpYW4tMTE6NDUxNDY1OTkyOTExODg2OToyNzQzNjc2",
            "pid": 2743676,
            "uid": 0,
            "cwd": "/root/openssl",
            "binary": "/usr/bin/openssl",
            "arguments": "engine dynamic -pre SO_PATH:./hostname.so -pre LOAD",
            "flags": "execve clone",
            "start_time": "2025-04-10T14:17:32.090308606Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://e9db8d37b2ebde52f63ed8055253e3829175d82285cefa1c06014a8ff59b11c8",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-03-04T11:51:22Z",
                    "pid": 5520,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "e9db8d37b2ebde52f63ed8055253e38",
            "parent_exec_id": "cHRjcy1kZWJpYW4tMTE6NDUxNDQ1NjExOTAyMTQzNzoyNzM4MTc4",
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
                    "inum": 4026533060,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533061,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533063,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533064,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533064,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532932,
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
                    "inum": 4026533065,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2743676,
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
            "exec_id": "cHRjcy1kZWJpYW4tMTE6NDUxNDQ1NjExOTAyMTQzNzoyNzM4MTc4",
            "pid": 2738178,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2025-04-10T14:14:08.280211442Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://e9db8d37b2ebde52f63ed8055253e3829175d82285cefa1c06014a8ff59b11c8",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2025-03-04T11:51:22Z",
                    "pid": 5508,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "e9db8d37b2ebde52f63ed8055253e38",
            "parent_exec_id": "cHRjcy1kZWJpYW4tMTE6NDUxNDQ1NjExMzYyNjk4NzoyNzM4MTc4",
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
                    "inum": 4026533060,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533061,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533063,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533064,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533064,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532932,
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
                    "inum": 4026533065,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2738178,
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
    "node_name": "cs-debian-11",
    "time": "2025-04-10T14:17:32.090307890Z",
    "aggregation_info": null
}
*/
