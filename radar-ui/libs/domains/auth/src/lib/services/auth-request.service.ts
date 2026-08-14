import { HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    GetLoginRequest,
    GetLoginResponse,
    GetTokenResponse
} from '../interfaces/contract/auth-api-contract.interface';

@Injectable({
    providedIn: 'root'
})
export class AuthRequestService {
    private readonly apiService = inject(ApiService);

    getLogin(request: GetLoginRequest): Observable<GetLoginResponse> {
        return this.apiService.post<GetLoginRequest, GetLoginResponse>('signin', request);
    }

    getTokens(headers: HttpHeaders): Observable<GetTokenResponse> {
        return this.apiService.get<ApiEmptyRequest, GetTokenResponse>('tokens', {}, headers);
    }
}
