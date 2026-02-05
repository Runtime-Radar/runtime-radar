//go:build tinygo.wasm

package main

import (
	"context"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "CS_RT_HACK_TOOLS"
	Name        = "Hacking tools"
	Description = "The detector detects if any known hacking tools were started, such as nmap, netcat, and peirates. In addition, the detector detects startup of utilities that are not used directly for hacking but help to identify suspicious activity, such as id or whoami. The latter functionality may be removed in future versions of the detector."
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
	hackTools = []glob.Glob{
		// sudo and su a rarely used in runtime, containers are either rooted or rootless with no need to escalate privileges for a normal user.
		glob.MustCompile("*/sudo"),
		glob.MustCompile("*/su"),

		// These tools are supposed to be mostly malicious in general or in the context of containers and k8s environment.
		glob.MustCompile("*/nc"),
		glob.MustCompile("*/netcat"),
		glob.MustCompile("*/ncat"),
		glob.MustCompile("*/curl"),
		glob.MustCompile("*/wget"),
		glob.MustCompile("*/nmap"),
		glob.MustCompile("*/vanish"),
		glob.MustCompile("*/vanish[0-9]"),
		glob.MustCompile("*/peirates"),
		glob.MustCompile("*/linpeas.sh"),
		glob.MustCompile("*/nsenter"),
		glob.MustCompile("*/unshare"),
		glob.MustCompile("*/kubectl"),
		glob.MustCompile("*/arp"),
		glob.MustCompile("*/dd"),
		glob.MustCompile("*/ftp"),
		glob.MustCompile("*/svn"),
		glob.MustCompile("*/git-remote-*"),

		// Anything run from /tmp is a hack tool by definition, right?
		glob.MustCompile("/tmp/*"),

		// Networking and pivoting
		glob.MustCompile("*/socat"),
		glob.MustCompile("*/chisel"),
		glob.MustCompile("*/proxychains"),
		glob.MustCompile("*/ssf"),
		glob.MustCompile("*/ssfd"),
		glob.MustCompile("*/sshuttle"),
		glob.MustCompile("*/proxy"), // ligolo-ng
		glob.MustCompile("*/agent"), // ligolo-ng
		glob.MustCompile("*/gost"),
		glob.MustCompile("*/3proxy"),
	}

	// These tools are mostly legal but may halp in incident investigations as a good indication of potentially malicious actions.
	// Most of them added as a result of Standoff hackers activity analysis, and can be removed in later versions.
	suspiciousTools = []glob.Glob{
		glob.MustCompile("*/apt"),
		glob.MustCompile("*/apt-get"),
		glob.MustCompile("*/apk"),
		glob.MustCompile("*/id"),
		glob.MustCompile("*/whoami"),
		glob.MustCompile("*/env"),
		glob.MustCompile("*/printenv"),
		glob.MustCompile("*/ip"),
		glob.MustCompile("*/ifconfig"),
		glob.MustCompile("*/netstat"),
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
		bin := exec.GetProcess().GetBinary()

		for _, suspTool := range suspiciousTools {
			if suspTool.Match(bin) {
				resp.Severity = api.DetectResp_LOW // <-- threat detected

				return resp, nil
			}
		}

		for _, hackTool := range hackTools {
			if hackTool.Match(bin) {
				resp.Severity = api.DetectResp_CRITICAL // <-- threat detected

				return resp, nil
			}
		}

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
            "exec_id": "a2luZC1jb250cm9sLXBsYW5lOjEyNTE4NDc5MTIyMDY0OTo3OTEzODI=",
            "pid": 791382,
            "uid": 0,
            "cwd": "/tmp",
            "binary": "/usr/bin/nc",
            "arguments": "-vz postgres 5432",
            "flags": "execve clone",
            "start_time": "2023-11-21T22:09:06.537928399Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://9462ad558a049ce83a6d6f10fa8c2f1a63a26cbbc967721cd339ad7889d9b4de",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:b55e2651b71408015f8068dd74e1d04404a8fa607dd2cfe284b4824c11f4d9bd",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2023-11-18T21:50:00Z",
                    "pid": 451
                },
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "9462ad558a049ce83a6d6f10fa8c2f1",
            "parent_exec_id": "a2luZC1jb250cm9sLXBsYW5lOjEyMzgxNDY3NTk2MzQ5Nzo3ODE0NzY=",
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
                    "inum": 4026533031
                },
                "ipc": {
                    "inum": 4026533032
                },
                "mnt": {
                    "inum": 4026533082
                },
                "pid": {
                    "inum": 4026533083
                },
                "pid_for_children": {
                    "inum": 4026533083
                },
                "net": {
                    "inum": 4026532596
                },
                "cgroup": {
                    "inum": 4026532330,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 791382,
            "process_credentials": {
                "uid": 0,
                "gid": 0,
                "euid": 0,
                "egid": 0,
                "suid": 0,
                "sgid": 0,
                "fsuid": 0,
                "fsgid": 0
            }
        },
        "parent": {
            "exec_id": "a2luZC1jb250cm9sLXBsYW5lOjEyMzgxNDY3NTk2MzQ5Nzo3ODE0NzY=",
            "pid": 781476,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "flags": "execve rootcwd clone",
            "start_time": "2023-11-21T21:46:16.422671446Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "containerd://9462ad558a049ce83a6d6f10fa8c2f1a63a26cbbc967721cd339ad7889d9b4de",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "docker.io/library/debian@sha256:b55e2651b71408015f8068dd74e1d04404a8fa607dd2cfe284b4824c11f4d9bd",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2023-11-18T21:50:00Z",
                    "pid": 437
                },
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "9462ad558a049ce83a6d6f10fa8c2f1",
            "parent_exec_id": "a2luZC1jb250cm9sLXBsYW5lOjEyMzgxNDY0NjgyNzM1Njo3ODE0NjY=",
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
                    "inum": 4026533031
                },
                "ipc": {
                    "inum": 4026533032
                },
                "mnt": {
                    "inum": 4026533082
                },
                "pid": {
                    "inum": 4026533083
                },
                "pid_for_children": {
                    "inum": 4026533083
                },
                "net": {
                    "inum": 4026532596
                },
                "cgroup": {
                    "inum": 4026532330,
                    "is_host": true
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 781476,
            "process_credentials": {
                "uid": 0,
                "gid": 0,
                "euid": 0,
                "egid": 0,
                "suid": 0,
                "sgid": 0,
                "fsuid": 0,
                "fsgid": 0
            }
        }
    },
    "node_name": "kind-control-plane",
    "time": "2023-11-21T22:09:06.537928399Z"
}

*/
