import { createAction, props } from '@ngrx/store';

import {
    RuntimeEventProcessorHistoryControl,
    RuntimeMonitorConfig,
    RuntimeMonitorConfigExtended,
    RuntimeState
} from '../interfaces';

export const LOAD_RUNTIME_CONFIG_TODO_ACTION = createAction('[Runtime] Load Config');

export const DEACTIVATE_RUNTIME_CONFIG_TODO_ACTION = createAction('[Runtime] Deactivate Config');

export const CREATE_RUNTIME_CONFIG_TODO_ACTION = createAction(
    '[Runtime] Create Config',
    props<{ config: RuntimeMonitorConfig; historyControl: RuntimeEventProcessorHistoryControl }>()
);

export const CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION = createAction('[Runtime] Create Notification');

export const RESET_RUNTIME_CONFIG_TODO_ACTION = createAction('[Runtime] Reset Config');

export const CHECK_RUNTIME_CHANGES_TODO_ACTION = createAction(
    '[Runtime] Check Changes',
    props<{ config: RuntimeMonitorConfigExtended }>()
);

export const GET_RUNTIME_CONFIG_STATUS_TODO_ACTION = createAction('[Runtime] Get Config Status');

export const SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION = createAction('[Runtime] Switch Expert Mode');

export const HIDE_RUNTIME_OVERLAY_TODO_ACTION = createAction('[Runtime] Hide Overlay');

export const UPDATE_RUNTIME_STATE_DOC_ACTION = createAction(
    '[Runtime] (Doc) Update State',
    props<Partial<RuntimeState>>()
);
