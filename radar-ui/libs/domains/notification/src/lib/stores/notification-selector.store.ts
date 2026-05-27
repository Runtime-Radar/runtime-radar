import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { RuleType } from '@cs/domains/rule';

import { Notification } from '../interfaces/contract/notification-contract.interface';
import { NotificationEntityState, NotificationState } from '../interfaces/state/notification-state.interface';
import { notificationEntitySelector, notificationReducer } from './notification-reducer.store';

export const NOTIFICATION_DOMAIN_KEY = 'notification';

export interface NotificationDomainState {
    readonly domain: NotificationState;
}

const RULE_EVENT_TYPE_RELATIONS = new Map<RuleType, string>([[RuleType.TYPE_RUNTIME, 'runtime_event']]);

const selectNotificationDomainState = createFeatureSelector<NotificationDomainState>(NOTIFICATION_DOMAIN_KEY);
const selectNotificationState = createSelector(
    selectNotificationDomainState,
    (state: NotificationDomainState) => state.domain
);
const selectNotificationEntityState = createSelector(selectNotificationState, (state: NotificationState) => state.list);

export const getNotificationLoadStatus = createSelector(
    selectNotificationState,
    (state: NotificationState) => state.loadStatus
);

export const getNotificationLastUpdate = createSelector(
    selectNotificationState,
    (state: NotificationState) => state.lastUpdate
);

export const getNotifications = createSelector(selectNotificationEntityState, (state: NotificationEntityState) =>
    notificationEntitySelector.selectAll(state)
);

/* eslint @typescript-eslint/no-unnecessary-type-assertion: "off" */
export const getNotificationsByIntegrationId = (id: string) =>
    createSelector(
        selectNotificationEntityState,
        (state: NotificationEntityState) =>
            Object.values(state.entities).filter(
                (item) => item !== undefined && item.integration_id === id
            ) as Notification[]
    );

/* eslint @typescript-eslint/no-unnecessary-type-assertion: "off" */
export const getNotificationsByEventType = (type: RuleType) =>
    createSelector(
        selectNotificationEntityState,
        (state: NotificationEntityState) =>
            Object.values(state.entities).filter(
                (item) => item !== undefined && item.event_type === RULE_EVENT_TYPE_RELATIONS.get(type)
            ) as Notification[]
    );

export const notificationDomainReducer: ActionReducerMap<NotificationDomainState> = {
    domain: notificationReducer
};
