import { createEntityAdapter } from '@ngrx/entity';
import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import { Notification } from '../interfaces/contract/notification-contract.interface';
import { NotificationState } from '../interfaces/state/notification-state.interface';
import {
    DELETE_NOTIFICATION_ENTITIES_DOC_ACTION,
    DELETE_NOTIFICATION_ENTITY_DOC_ACTION,
    SET_ALL_NOTIFICATION_ENTITIES_DOC_ACTION,
    SET_NOTIFICATION_ENTITY_DOC_ACTION,
    UPDATE_NOTIFICATION_ENTITY_DOC_ACTION,
    UPDATE_NOTIFICATION_STATE_DOC_ACTION
} from './notification-action.store';

export const adapter = createEntityAdapter<Notification>();

const INITIAL_STATE: NotificationState = {
    loadStatus: LoadStatus.INIT,
    lastUpdate: 0,
    list: adapter.getInitialState()
};

const reducer: ActionReducer<NotificationState, Action> = createReducer(
    INITIAL_STATE,
    on(UPDATE_NOTIFICATION_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values })),
    on(SET_ALL_NOTIFICATION_ENTITIES_DOC_ACTION, (state, { list }) => ({
        ...state,
        list: adapter.setAll(list, state.list)
    })),
    on(SET_NOTIFICATION_ENTITY_DOC_ACTION, (state, { item }) => ({
        ...state,
        list: adapter.setOne(item, state.list)
    })),
    on(UPDATE_NOTIFICATION_ENTITY_DOC_ACTION, (state, { item }) => ({
        ...state,
        list: adapter.updateOne(item, state.list)
    })),
    on(DELETE_NOTIFICATION_ENTITIES_DOC_ACTION, (state, { ids }) => ({
        ...state,
        list: adapter.removeMany(ids, state.list)
    })),
    on(DELETE_NOTIFICATION_ENTITY_DOC_ACTION, (state, { id }) => ({
        ...state,
        list: adapter.removeOne(id, state.list)
    }))
);

export const notificationEntitySelector = adapter.getSelectors();

export function notificationReducer(state: NotificationState | undefined, action: Action): NotificationState {
    return reducer(state, action);
}
