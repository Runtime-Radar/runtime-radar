import { RuntimeEventProcessorHistoryControl } from './runtime-event-processor-contract.interface';

export type RuntimeMonitorTracingPolicies = {
    [key: string]: RuntimeMonitorTracingPolicy;
};

export interface RuntimeMonitorTracingPolicy {
    name: string;
    enabled: boolean;
    description?: string;
    yaml?: string;
}

export interface RuntimeMonitorPermission {
    arguments_regex: string[];
    binary_regex: string[];
    namespace: string[];
    health_check: boolean;
    pid: number[];
    pid_set: number[];
    event_set: string[];
    pod_regex: string[];
    labels: string[];
}

export interface RuntimeMonitorAggregationOptions {
    window_size: string;
    channel_buffer_size: string;
}

export interface RuntimeMonitorConfig {
    version: string;
    tracing_policies: RuntimeMonitorTracingPolicies;
    allow_list: Partial<RuntimeMonitorPermission>[];
    deny_list: Partial<RuntimeMonitorPermission>[];
    aggregation_options?: RuntimeMonitorAggregationOptions;
}

export interface RuntimeMonitorConfigStatus {
    default: boolean;
    default_tracing_policies: boolean;
    last_init_error: string;
    node_name: string;
}

export interface RuntimeMonitorConfigExtended extends RuntimeMonitorConfig {
    historyControl: RuntimeEventProcessorHistoryControl;
}

export interface RuntimeMonitor {
    id: string;
    config: RuntimeMonitorConfig;
}
