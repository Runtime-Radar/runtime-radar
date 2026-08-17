import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { HttpErrorResponse } from '@angular/common/http';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, of } from 'rxjs';
import { catchError, filter, map, switchMap, take, tap } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { SIGN_OUT_EVENT_ACTION } from '@cs/domains/auth';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';

import { TokenRequestService } from '../services/token-request.service';
import { getTokenLoadStatus } from './token-selector.store';
import {
    CREATE_TOKEN_ENTITY_TODO_ACTION,
    DELETE_ALL_TOKEN_ENTITIES_DOC_ACTION,
    DELETE_TOKEN_ENTITY_DOC_ACTION,
    DELETE_TOKEN_ENTITY_TODO_ACTION,
    LOAD_TOKEN_ENTITIES_TODO_ACTION,
    POLLING_LOAD_TOKEN_ENTITIES_TODO_ACTION,
    REVOKE_TOKEN_ENTITIES_TODO_ACTION,
    SET_ALL_TOKEN_ENTITIES_DOC_ACTION,
    SET_TOKEN_ENTITY_DOC_ACTION,
    UPDATE_TOKEN_STATE_DOC_ACTION,
    UPSERT_TOKEN_ENTITIES_DOC_ACTION
} from './token-action.store';
import { Token, TokenState } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class TokenEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly i18nService = inject(I18nService);
    private readonly store = inject<Store<TokenState>>(Store);
    private readonly tokenRequestService = inject(TokenRequestService);

    readonly loadTokens$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_TOKEN_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.tokenRequestService.getTokens().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((list) => {
                if (list === undefined) {
                    return [
                        UPDATE_TOKEN_STATE_DOC_ACTION({
                            loadStatus: LoadStatus.ERROR
                        })
                    ];
                }

                return [
                    SET_ALL_TOKEN_ENTITIES_DOC_ACTION({ list }),
                    UPDATE_TOKEN_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];
            })
        )
    );

    readonly reloadTokens$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.store.select(getTokenLoadStatus).pipe(take(1))),
            filter((status) => status !== LoadStatus.INIT),
            map(() => LOAD_TOKEN_ENTITIES_TODO_ACTION())
        )
    );

    readonly pollingLoadTokens$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_TOKEN_ENTITIES_TODO_ACTION),
            switchMap(() => this.tokenRequestService.getTokens().pipe(take(1))),
            switchMap((list) => [
                SET_ALL_TOKEN_ENTITIES_DOC_ACTION({ list }),
                UPDATE_TOKEN_STATE_DOC_ACTION({ lastUpdate: this.dateAdapter.today().toMillis() })
            ])
        )
    );

    readonly createToken$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_TOKEN_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.tokenRequestService.createToken(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.NAME_MUST_BE_UNIQUE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Token.Pseudo.Notification.NameMustBeUnique')
                            });
                        }

                        return of({} as Token);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) => SET_TOKEN_ENTITY_DOC_ACTION({ item })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Token.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly deleteToken$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_TOKEN_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.tokenRequestService.deleteToken(action.id).pipe(
                    take(1),
                    catchError(() => {
                        this.toastService.show({
                            style: KbqToastStyle.Warning,
                            title: this.i18nService.translate('Token.Pseudo.Notification.DeleteFailed')
                        });

                        return of('');
                    })
                )
            ),
            filter((id) => !!id),
            map((id) => DELETE_TOKEN_ENTITY_DOC_ACTION({ id })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Token.Pseudo.Notification.Deleted')
                });
            })
        )
    );

    readonly revokeTokens$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(REVOKE_TOKEN_ENTITIES_TODO_ACTION),
            switchMap(() => this.tokenRequestService.revokeTokens().pipe(take(1))),
            filter((isValid) => isValid),
            switchMap(() => this.tokenRequestService.getTokens().pipe(take(1))),
            map((list) => UPSERT_TOKEN_ENTITIES_DOC_ACTION({ list })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Token.Pseudo.Notification.Revoked')
                });
            })
        )
    );

    readonly clearState$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SIGN_OUT_EVENT_ACTION),
            switchMap(() => [
                DELETE_ALL_TOKEN_ENTITIES_DOC_ACTION(),
                UPDATE_TOKEN_STATE_DOC_ACTION({
                    loadStatus: LoadStatus.INIT,
                    lastUpdate: 0
                })
            ])
        )
    );
}
