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
	ID          = "CS_RT_HACK_TOOLS"
	Name        = "Hacking tools"
	Description = "The detector detects startups of tools typically associated with malicious activities, such as automated exploitation of known vulnerabilities, brute-force attacks, reconnaissance, and information collection."
	Version     = 2
	Author      = "Runtime Radar Team"
	License     = "Apache License 2.0"
)

const (
	ExecAccountDiscoveryToolsNoArgs     = "Detected that the account discovery tool `%s` was started"
	ExecAccountDiscoveryToolsDefault    = "Detected that the account discovery tool `%s` was started using the `%s` arguments"
	ExecBruteforceToolsNoArgs           = "Detected that the bruteforcing tool `%s` was started"
	ExecBruteforceToolsDefault          = "Detected that the bruteforcing tool `%s` was started using the `%s` arguments"
	ExecContainersToolsNoArgs           = "Detected that the `%s` tool for exploiting the container infrastructure was started"
	ExecContainersToolsDefault          = "Detected that the `%s` tool for exploiting the container infrastructure was started using the `%s` arguments"
	ExecDataRemovalToolsNoArgs          = "Detected that the data removal tool `%s` was started"
	ExecDataRemovalToolsDefault         = "Detected that the data removal tool `%s` was started using the `%s` arguments"
	ExecDnsToolsNoArgs                  = "Detected that the DNS management tool `%s` was started"
	ExecDnsToolsDefault                 = "Detected that the DNS management tool `%s` was started using the `%s` arguments"
	ExecFilelessExecToolsNoArgs         = "Detected that the `%s` tool for fileless process execution was started"
	ExecFilelessExecToolsDefault        = "Detected that the `%s` tool for fileless process execution was started using the `%s` arguments"
	ExecLdapToolsNoArgs                 = "Detected that the LDAP management tool `%s` was started"
	ExecLdapToolsDefault                = "Detected that the LDAP management tool `%s` was started using the `%s` arguments"
	ExecMitmToolsNoArgs                 = "Detected that the MITM attack tool `%s` was started"
	ExecMitmToolsDefault                = "Detected that the MITM attack tool `%s` was started using the `%s` arguments"
	ExecNetcatLikeToolsNoArgs           = "Detected that the netcat tool `%s` or a similar tool was started"
	ExecNetcatLikeToolsDefault          = "Detected that the netcat tool `%s` or a similar tool was started using the `%s` arguments"
	ExecNetworkScannersNoArgs           = "Detected that the network scanner `%s` was started"
	ExecNetworkScannersDefault          = "Detected that the network scanner `%s` was started using the `%s` arguments"
	ExecNetworkTrafficToolsNoArgs       = "Detected that the network traffic management tool `%s` was started"
	ExecNetworkTrafficToolsDefault      = "Detected that the network traffic management tool `%s` was started using the `%s` arguments"
	ExecPentestToolsNoArgs              = "Detected that the pentesting tool `%s` was started"
	ExecPentestToolsDefault             = "Detected that the pentesting tool `%s` was started using the `%s` arguments"
	ExecPostExploitationToolsNoArgs     = "Detected that the postexploitation tool `%s` was started"
	ExecPostExploitationToolsDefault    = "Detected that the postexploitation tool `%s` was started using the `%s` arguments"
	ExecPrivilegeEscalationToolsNoArgs  = "Detected that the privilege escalation tool `%s` was started"
	ExecPrivilegeEscalationToolsDefault = "Detected that the privilege escalation tool `%s` was started using the `%s` arguments"
	ExecProxyToolsNoArgs                = "Detected that the `%s` tool for proxying and tunneling connections was started"
	ExecProxyToolsDefault               = "Detected that the `%s` tool for proxying and tunneling connections was started using the `%s` arguments"
	ExecSocialEngineeringToolsNoArgs    = "Detected that the `%s` tool for social engineering was started"
	ExecSocialEngineeringToolsDefault   = "Detected that the `%s` tool for social engineering was started using the `%s` arguments"
	ExecSqlToolsNoArgs                  = "Detected that the SQL management tool `%s` was started"
	ExecSqlToolsDefault                 = "Detected that the SQL management tool `%s` was started using the `%s` arguments"
	ExecSystemInfoDiscoveryToolsNoArgs  = "Detected that the system discovery tool `%s` was started"
	ExecSystemInfoDiscoveryToolsDefault = "Detected that the system discovery tool `%s` was started using the `%s` arguments"
	ExecWebToolsNoArgs                  = "Detected that the web application management tool `%s` was started"
	ExecWebToolsDefault                 = "Detected that the web application management tool `%s` was started using the `%s` arguments"
	ExecWifiToolsNoArgs                 = "Detected that the Wi-Fi management tool `%s` was started"
	ExecWifiToolsDefault                = "Detected that the Wi-Fi management tool `%s` was started using the `%s` arguments"
	ExecWindowsToolsNoArgs              = "Detected that the Windows management tool `%s` was started"
	ExecWindowsToolsDefault             = "Detected that the Windows management tool `%s` was started using the `%s` arguments"
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
	mitreTactics = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1609",
				"T1053.007",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1098.006",
				"T1543",
				"T1543.005",
				"T1053.007",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1548.001",
				"T1098.006",
				"T1543",
				"T1543.005",
				"T1611",
				"T1068",
				"T1053.007",
			},
		},
		{
			Id: "TA0005",
			Techniques: []string{
				"T1548.001",
				"T1036.005",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1552.007",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1613",
				"T1069.002",
				"T1016.001",
				"T1049",
				"T1033",
			},
		},
		{
			Id: "TA0040",
			Techniques: []string{
				"T1485",
			},
		},
	}

	mitreTacticsContainers = []*api.MitreTactic{
		{
			Id: "TA0002",
			Techniques: []string{
				"T1609",
				"T1053.007",
			},
		},
		{
			Id: "TA0003",
			Techniques: []string{
				"T1098.006",
				"T1543.005",
				"T1053.007",
			},
		},
		{
			Id: "TA0004",
			Techniques: []string{
				"T1098.006",
				"T1543.005",
				"T1611",
				"T1068",
				"T1053.007",
			},
		},
		{
			Id: "TA0006",
			Techniques: []string{
				"T1552.007",
			},
		},
		{
			Id: "TA0007",
			Techniques: []string{
				"T1613",
			},
		},
	}

	mitreTacticsLdap = []*api.MitreTactic{
		{
			Id: "TA0007",
			Techniques: []string{
				"T1069.002",
			},
		},
	}

	mitreTacticsDataRemoval = []*api.MitreTactic{
		{
			Id: "TA0040",
			Techniques: []string{
				"T1485",
			},
		},
	}
)

type hackTool struct {
	pattern glob.Glob
	url     string
	class   string
}

var (
	hackTools = []hackTool{
		// Account discovery tools
		{pattern: glob.MustCompile("*/hashcat"), url: "https://hashcat.net/hashcat/", class: "AccountDiscovery"},
		{pattern: glob.MustCompile("*/patator*"), url: "https://github.com/lanjelot/patator", class: "AccountDiscovery"},
		{pattern: glob.MustCompile("*/unshadow"), url: "https://manpages.ubuntu.com/manpages/noble/man8/unshadow.8.html", class: "AccountDiscovery"},

		// Bruteforce tools
		{pattern: glob.MustCompile("*/enumiax"), url: "https://github.com/foreni-packages/enumiax", class: "Bruteforce"},
		{pattern: glob.MustCompile("*/hydra"), url: "https://github.com/vanhauser-thc/thc-hydra", class: "Bruteforce"},
		{pattern: glob.MustCompile("*/sshprank"), url: "https://github.com/noptrix/sshprank", class: "Bruteforce"},

		// Containers tools
		{pattern: glob.MustCompile("*/cdk"), url: "https://github.com/cdk-team/CDK", class: "Containers"},
		{pattern: glob.MustCompile("*/deepce*"), url: "https://github.com/stealthcopter/deepce", class: "Containers"},
		{pattern: glob.MustCompile("*/k8spider*"), url: "https://github.com/Esonhugh/k8spider", class: "Containers"},
		{pattern: glob.MustCompile("*/kdigger*"), url: "https://github.com/quarkslab/kdigger", class: "Containers"},
		{pattern: glob.MustCompile("*/peirates"), url: "https://github.com/inguardians/peirates", class: "Containers"},

		// Data removal tools
		{pattern: glob.MustCompile("*/vanish*"), url: "https://dwukn.vercel.app/projects/vanish", class: "DataRemoval"},
		{pattern: glob.MustCompile("*/vanish[0-9]*"), url: "https://dwukn.vercel.app/projects/vanish", class: "DataRemoval"},
		{pattern: glob.MustCompile("*/vx"), url: "https://dwukn.vercel.app/projects/vanish", class: "DataRemoval"},

		// DNS tools
		{pattern: glob.MustCompile("*/dnsenum"), url: "https://github.com/SparrowOchon/dnsenum2", class: "DNS"},
		{pattern: glob.MustCompile("*/dnsmap"), url: "https://github.com/resurrecting-open-source-projects/dnsmap", class: "DNS"},
		{pattern: glob.MustCompile("*/dnsrecon"), url: "https://github.com/darkoperator/dnsrecon", class: "DNS"},
		{pattern: glob.MustCompile("*/sonar*"), url: "https://github.com/jrozner/sonar", class: "DNS"},

		// Fileless execution tools
		{pattern: glob.MustCompile("*/fee"), url: "https://pypi.org/project/fee/", class: "FilelessExecution"},

		// LDAP tools
		{pattern: glob.MustCompile("*/ldapper"), url: "https://github.com/Synzack/ldapper", class: "LDAP"},

		// MITM tools
		{pattern: glob.MustCompile("*/arp.spoof"), url: "https://github.com/hackerschoice/dsniff", class: "MITM"},
		{pattern: glob.MustCompile("*/bdfproxy"), url: "https://github.com/secretsquirrel/BDFProxy", class: "MITM"},
		{pattern: glob.MustCompile("*/bettercap"), url: "https://github.com/bettercap/bettercap", class: "MITM"},
		{pattern: glob.MustCompile("*/ettercap"), url: "https://www.ettercap-project.org", class: "MITM"},
		{pattern: glob.MustCompile("*/evilginx"), url: "https://github.com/kgretzky/evilginx2", class: "MITM"},
		{pattern: glob.MustCompile("*/net.sniff"), url: "", class: "MITM"},
		{pattern: glob.MustCompile("*/responder*"), url: "", class: "MITM"},
		{pattern: glob.MustCompile("*/ticker.commands"), url: "", class: "MITM"},
		{pattern: glob.MustCompile("*/ticker.period"), url: "", class: "MITM"},
		{pattern: glob.MustCompile("*/wifi.recon"), url: "", class: "MITM"},

		// Netcat-like tools
		{pattern: glob.MustCompile("*/nc"), url: "", class: "NetcatLike"},
		{pattern: glob.MustCompile("*/ncat"), url: "", class: "NetcatLike"},
		{pattern: glob.MustCompile("*/nc.traditional"), url: "", class: "NetcatLike"},
		{pattern: glob.MustCompile("*/netcat"), url: "", class: "NetcatLike"},
		{pattern: glob.MustCompile("*/pwncat"), url: "", class: "NetcatLike"},

		// Network scanners
		{pattern: glob.MustCompile("*/[an]map"), url: "", class: "NetworkScanners"},
		{pattern: glob.MustCompile("*/zenmap"), url: "", class: "NetworkScanners"},
		{pattern: glob.MustCompile("*/masscan"), url: "", class: "NetworkScanners"},
		{pattern: glob.MustCompile("*/fscan"), url: "", class: "NetworkScanners"},

		// Network traffic tools
		{pattern: glob.MustCompile("*/hping*"), url: "", class: "NetworkTraffic"},
		{pattern: glob.MustCompile("*/socat"), url: "", class: "NetworkTraffic"},
		{pattern: glob.MustCompile("*/tcpdump"), url: "", class: "NetworkTraffic"},
		{pattern: glob.MustCompile("*/wireshark"), url: "", class: "NetworkTraffic"},
		{pattern: glob.MustCompile("*/xplico"), url: "", class: "NetworkTraffic"},

		// Pentest tools
		{pattern: glob.MustCompile("*/msfconsole"), url: "", class: "Pentest"},
		{pattern: glob.MustCompile("*/msfpc"), url: "", class: "Pentest"},
		{pattern: glob.MustCompile("*/msfvenom"), url: "", class: "Pentest"},
		{pattern: glob.MustCompile("*/pypykatz*"), url: "", class: "Pentest"},

		// Post exploitation tools
		{pattern: glob.MustCompile("*/empire"), url: "", class: "PostExploitation"},
		{pattern: glob.MustCompile("*/ghost"), url: "", class: "PostExploitation"},

		// Privilege escalation tools
		{pattern: glob.MustCompile("*/beroot"), url: "", class: "PrivilegeEscalation"},
		{pattern: glob.MustCompile("*/linpeas*"), url: "", class: "PrivilegeEscalation"},
		{pattern: glob.MustCompile("*/linprivchecker"), url: "", class: "PrivilegeEscalation"},
		{pattern: glob.MustCompile("*/linuxprivchecker"), url: "", class: "PrivilegeEscalation"},
		{pattern: glob.MustCompile("*/privesc"), url: "", class: "PrivilegeEscalation"},
		{pattern: glob.MustCompile("*/traitor"), url: "", class: "PrivilegeEscalation"},

		// Proxies
		{pattern: glob.MustCompile("*/3proxy"), url: "", class: "Proxy"},
		// ligolo-ng
		{pattern: glob.MustCompile("*/agent"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/chisel"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/gost"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/graftcp-local"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/mgraftcp"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/pivotnacci"), url: "", class: "Proxy"},
		// ligolo-ng
		{pattern: glob.MustCompile("*/proxy"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/proxychains"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/proxychains[0-9]*"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/proxytunnel"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/regeorg"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/ssf"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/ssfd"), url: "", class: "Proxy"},
		{pattern: glob.MustCompile("*/sshuttle"), url: "", class: "Proxy"},

		// Social engineering tools
		{pattern: glob.MustCompile("*/setoolkit"), url: "https://github.com/trustedsec/social-engineer-toolkit", class: "SocialEngineering"},

		// SQL tools
		{pattern: glob.MustCompile("*/sqlmap"), url: "https://sqlmap.org", class: "SQL"},

		// System info discovery tools
		{pattern: glob.MustCompile("*/enum4linux"), url: "https://labs.portcullis.co.uk/tools/enum4linux/", class: "SystemInfoDiscovery"},
		{pattern: glob.MustCompile("*/lynis"), url: "https://github.com/CISOfy/Lynis", class: "SystemInfoDiscovery"},
		{pattern: glob.MustCompile("*/mxtract"), url: "https://github.com/rek7/mXtract", class: "SystemInfoDiscovery"},
		{pattern: glob.MustCompile("*/volatility*"), url: "https://github.com/volatilityfoundation/volatility", class: "SystemInfoDiscovery"},

		// Web tools
		{pattern: glob.MustCompile("*/beef"), url: "https://github.com/beefproject/beef", class: "Web"},
		{pattern: glob.MustCompile("*/commix"), url: "https://github.com/commixproject/commix", class: "Web"},
		{pattern: glob.MustCompile("*/dirb"), url: "https://dirb.sourceforge.net", class: "Web"},
		{pattern: glob.MustCompile("*/dirbuster"), url: "https://www.kali.org/tools/dirbuster/", class: "Web"},
		{pattern: glob.MustCompile("*/dirsearch"), url: "", class: "Web"},
		{pattern: glob.MustCompile("*/gobuster"), url: "", class: "Web"},
		{pattern: glob.MustCompile("*/nikto*"), url: "", class: "Web"},
		{pattern: glob.MustCompile("*/openvas"), url: "", class: "Web"},
		{pattern: glob.MustCompile("*/phpsploit"), url: "", class: "Web"},
		{pattern: glob.MustCompile("*/skipfish"), url: "", class: "Web"},
		{pattern: glob.MustCompile("*/wpscan"), url: "", class: "Web"},

		// Wi-Fi tools
		{pattern: glob.MustCompile("*/aircrack-ng"), url: "", class: "Wi-Fi"},
		{pattern: glob.MustCompile("*/airmon-ng"), url: "", class: "Wi-Fi"},
		{pattern: glob.MustCompile("*/airodump-ng"), url: "", class: "Wi-Fi"},
		{pattern: glob.MustCompile("*/airolib-ng"), url: "", class: "Wi-Fi"},
		{pattern: glob.MustCompile("*/airgeddon"), url: "", class: "Wi-Fi"},
		{pattern: glob.MustCompile("*/rsf"), url: "", class: "Wi-Fi"},

		// Windows tools
		{pattern: glob.MustCompile("*/atexec"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/dcomexec"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/esentutl"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/getadusers"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/getarch"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/getnpusers"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/getosandsmbproperties"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/getuserspns"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/lookupsid"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/mmcexec"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/ntfs-read"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/ntlmrelayx"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/samrdump"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/secretsdump"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/smbclient"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/smbexec"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/smbrelayx"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/smbserver"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/ticketer"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/wmiexec"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/wmipersist"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/wmiquery"), url: "https://github.com/fortra/impacket", class: "Windows"},
		{pattern: glob.MustCompile("*/nbtscan"), url: "", class: "Windows"},
		{pattern: glob.MustCompile("*/spraykatz"), url: "", class: "Windows"},
	}
)

// init is required for TinyGo to compile to Wasm.
func init() {
	api.RegisterDetector(Detector{})
}

type Detector struct{}

func (d Detector) Info(ctx context.Context, req *api.InfoReq) (*api.InfoResp, error) {
	return &api.InfoResp{
		Id:             ID,
		Name:           Name,
		Description:    Description,
		Version:        Version,
		Author:         Author,
		License:        License,
		TacticsCovered: mitreTactics,
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

		for _, tool := range hackTools {
			if !tool.pattern.Match(binary) {
				continue
			}
			switch {
			case (tool.class == "AccountDiscovery") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecAccountDiscoveryToolsNoArgs, binary)
			case (tool.class == "AccountDiscovery") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecAccountDiscoveryToolsDefault, binary, args)
			case (tool.class == "Bruteforce") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecBruteforceToolsNoArgs, binary)
			case (tool.class == "Bruteforce") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecBruteforceToolsDefault, binary, args)
			case (tool.class == "Containers") && (args == ""):
				resp.TacticsCovered = mitreTacticsContainers
				resp.Reason = fmt.Sprintf(ExecContainersToolsNoArgs, binary)
			case (tool.class == "Containers") && (args != ""):
				resp.TacticsCovered = mitreTacticsContainers
				resp.Reason = fmt.Sprintf(ExecContainersToolsDefault, binary, args)
			case (tool.class == "DataRemoval") && (args == ""):
				resp.TacticsCovered = mitreTacticsDataRemoval
				resp.Reason = fmt.Sprintf(ExecDataRemovalToolsNoArgs, binary)
			case (tool.class == "DataRemoval") && (args != ""):
				resp.TacticsCovered = mitreTacticsDataRemoval
				resp.Reason = fmt.Sprintf(ExecDataRemovalToolsDefault, binary, args)
			case (tool.class == "DNS") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecDnsToolsNoArgs, binary)
			case (tool.class == "DNS") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecDnsToolsDefault, binary, args)
			case (tool.class == "FilelessExecution") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecFilelessExecToolsNoArgs, binary)
			case (tool.class == "FilelessExecution") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecFilelessExecToolsDefault, binary, args)
			case (tool.class == "LDAP") && (args == ""):
				resp.TacticsCovered = mitreTacticsLdap
				resp.Reason = fmt.Sprintf(ExecLdapToolsNoArgs, binary)
			case (tool.class == "LDAP") && (args != ""):
				resp.TacticsCovered = mitreTacticsLdap
				resp.Reason = fmt.Sprintf(ExecLdapToolsDefault, binary, args)
			case (tool.class == "MITM") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecMitmToolsNoArgs, binary)
			case (tool.class == "MITM") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecMitmToolsDefault, binary, args)
			case (tool.class == "NetcatLike") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecNetcatLikeToolsNoArgs, binary)
			case (tool.class == "NetcatLike") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecNetcatLikeToolsDefault, binary, args)
			case (tool.class == "NetworkScanners") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecNetworkScannersNoArgs, binary)
			case (tool.class == "NetworkScanners") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecNetworkScannersDefault, binary, args)
			case (tool.class == "NetworkTraffic") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecNetworkTrafficToolsNoArgs, binary)
			case (tool.class == "NetworkTraffic") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecNetworkTrafficToolsDefault, binary, args)
			case (tool.class == "Pentest") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecPentestToolsNoArgs, binary)
			case (tool.class == "Pentest") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecPentestToolsDefault, binary, args)
			case (tool.class == "PostExploitation") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecPostExploitationToolsNoArgs, binary)
			case (tool.class == "PostExploitation") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecPostExploitationToolsDefault, binary, args)
			case (tool.class == "PrivilegeEscalation") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecPrivilegeEscalationToolsNoArgs, binary)
			case (tool.class == "PrivilegeEscalation") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecPrivilegeEscalationToolsDefault, binary, args)
			case (tool.class == "Proxy") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecProxyToolsNoArgs, binary)
			case (tool.class == "Proxy") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecProxyToolsDefault, binary, args)
			case (tool.class == "SocialEngineering") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSocialEngineeringToolsNoArgs, binary)
			case (tool.class == "SocialEngineering") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSocialEngineeringToolsDefault, binary, args)
			case (tool.class == "SQL") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSqlToolsNoArgs, binary)
			case (tool.class == "SQL") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSqlToolsDefault, binary, args)
			case (tool.class == "SystemInfoDiscovery") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSystemInfoDiscoveryToolsNoArgs, binary)
			case (tool.class == "SystemInfoDiscovery") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecSystemInfoDiscoveryToolsDefault, binary, args)
			case (tool.class == "Web") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecWebToolsNoArgs, binary)
			case (tool.class == "Web") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecWebToolsDefault, binary, args)
			case (tool.class == "Wi-Fi") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecWifiToolsNoArgs, binary)
			case (tool.class == "Wi-Fi") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecWifiToolsDefault, binary, args)
			case (tool.class == "Windows") && (args == ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecWindowsToolsNoArgs, binary)
			case (tool.class == "Windows") && (args != ""):
				resp.TacticsCovered = mitreTactics
				resp.Reason = fmt.Sprintf(ExecWindowsToolsDefault, binary, args)
			}
			resp.Severity = api.DetectResp_CRITICAL // <-- threat detected

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

// tacticTechniques is a constructor for *api.MitreTactic which makes its initialization less verbose.
func tacticTechniques(tacticID string, techniqueIDs ...string) *api.MitreTactic {
	return &api.MitreTactic{
		Id:         tacticID,
		Techniques: techniqueIDs,
	}
}
