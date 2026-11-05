import { Store } from '@ngrx/store';
import { jwtDecode } from 'jwt-decode';
import {
    HttpErrorResponse,
    HttpEvent,
    HttpHandler,
    HttpHeaders,
    HttpInterceptor,
    HttpRequest,
    HttpStatusCode
} from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, catchError, filter, map, switchMap, tap, throwError } from 'rxjs';

import { API_PATH, ApiErrorCode, ApiPathService, ApiUtilsService as apiUtils } from '@cs/api';

import { AuthJwtData } from '../interfaces/contract/auth-jwt-contract.interface';
import { AuthLocalStorageService } from '../services/auth-local-storage.services';
import { AuthRequestService } from '../services/auth-request.service';
import { AuthState } from '../interfaces/state/auth-state.interface';
import { AUTH_BASIC_TOKEN, AuthTokenName } from '../constants/auth.constant';
import { EXPIRE_AUTH_TOKENS_TODO_ACTION, EXPIRE_PASSWORD_TODO_ACTION } from '../stores/auth-action.store';

@Injectable({
    providedIn: 'root'
})
export class AuthHeadersInterceptor implements HttpInterceptor {
    private readonly apiPath = inject(API_PATH);
    private readonly apiPathService = inject(ApiPathService);
    private readonly authLocalStorageService = inject(AuthLocalStorageService);
    private readonly authRequestService = inject(AuthRequestService);
    private readonly store = inject<Store<AuthState>>(Store);

    intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
        // matching the api path alone would also match it inside a foreign host, sending the token there
        if (!request.url.includes(this.apiPath) || !this.apiPathService.isOwnRequest(request.url)) {
            return next.handle(request);
        }

        const tokens = this.authLocalStorageService.getTokens();
        const originHeaders = request.headers.delete('Access-Control-Allow-Origin');
        const accessHeaders = originHeaders.set(AuthTokenName.ACCESS, tokens.accessToken || AUTH_BASIC_TOKEN);
        const requestHeader = originHeaders.get(AuthTokenName.ACCESS) ? originHeaders : accessHeaders;

        return next.handle(request.clone({ headers: requestHeader })).pipe(
            catchError((error: HttpErrorResponse) => {
                if ((error.status as HttpStatusCode) === HttpStatusCode.Unauthorized) {
                    const reason = apiUtils.getReasonCode(error);

                    if (reason === ApiErrorCode.ACCESS_TOKEN_EXPIRED) {
                        const refreshHeader = new HttpHeaders().set(AuthTokenName.ACCESS, tokens.refreshToken || '');

                        return this.authRequestService.getTokens(refreshHeader).pipe(
                            map((response) => {
                                const jwt: AuthJwtData = jwtDecode(response.access_token);
                                if (!jwt) {
                                    this.store.dispatch(EXPIRE_AUTH_TOKENS_TODO_ACTION());
                                }

                                return {
                                    accessToken: jwt ? `${response.token_type} ${response.access_token}` : '',
                                    refreshToken: jwt ? `${response.token_type} ${response.refresh_token}` : ''
                                };
                            }),
                            filter(({ accessToken, refreshToken }) => !!accessToken && !!refreshToken),
                            tap(({ accessToken, refreshToken }) => {
                                this.authLocalStorageService.setTokens(accessToken, refreshToken);
                            }),
                            switchMap(({ accessToken }) => {
                                const headers = new HttpHeaders().set(AuthTokenName.ACCESS, accessToken);

                                return next.handle(request.clone({ headers }));
                            })
                        );
                    } else if (
                        reason === ApiErrorCode.REFRESH_TOKEN_EXPIRED ||
                        reason === ApiErrorCode.UNAUTHENTICATED
                    ) {
                        this.store.dispatch(EXPIRE_AUTH_TOKENS_TODO_ACTION());
                    } else if (reason === ApiErrorCode.ACCESS_AND_REFRESH_TOKENS_CHANGED) {
                        this.store.dispatch(EXPIRE_PASSWORD_TODO_ACTION());
                    }
                }

                return throwError(() => error);
            })
        );
    }
}
