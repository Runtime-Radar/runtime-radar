import { Action } from '@ngrx/store';
import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Router } from '@angular/router';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { NEVER, Observable, of } from 'rxjs';
import { catchError, concatMap, filter, map, switchMap, take, tap } from 'rxjs/operators';

import { ApiPathService } from '@cs/api';
import { I18nService } from '@cs/i18n';
import { ROLE_LOAD_DONE_EVENT_ACTION } from '@cs/domains/role';
import { LoadStatus, RouterName, CoreUtilsService as utils } from '@cs/core';

import { CLUSTER_CREATE_FRAGMENT } from '../constants/cluster.constant';
import { ClusterRequestService } from '../services/cluster-request.service';
import {
    CREATE_CLUSTER_ENTITY_TODO_ACTION,
    DELETE_CLUSTER_ENTITY_DOC_ACTION,
    DELETE_CLUSTER_ENTITY_TODO_ACTION,
    LOAD_CLUSTER_ENTITIES_TODO_ACTION,
    POLLING_LOAD_CLUSTER_ENTITIES_TODO_ACTION,
    SET_ALL_CLUSTER_ENTITIES_DOC_ACTION,
    SET_ALL_REGISTERED_CLUSTER_ENTITIES_DOC_ACTION,
    SET_CLUSTER_ENTITY_DOC_ACTION,
    SWITCH_CLUSTER_ENTITY_TODO_ACTION,
    SWITCH_CLUSTER_EVENT_ACTION,
    UPDATE_CLUSTER_ENTITY_DOC_ACTION,
    UPDATE_CLUSTER_ENTITY_TODO_ACTION,
    UPDATE_CLUSTER_STATE_DOC_ACTION,
    UPDATE_REGISTERED_CLUSTER_ENTITY_DOC_ACTION,
    VALIDATE_CLUSTER_HOST_TODO_ACTION
} from './cluster-action.store';
import { Cluster, RegisteredCluster } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class ClusterEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly router = inject(Router);
    private readonly toastService = inject(KbqToastService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterRequestService = inject(ClusterRequestService);
    private readonly i18nService = inject(I18nService);

    readonly loadClusters$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_CLUSTER_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.clusterRequestService.getClusters().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((list) => {
                if (list === undefined) {
                    return [
                        UPDATE_CLUSTER_STATE_DOC_ACTION({
                            loadStatus: LoadStatus.ERROR
                        })
                    ];
                }

                return [
                    SET_ALL_CLUSTER_ENTITIES_DOC_ACTION({ list }),
                    UPDATE_CLUSTER_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];
            })
        )
    );

    readonly loadRegisteredClusters$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(ROLE_LOAD_DONE_EVENT_ACTION),
            switchMap(() =>
                this.clusterRequestService.getRegisteredClusters().pipe(
                    take(1),
                    catchError(() => of([]))
                )
            ),
            map((list) => {
                const central: RegisteredCluster = {
                    id: utils.generateUuid(),
                    name: '',
                    own_cs_url: ''
                };

                return [...list, central];
            }),
            switchMap((list) => [
                SET_ALL_REGISTERED_CLUSTER_ENTITIES_DOC_ACTION({ list }),
                VALIDATE_CLUSTER_HOST_TODO_ACTION({ list })
            ])
        )
    );

    readonly validateClusterHost$: Observable<Action> = createEffect(
        () =>
            this.actions$.pipe(
                ofType(VALIDATE_CLUSTER_HOST_TODO_ACTION),
                switchMap(({ list }) =>
                    this.apiPathService.host$.pipe(
                        take(1),
                        map((host) => list.some((item) => item.own_cs_url === host))
                    )
                ),
                tap((isHostExist) => {
                    if (!isHostExist) {
                        this.apiPathService.setHost('');
                        this.toastService.show({
                            style: KbqToastStyle.Warning,
                            title: this.i18nService.translate('Common.Pseudo.Notification.HostValidationFailed')
                        });
                    }
                }),
                concatMap(() => NEVER)
            ),
        { dispatch: false }
    );

    readonly pollingLoadClusters$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_CLUSTER_ENTITIES_TODO_ACTION),
            switchMap(() => this.clusterRequestService.getClusters().pipe(take(1))),
            switchMap((list) => [
                SET_ALL_CLUSTER_ENTITIES_DOC_ACTION({ list }),
                UPDATE_CLUSTER_STATE_DOC_ACTION({ lastUpdate: this.dateAdapter.today().toMillis() })
            ])
        )
    );

    readonly pollingLoadRegisteredClusters$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_CLUSTER_ENTITIES_TODO_ACTION),
            switchMap(() => this.clusterRequestService.getRegisteredClusters().pipe(take(1))),
            map((list) => {
                const central: RegisteredCluster = {
                    id: utils.generateUuid(),
                    name: '',
                    own_cs_url: ''
                };

                return [...list, central];
            }),
            switchMap((list) => [
                SET_ALL_REGISTERED_CLUSTER_ENTITIES_DOC_ACTION({ list }),
                VALIDATE_CLUSTER_HOST_TODO_ACTION({ list })
            ])
        )
    );

    readonly createCluster$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_CLUSTER_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.clusterRequestService.createCluster(action.item).pipe(
                    take(1),
                    catchError(() => {
                        this.toastService.show({
                            style: KbqToastStyle.Warning,
                            title: this.i18nService.translate('Cluster.Pseudo.Notification.CreationFailed')
                        });

                        return of({} as Cluster);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) => SET_CLUSTER_ENTITY_DOC_ACTION({ item })),
            tap((action) => {
                this.router.navigate([RouterName.CLUSTERS, action.item.id], { fragment: CLUSTER_CREATE_FRAGMENT });
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Cluster.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly updateCluster$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_CLUSTER_ENTITY_TODO_ACTION),
            switchMap((action) => this.clusterRequestService.updateCluster(action.id, action.item).pipe(take(1))),
            filter((item) => !!item.id),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Cluster.Pseudo.Notification.Updated')
                });
            }),
            switchMap((item) => [
                UPDATE_CLUSTER_ENTITY_DOC_ACTION({
                    item: {
                        id: item.id,
                        changes: item
                    }
                }),
                UPDATE_REGISTERED_CLUSTER_ENTITY_DOC_ACTION({
                    item: {
                        id: item.id,
                        changes: {
                            name: item.name
                        }
                    }
                })
            ])
        )
    );

    readonly deleteCluster$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_CLUSTER_ENTITY_TODO_ACTION),
            switchMap((action) => this.clusterRequestService.deleteCluster(action.id).pipe(take(1))),
            filter((id) => !!id),
            map((id) => DELETE_CLUSTER_ENTITY_DOC_ACTION({ id })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Cluster.Pseudo.Notification.Deleted')
                });
            })
        )
    );

    readonly switchCluster$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_ENTITY_TODO_ACTION),
            tap(({ url }) => {
                this.apiPathService.setHost(url);
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Common.Pseudo.Notification.ClusterSwitched')
                });
            }),
            map(() => SWITCH_CLUSTER_EVENT_ACTION())
        )
    );
}
