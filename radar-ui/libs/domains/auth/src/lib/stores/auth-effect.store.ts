import { HttpErrorResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { jwtDecode } from 'jwt-decode';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { NEVER, Observable, of } from 'rxjs';
import { catchError, concatMap, filter, map, switchMap, take, tap } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { ROLE_LOAD_DONE_EVENT_ACTION } from '@cs/domains/role';
import { UPDATE_USER_PASSWORD_EVENT_ACTION } from '@cs/domains/user';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';
import { LoadStatus, RouterName } from '@cs/core';

import { AUTH_DEFAULT_ORIGIN_PATH } from '../constants/auth.constant';
import { AuthJwtData } from '../interfaces/contract/auth-jwt-contract.interface';
import { AuthLocalStorageService } from '../services/auth-local-storage.services';
import { AuthRequestService } from '../services/auth-request.service';
import { getAuthOriginPath } from './auth-selector.store';
import {
    ALLOW_AUTH_EVENT_ACTION,
    APPLY_AUTH_TOKENS_TODO_ACTION,
    EXPIRE_AUTH_TOKENS_TODO_ACTION,
    EXPIRE_PASSWORD_TODO_ACTION,
    GET_LOCATION_PATH_TODO_ACTION,
    REDIRECT_TO_SWITCH_ROUTE_TODO_ACTION,
    RESET_AUTH_CREDENTIALS_DOC_ACTION,
    SIGN_IN_TODO_ACTION,
    SIGN_OUT_EVENT_ACTION,
    SIGN_OUT_TODO_ACTION,
    SUCCESS_SIGN_IN_TODO_ACTION,
    UPDATE_AUTH_STATE_DOC_ACTION
} from './auth-action.store';
import { AuthCredentials, AuthState, AuthTokens } from '../interfaces/state/auth-state.interface';

@Injectable({
    providedIn: 'root'
})
export class AuthEffectStore {
    readonly getLocationPath$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(GET_LOCATION_PATH_TODO_ACTION),
            map(({ location }) => location.pathname + location.search),
            map((path) =>
                UPDATE_AUTH_STATE_DOC_ACTION({
                    originPath: path.indexOf(RouterName.SIGN_IN) === 1 ? AUTH_DEFAULT_ORIGIN_PATH : path
                })
            )
        )
    );

    readonly applyTokens$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(APPLY_AUTH_TOKENS_TODO_ACTION),
            map(() => this.authLocalStorageService.getTokens()),
            map(({ accessToken }) => {
                const credentials = this.getCredentialsFromToken(accessToken);
                if (credentials) {
                    return SUCCESS_SIGN_IN_TODO_ACTION(credentials);
                }

                this.authLocalStorageService.removeTokens();

                return RESET_AUTH_CREDENTIALS_DOC_ACTION();
            })
        )
    );

    readonly updateTokens$: Observable<Action> = createEffect(
        () =>
            this.actions$.pipe(
                ofType(UPDATE_USER_PASSWORD_EVENT_ACTION),
                tap((action) => {
                    this.authLocalStorageService.setTokens(
                        `${action.tokens.token_type} ${action.tokens.access_token}`,
                        `${action.tokens.token_type} ${action.tokens.refresh_token}`
                    );
                }),
                concatMap(() => NEVER)
            ),
        { dispatch: false }
    );

    readonly setInProgressStatus$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SIGN_IN_TODO_ACTION),
            map(() =>
                UPDATE_AUTH_STATE_DOC_ACTION({
                    loadStatus: LoadStatus.IN_PROGRESS
                })
            )
        )
    );

    readonly signIn$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SIGN_IN_TODO_ACTION),
            switchMap(({ username, password }) =>
                this.authRequestService.getLogin({ username, password }).pipe(
                    map((response) => ({
                        accessToken: `${response?.token_type} ${response?.access_token}`,
                        refreshToken: `${response?.token_type} ${response?.refresh_token}`
                    })),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.UNCONFIRMED_USER) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Auth.Pseudo.Notification.InvalidCredentials')
                            });
                        } else if (apiUtils.getReasonCode(error) === ApiErrorCode.UNCONFIRMED_PASSWORD) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Auth.Pseudo.Notification.UnconfirmedPassword')
                            });
                        }

                        return of({ accessToken: '', refreshToken: '' });
                    })
                )
            ),
            tap((values: AuthTokens) => {
                if (values.accessToken && values.refreshToken) {
                    this.authLocalStorageService.setTokens(values.accessToken, values.refreshToken);
                }
            }),
            map((values?: AuthTokens) => {
                const credentials = this.getCredentialsFromToken(values?.accessToken);
                if (credentials) {
                    return SUCCESS_SIGN_IN_TODO_ACTION(credentials);
                }

                this.authLocalStorageService.removeTokens();

                return RESET_AUTH_CREDENTIALS_DOC_ACTION();
            })
        )
    );

    readonly successSignIn$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SUCCESS_SIGN_IN_TODO_ACTION),
            switchMap((credentials: AuthCredentials) => [
                ALLOW_AUTH_EVENT_ACTION(),
                UPDATE_AUTH_STATE_DOC_ACTION(credentials)
            ])
        )
    );

    readonly successRoleLoad$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(ROLE_LOAD_DONE_EVENT_ACTION),
            map(({ loadStatus }) => UPDATE_AUTH_STATE_DOC_ACTION({ loadStatus }))
        )
    );

    readonly navigateToOriginPath$: Observable<Action> = createEffect(
        () =>
            this.actions$.pipe(
                ofType(ROLE_LOAD_DONE_EVENT_ACTION),
                filter(({ loadStatus }) => loadStatus === LoadStatus.LOADED),
                switchMap(() => this.store.select(getAuthOriginPath).pipe(take(1))),
                tap((originPath) => {
                    this.router.navigateByUrl(originPath);
                }),
                concatMap(() => NEVER)
            ),
        { dispatch: false }
    );

    readonly signOutReset$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SIGN_OUT_TODO_ACTION),
            switchMap(() => [RESET_AUTH_CREDENTIALS_DOC_ACTION(), SIGN_OUT_EVENT_ACTION()])
        )
    );

    readonly signOutNavigate$: Observable<Action> = createEffect(
        () =>
            this.actions$.pipe(
                ofType(SIGN_OUT_TODO_ACTION),
                tap(() => {
                    this.authLocalStorageService.removeTokens();
                    this.router.navigate([RouterName.SIGN_IN]);
                }),
                concatMap(() => NEVER)
            ),
        { dispatch: false }
    );

    readonly expireTokens$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(EXPIRE_AUTH_TOKENS_TODO_ACTION),
            map(() => SIGN_OUT_TODO_ACTION()),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Warning,
                    title: this.i18nService.translate('Common.Pseudo.Notification.TokenExpired')
                });
            })
        )
    );

    readonly expirePassword$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(EXPIRE_PASSWORD_TODO_ACTION),
            map(() => SIGN_OUT_TODO_ACTION()),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Warning,
                    title: this.i18nService.translate('Common.Pseudo.Notification.PasswordExpired')
                });
            })
        )
    );

    readonly redirectToSwitchRoute$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(REDIRECT_TO_SWITCH_ROUTE_TODO_ACTION),
            map(() => RESET_AUTH_CREDENTIALS_DOC_ACTION()),
            tap(() => {
                this.router.navigate([RouterName.SWITCH]);
            })
        )
    );

    constructor(
        private readonly actions$: Actions,
        private readonly authLocalStorageService: AuthLocalStorageService,
        private readonly authRequestService: AuthRequestService,
        private readonly i18nService: I18nService,
        private readonly router: Router,
        private readonly store: Store<AuthState>,
        private readonly toastService: KbqToastService
    ) {}

    private getCredentialsFromToken(token?: string): AuthCredentials | undefined {
        if (!token) {
            return undefined;
        }

        const auth: AuthJwtData = jwtDecode(token);
        if (!auth) {
            return undefined;
        }

        return {
            userId: auth.user_id,
            username: auth.username,
            email: auth.email,
            roleId: auth.role.id,
            authType: auth.auth_type,
            passwordChangeTimestamp: auth.last_password_changed_at
        };
    }
}
