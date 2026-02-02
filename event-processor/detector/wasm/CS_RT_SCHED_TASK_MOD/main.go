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
	ID          = "CS_RT_SCHED_TASK_MOD"
	Name        = "Suspicious changes in task scheduler configuration files"
	Description = "The detector detects changes in configuration files of a task scheduler (for example, cron) if it was installed in a container or added by an attacker."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	// File access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/linux/fs.h#L100
	MAY_WRITE = 2

	// Memory page access permissions
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L11
	PROT_WRITE = 2
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},

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
	schedulerFiles = []glob.Glob{
		glob.MustCompile("/etc/crontab"),                          // system task scheduler
		glob.MustCompile("/etc/anacrontab"),                       // system task scheduler
		glob.MustCompile("/etc/cron.d/*"),                         // system task scheduler
		glob.MustCompile("/etc/cron.hourly/*"),                    // tasks with predefined hourly interval
		glob.MustCompile("/etc/cron.daily/*"),                     // tasks with predefined daily interval
		glob.MustCompile("/etc/cron.weekly/*"),                    // tasks with predefined weekly interval
		glob.MustCompile("/etc/cron.monthly/*"),                   // tasks with predefined monthly interval
		glob.MustCompile("/var/spool/cron/*"),                     // user task scheduler
		glob.MustCompile("/var/spool/anacron/*"),                  // user task scheduler
		glob.MustCompile("/etc/cron.deny"),                        // user access list for scheduler
		glob.MustCompile("/etc/cron.allow"),                       // user access list for scheduler
		glob.MustCompile("/var/spool/at/*"),                       // user task scheduler
		glob.MustCompile("/etc/at.deny"),                          // user access list for scheduler
		glob.MustCompile("/etc/at.allow"),                         // user access list for scheduler
		glob.MustCompile("/etc/systemd/system/*.timer"),           // systemd task scheduler
		glob.MustCompile("/usr/local/lib/systemd/system/*.timer"), // systemd task scheduler
		glob.MustCompile("/lib/systemd/system/*.timer"),           // systemd task scheduler
		glob.MustCompile("/usr/lib/systemd/system/*.timer"),       // systemd task scheduler
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
		// trigger when security function check for file write access
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != MAY_WRITE {
				return resp, nil
			}

			path = args[0].GetFileArg().GetPath()

		// trigger when security function check for memory page write access
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_WRITE == 0 {
				return resp, nil
			}

			path = args[0].GetFileArg().GetPath()

		// trigger when security function check if truncating a file is allowed
		// https://elixir.bootlin.com/linux/v6.10.6/source/security/security.c#L1923
		case "security_path_truncate":
			if len(args) < 1 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
			}

			path = args[0].GetPathArg().GetPath()

		default:
			return resp, nil
		}

		for _, file := range schedulerFiles {
			if file.Match(path) {
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNTE5ODczNTgwOTM0MTk6NjAyNjQy",
            "pid": 602642,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/vim",
            "arguments": "/etc/crontab",
            "flags": "execve rootcwd clone",
            "start_time": "2024-08-08T09:52:03.004302955Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "cri-o://426ccd7bdd6e9565a3f2767765b0c1fc160c8132c331884a6000759307b4fae2",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "31d5e503c34f4496a263fb3557575cf53e6a40add4c459370120c7454985f7b7",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-18T19:36:11Z",
                    "pid": 4183,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "426ccd7bdd6e9565a3f2767765b0c1f",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNTE4OTMwOTUxMjc5Mjc6NjAxNTc0",
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
                    "inum": 4026532264,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532265,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026532792,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532266,
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
                    "inum": 4026532794,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 602642,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNTE4OTMwOTUxMjc5Mjc6NjAxNTc0",
            "pid": 601574,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-08-08T09:50:28.741338076Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "default",
                "name": "test-pod-debian",
                "container": {
                    "id": "cri-o://426ccd7bdd6e9565a3f2767765b0c1fc160c8132c331884a6000759307b4fae2",
                    "name": "test-pod-debian",
                    "image": {
                        "id": "31d5e503c34f4496a263fb3557575cf53e6a40add4c459370120c7454985f7b7",
                        "name": "docker.io/library/debian:12.2-slim"
                    },
                    "start_time": "2024-05-18T19:36:11Z",
                    "pid": 4032,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "426ccd7bdd6e9565a3f2767765b0c1f",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjcxNTE4OTMwOTM3MzM1NTk6NjAxNTc0",
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
                    "inum": 4026532264,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026532265,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026532792,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026532793,
                    "is_host": false
                },
                "net": {
                    "inum": 4026532266,
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
                    "inum": 4026532794,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 601574,
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
        "function_name": "security_file_permission",
        "args": [
            {
                "file_arg": {
                    "mount": "",
                    "path": "/etc/crontab",
                    "flags": "",
                    "permission": "-rw-r--r--"
                },
                "label": ""
            },
            {
                "int_arg": 2,
                "label": ""
            }
        ],
        "return": {
            "int_arg": 0,
            "label": ""
        },
        "action": "KPROBE_ACTION_POST",
        "kernel_stack_trace": [],
        "policy_name": "file-monitoring",
        "return_action": "KPROBE_ACTION_POST",
        "message": "",
        "tags": [],
        "user_stack_trace": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-08-08T09:52:07.009362950Z",
    "aggregation_info": null
}

*/
