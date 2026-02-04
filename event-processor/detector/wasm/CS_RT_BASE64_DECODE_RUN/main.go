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
	ID          = "CS_RT_BASE64_DECODE_RUN"
	Name        = "BASE64: data decoding"
	Description = "The detector detects if utilities for decoding arbitrary text from Base64 are used."
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
	base64Utils = []glob.Glob{
		glob.MustCompile("*/base64"),
		// Utils down below are included in basez package and can be used by an attacker to avoid detection.
		// https://manpages.debian.org/unstable/basez/base64plain.1.en.html
		glob.MustCompile("*/base64plain"),
		glob.MustCompile("*/base64url"),
		glob.MustCompile("*/base64mime"),
		glob.MustCompile("*/base64pem"),
	}

	decodeArgs = regexp.MustCompile(`-(?:d|D|-decode)`)
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

		for _, b64 := range base64Utils {
			if b64.Match(binary) && decodeArgs.MatchString(args) {
				resp.Severity = api.DetectResp_LOW // <-- threat detected
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
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZTo4NTMzNzg0OTcwMTIwODc3OjE5OTI5MzQ=",
            "pid": 1992934,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/base64",
            "arguments": "-d",
            "flags": "execve rootcwd clone",
            "start_time": "2024-12-12T09:55:30.420477932Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-raw",
                "container": {
                    "id": "containerd://2ee9cb5d00f3c58ea29f9380aaa86c86e70efc424bfa92c33d8ee0d50e0b6d00",
                    "name": "test-pod-debian-raw",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-10-10T10:11:11Z",
                    "pid": 6051,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "2ee9cb5d00f3c58ea29f9380aaa86c8",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZTo4MjkzMzM5MDcyMTY2OTc5OjMxODA2Mzg=",
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
                    "inum": 4026534481,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534482,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534484,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534485,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534485,
                    "is_host": false
                },
                "net": {
                    "inum": 4026534422,
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
                    "inum": 4026534486,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 1992934,
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
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZTo4MjkzMzM5MDcyMTY2OTc5OjMxODA2Mzg=",
            "pid": 3180638,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-12-09T15:07:53.152721975Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian-raw",
                "container": {
                    "id": "containerd://2ee9cb5d00f3c58ea29f9380aaa86c86e70efc424bfa92c33d8ee0d50e0b6d00",
                    "name": "test-pod-debian-raw",
                    "image": {
                        "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-10-10T10:11:11Z",
                    "pid": 5874,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "2ee9cb5d00f3c58ea29f9380aaa86c8",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZTo4MjkzMzM5MDY0MzE0NDQzOjMxODA2Mzg=",
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
                    "inum": 4026534481,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026534482,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026534484,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026534485,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026534485,
                    "is_host": false
                },
                "net": {
                    "inum": 4026534422,
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
                    "inum": 4026534486,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 3180638,
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
    "time": "2024-12-12T09:55:30.420475628Z",
    "aggregation_info": null
}
*/
