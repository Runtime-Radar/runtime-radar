import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { HttpErrorResponse } from '@angular/common/http';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, of } from 'rxjs';
import { catchError, filter, map, switchMap, take, tap } from 'rxjs/operators';

import { DELETE_CONNECTED_INTEGRATION_EVENT_ACTION } from '@cs/domains/integration';
import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';

import { Notification } from '../interfaces/contract/notification-contract.interface';
import { NotificationRequestService } from '../services/notification-request.service';
import { NotificationState } from '../interfaces/state/notification-state.interface';
import {
    CREATE_NOTIFICATION_ENTITY_TODO_ACTION,
    DELETE_NOTIFICATION_ENTITIES_DOC_ACTION,
    DELETE_NOTIFICATION_ENTITY_DOC_ACTION,
    DELETE_NOTIFICATION_ENTITY_TODO_ACTION,
    LOAD_NOTIFICATION_ENTITIES_TODO_ACTION,
    POLLING_LOAD_NOTIFICATION_ENTITIES_TODO_ACTION,
    SET_ALL_NOTIFICATION_ENTITIES_DOC_ACTION,
    SET_NOTIFICATION_ENTITY_DOC_ACTION,
    UPDATE_NOTIFICATION_ENTITY_DOC_ACTION,
    UPDATE_NOTIFICATION_ENTITY_TODO_ACTION,
    UPDATE_NOTIFICATION_STATE_DOC_ACTION
} from './notification-action.store';
import { getNotificationLoadStatus, getNotificationsByIntegrationId } from './notification-selector.store';

@Injectable({
    providedIn: 'root'
})
export class NotificationEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly i18nService = inject(I18nService);
    private readonly notificationRequestService = inject(NotificationRequestService);
    private readonly store = inject<Store<NotificationState>>(Store);

    readonly loadNotifications$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_NOTIFICATION_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.notificationRequestService.getNotifications().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((list) => {
                if (list === undefined) {
                    return [
                        UPDATE_NOTIFICATION_STATE_DOC_ACTION({
                            loadStatus: LoadStatus.ERROR
                        })
                    ];
                }

                return [
                    SET_ALL_NOTIFICATION_ENTITIES_DOC_ACTION({ list }),
                    UPDATE_NOTIFICATION_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];
            })
        )
    );

    readonly reloadNotifications$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.store.select(getNotificationLoadStatus).pipe(take(1))),
            filter((status) => status !== LoadStatus.INIT),
            map(() => LOAD_NOTIFICATION_ENTITIES_TODO_ACTION())
        )
    );

    readonly pollingLoadNotifications$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_NOTIFICATION_ENTITIES_TODO_ACTION),
            switchMap(() => this.notificationRequestService.getNotifications().pipe(take(1))),
            switchMap((list) => [
                SET_ALL_NOTIFICATION_ENTITIES_DOC_ACTION({ list }),
                UPDATE_NOTIFICATION_STATE_DOC_ACTION({ lastUpdate: this.dateAdapter.today().toMillis() })
            ])
        )
    );

    readonly createNotification$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_NOTIFICATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.notificationRequestService.createNotification(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.NAME_MUST_BE_UNIQUE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Integration.Pseudo.Notification.NameMustBeUnique')
                            });
                        }

                        if (apiUtils.getReasonCode(error) === ApiErrorCode.EMAIL_SERVER_INACCESSIBLE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate(
                                    'Integration.Pseudo.Notification.EmailServerInaccessible'
                                )
                            });
                        }

                        return of({} as Notification);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) => SET_NOTIFICATION_ENTITY_DOC_ACTION({ item })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.RecipientsCreated')
                });
            })
        )
    );

    readonly updateNotification$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_NOTIFICATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.notificationRequestService.updateNotification(action.id, action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.NAME_MUST_BE_UNIQUE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Integration.Pseudo.Notification.NameMustBeUnique')
                            });
                        }

                        if (apiUtils.getReasonCode(error) === ApiErrorCode.EMAIL_SERVER_INACCESSIBLE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate(
                                    'Integration.Pseudo.Notification.EmailServerInaccessible'
                                )
                            });
                        }

                        return of({} as Notification);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) =>
                UPDATE_NOTIFICATION_ENTITY_DOC_ACTION({
                    item: {
                        id: item.id,
                        changes: item
                    }
                })
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.RecipientsUpdated')
                });
            })
        )
    );

    readonly deleteNotification$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_NOTIFICATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.notificationRequestService.deleteNotification(action.id).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.NOTIFICATION_IN_USE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Integration.Pseudo.Notification.NotificationInUse')
                            });
                        }

                        return of('');
                    })
                )
            ),
            filter((id) => !!id),
            map((id) => DELETE_NOTIFICATION_ENTITY_DOC_ACTION({ id })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.RecipientsDeleted')
                });
            })
        )
    );

    readonly deleteConnectedNotifications$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_CONNECTED_INTEGRATION_EVENT_ACTION),
            switchMap((action) =>
                this.store.select(getNotificationsByIntegrationId(action.integrationId)).pipe(take(1))
            ),
            map((notifications) => notifications.map((item) => item.id)),
            map((ids) => DELETE_NOTIFICATION_ENTITIES_DOC_ACTION({ ids }))
        )
    );
}
