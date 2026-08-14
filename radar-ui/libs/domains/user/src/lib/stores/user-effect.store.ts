import { Action } from '@ngrx/store';
import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { HttpErrorResponse } from '@angular/common/http';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, of } from 'rxjs';
import { catchError, filter, map, switchMap, take, tap } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';
import { AuthStoreService, GetTokenResponse } from '@cs/domains/auth';

import { User } from '../interfaces';
import { UserRequestService } from '../services/user-request.service';
import {
    CREATE_USER_ENTITY_TODO_ACTION,
    DELETE_USER_ENTITY_DOC_ACTION,
    DELETE_USER_ENTITY_TODO_ACTION,
    LOAD_USER_ENTITIES_TODO_ACTION,
    POLLING_LOAD_USER_ENTITIES_TODO_ACTION,
    SET_ALL_USER_ENTITIES_DOC_ACTION,
    SET_USER_ENTITY_DOC_ACTION,
    UPDATE_USER_ENTITY_DOC_ACTION,
    UPDATE_USER_ENTITY_TODO_ACTION,
    UPDATE_USER_PASSWORD_EVENT_ACTION,
    UPDATE_USER_PASSWORD_TODO_ACTION,
    UPDATE_USER_STATE_DOC_ACTION
} from './user-action.store';

@Injectable({
    providedIn: 'root'
})
export class UserEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly authStoreService = inject(AuthStoreService);
    private readonly i18nService = inject(I18nService);
    private readonly userRequestService = inject(UserRequestService);

    readonly loadUsers$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_USER_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.userRequestService.getUsers().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((list) => {
                if (list === undefined) {
                    return [
                        UPDATE_USER_STATE_DOC_ACTION({
                            loadStatus: LoadStatus.ERROR
                        })
                    ];
                }

                return [
                    SET_ALL_USER_ENTITIES_DOC_ACTION({ list }),
                    UPDATE_USER_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];
            })
        )
    );

    readonly pollingLoadUsers$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_USER_ENTITIES_TODO_ACTION),
            switchMap(() => this.userRequestService.getUsers().pipe(take(1))),
            switchMap((list) => [
                SET_ALL_USER_ENTITIES_DOC_ACTION({ list }),
                UPDATE_USER_STATE_DOC_ACTION({ lastUpdate: this.dateAdapter.today().toMillis() })
            ])
        )
    );

    readonly createUser$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_USER_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.userRequestService.createUser(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handlePasswordErrors(error);
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.USER_ALREADY_EXIST) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('User.Pseudo.Notification.UserAlreadyExist')
                            });
                        }

                        return of({} as User);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) => SET_USER_ENTITY_DOC_ACTION({ item })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('User.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly updateUser$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_USER_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.userRequestService
                    .updateUser(action.id, {
                        email: action.item.email,
                        role_id: action.item.roleId
                    })
                    .pipe(take(1))
            ),
            filter((item) => !!item.id),
            map((item) =>
                UPDATE_USER_ENTITY_DOC_ACTION({
                    item: {
                        id: item.id,
                        changes: item
                    }
                })
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('User.Pseudo.Notification.Updated')
                });
            })
        )
    );

    readonly updatePassword$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_USER_ENTITY_TODO_ACTION),
            filter((action) => !!action.item.password),
            switchMap((action) =>
                this.userRequestService.updatePassword(action.id, action.item.password).pipe(
                    take(1),
                    map((tokens) => ({
                        id: action.id,
                        tokens
                    })),
                    catchError((error: HttpErrorResponse) => {
                        this.handlePasswordErrors(error);

                        return of({
                            id: action.id,
                            tokens: {} as GetTokenResponse
                        });
                    })
                )
            ),
            filter(({ tokens }) => !!tokens.access_token),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('User.Pseudo.Notification.PasswordUpdated')
                });
            }),
            switchMap(({ id, tokens }) =>
                this.authStoreService.credentials$.pipe(
                    take(1),
                    map((credentials) => ({
                        isTokenUpdateNeeded: id === credentials.userId,
                        tokens
                    }))
                )
            ),
            filter(({ isTokenUpdateNeeded }) => isTokenUpdateNeeded),
            map(({ tokens }) => UPDATE_USER_PASSWORD_EVENT_ACTION({ tokens }))
        )
    );

    readonly updateOwnPassword$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_USER_PASSWORD_TODO_ACTION),
            switchMap((action) =>
                this.authStoreService.credentials$.pipe(
                    take(1),
                    map((credentials) => ({
                        userId: credentials.userId,
                        password: action.password
                    }))
                )
            ),
            switchMap(({ userId, password }) =>
                this.userRequestService.updatePassword(userId, password).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        this.handlePasswordErrors(error);

                        return of({} as GetTokenResponse);
                    })
                )
            ),
            filter((tokens) => !!tokens.access_token),
            map((tokens) => UPDATE_USER_PASSWORD_EVENT_ACTION({ tokens })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('User.Pseudo.Notification.PasswordUpdated')
                });
            })
        )
    );

    readonly deleteUser$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_USER_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.userRequestService.deleteUser(action.id).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.LAST_ADMIN_REMOVING_DENIED) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('User.Pseudo.Notification.LastAdminRemovingDenied')
                            });
                        }

                        return of('');
                    })
                )
            ),
            filter((id) => !!id),
            map((id) => DELETE_USER_ENTITY_DOC_ACTION({ id })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('User.Pseudo.Notification.Deleted')
                });
            })
        )
    );

    private handlePasswordErrors(error: HttpErrorResponse) {
        if (apiUtils.getReasonCode(error) === ApiErrorCode.PASSWORD_FOUND_IN_PASS_LIST) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('User.Pseudo.Notification.PasswordFoundInPassList')
            });
        } else if (apiUtils.getReasonCode(error) === ApiErrorCode.PASSWORD_HAS_BEEN_USED_BEFORE) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('User.Pseudo.Notification.PasswordHasBeenUsedBefore')
            });
        } else if (apiUtils.getReasonCode(error) === ApiErrorCode.USER_DOES_NOT_EXIST) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('User.Pseudo.Notification.UserDoesNotExist')
            });
        } else if (apiUtils.getReasonCode(error) === ApiErrorCode.PASSWORD_UPDATE_FAILED) {
            this.toastService.show({
                style: KbqToastStyle.Warning,
                title: this.i18nService.translate('User.Pseudo.Notification.PasswordUpdateFailed')
            });
        }
    }
}
