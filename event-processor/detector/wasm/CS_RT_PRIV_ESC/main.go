//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_PRIV_ESC"
	Name        = "Privilege escalation"
	Description = "The detector detects if a non-root user makes the commit_creds calls with the UID/EUID == 0 or GID/EGID == 0, which may indicate an attacker's attempt to escalate their privileges."
	Version     = 1
	Author      = "Runtime Radar Team"

	License = "Apache License 2.0"
)

const (
	// UID of superuser
	rootUID = 0
	// GID of superuser
	rootGID = 0
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"commit_creds"},

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
		functionName := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		if functionName != "commit_creds" {
			return resp, nil // <-- return
		}

		// We are taking into account only processes which are running from regular user
		if kprobe.GetProcess().GetUid().GetValue() != rootUID {
			for _, arg := range args {
				switch a := arg.GetArg().(type) {
				case *tetragon.KprobeArgument_ProcessCredentialsArg:
					// Checking EUID will also cover cases with setuid binary
					if a.ProcessCredentialsArg.GetUid().GetValue() == rootUID ||
						a.ProcessCredentialsArg.GetEuid().GetValue() == rootUID {
						resp.Severity = api.DetectResp_CRITICAL // <-- threat detected

						return resp, nil // <-- return
					}
				}
			}
		}

		if kprobe.GetProcess().GetProcessCredentials().GetGid().GetValue() != rootGID {
			for _, arg := range args {
				switch a := arg.GetArg().(type) {
				case *tetragon.KprobeArgument_ProcessCredentialsArg:
					// Checking EGID will also cover cases with setgid binary
					if a.ProcessCredentialsArg.GetGid().GetValue() == rootGID ||
						a.ProcessCredentialsArg.GetEgid().GetValue() == rootGID {
						resp.Severity = api.DetectResp_CRITICAL // <-- threat detected

						return resp, nil // <-- return
					}
				}
			}
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
            "exec_id": "a2luZC1jb250cm9sLXBsYW5lOjg2NzQ3MDIxMzk5NjI2NToyMjA1Mzk2",
            "pid": 2205396,
            "uid": 5,
            "cwd": "/usr/games",
            "binary": "/usr/bin/su",
            "arguments": "-",
            "flags": "execve clone",
            "start_time": "2023-10-26T15:23:43.191924497Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://f1a282e1a9daf70be7d466c6c064884f5f6cad37add2fa877a94e384b6e13e91",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:b55e2651b71408015f8068dd74e1d04404a8fa607dd2cfe284b4824c11f4d9bd",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2023-10-26T15:04:37Z",
                    "pid": 62
                },
                "workload": "test-pod-debian"
            },
            "docker": "f1a282e1a9daf70be7d466c6c064884",
            "parent_exec_id": "a2luZC1jb250cm9sLXBsYW5lOjg2NzQ1NzY4MDkxNTE0MToyMjA1Mjc0",
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
                ]
            },
            "ns": {
                "uts": {
                    "inum": 4026533028
                },
                "ipc": {
                    "inum": 4026533029
                },
                "mnt": {
                    "inum": 4026533034
                },
                "pid": {
                    "inum": 4026533039
                },
                "pid_for_children": {
                    "inum": 4026533039
                },
                "net": {
                    "inum": 4026532910
                },
                "cgroup": {
                    "inum": 4026532390,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2205396
        },
        "parent": {
            "exec_id": "a2luZC1jb250cm9sLXBsYW5lOjg2NzQ1NzY4MDkxNTE0MToyMjA1Mjc0",
            "pid": 2205274,
            "uid": 5,
            "cwd": "/usr/games",
            "binary": "/bin/bash",
            "flags": "execve clone",
            "start_time": "2023-10-26T15:23:30.658843773Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://f1a282e1a9daf70be7d466c6c064884f5f6cad37add2fa877a94e384b6e13e91",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:b55e2651b71408015f8068dd74e1d04404a8fa607dd2cfe284b4824c11f4d9bd",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2023-10-26T15:04:37Z",
                    "pid": 60
                },
                "workload": "test-pod-debian"
            },
            "docker": "f1a282e1a9daf70be7d466c6c064884",
            "parent_exec_id": "a2luZC1jb250cm9sLXBsYW5lOjg2NzQ1NzY3ODc3NDM3ODoyMjA1Mjcz",
            "cap": {},
            "ns": {
                "uts": {
                    "inum": 4026533028
                },
                "ipc": {
                    "inum": 4026533029
                },
                "mnt": {
                    "inum": 4026533034
                },
                "pid": {
                    "inum": 4026533039
                },
                "pid_for_children": {
                    "inum": 4026533039
                },
                "net": {
                    "inum": 4026532910
                },
                "cgroup": {
                    "inum": 4026532390,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2205274
        },
        "function_name": "commit_creds",
        "args": [
            {
                "process_credentials_arg": {
                    "uid": 0,
                    "gid": 0,
                    "euid": 0,
                    "egid": 60,
                    "suid": 0,
                    "sgid": 60,
                    "fsuid": 0,
                    "fsgid": 60,
                    "caps": {
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
                        ]
                    },
                    "user_ns": {
                        "level": 0,
                        "uid": 0,
                        "gid": 0,
                        "ns": {
                            "inum": 4026531837,
                            "is_host": true
                        }
                    }
                }
            }
        ],
        "action": "KPROBE_ACTION_POST"
    },
    "node_name": "kind-control-plane",
    "time": "2023-10-26T15:23:46.828531551Z"
}

*/
