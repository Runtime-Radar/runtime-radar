import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Router } from '@angular/router';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { Observable, filter, map, switchMap, take, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL, RouterName } from '@cs/core';

import { RuleState } from '../interfaces/state/rule-state.interface';
import { LOAD_RULE_ENTITIES_TODO_ACTION, POLLING_LOAD_RULE_ENTITIES_TODO_ACTION } from '../stores/rule-action.store';
import { getRuleLastUpdate, getRuleLoadStatus } from '../stores/rule-selector.store';

const ruleActivate = (): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    const router = inject(Router);

    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<RuleState>>(Store);

    return store.select(getRuleLoadStatus).pipe(
        tap((status) => {
            if (status === LoadStatus.INIT) {
                store.dispatch(LOAD_RULE_ENTITIES_TODO_ACTION());
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
            store.select(getRuleLastUpdate).pipe(
                take(1),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_RULE_ENTITIES_TODO_ACTION());
                    }
                }),
                map(() => isLoaded)
            )
        )
    );
};

export const ruleActivateGuard = () => ruleActivate();

export const ruleActivateChildGuard = () => ruleActivate();
