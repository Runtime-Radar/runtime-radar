import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { HttpErrorResponse } from '@angular/common/http';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, forkJoin, of } from 'rxjs';
import { catchError, concatMap, filter, map, scan, switchMap, take, tap } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';

import { IntegrationEmailRequestService } from '../services/integration-email-request.service';
import { IntegrationState } from '../interfaces/state/integration-state.interface';
import { IntegrationSyslogRequestService } from '../services/integration-syslog-request.service';
import { IntegrationWebhookRequestService } from '../services/integration-webhook-request.service';
import { getIntegrationLoadStatus } from './integration-selector.store';
import {
    APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION,
    CREATE_EMAIL_INTEGRATION_ENTITY_TODO_ACTION,
    CREATE_SYSLOG_INTEGRATION_ENTITY_TODO_ACTION,
    CREATE_WEBHOOK_INTEGRATION_ENTITY_TODO_ACTION,
    DELETE_CONNECTED_INTEGRATION_EVENT_ACTION,
    DELETE_EMAIL_INTEGRATION_ENTITY_TODO_ACTION,
    DELETE_INTEGRATION_ENTITY_DOC_ACTION,
    DELETE_SYSLOG_INTEGRATION_ENTITY_TODO_ACTION,
    DELETE_WEBHOOK_INTEGRATION_ENTITY_TODO_ACTION,
    LOAD_INTEGRATION_ENTITIES_TODO_ACTION,
    POLLING_LOAD_INTEGRATION_ENTITIES_TODO_ACTION,
    SET_ALL_INTEGRATION_ENTITIES_DOC_ACTION,
    SET_INTEGRATION_ENTITY_DOC_ACTION,
    SET_INTEGRATION_LOADED_TYPE_DOC_ACTION,
    UPDATE_EMAIL_INTEGRATION_ENTITY_TODO_ACTION,
    UPDATE_INTEGRATION_ENTITY_DOC_ACTION,
    UPDATE_INTEGRATION_STATE_DOC_ACTION,
    UPDATE_SYSLOG_INTEGRATION_ENTITY_TODO_ACTION,
    UPDATE_WEBHOOK_INTEGRATION_ENTITY_TODO_ACTION
} from './integration-action.store';
import { IntegrationEmail, IntegrationSyslog, IntegrationType, IntegrationWebhook } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class IntegrationEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly i18nService = inject(I18nService);
    private readonly integrationEmailRequestService = inject(IntegrationEmailRequestService);
    private readonly integrationSyslogRequestService = inject(IntegrationSyslogRequestService);
    private readonly integrationWebhookRequestService = inject(IntegrationWebhookRequestService);
    private readonly store = inject<Store<IntegrationState>>(Store);

    readonly loadEmailIntegrations$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_INTEGRATION_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.integrationEmailRequestService.getEmailIntegrations().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((email) => {
                if (email === undefined) {
                    return [APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION({ isLoaded: false })];
                }

                return [
                    APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION({ isLoaded: true }),
                    SET_INTEGRATION_LOADED_TYPE_DOC_ACTION({ integrationType: IntegrationType.EMAIL }),
                    SET_ALL_INTEGRATION_ENTITIES_DOC_ACTION({ email })
                ];
            })
        )
    );

    readonly loadSyslogIntegrations$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_INTEGRATION_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.integrationSyslogRequestService.getSyslogIntegrations().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((syslog) => {
                if (syslog === undefined) {
                    return [APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION({ isLoaded: false })];
                }

                return [
                    APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION({ isLoaded: true }),
                    SET_INTEGRATION_LOADED_TYPE_DOC_ACTION({ integrationType: IntegrationType.SYSLOG }),
                    SET_ALL_INTEGRATION_ENTITIES_DOC_ACTION({ syslog })
                ];
            })
        )
    );

    readonly loadWebhookIntegrations$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_INTEGRATION_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.integrationWebhookRequestService.getWebhookIntegrations().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((webhook) => {
                if (webhook === undefined) {
                    return [APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION({ isLoaded: false })];
                }

                return [
                    APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION({ isLoaded: true }),
                    SET_INTEGRATION_LOADED_TYPE_DOC_ACTION({ integrationType: IntegrationType.WEBHOOK }),
                    SET_ALL_INTEGRATION_ENTITIES_DOC_ACTION({ webhook })
                ];
            })
        )
    );

    readonly applyLoadStatus$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(APPLY_INTEGRATION_LOAD_STATUS_TODO_ACTION),
            scan((statuses, { isLoaded }) => [...statuses, isLoaded], [] as boolean[]),
            filter((statuses) => statuses.length === Object.keys(IntegrationType).length),
            map((statuses) => {
                const isSuccess = statuses.some((item) => !!item);

                return UPDATE_INTEGRATION_STATE_DOC_ACTION({
                    loadStatus: isSuccess ? LoadStatus.LOADED : LoadStatus.ERROR,
                    lastUpdate: isSuccess ? this.dateAdapter.today().toMillis() : 0
                });
            })
        )
    );

    readonly reloadIntegrations$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.store.select(getIntegrationLoadStatus).pipe(take(1))),
            filter((status) => status !== LoadStatus.INIT),
            map(() => LOAD_INTEGRATION_ENTITIES_TODO_ACTION())
        )
    );

    readonly pollingLoadIntegrations$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_INTEGRATION_ENTITIES_TODO_ACTION),
            concatMap(() =>
                forkJoin([
                    this.integrationEmailRequestService.getEmailIntegrations().pipe(take(1)),
                    this.integrationSyslogRequestService.getSyslogIntegrations().pipe(take(1)),
                    this.integrationWebhookRequestService.getWebhookIntegrations().pipe(take(1))
                ])
            ),
            switchMap(([email, syslog, webhook]) => [
                UPDATE_INTEGRATION_STATE_DOC_ACTION({ lastUpdate: this.dateAdapter.today().toMillis() }),
                SET_ALL_INTEGRATION_ENTITIES_DOC_ACTION({
                    email,
                    syslog,
                    webhook
                })
            ])
        )
    );

    readonly createEmailIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_EMAIL_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.integrationEmailRequestService.createEmailIntegration(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handleWarningToastMessages(error);

                        return of({} as IntegrationEmail);
                    })
                )
            ),
            filter((email) => !!email.id),
            map((email) => SET_INTEGRATION_ENTITY_DOC_ACTION({ email })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly createSyslogIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_SYSLOG_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.integrationSyslogRequestService.createSyslogIntegration(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handleWarningToastMessages(error);

                        return of({} as IntegrationSyslog);
                    })
                )
            ),
            filter((syslog) => !!syslog.id),
            map((syslog) => SET_INTEGRATION_ENTITY_DOC_ACTION({ syslog })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly createWebhookIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_WEBHOOK_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.integrationWebhookRequestService.createWebhookIntegration(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handleWarningToastMessages(error);

                        return of({} as IntegrationWebhook);
                    })
                )
            ),
            filter((webhook) => !!webhook.id),
            map((webhook) => SET_INTEGRATION_ENTITY_DOC_ACTION({ webhook })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly updateEmailIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_EMAIL_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.integrationEmailRequestService.updateEmailIntegration(action.id, action.item).pipe(take(1))
            ),
            filter((email) => !!email.id),
            map((email) =>
                UPDATE_INTEGRATION_ENTITY_DOC_ACTION({
                    email: {
                        id: email.id,
                        changes: email
                    }
                })
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Updated')
                });
            })
        )
    );

    readonly updateSyslogIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_SYSLOG_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.integrationSyslogRequestService.updateSyslogIntegration(action.id, action.item).pipe(take(1))
            ),
            filter((syslog) => !!syslog.id),
            map((syslog) =>
                UPDATE_INTEGRATION_ENTITY_DOC_ACTION({
                    syslog: {
                        id: syslog.id,
                        changes: syslog
                    }
                })
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Updated')
                });
            })
        )
    );

    readonly updateWebhookIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_WEBHOOK_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.integrationWebhookRequestService.updateWebhookIntegration(action.id, action.item).pipe(take(1))
            ),
            filter((webhook) => !!webhook.id),
            map((webhook) =>
                UPDATE_INTEGRATION_ENTITY_DOC_ACTION({
                    webhook: {
                        id: webhook.id,
                        changes: webhook
                    }
                })
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Updated')
                });
            })
        )
    );

    readonly deleteEmailIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_EMAIL_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((item) =>
                this.integrationEmailRequestService.deleteEmailIntegration(item.id).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handleWarningToastMessages(error);

                        return of('');
                    })
                )
            ),
            filter((emailId) => !!emailId),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Deleted')
                });
            }),
            switchMap((emailId) => [
                DELETE_INTEGRATION_ENTITY_DOC_ACTION({ emailId }),
                DELETE_CONNECTED_INTEGRATION_EVENT_ACTION({ integrationId: emailId })
            ])
        )
    );

    readonly deleteSyslogIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_SYSLOG_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((item) =>
                this.integrationSyslogRequestService.deleteSyslogIntegration(item.id).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handleWarningToastMessages(error);

                        return of('');
                    })
                )
            ),
            filter((syslogId) => !!syslogId),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Deleted')
                });
            }),
            switchMap((syslogId) => [
                DELETE_INTEGRATION_ENTITY_DOC_ACTION({ syslogId }),
                DELETE_CONNECTED_INTEGRATION_EVENT_ACTION({ integrationId: syslogId })
            ])
        )
    );

    readonly deleteWebhookIntegration$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_WEBHOOK_INTEGRATION_ENTITY_TODO_ACTION),
            switchMap((item) =>
                this.integrationWebhookRequestService.deleteWebhookIntegration(item.id).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handleWarningToastMessages(error);

                        return of('');
                    })
                )
            ),
            filter((webhookId) => !!webhookId),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Integration.Pseudo.Notification.Deleted')
                });
            }),
            switchMap((webhookId) => [
                DELETE_INTEGRATION_ENTITY_DOC_ACTION({ webhookId }),
                DELETE_CONNECTED_INTEGRATION_EVENT_ACTION({ integrationId: webhookId })
            ])
        )
    );

    private handleWarningToastMessages(error: HttpErrorResponse) {
        if (apiUtils.getReasonCode(error) === ApiErrorCode.NOTIFICATION_IN_USE) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('Integration.Pseudo.Notification.IntegrationInUse')
            });
        }

        if (apiUtils.getReasonCode(error) === ApiErrorCode.NAME_MUST_BE_UNIQUE) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('Integration.Pseudo.Notification.IntegrationNameMustBeUnique')
            });
        }

        if (apiUtils.getReasonCode(error) === ApiErrorCode.INTEGRATION_INACCESSIBLE) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('Integration.Pseudo.Notification.IntegrationInaccessible')
            });
        }
    }
}
