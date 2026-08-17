import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { Observable, combineLatest, filter, map, switchMap, take, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL, RouterName } from '@cs/core';

import { IntegrationState } from '../interfaces/state/integration-state.interface';
import {
    LOAD_INTEGRATION_ENTITIES_TODO_ACTION,
    POLLING_LOAD_INTEGRATION_ENTITIES_TODO_ACTION
} from '../stores/integration-action.store';
import {
    getIntegrationLastUpdate,
    getIntegrationLoadStatus,
    getIntegrationLoadedTypes
} from '../stores/integration-selector.store';

const integrationActivate = (): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    const router = inject(Router);

    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<IntegrationState>>(Store);

    return combineLatest([store.select(getIntegrationLoadStatus), store.select(getIntegrationLoadedTypes)]).pipe(
        tap(([status, types]) => {
            if (status === LoadStatus.INIT && !types.length) {
                store.dispatch(LOAD_INTEGRATION_ENTITIES_TODO_ACTION());
            }
        }),
        filter(([status, types]) => !!types.length || [LoadStatus.LOADED, LoadStatus.ERROR].includes(status)),
        map(([status, types]) => !!types.length || status === LoadStatus.LOADED),
        tap((isLoaded) => {
            if (!isLoaded) {
                router.navigate([RouterName.ERROR]);
            }
        }),
        switchMap((isLoaded) =>
            store.select(getIntegrationLastUpdate).pipe(
                take(1),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_INTEGRATION_ENTITIES_TODO_ACTION());
                    }
                }),
                map(() => isLoaded)
            )
        )
    );
};

export const integrationActivateGuard = () => integrationActivate();

export const integrationActivateChildGuard = () => integrationActivate();
