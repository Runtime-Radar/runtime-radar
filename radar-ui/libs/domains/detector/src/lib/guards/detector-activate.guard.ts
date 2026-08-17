import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { take } from 'rxjs/operators';
import { Observable, filter, map, switchMap, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL, RouterName } from '@cs/core';

import { getDetectorConfig } from '../stores/detector-selector.store';
import { DetectorState, DetectorType } from '../interfaces';
import {
    LOAD_DETECTOR_ENTITIES_TODO_ACTION,
    POLLING_LOAD_DETECTOR_ENTITIES_TODO_ACTION
} from '../stores/detector-action.store';

const detectorActivate = (type: DetectorType): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    const router = inject<Router>(Router);

    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<DetectorState>>(Store);

    return store.select(getDetectorConfig(type)).pipe(
        map((config) => (config ? config.loadStatus : LoadStatus.INIT)),
        tap((status) => {
            if (status === LoadStatus.INIT) {
                store.dispatch(LOAD_DETECTOR_ENTITIES_TODO_ACTION({ detectorType: type }));
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
            store.select(getDetectorConfig(type)).pipe(
                take(1),
                map((config) => (config ? config.lastUpdate : 0)),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_DETECTOR_ENTITIES_TODO_ACTION({ detectorType: type }));
                    }
                }),
                map(() => isLoaded)
            )
        )
    );
};

export const detectorRuntimeActivateGuard = () => detectorActivate(DetectorType.RUNTIME);
