import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import { GetAppVersionResponse, GetCentralUrlResponse } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class LicenseRequestService {
    private readonly apiService = inject(ApiService);

    getAppVersion(): Observable<string> {
        return this.apiService
            .get<ApiEmptyRequest, GetAppVersionResponse>('info/version')
            .pipe(map((response) => response.version));
    }

    /** @external */
    getCentralUrl(): Observable<string> {
        return this.apiService
            .get<ApiEmptyRequest, GetCentralUrlResponse>('info/central-cs-url')
            .pipe(map((response) => response.url));
    }
}
