import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Injectable } from '@angular/core';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Observable, filter, of } from 'rxjs';
import { catchError, map, switchMap, take } from 'rxjs/operators';

import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';

import { KubeManagerRequestService } from '../services/kube-manager-request.service';
import { KubeManagerState } from '../interfaces';
import { getKubeManagerLoadStatus } from './kube-manager-selector.store';
import {
    LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION,
    LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION,
    SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION,
    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION
} from './kube-manager-action.store';

@Injectable({
    providedIn: 'root'
})
export class KubeManagerEffectStore {
    readonly loadPods$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.kubeManagerRequestService.getPods().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((list) => {
                if (list === undefined) {
                    return [
                        UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                            loadStatus: LoadStatus.ERROR
                        })
                    ];
                }

                return [
                    SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION({ list }),
                    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];
            })
        )
    );

    readonly reloadPods$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.store.select(getKubeManagerLoadStatus).pipe(take(1))),
            filter((status) => status !== LoadStatus.INIT),
            map(() => LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION())
        )
    );

    readonly setInProgressStatus$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION, LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION),
            map(() =>
                UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                    loadStatus: LoadStatus.IN_PROGRESS
                })
            )
        )
    );

    readonly loadPodsWithFilters$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION),
            switchMap(({ filters }) => this.kubeManagerRequestService.getPods(filters).pipe(take(1))),
            switchMap((list) => [
                SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION({ list }),
                UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                    loadStatus: LoadStatus.LOADED
                })
            ])
        )
    );

    constructor(
        private readonly actions$: Actions,
        private readonly dateAdapter: DateAdapter<DateTime>,
        private readonly kubeManagerRequestService: KubeManagerRequestService,
        private readonly store: Store<KubeManagerState>
    ) {}
}
