import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Store } from '@ngrx/store';
import { inject } from '@angular/core';
import { Observable, map, switchMap, take, tap } from 'rxjs';

import { LoadStatus, POLLING_INTERVAL } from '@cs/core';

import { RuleState } from '../interfaces/state/rule-state.interface';
import { LOAD_RULE_ENTITIES_TODO_ACTION, POLLING_LOAD_RULE_ENTITIES_TODO_ACTION } from '../stores/rule-action.store';
import { getRuleLastUpdate, getRuleLoadStatus } from '../stores/rule-selector.store';

const ruleLazyActivate = (): Observable<boolean> => {
    const dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);

    const pollingInterval = inject(POLLING_INTERVAL);
    const store = inject<Store<RuleState>>(Store);

    return store.select(getRuleLoadStatus).pipe(
        tap((status) => {
            if (status === LoadStatus.INIT) {
                store.dispatch(LOAD_RULE_ENTITIES_TODO_ACTION());
            }
        }),
        switchMap(() =>
            store.select(getRuleLastUpdate).pipe(
                take(1),
                tap((lastUpdate) => {
                    const nextUpdate = dateAdapter.today().toMillis() - pollingInterval;
                    if (lastUpdate && nextUpdate > lastUpdate) {
                        store.dispatch(POLLING_LOAD_RULE_ENTITIES_TODO_ACTION());
                    }
                }),
                map(() => true)
            )
        )
    );
};

export const ruleLazyActivateGuard = () => ruleLazyActivate();

export const ruleLazyActivateChildGuard = () => ruleLazyActivate();
