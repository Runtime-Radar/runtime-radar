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
	ID          = "CS_RT_CVE_2025_1974"
	Name        = "Ingress Nightmare vulnerability (malicious code execution)"
	Description = "The detector detects if a shared library was loaded from the directory /tmp/nginx/client-body/, which may indicate that malicious code was loaded and executed to exploit the Ingress Nightmare vulnerability (CVE-2025-1974)."
	Version     = 1
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	// Memory page access permissions.
	// https://elixir.bootlin.com/linux/v6.10-rc6/source/include/uapi/asm-generic/mman-common.h#L10
	PROT_READ = 1
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"security_mmap_file"},

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
	// NGINX binary.
	nginxBinary = glob.MustCompile("*/nginx")

	// Path to request body buffered by NGINX.
	libPath = glob.MustCompile("/tmp/nginx/client-body/*")
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

		binary := kprobe.GetProcess().GetBinary()
		function := kprobe.GetFunctionName()
		args := kprobe.GetArgs()

		path := ""

		switch function {
		// Trigger when security function check for memory page write access.
		// https://tetragon.io/docs/use-cases/filename-access/
		case "security_mmap_file":
			if len(args) < 2 {
				return nil, fmt.Errorf("unexpected args len, got %d, want >= 2", len(args))
			} else if prot := args[1].GetUintArg(); prot&PROT_READ == 0 {
				return resp, nil
			}

		default:
			return resp, nil
		}

		path = args[0].GetFileArg().GetPath()

		if nginxBinary.Match(binary) && libPath.Match(path) {
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
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZToyMDQ2MjI5Mjk5MjY5MzY5NzoyMDgyNTM2",
            "pid": 2082536,
            "uid": 101,
            "cwd": "/etc/nginx",
            "binary": "/usr/bin/nginx",
            "arguments": "-c /tmp/nginx/nginx-cfg1787549034 -t",
            "flags": "execve clone",
            "start_time": "2025-04-29T11:27:19.395558820Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "ingress-nginx",
                "name": "ingress-nginx-controller-56748c4f7c-nfsq8",
                "container": {
                    "id": "containerd://60c397c08e0c4dc29b6623f864451415e6b9999cac34608572514e94f514a98b",
                    "name": "controller",
                    "image": {
                        "id": "registry.k8s.io/ingress-nginx/controller@sha256:9724476b928967173d501040631b23ba07f47073999e80e34b120e8db5f234d5",
                        "name": "sha256:11b916a025f028a5868d51d66773e876910636a1c261e919d72864dbb9bfc860"
                    },
                    "start_time": "2025-04-29T11:12:32Z",
                    "pid": 730,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app.kubernetes.io/component": "controller",
                    "app.kubernetes.io/instance": "ingress-nginx",
                    "app.kubernetes.io/name": "ingress-nginx",
                    "app.kubernetes.io/part-of": "ingress-nginx",
                    "app.kubernetes.io/version": "1.12.0-beta.0",
                    "pod-template-hash": "56748c4f7c"
                },
                "workload": "ingress-nginx-controller",
                "workload_kind": "Deployment"
            },
            "docker": "60c397c08e0c4dc29b6623f86445141",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZToyMDQ2MTQwNjE3MDAyMDM3NToyMDU1MjUx",
            "refcnt": 1,
            "cap": {
                "permitted": [
                    "CAP_NET_BIND_SERVICE"
                ],
                "effective": [
                    "CAP_NET_BIND_SERVICE"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026533517,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533518,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533520,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533521,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533521,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533119,
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
                    "inum": 4026533522,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2082536,
            "process_credentials": {
                "uid": 101,
                "gid": 82,
                "euid": 101,
                "egid": 82,
                "suid": 101,
                "sgid": 82,
                "fsuid": 101,
                "fsgid": 82,
                "securebits": [],
                "caps": null,
                "user_ns": null
            },
            "binary_properties": null,
            "user": null
        },
        "parent": {
            "exec_id": "cHRjcy1tYXN0ZXItbm9kZToyMDQ2MTQwNjE3MDAyMDM3NToyMDU1MjUx",
            "pid": 2055251,
            "uid": 101,
            "cwd": "/etc/nginx",
            "binary": "/nginx-ingress-controller",
            "arguments": "--election-id=ingress-nginx-leader --controller-class=k8s.io/ingress-nginx --ingress-class=nginx --configmap=ingress-nginx/ingress-nginx-controller --validating-webhook=:8443 --validating-webhook-certificate=/usr/local/certificates/cert --validating-webhook-key=/usr/local/certificates/key",
            "flags": "execve clone",
            "start_time": "2025-04-29T11:12:32.572886816Z",
            "auid": 4294967295,
            "pod": {
                "namespace": "ingress-nginx",
                "name": "ingress-nginx-controller-56748c4f7c-nfsq8",
                "container": {
                    "id": "containerd://60c397c08e0c4dc29b6623f864451415e6b9999cac34608572514e94f514a98b",
                    "name": "controller",
                    "image": {
                        "id": "registry.k8s.io/ingress-nginx/controller@sha256:9724476b928967173d501040631b23ba07f47073999e80e34b120e8db5f234d5",
                        "name": "sha256:11b916a025f028a5868d51d66773e876910636a1c261e919d72864dbb9bfc860"
                    },
                    "start_time": "2025-04-29T11:12:32Z",
                    "pid": 7,
                    "maybe_exec_probe": false
                },
                "pod_labels": {
                    "app.kubernetes.io/component": "controller",
                    "app.kubernetes.io/instance": "ingress-nginx",
                    "app.kubernetes.io/name": "ingress-nginx",
                    "app.kubernetes.io/part-of": "ingress-nginx",
                    "app.kubernetes.io/version": "1.12.0-beta.0",
                    "pod-template-hash": "56748c4f7c"
                },
                "workload": "ingress-nginx-controller",
                "workload_kind": "Deployment"
            },
            "docker": "60c397c08e0c4dc29b6623f86445141",
            "parent_exec_id": "cHRjcy1tYXN0ZXItbm9kZToyMDQ2MTQwNjE2NDA5NDAwMToyMDU1MjI3",
            "refcnt": 0,
            "cap": {
                "permitted": [
                    "CAP_NET_BIND_SERVICE"
                ],
                "effective": [
                    "CAP_NET_BIND_SERVICE"
                ],
                "inheritable": []
            },
            "ns": {
                "uts": {
                    "inum": 4026533517,
                    "is_host": false
                },
                "ipc": {
                    "inum": 4026533518,
                    "is_host": false
                },
                "mnt": {
                    "inum": 4026533520,
                    "is_host": false
                },
                "pid": {
                    "inum": 4026533521,
                    "is_host": false
                },
                "pid_for_children": {
                    "inum": 4026533521,
                    "is_host": false
                },
                "net": {
                    "inum": 4026533119,
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
                    "inum": 4026533522,
                    "is_host": false
                },
                "user": {
                    "inum": 4026531837,
                    "is_host": true
                }
            },
            "tid": 2055251,
            "process_credentials": {
                "uid": 101,
                "gid": 82,
                "euid": 101,
                "egid": 82,
                "suid": 101,
                "sgid": 82,
                "fsuid": 101,
                "fsgid": 82,
                "securebits": [],
                "caps": null,
                "user_ns": null
            },
            "binary_properties": null,
            "user": null
        },
        "function_name": "security_mmap_file",
        "args": [
            {
                "file_arg": {
                    "mount": "",
                    "path": "/tmp/nginx/client-body/0000000002 (deleted)",
                    "flags": "",
                    "permission": "-rw-------"
                },
                "label": ""
            },
            {
                "uint_arg": 5,
                "label": ""
            },
            {
                "uint_arg": 2,
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
    "node_name": "cs-master-node",
    "time": "2025-04-29T11:27:19.484544474Z",
    "aggregation_info": null
}
*/
