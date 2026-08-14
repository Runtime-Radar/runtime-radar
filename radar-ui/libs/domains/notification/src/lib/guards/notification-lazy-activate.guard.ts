import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { Observable, map, switchMap, take, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL } from '@cs/core';

import { NotificationState } from '../interfaces/state/notification-state.interface';
import {
    LOAD_NOTIFICATION_ENTITIES_TODO_ACTION,
    POLLING_LOAD_NOTIFICATION_ENTITIES_TODO_ACTION
} from '../stores/notification-action.store';
import { getNotificationLastUpdate, getNotificationLoadStatus } from '../stores/notification-selector.store';

const notificationLazyActivate = (): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<NotificationState>>(Store);

    return store.select(getNotificationLoadStatus).pipe(
        tap((status) => {
            if (status === LoadStatus.INIT) {
                store.dispatch(LOAD_NOTIFICATION_ENTITIES_TODO_ACTION());
            }
        }),
        switchMap(() =>
            store.select(getNotificationLastUpdate).pipe(
                take(1),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_NOTIFICATION_ENTITIES_TODO_ACTION());
                    }
                }),
                map(() => true)
            )
        )
    );
};

export const notificationLazyActivateGuard = () => notificationLazyActivate();

export const notificationLazyActivateChildGuard = () => notificationLazyActivate();
