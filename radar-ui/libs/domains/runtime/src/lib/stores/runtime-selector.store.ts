import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { RuntimeState } from '../interfaces';
import { runtimeReducer } from './runtime-reducer.store';

export const RUNTIME_DOMAIN_KEY = 'runtime';

export interface RuntimeDomainState {
    readonly domain: RuntimeState;
}

const selectRuntimeDomainState = createFeatureSelector<RuntimeDomainState>(RUNTIME_DOMAIN_KEY);
const selectRuntimeState = createSelector(selectRuntimeDomainState, (state: RuntimeDomainState) => state.domain);

export const getRuntimeLoadStatus = createSelector(selectRuntimeState, (state: RuntimeState) => state.loadStatus);

export const getRuntimeConfigStatus = createSelector(selectRuntimeState, (state: RuntimeState) => state.configStatus);

export const getRuntimeGrafanaUrl = createSelector(selectRuntimeState, (state: RuntimeState) => state.grafanaUrl);

export const getRuntimeMonitorConfig = createSelector(selectRuntimeState, (state: RuntimeState) => state.config);

export const getRuntimeMonitorConfigStatus = createSelector(selectRuntimeState, (state: RuntimeState) => state.status);

export const getRuntimeEventProcessorHistoryControl = createSelector(
    selectRuntimeState,
    (state: RuntimeState) => state.historyControl
);

export const getRuntimeHasChanges = createSelector(selectRuntimeState, (state: RuntimeState) => state.hasChanges);

export const getRuntimeHasPoliciesChanges = createSelector(
    selectRuntimeState,
    (state: RuntimeState) => state.hasPoliciesChanges
);

export const getRuntimeIsExpertMode = createSelector(selectRuntimeState, (state: RuntimeState) => state.isExpertMode);

export const getRuntimeIsOverlayed = createSelector(selectRuntimeState, (state: RuntimeState) => state.isOverlayed);

export const runtimeDomainReducer: ActionReducerMap<RuntimeDomainState> = {
    domain: runtimeReducer
};
