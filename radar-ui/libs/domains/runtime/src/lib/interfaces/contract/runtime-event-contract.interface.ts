import { RuleSeverity } from '@cs/domains/rule';
import { Detector, DetectorMitreTactic } from '@cs/domains/detector';

export enum RuntimeEventType {
    EXEC = 'process_exec',
    EXIT = 'process_exit',
    KPROBE = 'process_kprobe',
    LOADER = 'process_loader',
    TRACEPOINT = 'process_tracepoint',
    UPROBE = 'process_uprobe'
}

export enum RuntimeCapabilityType {
    CAP_CHOWN = 'CAP_CHOWN',
    DAC_OVERRIDE = 'DAC_OVERRIDE',
    CAP_DAC_READ_SEARCH = 'CAP_DAC_READ_SEARCH',
    CAP_FOWNER = 'CAP_FOWNER',
    CAP_FSETID = 'CAP_FSETID',
    CAP_KILL = 'CAP_KILL',
    CAP_SETGID = 'CAP_SETGID',
    CAP_SETUID = 'CAP_SETUID',
    CAP_SETPCAP = 'CAP_SETPCAP',
    CAP_LINUX_IMMUTABLE = 'CAP_LINUX_IMMUTABLE',
    CAP_NET_BIND_SERVICE = 'CAP_NET_BIND_SERVICE',
    CAP_NET_BROADCAST = 'CAP_NET_BROADCAST',
    CAP_NET_ADMIN = 'CAP_NET_ADMIN',
    CAP_NET_RAW = 'CAP_NET_RAW',
    CAP_IPC_LOCK = 'CAP_IPC_LOCK',
    CAP_IPC_OWNER = 'CAP_IPC_OWNER',
    CAP_SYS_MODULE = 'CAP_SYS_MODULE',
    CAP_SYS_RAWIO = 'CAP_SYS_RAWIO',
    CAP_SYS_CHROOT = 'CAP_SYS_CHROOT',
    CAP_SYS_PTRACE = 'CAP_SYS_PTRACE',
    CAP_SYS_PACCT = 'CAP_SYS_PACCT',
    CAP_SYS_ADMIN = 'CAP_SYS_ADMIN',
    CAP_SYS_BOOT = 'CAP_SYS_BOOT',
    CAP_SYS_NICE = 'CAP_SYS_NICE',
    CAP_SYS_RESOURCE = 'CAP_SYS_RESOURCE',
    CAP_SYS_TIME = 'CAP_SYS_TIME',
    CAP_SYS_TTY_CONFIG = 'CAP_SYS_TTY_CONFIG',
    CAP_MKNOD = 'CAP_MKNOD',
    CAP_LEASE = 'CAP_LEASE',
    CAP_AUDIT_WRITE = 'CAP_AUDIT_WRITE',
    CAP_AUDIT_CONTROL = 'CAP_AUDIT_CONTROL',
    CAP_SETFCAP = 'CAP_SETFCAP',
    CAP_MAC_OVERRIDE = 'CAP_MAC_OVERRIDE',
    CAP_MAC_ADMIN = 'CAP_MAC_ADMIN',
    CAP_SYSLOG = 'CAP_SYSLOG',
    CAP_WAKE_ALARM = 'CAP_WAKE_ALARM',
    CAP_BLOCK_SUSPEND = 'CAP_BLOCK_SUSPEND',
    CAP_AUDIT_READ = 'CAP_AUDIT_READ',
    CAP_PERFMON = 'CAP_PERFMON',
    CAP_BPF = 'CAP_BPF',
    CAP_CHECKPOINT_RESTORE = 'CAP_CHECKPOINT_RESTORE'
}

export interface RuntimeEventProcessEntityPod {
    namespace: string;
    name: string;
    container: RuntimeEventProcessPodContainer;
}

export interface RuntimeEventProcessPodContainer {
    id: string;
    name: string;
    maybe_exec_probe: boolean;
    pid: number;
    start_time: string; // RFC3339
    image: RuntimeEventProcessPodContainerImage;
}

export interface RuntimeEventProcessPodContainerImage {
    id: string;
    name: string;
}

export interface RuntimeEventProcessEntityCap {
    permitted: RuntimeCapabilityType[];
    effective: RuntimeCapabilityType[];
    inheritable: RuntimeCapabilityType[];
}

export interface RuntimeEventProcessEntity {
    binary: string;
    arguments: string;
    exec_id: string;
    parent_exec_id: string;
    start_time: string; // RFC3339
    pid: number;
    uid: number;
    cwd: string;
    pod: RuntimeEventProcessEntityPod | null;
    cap: RuntimeEventProcessEntityCap;
}

export interface RuntimeEventProcess {
    process: RuntimeEventProcessEntity;
    parent: RuntimeEventProcessEntity | null;
    function_name?: string;
}

export type RuntimeEventEntity = {
    node_name: string;
    time: string;
} & Partial<{
    [key in RuntimeEventType]: RuntimeEventProcess;
}>;

export interface RuntimeEventThreat {
    detector: Detector;
    severity: RuleSeverity;
    reason: string;
    tactics_covered: DetectorMitreTactic[];
}

export interface RuntimeDetectError {
    detector: Detector;
    error: string;
}

export interface RuntimeEvent {
    id: string;
    tetragon_version: string;
    event: RuntimeEventEntity;
    threats: RuntimeEventThreat[];
    block_by: string[];
    notify_by: string[];
    incident_severity: RuleSeverity;
    is_incident: boolean;
    detect_errors: RuntimeDetectError[];
}
