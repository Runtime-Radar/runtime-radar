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

	//go:embed tracingpolicy/socket.yaml
	socket string

	//go:embed tracingpolicy/io-streams.yaml
	ioStreams string

	//go:embed tracingpolicy/io-uring.yml
	ioUring string
)

var (
	DefaultConfig = &Config{
		Base: Base{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		Config: &ConfigJSON{
			Version: string(ConfigVersion),
			TracingPolicies: map[string]*api.TracingPolicy{
				"connect": {
					Name:        "Outgoing TCP connections",
					Description: "This source tracks calls to the Linux kernel functions tcp_connect(), tcp_close(), and tcp_sendmsg() to detect outgoing TCP connections, including connection establishment, connection termination, and TCP packet transmission. Enabling this source can significantly increase the event flow and system load. To optimize performance, use precise filters, for example, limit monitoring to specific pods.",
					Yaml:        connect,
					Enabled:     false,
				},
				"permissions": {
					Name:        "File and process permission changes",
					Description: "This source tracks calls to the Linux kernel function commit_creds(), which may indicate an attacker attempting to escalate process privileges, including obtaining superuser (root) privileges. The source also tracks calls to the LSM function security_path_chmod() with permission sets that include execute permissions.",
					Yaml:        permissions,
					Enabled:     false,
				},
				"file-monitoring": {
					Name:        "Access to important system files",
					Description: "This source tracks calls to the Linux kernel functions security_file_permission(), security_mmap_file(), security_path_truncate(), security_path_link(), and security_path_rename() targeting important system directories and files. These include, for example, /boot, /root/.ssh, /etc/shadow, /etc/profile, /etc/sudoers, and /etc/pam.conf. Different actions are tracked for different files, including read-only, write-only, read and write, and hard link creation. To examine how the source works in detail, enable expert mode.",
					Yaml:        fileMonitoring,
					Enabled:     false,
				},
				"ptrace": {
					Name:        "Using tools for debugging and reverse engineering (ptrace)",
					Description: "This source tracks the ptrace() system call, which may indicate an attacker attempting to inject malicious code into another process.",
					Yaml:        ptrace,
					Enabled:     false,
				},
				"mount": {
					Name:        "Device mounting",
					Description: "This source tracks the mount() system call to detect suspicious events related to device mounting.",
					Yaml:        mount,
					Enabled:     false,
				},
				"kernel-modules": {
					Name:        "Loading and unloading kernel modules",
					Description: "This source tracks calls to the Linux kernel functions do_init_module(), free_module(), security_kernel_module_request(), and security_kernel_read_file() to record all kernel module loading and unloading events, including automatic events initiated by the system. This helps detect signs of unauthorized module manipulation and other illegitimate activity.",
					Yaml:        kernelModules,
					Enabled:     false,
				},
				"socket": {
					Name:        "Socket monitoring",
					Description: "This source tracks the inet_csk_listen_start call revealing possible activity of unwanted networking tools. The source also detects legitimate activity in a container; therefore, in addition to opening of a socket, detectors must also consider other event parameters.",
					Yaml:        socket,
					Enabled:     false,
				},
				"io-streams": {
					Name:        "Standard input/output stream activity",
					Description: "This source tracks calls to the Linux kernel function do_dup2(), which copies the standard input (STDIN) file descriptor, as well as named pipe (S_IFIFO) file creation through the LSM function security_path_mknod(). Such actions often indicate that an attacker is attempting to run a reverse shell, hidden communication channel, or another attack tool.",
					Yaml:        ioStreams,
					Enabled:     false,
				},
				"io-uring": {
					Name:        "Creating and using an io_uring interface instance",
					Description: "This source tracks the io_uring_setup() and io_uring_enter() system calls, which indicate use of the io_uring interface on the system. This helps detect suspicious activity related to unauthorized access to asynchronous input/output operations or potential exploitation of vulnerabilities in Linux kernel mechanisms.",
					Yaml:        ioUring,
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
