import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import { CreateTokenRequest, CreateTokenResponse, EmptyTokenResponse, GetTokensResponse, Token } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class TokenRequestService {
    private readonly apiService = inject(ApiService);

    getTokens(): Observable<Token[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetTokensResponse>('access-token/page/1?page_size=100')
            .pipe(map((response) => response.access_tokens));
    }

    getToken(id: string): Observable<Token> {
        return this.apiService.get<ApiEmptyRequest, Token>(`access-token/${id}`);
    }

    createToken(request: CreateTokenRequest): Observable<Token> {
        return this.apiService.post<CreateTokenRequest, CreateTokenResponse>('access-token', request).pipe(
            filter((response) => !!response.id),
            switchMap((response) =>
                this.getToken(response.id).pipe(
                    take(1),
                    map((token) => ({
                        ...token,
                        access_token: response.access_token
                    }))
                )
            )
        );
    }

    deleteToken(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyTokenResponse>(`access-token/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }

    revokeTokens(): Observable<boolean> {
        return this.apiService
            .post<ApiEmptyRequest, EmptyTokenResponse>('access-token/invalidate-access-tokens', {})
            .pipe(map((response) => response && !Object.keys(response).length));
    }
}
