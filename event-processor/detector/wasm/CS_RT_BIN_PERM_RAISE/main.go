//go:build tinygo.wasm

package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_BIN_PERM_RAISE"
	Name        = "Change of file access permisions"
	Description = "The detector detects if permissions to execute files in the boot, dev, home, media, mnt, run, sys, tmp, and var directories were granted."
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
		"PROCESS_EXEC":   {},
		"PROCESS_KPROBE": {"security_path_chmod"},

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
	chmod       = glob.MustCompile(`*/chmod`)
	execPerm    = regexp.MustCompile(`^(?:[\-Rcfv\s]+)?(?:(?:[0-7]?[0-7][1357][0-7]\s|[0-7]?[1357][0-7][0-7]\s|[0-7]?[0-7][0-7][1357]\s)|(?:[ugoa]*[+=]x\s))`)
	suspDirArgs = regexp.MustCompile(`^(?:[\-Rcfv\s]+)?(?:(?:[ugoa]*[-=+][rwxX])|(?:[0-7]{3,4}))(?:.*)\/?(?:boot|dev|home|media|mnt|run|sys|tmp|var)\/`)
	relPath     = regexp.MustCompile(`^(?:[\-Rcfv\s]+)?(?:(?:[ugoa]*[-=+][rwxX])|(?:[0-7]{3,4}))(?:.*)\s(?:[^\/])`)

	suspDirCwd = []glob.Glob{
		glob.MustCompile("/boot*"),
		glob.MustCompile("/dev*"),
		glob.MustCompile("/home*"),
		glob.MustCompile("/media*"),
		glob.MustCompile("/mnt*"),
		glob.MustCompile("/run*"),
		glob.MustCompile("/srv*"),
		glob.MustCompile("/sys*"),
		glob.MustCompile("/tmp*"),
		glob.MustCompile("/var*"),
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
		cwd := exec.GetProcess().GetCwd()

		// Check binary.
		if !chmod.Match(binary) {
			return resp, nil
		}

		// Check if execute access permissions are in args.
		execPermSet := false
		if execPerm.MatchString(args) {
			execPermSet = true
		} else {
			return resp, nil
		}

		// Check if suspicious directory is in args.
		suspiciousDir := false
		if suspDirArgs.MatchString(args) {
			suspiciousDir = true
		}

		// Next sequence is for cases like: cd /tmp && chmod +x ./xmrig).
		if !suspiciousDir {
			// Check if current directory is suspicious.
			suspiciousDirInCwd := false
			for _, d := range suspDirCwd {
				if d.Match(cwd) {
					suspiciousDirInCwd = true
				}
			}

			// Check if path in args is relative.
			if relPath.MatchString(args) && suspiciousDirInCwd {
				suspiciousDir = true
			}
		}

		// Trigger only on executable bits in arguments and suspicious directories.
		if execPermSet && suspiciousDir {
			resp.Severity = api.DetectResp_MEDIUM // <-- threat detected
			return resp, nil
		}

		return resp, nil

	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()
		path := ""

		switch function {
		case "security_path_chmod":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			}

			path = args[0].GetPathArg().GetPath()
		default:
			return resp, nil
		}

		for _, dir := range suspDirCwd {
			if dir.Match(path) {
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected

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

PROCESS_EXEC

{
    "process_exec": {
        "process": {
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZTo2MDI5MzkyNDczOTkyNTY1OjI5ODg4OTU=",
            "pid": 2988895,
            "uid": 0,
            "cwd": "/root",
            "binary": "/usr/bin/chmod",
            "arguments": "+x xmrig",
            "flags": "execve clone",
            "start_time": "2024-11-13T10:15:07.462453463Z",
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
                    "pid": 4630,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "2ee9cb5d00f3c58ea29f9380aaa86c8",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZTo1OTU4NjE5OTE0NDEwNjYyOjg3NjcyNg==",
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
            "tid": 2988895,
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
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZTo1OTU4NjE5OTE0NDEwNjYyOjg3NjcyNg==",
            "pid": 876726,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-11-12T14:35:34.902871824Z",
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
                    "pid": 4609,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian-raw",
                "workload_kind": "Pod"
            },
            "docker": "2ee9cb5d00f3c58ea29f9380aaa86c8",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZTo1OTU4NjE5ODY4NDU0Njk2Ojg3NjcyNg==",
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
            "tid": 876726,
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
    "node_name": "experts-k8s-cs",
    "time": "2024-11-13T10:15:07.462452981Z",
    "aggregation_info": null
}

PROCESS_KPROBE

{
    "process_kprobe": {
        "process": {
            "exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODEwOTkyMTM4Njc5NjQ3NjoyNzQ1MTE3",
            "pid": 2745117,
            "uid": 0,
            "cwd": "/tmp",
            "binary": "/usr/bin/chmod",
            "arguments": "o+x reg-file",
            "flags": "execve clone",
            "start_time": "2025-12-10T19:03:39.717065187Z",
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
                    "pid": 174,
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
            "parent_exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODEwNzg5OTk2NDM0Nzg5MDoyNzAxOTgy",
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
            "tid": 2745117,
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
            "exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODEwNzg5OTk2NDM0Nzg5MDoyNzAxOTgy",
            "pid": 2701982,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2025-12-10T18:29:58.294617147Z",
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
                    "pid": 82,
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
            "parent_exec_id": "dnNoLWs4cy5hcHBzZWMtc3RhbmQucHRzZWN1cml0eS5jbG91ZDoxODEwNzg5OTk2MzIwNzExMDoyNzAxOTgy",
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
            "tid": 2701982,
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
        "function_name": "security_path_chmod",
        "args": [
            {
                "path_arg": {
                    "mount": "",
                    "path": "/tmp/reg-file",
                    "flags": "",
                    "permission": "-rw-r--rw-"
                },
                "label": "file path"
            },
            {
                "uint_arg": 423,
                "label": "file mode"
            }
        ],
        "return": null,
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "permissions-manipulation",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": [],
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2025-12-10T19:03:39.717614198Z",
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
