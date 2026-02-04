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
	ID          = "CS_RT_K8S_SA_TOKEN_READ"
	Name        = "Reading of Kubernetes authentication token"
	Description = "The detector detects reading of the Kubernetes authentication token, which may indicate that the Kubernetes account was compromised."
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
		"PROCESS_KPROBE": {"security_file_permission", "security_mmap_file"},

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
	// Following code is commented and will probably be removed later on, because listed tools run on host itself, out of container context,
	// and will be filtered out by monitoring agent most of the time because of runtime filters.
	// The names of tools coded here can also be used by an attacker to avoid detection.
	//
	// legitUtilPatterns contains list of utils which are allowed to read ServiceAccount token.
	// legitUtilPatterns = []glob.Glob{
	// 	glob.MustCompile("*/flanneld"),
	// 	glob.MustCompile("*/kube-proxy"),
	// 	glob.MustCompile("*/etcd"),
	// 	glob.MustCompile("*/kube-apiserver"),
	// 	glob.MustCompile("*/coredns"),
	// 	glob.MustCompile("*/kube-controller"),
	// 	glob.MustCompile("*/kubectl"),
	// 	glob.MustCompile("*/storage-provisioner"),
	// }

	// fileRegex is a regex for path to a file where ServiceAccount token is stored.
	fileRegex = regexp.MustCompile("secrets/kubernetes.io/serviceaccount.+token$")
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

		containerIDPart := kprobe.GetProcess().GetDocker()

		// binary := kprobe.GetProcess().GetBinary()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		// Process events only from runtime env
		if containerIDPart == "" {
			return resp, nil
		}

		switch function {
		case "security_file_permission":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if mask := args[1].GetIntArg(); mask != 4 { // need MAY_READ
				return resp, nil
			}
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&1 != 1 { // need PROT_READ
				return resp, nil
			}
		default:
			return resp, nil
		}

		// This code commented intentionally, see the comment above.
		// for _, p := range legitUtilPatterns {
		// 	if p.Match(binary) {
		// 		return resp, nil
		// 	}
		// }

		path := args[0].GetFileArg().GetPath()

		if fileRegex.MatchString(path) {
			resp.Severity = api.DetectResp_HIGH // <-- threat detected
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
      "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjI5NDI3MzM0NjUyMTQ0MDc6MjQ4Mjc1MA==",
      "pid": 2482750,
      "uid": 0,
      "cwd": "/",
      "binary": "/usr/bin/cat",
      "arguments": "/var/run/secrets/kubernetes.io/serviceaccount/token",
      "flags": "execve rootcwd clone",
      "start_time": "2024-03-18T12:39:20.258664559Z",
      "auid": 4294967295,
      "pod": {
        "namespace": "default",
        "name": "test-pod-debian-cap",
        "labels": [],
        "container": {
          "id": "cri-o://635bb04d5b6d67585f78358b6d54d0373576ed0e0e24cf02c21c86a88549b903",
          "name": "test-pod-debian-cap",
          "image": {
            "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
            "name": "docker.io/library/debian:12.2-slim"
          },
          "start_time": "2024-03-06T11:01:56Z",
          "pid": 3662,
          "maybe_exec_probe": false
        },
        "pod_labels": {},
        "workload": "test-pod-debian-cap",
        "workload_kind": "Pod"
      },
      "docker": "635bb04d5b6d67585f78358b6d54d03",
      "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjI5NDI3MTQxMDQ5ODA2ODg6MjQ4MjUzNg==",
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
          "CAP_SYS_CHROOT",
          "CAP_SYS_ADMIN",
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
          "CAP_SYS_CHROOT",
          "CAP_SYS_ADMIN",
          "CAP_SETFCAP"
        ],
        "inheritable": []
      },
      "ns": {
        "uts": {
          "inum": 4026534057,
          "is_host": false
        },
        "ipc": {
          "inum": 4026534058,
          "is_host": false
        },
        "mnt": {
          "inum": 4026535084,
          "is_host": false
        },
        "pid": {
          "inum": 4026535085,
          "is_host": false
        },
        "pid_for_children": {
          "inum": 4026535085,
          "is_host": false
        },
        "net": {
          "inum": 4026534059,
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
          "inum": 4026535086,
          "is_host": false
        },
        "user": {
          "inum": 4026531837,
          "is_host": true
        }
      },
      "tid": 2482750,
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
      "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjI5NDI3MTQxMDQ5ODA2ODg6MjQ4MjUzNg==",
      "pid": 2482536,
      "uid": 0,
      "cwd": "/",
      "binary": "/usr/bin/bash",
      "arguments": "",
      "flags": "execve rootcwd",
      "start_time": "2024-03-18T12:39:00.898431349Z",
      "auid": 4294967295,
      "pod": {
        "namespace": "default",
        "name": "test-pod-debian-cap",
        "labels": [],
        "container": {
          "id": "cri-o://635bb04d5b6d67585f78358b6d54d0373576ed0e0e24cf02c21c86a88549b903",
          "name": "test-pod-debian-cap",
          "image": {
            "id": "docker.io/library/debian@sha256:2bc5c236e9b262645a323e9088dfa3bb1ecb16cc75811daf40a23a824d665be9",
            "name": "docker.io/library/debian:12.2-slim"
          },
          "start_time": "2024-03-06T11:01:56Z",
          "pid": 3653,
          "maybe_exec_probe": false
        },
        "pod_labels": {},
        "workload": "test-pod-debian-cap",
        "workload_kind": "Pod"
      },
      "docker": "635bb04d5b6d67585f78358b6d54d03",
      "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjI5NDI3MTQxMDI3NDQ4NDg6MjQ4MjUzNg==",
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
          "CAP_SYS_CHROOT",
          "CAP_SYS_ADMIN",
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
          "CAP_SYS_CHROOT",
          "CAP_SYS_ADMIN",
          "CAP_SETFCAP"
        ],
        "inheritable": []
      },
      "ns": {
        "uts": {
          "inum": 4026534057,
          "is_host": false
        },
        "ipc": {
          "inum": 4026534058,
          "is_host": false
        },
        "mnt": {
          "inum": 4026535084,
          "is_host": false
        },
        "pid": {
          "inum": 4026535085,
          "is_host": false
        },
        "pid_for_children": {
          "inum": 4026535085,
          "is_host": false
        },
        "net": {
          "inum": 4026534059,
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
          "inum": 4026535086,
          "is_host": false
        },
        "user": {
          "inum": 4026531837,
          "is_host": true
        }
      },
      "tid": 2482536,
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
    "function_name": "security_file_permission",
    "args": [
      {
        "file_arg": {
          "mount": "",
          "path": "/run/secrets/kubernetes.io/serviceaccount/..2024_03_18_12_27_42.2985670933/token",
          "flags": ""
        },
        "label": ""
      },
      {
        "int_arg": 4,
        "label": ""
      }
    ],
    "return": {
      "int_arg": 0,
      "label": ""
    },
    "action": "KPROBE_ACTION_POST",
    "stack_trace": [],
    "policy_name": "file-monitoring"
  },
  "node_name": "experts-k8s-cs",
  "time": "2024-03-18T12:39:20.259331822Z",
  "aggregation_info": null
}
*/
