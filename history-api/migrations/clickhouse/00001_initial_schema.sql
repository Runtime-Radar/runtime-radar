-- +goose Up
-- Initial schema for history-api ClickHouse: runtime_events table

CREATE TABLE IF NOT EXISTS runtime_events (
    id UUID,
    created_at DateTime64(9),
    registered_at DateTime64(9),

    -- Common event attributes
    event_type Enum8('UNDEF' = 0, 'PROCESS_EXEC' = 1, 'PROCESS_EXIT' = 5, 'PROCESS_KPROBE' = 9, 'PROCESS_TRACEPOINT' = 10, 'PROCESS_LOADER' = 11, 'PROCESS_UPROBE' = 12),
    node_name String,
    source_event String,
    threats Nullable(String),
    threats_detectors Array(String),
    threats_severities Array(Int32),
    detect_errors Nullable(String),
    tetragon_version String,
    is_incident UInt8,
    incident_severity UInt8,
    block_by Array(String),
    notify_by Array(String),

    -- Common process attributes
    process_exec_id String,
    process_pid UInt32,
    process_uid UInt32,
    process_cwd String,
    process_binary String,
    process_arguments String,
    process_flags String,
    process_start_time Nullable(DateTime64(9)),
    process_auid UInt32,
    process_pod_namespace String,
    process_pod_name String,
    process_pod_container_id String,
    process_pod_container_name String,
    process_pod_container_image_id String,
    process_pod_container_image_name String,
    process_pod_container_start_time Nullable(DateTime64(9)),
    process_pod_container_pid UInt32,
    process_pod_container_maybe_exec_probe UInt8,
    process_pod_pod_labels Nullable(String),
    process_pod_workload String,
    process_pod_workload_kind String,
    process_docker String,
    process_parent_exec_id String,
    process_refcnt UInt32,
    process_cap_permitted Array(Enum8('CAP_CHOWN' = 0, 'DAC_OVERRIDE' = 1, 'CAP_DAC_READ_SEARCH' = 2, 'CAP_FOWNER' = 3, 'CAP_FSETID' = 4, 'CAP_KILL' = 5, 'CAP_SETGID' = 6, 'CAP_SETUID' = 7, 'CAP_SETPCAP' = 8, 'CAP_LINUX_IMMUTABLE' = 9, 'CAP_NET_BIND_SERVICE' = 10, 'CAP_NET_BROADCAST' = 11, 'CAP_NET_ADMIN' = 12, 'CAP_NET_RAW' = 13, 'CAP_IPC_LOCK' = 14, 'CAP_IPC_OWNER' = 15, 'CAP_SYS_MODULE' = 16, 'CAP_SYS_RAWIO' = 17, 'CAP_SYS_CHROOT' = 18, 'CAP_SYS_PTRACE' = 19, 'CAP_SYS_PACCT' = 20, 'CAP_SYS_ADMIN' = 21, 'CAP_SYS_BOOT' = 22, 'CAP_SYS_NICE' = 23, 'CAP_SYS_RESOURCE' = 24, 'CAP_SYS_TIME' = 25, 'CAP_SYS_TTY_CONFIG' = 26, 'CAP_MKNOD' = 27, 'CAP_LEASE' = 28, 'CAP_AUDIT_WRITE' = 29, 'CAP_AUDIT_CONTROL' = 30, 'CAP_SETFCAP' = 31, 'CAP_MAC_OVERRIDE' = 32, 'CAP_MAC_ADMIN' = 33, 'CAP_SYSLOG' = 34, 'CAP_WAKE_ALARM' = 35, 'CAP_BLOCK_SUSPEND' = 36, 'CAP_AUDIT_READ' = 37, 'CAP_PERFMON' = 38, 'CAP_BPF' = 39, 'CAP_CHECKPOINT_RESTORE' = 40)),
    process_cap_effective Array(Enum8('CAP_CHOWN' = 0, 'DAC_OVERRIDE' = 1, 'CAP_DAC_READ_SEARCH' = 2, 'CAP_FOWNER' = 3, 'CAP_FSETID' = 4, 'CAP_KILL' = 5, 'CAP_SETGID' = 6, 'CAP_SETUID' = 7, 'CAP_SETPCAP' = 8, 'CAP_LINUX_IMMUTABLE' = 9, 'CAP_NET_BIND_SERVICE' = 10, 'CAP_NET_BROADCAST' = 11, 'CAP_NET_ADMIN' = 12, 'CAP_NET_RAW' = 13, 'CAP_IPC_LOCK' = 14, 'CAP_IPC_OWNER' = 15, 'CAP_SYS_MODULE' = 16, 'CAP_SYS_RAWIO' = 17, 'CAP_SYS_CHROOT' = 18, 'CAP_SYS_PTRACE' = 19, 'CAP_SYS_PACCT' = 20, 'CAP_SYS_ADMIN' = 21, 'CAP_SYS_BOOT' = 22, 'CAP_SYS_NICE' = 23, 'CAP_SYS_RESOURCE' = 24, 'CAP_SYS_TIME' = 25, 'CAP_SYS_TTY_CONFIG' = 26, 'CAP_MKNOD' = 27, 'CAP_LEASE' = 28, 'CAP_AUDIT_WRITE' = 29, 'CAP_AUDIT_CONTROL' = 30, 'CAP_SETFCAP' = 31, 'CAP_MAC_OVERRIDE' = 32, 'CAP_MAC_ADMIN' = 33, 'CAP_SYSLOG' = 34, 'CAP_WAKE_ALARM' = 35, 'CAP_BLOCK_SUSPEND' = 36, 'CAP_AUDIT_READ' = 37, 'CAP_PERFMON' = 38, 'CAP_BPF' = 39, 'CAP_CHECKPOINT_RESTORE' = 40)),
    process_cap_inheritable Array(Enum8('CAP_CHOWN' = 0, 'DAC_OVERRIDE' = 1, 'CAP_DAC_READ_SEARCH' = 2, 'CAP_FOWNER' = 3, 'CAP_FSETID' = 4, 'CAP_KILL' = 5, 'CAP_SETGID' = 6, 'CAP_SETUID' = 7, 'CAP_SETPCAP' = 8, 'CAP_LINUX_IMMUTABLE' = 9, 'CAP_NET_BIND_SERVICE' = 10, 'CAP_NET_BROADCAST' = 11, 'CAP_NET_ADMIN' = 12, 'CAP_NET_RAW' = 13, 'CAP_IPC_LOCK' = 14, 'CAP_IPC_OWNER' = 15, 'CAP_SYS_MODULE' = 16, 'CAP_SYS_RAWIO' = 17, 'CAP_SYS_CHROOT' = 18, 'CAP_SYS_PTRACE' = 19, 'CAP_SYS_PACCT' = 20, 'CAP_SYS_ADMIN' = 21, 'CAP_SYS_BOOT' = 22, 'CAP_SYS_NICE' = 23, 'CAP_SYS_RESOURCE' = 24, 'CAP_SYS_TIME' = 25, 'CAP_SYS_TTY_CONFIG' = 26, 'CAP_MKNOD' = 27, 'CAP_LEASE' = 28, 'CAP_AUDIT_WRITE' = 29, 'CAP_AUDIT_CONTROL' = 30, 'CAP_SETFCAP' = 31, 'CAP_MAC_OVERRIDE' = 32, 'CAP_MAC_ADMIN' = 33, 'CAP_SYSLOG' = 34, 'CAP_WAKE_ALARM' = 35, 'CAP_BLOCK_SUSPEND' = 36, 'CAP_AUDIT_READ' = 37, 'CAP_PERFMON' = 38, 'CAP_BPF' = 39, 'CAP_CHECKPOINT_RESTORE' = 40)),
    process_ns_uts_inum UInt32,
    process_ns_uts_is_host UInt8,
    process_ns_ipc_inum UInt32,
    process_ns_ipc_is_host UInt8,
    process_ns_mnt_inum UInt32,
    process_ns_mnt_is_host UInt8,
    process_ns_pid_inum UInt32,
    process_ns_pid_is_host UInt8,
    process_ns_pid_for_children_inum UInt32,
    process_ns_pid_for_children_is_host UInt8,
    process_ns_net_inum UInt32,
    process_ns_net_is_host UInt8,
    process_ns_time_inum UInt32,
    process_ns_time_is_host UInt8,
    process_ns_time_for_children_inum UInt32,
    process_ns_time_for_children_is_host UInt8,
    process_ns_cgroup_inum UInt32,
    process_ns_cgroup_is_host UInt8,
    process_ns_user_inum UInt32,
    process_ns_user_is_host UInt8,
    process_tid UInt32,

    -- Process's parent attributes
    parent_exec_id String,
    parent_pid UInt32,
    parent_uid UInt32,
    parent_cwd String,
    parent_binary String,
    parent_arguments String,
    parent_flags String,
    parent_start_time Nullable(DateTime64(9)),
    parent_auid UInt32,
    parent_pod_namespace String,
    parent_pod_name String,
    parent_pod_container_id String,
    parent_pod_container_name String,
    parent_pod_container_image_id String,
    parent_pod_container_image_name String,
    parent_pod_container_start_time Nullable(DateTime64(9)),
    parent_pod_container_pid UInt32,
    parent_pod_container_maybe_exec_probe UInt8,
    parent_pod_pod_labels Nullable(String),
    parent_pod_workload String,
    parent_pod_workload_kind String,
    parent_docker String,
    parent_parent_exec_id String,
    parent_refcnt UInt32,
    parent_cap_permitted Array(Enum8('CAP_CHOWN' = 0, 'DAC_OVERRIDE' = 1, 'CAP_DAC_READ_SEARCH' = 2, 'CAP_FOWNER' = 3, 'CAP_FSETID' = 4, 'CAP_KILL' = 5, 'CAP_SETGID' = 6, 'CAP_SETUID' = 7, 'CAP_SETPCAP' = 8, 'CAP_LINUX_IMMUTABLE' = 9, 'CAP_NET_BIND_SERVICE' = 10, 'CAP_NET_BROADCAST' = 11, 'CAP_NET_ADMIN' = 12, 'CAP_NET_RAW' = 13, 'CAP_IPC_LOCK' = 14, 'CAP_IPC_OWNER' = 15, 'CAP_SYS_MODULE' = 16, 'CAP_SYS_RAWIO' = 17, 'CAP_SYS_CHROOT' = 18, 'CAP_SYS_PTRACE' = 19, 'CAP_SYS_PACCT' = 20, 'CAP_SYS_ADMIN' = 21, 'CAP_SYS_BOOT' = 22, 'CAP_SYS_NICE' = 23, 'CAP_SYS_RESOURCE' = 24, 'CAP_SYS_TIME' = 25, 'CAP_SYS_TTY_CONFIG' = 26, 'CAP_MKNOD' = 27, 'CAP_LEASE' = 28, 'CAP_AUDIT_WRITE' = 29, 'CAP_AUDIT_CONTROL' = 30, 'CAP_SETFCAP' = 31, 'CAP_MAC_OVERRIDE' = 32, 'CAP_MAC_ADMIN' = 33, 'CAP_SYSLOG' = 34, 'CAP_WAKE_ALARM' = 35, 'CAP_BLOCK_SUSPEND' = 36, 'CAP_AUDIT_READ' = 37, 'CAP_PERFMON' = 38, 'CAP_BPF' = 39, 'CAP_CHECKPOINT_RESTORE' = 40)),
    parent_cap_effective Array(Enum8('CAP_CHOWN' = 0, 'DAC_OVERRIDE' = 1, 'CAP_DAC_READ_SEARCH' = 2, 'CAP_FOWNER' = 3, 'CAP_FSETID' = 4, 'CAP_KILL' = 5, 'CAP_SETGID' = 6, 'CAP_SETUID' = 7, 'CAP_SETPCAP' = 8, 'CAP_LINUX_IMMUTABLE' = 9, 'CAP_NET_BIND_SERVICE' = 10, 'CAP_NET_BROADCAST' = 11, 'CAP_NET_ADMIN' = 12, 'CAP_NET_RAW' = 13, 'CAP_IPC_LOCK' = 14, 'CAP_IPC_OWNER' = 15, 'CAP_SYS_MODULE' = 16, 'CAP_SYS_RAWIO' = 17, 'CAP_SYS_CHROOT' = 18, 'CAP_SYS_PTRACE' = 19, 'CAP_SYS_PACCT' = 20, 'CAP_SYS_ADMIN' = 21, 'CAP_SYS_BOOT' = 22, 'CAP_SYS_NICE' = 23, 'CAP_SYS_RESOURCE' = 24, 'CAP_SYS_TIME' = 25, 'CAP_SYS_TTY_CONFIG' = 26, 'CAP_MKNOD' = 27, 'CAP_LEASE' = 28, 'CAP_AUDIT_WRITE' = 29, 'CAP_AUDIT_CONTROL' = 30, 'CAP_SETFCAP' = 31, 'CAP_MAC_OVERRIDE' = 32, 'CAP_MAC_ADMIN' = 33, 'CAP_SYSLOG' = 34, 'CAP_WAKE_ALARM' = 35, 'CAP_BLOCK_SUSPEND' = 36, 'CAP_AUDIT_READ' = 37, 'CAP_PERFMON' = 38, 'CAP_BPF' = 39, 'CAP_CHECKPOINT_RESTORE' = 40)),
    parent_cap_inheritable Array(Enum8('CAP_CHOWN' = 0, 'DAC_OVERRIDE' = 1, 'CAP_DAC_READ_SEARCH' = 2, 'CAP_FOWNER' = 3, 'CAP_FSETID' = 4, 'CAP_KILL' = 5, 'CAP_SETGID' = 6, 'CAP_SETUID' = 7, 'CAP_SETPCAP' = 8, 'CAP_LINUX_IMMUTABLE' = 9, 'CAP_NET_BIND_SERVICE' = 10, 'CAP_NET_BROADCAST' = 11, 'CAP_NET_ADMIN' = 12, 'CAP_NET_RAW' = 13, 'CAP_IPC_LOCK' = 14, 'CAP_IPC_OWNER' = 15, 'CAP_SYS_MODULE' = 16, 'CAP_SYS_RAWIO' = 17, 'CAP_SYS_CHROOT' = 18, 'CAP_SYS_PTRACE' = 19, 'CAP_SYS_PACCT' = 20, 'CAP_SYS_ADMIN' = 21, 'CAP_SYS_BOOT' = 22, 'CAP_SYS_NICE' = 23, 'CAP_SYS_RESOURCE' = 24, 'CAP_SYS_TIME' = 25, 'CAP_SYS_TTY_CONFIG' = 26, 'CAP_MKNOD' = 27, 'CAP_LEASE' = 28, 'CAP_AUDIT_WRITE' = 29, 'CAP_AUDIT_CONTROL' = 30, 'CAP_SETFCAP' = 31, 'CAP_MAC_OVERRIDE' = 32, 'CAP_MAC_ADMIN' = 33, 'CAP_SYSLOG' = 34, 'CAP_WAKE_ALARM' = 35, 'CAP_BLOCK_SUSPEND' = 36, 'CAP_AUDIT_READ' = 37, 'CAP_PERFMON' = 38, 'CAP_BPF' = 39, 'CAP_CHECKPOINT_RESTORE' = 40)),
    parent_ns_uts_inum UInt32,
    parent_ns_uts_is_host UInt8,
    parent_ns_ipc_inum UInt32,
    parent_ns_ipc_is_host UInt8,
    parent_ns_mnt_inum UInt32,
    parent_ns_mnt_is_host UInt8,
    parent_ns_pid_inum UInt32,
    parent_ns_pid_is_host UInt8,
    parent_ns_pid_for_children_inum UInt32,
    parent_ns_pid_for_children_is_host UInt8,
    parent_ns_net_inum UInt32,
    parent_ns_net_is_host UInt8,
    parent_ns_time_inum UInt32,
    parent_ns_time_is_host UInt8,
    parent_ns_time_for_children_inum UInt32,
    parent_ns_time_for_children_is_host UInt8,
    parent_ns_cgroup_inum UInt32,
    parent_ns_cgroup_is_host UInt8,
    parent_ns_user_inum UInt32,
    parent_ns_user_is_host UInt8,
    parent_tid UInt32,
    policy_name String,

    -- Extra exit attributes
    exit_signal String,
    exit_status UInt32,
    exit_time Nullable(DateTime64(9)),

    -- Extra kprobe attributes
    kprobe_function_name String,
    kprobe_action Enum8('KPROBE_ACTION_UNKNOWN' = 0, 'KPROBE_ACTION_POST' = 1, 'KPROBE_ACTION_FOLLOWFD' = 2, 'KPROBE_ACTION_SIGKILL' = 3, 'KPROBE_ACTION_UNFOLLOWFD' = 4, 'KPROBE_ACTION_OVERRIDE' = 5, 'KPROBE_ACTION_COPYFD' = 6, 'KPROBE_ACTION_GETURL' = 7, 'KPROBE_ACTION_DNSLOOKUP' = 8, 'KPROBE_ACTION_NOPOST' = 9, 'KPROBE_ACTION_SIGNAL' = 10),

    -- Extra tracepoint attributes
    tracepoint_subsys String,
    tracepoint_event String,

    -- Extra loader attributes
    loader_path String,
    loader_buildid String,

    -- Extra uprobe attributes
    uprobe_path String,
    uprobe_symbol String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(registered_at)
ORDER BY (registered_at)
SETTINGS index_granularity = 8192;

-- Data-skipping indexes are added separately so that a table created by an earlier release picks
-- them up as well. ClickHouse does not index already stored rows here; existing parts get indexed
-- during their next merge.

-- According to the research on data from Standoffs, bloom filter index can reduce latency of queries performing searches in array columns (for example, using has() function).
-- Cardinality and distribution of arrays' elements by granulas were taken into account.
-- The effectiveness of index can be different depending on data distribution in production environment.
-- Granularity was chosen by analyzing the number of unique values within different numbers of granules.
ALTER TABLE runtime_events ADD INDEX IF NOT EXISTS idx_runtime_events_threats_detectors (threats_detectors) TYPE bloom_filter GRANULARITY 3;
-- Currently, clickhouse doesn't use data-skipping indexes when executing queries with WHERE ... OR ... conditions on multiple indexed columns.
-- See https://github.com/ClickHouse/ClickHouse/issues/8168 for details.
-- There's workaround which implies creating composite index on fields which take part in OR query.
ALTER TABLE runtime_events ADD INDEX IF NOT EXISTS idx_runtime_events_block_by_notify_by (block_by, notify_by) TYPE bloom_filter GRANULARITY 3;

-- Number of tokens (n) for tokenbf_v1 was chosen by exploring data distribution within tables filled during Standoffs.
-- According to the results, process_pod_name can contain up to 55 unique tokens (n) per granule. Actually, 100 was used as n in order to have some reserve.
-- As most of other columns have lower cardinality, calculated parameters can be generalized and used to index those columns as well.
-- This value may be changed later after exploring tables with production data.
--
-- Number of hash functions (k) is kept low because higher values can affect inserts' bandwidth.
--
-- Index has granularity = 1 as it isn't very expensive to index every granule.
ALTER TABLE runtime_events ADD INDEX IF NOT EXISTS idx_runtime_events_process_pod_name (process_pod_name) TYPE tokenbf_v1(256, 2, 0) GRANULARITY 1;
ALTER TABLE runtime_events ADD INDEX IF NOT EXISTS idx_runtime_events_process_pod_namespace (process_pod_namespace) TYPE tokenbf_v1(256, 2, 0) GRANULARITY 1;

-- Size of set and index's granularity makes it potentially useful for the fields with low cardinality, such as boolean values.
-- Index has granularity = 1 as it isn't very expensive to index every granule.
ALTER TABLE runtime_events ADD INDEX IF NOT EXISTS idx_runtime_events_is_incident (is_incident) TYPE set(0) GRANULARITY 1;

-- +goose Down
DROP TABLE IF EXISTS runtime_events;
