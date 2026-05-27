import { createEntityAdapter } from '@ngrx/entity';
import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import { KubeManagerPod, KubeManagerState } from '../interfaces';
import {
    SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION,
    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION
} from './kube-manager-action.store';

export const adapter = createEntityAdapter<KubeManagerPod>({
    selectId: (pod) => pod.uid
});

const INITIAL_STATE: KubeManagerState = {
    loadStatus: LoadStatus.INIT,
    lastUpdate: 0,
    list: adapter.getInitialState()
};

const reducer: ActionReducer<KubeManagerState, Action> = createReducer(
    INITIAL_STATE,
    on(UPDATE_KUBE_MANAGER_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values })),
    on(SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION, (state, { list }) => ({
        ...state,
        list: adapter.setAll(list, state.list)
    }))
);

export const kubeManagerEntitySelector = adapter.getSelectors();

export function kubeManagerReducer(state: KubeManagerState | undefined, action: Action): KubeManagerState {
    return reducer(state, action);
}
