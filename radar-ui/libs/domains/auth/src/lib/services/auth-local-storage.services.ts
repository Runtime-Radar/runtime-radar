import { Injectable, inject } from '@angular/core';

import { CoreWindowService } from '@cs/core';

import { AuthTokenName } from '../constants/auth.constant';
import { AuthTokens } from '../interfaces/state/auth-state.interface';

@Injectable({
    providedIn: 'root'
})
export class AuthLocalStorageService {
    private readonly coreWindowService = inject(CoreWindowService);

    getTokens(): AuthTokens {
        return {
            accessToken: this.coreWindowService.localStorage.getItem(AuthTokenName.ACCESS) || '',
            refreshToken: this.coreWindowService.localStorage.getItem(AuthTokenName.REFRESH) || ''
        };
    }

    setTokens(accessToken: string, refreshToken: string) {
        this.coreWindowService.localStorage.setItem(AuthTokenName.ACCESS, accessToken);
        this.coreWindowService.localStorage.setItem(AuthTokenName.REFRESH, refreshToken);
    }

    removeTokens() {
        this.coreWindowService.localStorage.removeItem(AuthTokenName.ACCESS);
        this.coreWindowService.localStorage.removeItem(AuthTokenName.REFRESH);
    }
}
