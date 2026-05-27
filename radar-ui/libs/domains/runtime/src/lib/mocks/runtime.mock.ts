import { RuntimeEventProcessor, RuntimeEventProcessorHistoryControl, RuntimeMonitorConfig } from '../interfaces';

export const RUNTIME_MONITOR_CONFIG: RuntimeMonitorConfig = {
    version: '1.0',
    tracing_policies: {},
    allow_list: [],
    deny_list: []
};

export const RUNTIME_EVENT_PROCESSOR: RuntimeEventProcessor = {
    id: 'ep1',
    config: {
        version: '1',
        history_control: RuntimeEventProcessorHistoryControl.ALL
    }
};
