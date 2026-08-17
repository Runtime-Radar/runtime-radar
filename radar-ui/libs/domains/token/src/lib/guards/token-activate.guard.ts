import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { Observable, filter, map, switchMap, take, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL, RouterName } from '@cs/core';

import { TokenState } from '../interfaces';
import { LOAD_TOKEN_ENTITIES_TODO_ACTION, POLLING_LOAD_TOKEN_ENTITIES_TODO_ACTION } from '../stores/token-action.store';
import { getTokenLastUpdate, getTokenLoadStatus } from '../stores/token-selector.store';

const tokenActivate = (): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    const router = inject(Router);

    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<TokenState>>(Store);

    return store.select(getTokenLoadStatus).pipe(
        tap((status) => {
            if (status === LoadStatus.INIT) {
                store.dispatch(LOAD_TOKEN_ENTITIES_TODO_ACTION());
            }
        }),
        filter((status) => [LoadStatus.LOADED, LoadStatus.ERROR].includes(status)),
        map((status) => status === LoadStatus.LOADED),
        tap((isLoaded) => {
            if (!isLoaded) {
                router.navigate([RouterName.ERROR]);
            }
        }),
        switchMap((isLoaded) =>
            store.select(getTokenLastUpdate).pipe(
                take(1),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_TOKEN_ENTITIES_TODO_ACTION());
                    }
                }),
                map(() => isLoaded)
            )
        )
    );
};

export const tokenActivateGuard = () => tokenActivate();

export const tokenActivateChildGuard = () => tokenActivate();
