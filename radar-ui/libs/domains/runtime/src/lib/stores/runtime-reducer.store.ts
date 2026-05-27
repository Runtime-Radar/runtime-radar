import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import { UPDATE_RUNTIME_STATE_DOC_ACTION } from './runtime-action.store';
import { RuntimeConfigStatus, RuntimeState } from '../interfaces';

const INITIAL_STATE: RuntimeState = {
    loadStatus: LoadStatus.INIT,
    configStatus: RuntimeConfigStatus.INIT,
    hasChanges: false,
    hasPoliciesChanges: false,
    isExpertMode: false,
    isOverlayed: false,
    grafanaUrl: '',
    config: {
        version: '',
        tracing_policies: {},
        allow_list: [],
        deny_list: []
    }
};

const reducer: ActionReducer<RuntimeState, Action> = createReducer(
    INITIAL_STATE,
    on(UPDATE_RUNTIME_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values }))
);

export function runtimeReducer(state: RuntimeState | undefined, action: Action): RuntimeState {
    return reducer(state, action);
}
