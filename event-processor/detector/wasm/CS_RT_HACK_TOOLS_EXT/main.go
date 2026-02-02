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
	ID          = "CS_RT_HACK_TOOLS_EXT"
	Name        = "Hacking tools (extended version)"
	Description = "The detector detects if any known hacking tools were started, such as nmap, netcat, and peirates. The detector may be triggered simultaneously with CS_RT_HACK_TOOLS. They may be merged later."
	Version     = 2
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
	hackTools = []*regexp.Regexp{
		// sudo and su are rarely used in runtime, containers are either rooted or rootless with no need to escalate privileges for a normal user.
		regexp.MustCompile(`/su(do)?$`),

		// These tools are supposed to be mostly malicious in general or in the context of containers and k8s environment.
		// network scan
		regexp.MustCompile(`/(a|n|zen)map$`),
		regexp.MustCompile(`/masscan$`),
		regexp.MustCompile(`/fscan$`),

		// proxies
		regexp.MustCompile(`/3proxy$`),
		regexp.MustCompile(`/agent$`), // ligolo-ng
		regexp.MustCompile(`/chisel$`),
		regexp.MustCompile(`/graftcp-local$`),
		regexp.MustCompile(`/mgraftcp$`),
		regexp.MustCompile(`/pivotnacci$`),
		regexp.MustCompile(`/proxy$`), // ligolo-ng
		regexp.MustCompile(`/proxychains[0-9]*$`),
		regexp.MustCompile(`/proxytunnel$`),
		regexp.MustCompile(`/regeorg$`),
		regexp.MustCompile(`/ssfd?$`),
		regexp.MustCompile(`/sshuttle$`),

		// remote fileless execution
		regexp.MustCompile(`/fee$`),

		// bruteforce
		regexp.MustCompile(`/enumiax$`),
		regexp.MustCompile(`/hydra$`),

		// system information discovery
		regexp.MustCompile(`/enum4linux$`),
		regexp.MustCompile(`/lynis$`),
		regexp.MustCompile(`/mxtract$`),
		regexp.MustCompile(`/volatility\S*$`),

		// wi-fi
		regexp.MustCompile(`/air(crack|mon|odump|olib)-ng$`),
		regexp.MustCompile(`/airgeddon$`),
		regexp.MustCompile(`/rsf$`),

		// mitm
		regexp.MustCompile(`/arp\.spoof$`),
		regexp.MustCompile(`/bdfproxy$`),
		regexp.MustCompile(`/bettercap$`),
		regexp.MustCompile(`/ettercap$`),
		regexp.MustCompile(`/evilginx$`),
		regexp.MustCompile(`/net\.sniff$`),
		regexp.MustCompile(`/responder\S*$`),
		regexp.MustCompile(`/ticker\.(commands|period)$`),
		regexp.MustCompile(`/wifi\.recon$`),

		// pentest utils
		regexp.MustCompile(`/msf(console|pc|venom)$`),
		regexp.MustCompile(`/pypykatz\S*$`),

		// web
		regexp.MustCompile(`/beef$`),
		regexp.MustCompile(`/commix$`),
		regexp.MustCompile(`/dir(b|buster|search)$`),
		regexp.MustCompile(`/gobuster$`),
		regexp.MustCompile(`/nikto\S*$`),
		regexp.MustCompile(`/openvas$`),
		regexp.MustCompile(`/phpsploit$`),
		regexp.MustCompile(`/skipfish$`),
		regexp.MustCompile(`/wpscan$`),

		// sql
		regexp.MustCompile(`/sqlmap$`),

		// accounts
		regexp.MustCompile(`/hashcat$`),
		regexp.MustCompile(`/patator\S*$`),
		regexp.MustCompile(`/unshadow$`),

		// social engineering
		regexp.MustCompile(`/setoolkit$`),

		// post exploitation
		regexp.MustCompile(`/empire$`),
		regexp.MustCompile(`/ghost$`),

		// pcap analysis
		regexp.MustCompile(`/xplico$`),

		// ssh
		regexp.MustCompile(`/sshprank$`),

		// windows
		regexp.MustCompile(`/spraykatz$`),

		// netcat
		regexp.MustCompile(`/nc(at|\.traditional)?$`),
		regexp.MustCompile(`/netcat$`),
		regexp.MustCompile(`/pwncat$`),

		// ldap
		regexp.MustCompile(`/ldapper$`),

		// linux privesc
		regexp.MustCompile(`/beroot`),
		regexp.MustCompile(`/linpeas\S*$`),
		regexp.MustCompile(`/lin(ux)?privchecker$`),
		regexp.MustCompile(`/privesc$`),
		regexp.MustCompile(`/traitor$`),

		// lateral movement
		regexp.MustCompile(`/atexec$`),
		regexp.MustCompile(`/dcomexec$`),
		regexp.MustCompile(`/esentutl$`),
		regexp.MustCompile(`/get(adusers|arch|npusers|osandsmbproperties|userspns)$`),
		regexp.MustCompile(`/lookupsid$`),
		regexp.MustCompile(`/mmcexec$`),
		regexp.MustCompile(`/ntfs-read$`),
		regexp.MustCompile(`/ntlmrelayx$`),
		regexp.MustCompile(`/samrdump$`),
		regexp.MustCompile(`/secretsdump$`),
		regexp.MustCompile(`/smb(client|exec|relayx|server)$`),
		regexp.MustCompile(`/ticketer$`),
		regexp.MustCompile(`/wmi(exec|persist|query)$`),

		// persistence
		regexp.MustCompile(`/nohup$`),

		// netbios
		regexp.MustCompile(`/nbtscan$`),

		// dns
		regexp.MustCompile(`/dns(enum|map|recon)$`),

		// network traffic dumping
		regexp.MustCompile(`/tcpdump$`),
		regexp.MustCompile(`/wireshark$`),

		// network packets manipulation
		regexp.MustCompile(`/hping\S*$`),
		regexp.MustCompile(`/socat$`),

		// virus tools
		regexp.MustCompile(`/vanish\S*$`),

		// k8s and container runtimes
		regexp.MustCompile(`/kubectl$`),
		regexp.MustCompile(`/nsenter$`),
		regexp.MustCompile(`/peirates$`),
		regexp.MustCompile(`/unshare$`),
		regexp.MustCompile(`/cdk$`),
		regexp.MustCompile(`/deepce\S*$`),
	}

	// suspDirs contains directories that normally do not contain binaries.
	suspDirs = []glob.Glob{
		glob.MustCompile("/home/*"),
		glob.MustCompile("/tmp/*"),
		glob.MustCompile("/var/*"),
		glob.MustCompile("/boot/*"),
		glob.MustCompile("/media/*"),
		glob.MustCompile("/mnt/*"),
		glob.MustCompile("/srv/*"),
		glob.MustCompile("/sys/*"),
		glob.MustCompile("/dev/*"),
		glob.MustCompile("/run/*"),
		glob.MustCompile("/sys/*"),
	}

	runcInitBinPattern = glob.MustCompile("/dev/fd/*")

	nohupBin = glob.MustCompile(`*/nohup`)
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
		parentBinary := exec.GetParent().GetBinary()

		args := exec.GetProcess().GetArguments()
		parentArgs := exec.GetParent().GetArguments()

		// exclude runc init container
		if runcInitBinPattern.Match(binary) && args == "init" && parentBinary == "/proc/self/exe" && parentArgs == "init" {
			return resp, nil
		}

		// trigger on explicit hacker utility name
		for _, ht := range hackTools {
			if ht.MatchString(binary) {
				resp.Severity = api.DetectResp_CRITICAL // <-- threat detected
				return resp, nil
			}
		}

		// trigger on binary with setuid bit
		if suid := exec.GetProcess().GetBinaryProperties().GetSetuid(); suid != nil {
			if suid.Value == rootUID {
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected
				return resp, nil
			}
		}

		// trigger on binary with setgid bit
		if sgid := exec.GetProcess().GetBinaryProperties().GetSetgid(); sgid != nil {
			if sgid.Value == rootGID {
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected
				return resp, nil
			}
		}

		// trigger on process from suspicious directories
		for _, dir := range suspDirs {
			if dir.Match(binary) {
				resp.Severity = api.DetectResp_MEDIUM // <-- threat detected
				return resp, nil
			}
		}

		// trigger on binary if its parent binary is nohup
		if nohupBin.Match(parentBinary) {
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODcxMzAxODE0NzIwMTU6MTcxMzg1MA==",
            "pid": 1713850,
            "uid": 0,
            "cwd": "/",
            "binary": "/tmp/linpeas.sh",
            "arguments": "/tmp/linpeas.sh",
            "flags": "execve rootcwd clone",
            "start_time": "2024-03-22T12:19:16.974922315Z",
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
                    "pid": 60345,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODY1NjA1MDI5Mzg4NDQ6MTcwODI4Mg==",
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
            "tid": 1713850,
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
            "exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODY1NjA1MDI5Mzg4NDQ6MTcwODI4Mg==",
            "pid": 1708282,
            "uid": 0,
            "cwd": "/",
            "binary": "/usr/bin/bash",
            "arguments": "",
            "flags": "execve rootcwd",
            "start_time": "2024-03-22T12:09:47.296389046Z",
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
                    "pid": 60319,
                    "maybe_exec_probe": false
                },
                "pod_labels": {},
                "workload": "test-pod-debian",
                "workload_kind": "Pod"
            },
            "docker": "332cff0fb99f03f8b4fb9633f245ba2",
            "parent_exec_id": "cHRleHBlcnRzLWs4cy1wdGNzOjMyODY1NjA1MDE4MjQ1Mjc6MTcwODI4Mg==",
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
            "tid": 1708282,
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
        "ancestors": []
    },
    "node_name": "experts-k8s-cs",
    "time": "2024-03-22T12:19:16.974922129Z",
    "aggregation_info": null
}

*/
