package model

import (
	"database/sql/driver"
	_ "embed"
	"encoding/json"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/api"
)

const (
	ConfigVersion Version = "1"
)

var (
	//go:embed tracingpolicy/connect.yaml
	connect string

	//go:embed tracingpolicy/permissions.yaml
	permissions string

	//go:embed tracingpolicy/file-monitoring.yaml
	fileMonitoring string

	//go:embed tracingpolicy/ptrace.yaml
	ptrace string

	//go:embed tracingpolicy/mount.yaml
	mount string

	//go:embed tracingpolicy/kernel-modules.yaml
	kernelModules string

	//go:embed tracingpolicy/listen-socket.yaml
	listenSocket string

	//go:embed tracingpolicy/io-streams.yaml
	ioStreams string

	//go:embed tracingpolicy/io-uring.yml
	ioUring string

	//go:embed tracingpolicy/umh.yaml
	umh string

	//go:embed tracingpolicy/rootkit.yml
	rootkit string
)

var (
	DefaultConfig = &Config{
		Base: Base{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		Config: &ConfigJSON{
			Version: string(ConfigVersion),
			TracingPolicies: map[string]*api.TracingPolicy{
				"connect": {
					Name:        "Outgoing TCP connections",
					Description: "This source tracks the tcp_connect, tcp_close, and tcp_sendmsg functions allowing detection of outgoing TCP connections (including the connection establishment, termination, and sending of TCP packets). Enabling this policy can significantly increase the event flow and load on the system. In this case, you should narrow the flow using more accurate filters (for example, specify only certain pods).",
					Yaml:        connect,
					Enabled:     false,
				},
				"permissions": {
					Name:        "File and process access rights actions",
					Description: "The source monitors calls to the Linux kernel function `commit_creds()`, which may indicate an attacker's attempts to escalate process privileges, including gaining superuser (root) rights. Additionally, the source monitors calls to the LSM function `security_path_chmod()` with permission sets that include execute rights.",
					Yaml:        permissions,
					Enabled:     false,
				},
				"file-monitoring": {
					Name:        "Access to important system files",
					Description: "This source tracks the security_file_permission, security_mmap_file, and security_path_truncate calls to a number of files, such as /boot, /root/.ssh, /etc/shadow, /etc/profile, /etc/sudoers, /etc/pam.conf, and others. Some files and directories are tracked for read, others are tracked for write, and others are tracked in all cases. To examine the source in detail, switch to expert mode.",
					Yaml:        fileMonitoring,
					Enabled:     false,
				},
				"ptrace": {
					Name:        "Using tools for debugging and reverse engineering (ptrace)",
					Description: "This source tracks the ptrace system call, which may indicate attacker activity in the target system.",
					Yaml:        ptrace,
					Enabled:     false,
				},
				"mount": {
					Name:        "Device mounting",
					Description: "This source tracks the mount() call allowing detection of potentially unwanted events related to device mounting.",
					Yaml:        mount,
					Enabled:     false,
				},
				"kernel-modules": {
					Name:        "Loading and unloading of kernel modules",
					Description: "This source tracks the do_init_module, free_module, security_kernel_module_request, and security_kernel_read_file calls allowing detection of explicit or implicit (automatic) loading and unloading of modules, as well as attempts to manipulate modules and other malicious activity.",
					Yaml:        kernelModules,
					Enabled:     false,
				},
				"listen-socket": {
					Name:        "Opening of a socket for incoming connections",
					Description: "This source tracks the inet_csk_listen_start call revealing possible activity of unwanted networking tools. The source also detects legitimate activity in a container; therefore, in addition to opening of a socket, detectors must also consider other event parameters.",
					Yaml:        listenSocket,
					Enabled:     false,
				},
				"io-streams": {
					Name:        "Standard I/O stream actions",
					Description: "The source monitors calls to the Linux kernel function `do_dup2()`, which duplicates the standard input (STDIN) file descriptor, as well as the creation of named pipe files (S_IFIFO) via the LSM function `security_path_mknod()`. Such actions often indicate that an attacker is attempting to launch a reverse shell, a covert communication channel, or another attack tool.",
					Yaml:        ioStreams,
					Enabled:     false,
				},
				"io-uring": {
					Name:        "Monitoring of the io_uring interface",
					Description: "This source tracks calls to io_uring_setup and io_uring_enter, providing monitoring of the creation and usage of an io_uring interface instance.",
					Yaml:        ioUring,
					Enabled:     false,
				},
				"umh": {
					Name:        "Usermode helper API usage",
					Description: "The source monitors calls to the Linux kernel functions call_usermodehelper_setup() and call_usermodehelper_exec(), which indicate the setup and launch of processes via the usermode helper API. This helps detect several different attacks that abuse this interface to launch processes at the host OS level.",
					Yaml:        umh,
					Enabled:     false,
				},
				"rootkit": {
					Name:        "Rootkit loading monitoring",
					Description: "The source monitors calls to the Linux kernel function kallsyms_lookup_name() used to obtain the address of the system call table, which may indicate rootkit activity on the target system. Additionally, to detect rootkits that use eBPF, the source monitors the loading of eBPF programs via the bpf_check() function.",
					Yaml:        rootkit,
					Enabled:     false,
				},
			},
			AllowList: []*tetragon.Filter{
				{PodRegex: []string{"deathstar"}},
			},
			DenyList: []*tetragon.Filter{},
			// If empty instead of nil, aggregator will be started and will complain with "Aggregator buffer is full. Consider increasing AggregatorOptions.channel_buffer_size."
			// AggregationOptions: &tetragon.AggregationOptions{},
		},
	}
)

type ConfigJSON api.Config_ConfigJSON

type Config struct {
	Base
	Config *ConfigJSON `gorm:"type:jsonb"`
}

// TableName method implements Tabler interface and makes GORM name the table of Config "runtime_monitor_configs" instead of just "configs".
// This is done in order to keep more generic "configs" available for possible use by dynamic config mechanism. However, this is not yet
// implemented and can be done differently. One of the possible scenarios is that "runtime_monitor_configs" will be slightly modified and become
// the base for storing dynamic configs of all components instead of having different config tables, in this case this method will be removed.
func (Config) TableName() string {
	return "runtime_monitor_configs"
}

func (s *ConfigJSON) Scan(src interface{}) error {
	b := src.([]byte)
	return json.Unmarshal(b, s)
}

func (s *ConfigJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}
