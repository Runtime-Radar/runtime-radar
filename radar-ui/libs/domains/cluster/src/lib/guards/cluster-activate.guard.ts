import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { Observable, filter, map, switchMap, take, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL, RouterName } from '@cs/core';

import { ClusterState } from '../interfaces';
import {
    LOAD_CLUSTER_ENTITIES_TODO_ACTION,
    POLLING_LOAD_CLUSTER_ENTITIES_TODO_ACTION
} from '../stores/cluster-action.store';
import { getClusterLastUpdate, getClusterLoadStatus } from '../stores/cluster-selector.store';

const clusterActivate = (): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    const router = inject(Router);

    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<ClusterState>>(Store);

    return store.select(getClusterLoadStatus).pipe(
        tap((status) => {
            if (status === LoadStatus.INIT) {
                store.dispatch(LOAD_CLUSTER_ENTITIES_TODO_ACTION());
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
            store.select(getClusterLastUpdate).pipe(
                take(1),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_CLUSTER_ENTITIES_TODO_ACTION());
                    }
                }),
                map(() => isLoaded)
            )
        )
    );
};

export const clusterActivateGuard = () => clusterActivate();

export const clusterActivateChildGuard = () => clusterActivate();
