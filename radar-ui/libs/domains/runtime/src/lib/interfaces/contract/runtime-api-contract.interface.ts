import { RuntimeEvent } from './runtime-event-contract.interface';
import { RuntimeEventProcessorConfig } from './runtime-event-processor-contract.interface';
import { RuntimeMonitorConfig } from './runtime-monitor-contract.interface';

export interface GetGrafanaUrlResponse {
    url: string;
}

export interface RuntimeDateTimeRange {
    from: string | null; // RFC3339
    to: string | null; // RFC3339
}

export interface RuntimeFilterRequest {
    event_type: string[]; // RuntimeEventType[]
    period: Partial<RuntimeDateTimeRange>;
    has_threats?: boolean;
    has_incident?: boolean;
    kprobe_function_name: string[];
    process_pod_namespace: string[];
    process_pod_name: string[];
    node_name: string[];
    process_pod_container_name: string[];
    process_pod_container_image_name: string[];
    process_binary: string[];
    process_arguments: string[];
    process_exec_id: string;
    process_parent_exec_id: string;
    threats_detectors: string[];
    rules: string[];
}

export enum RuntimeEventCursorDirection {
    LEFT = 'left',
    RIGHT = 'right'
}

export interface GetRuntimeEventCountResponse {
    count: number;
}

export interface GetRuntimeEventsRequest {
    cursor: string; // RFC3339
    slice_size: number;
}

export interface GetRuntimeEventsByFilterRequest extends GetRuntimeEventsRequest {
    filter: Partial<RuntimeFilterRequest>;
}

export interface GetRuntimeEventsResponse {
    runtime_events: RuntimeEvent[];
    left_cursor: string; // RFC3339
    right_cursor: string; // RFC3339
}

export interface CreateRuntimeMonitorRequest {
    config: RuntimeMonitorConfig;
}

export interface CreateRuntimeEventProcessorRequest {
    config: RuntimeEventProcessorConfig;
}

export type EmptyRuntimeResponse = Record<string, unknown>;
