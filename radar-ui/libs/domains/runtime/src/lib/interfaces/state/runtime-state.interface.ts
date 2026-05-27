import { LoadStatus } from '@cs/core';

import { RuntimeEventProcessorHistoryControl } from '../contract/runtime-event-processor-contract.interface';
import { RuntimeMonitorConfig, RuntimeMonitorConfigStatus } from '../contract/runtime-monitor-contract.interface';

export enum RuntimeConfigStatus {
    INIT,
    STAY,
    MODIFY
}

export interface RuntimeState {
    loadStatus: LoadStatus;
    configStatus: RuntimeConfigStatus;
    hasChanges: boolean;
    hasPoliciesChanges: boolean;
    isExpertMode: boolean;
    isOverlayed: boolean;
    grafanaUrl: string;
    historyControl?: RuntimeEventProcessorHistoryControl;
    config: RuntimeMonitorConfig;
    status?: RuntimeMonitorConfigStatus;
}
