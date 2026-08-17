import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    CreateIntegrationRequest,
    CreateIntegrationResponse,
    EmptyIntegrationResponse,
    GetIntegrationsResponse,
    IntegrationSyslog,
    UpdateIntegrationRequest
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class IntegrationSyslogRequestService {
    private readonly apiService = inject(ApiService);

    getSyslogIntegrations(): Observable<IntegrationSyslog[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetIntegrationsResponse<IntegrationSyslog>>('integration/syslog/list')
            .pipe(map((response) => response.integrations));
    }

    getSyslogIntegration(id: string): Observable<IntegrationSyslog> {
        return this.apiService.get<ApiEmptyRequest, IntegrationSyslog>(`integration/syslog/${id}`);
    }

    createSyslogIntegration(request: CreateIntegrationRequest<IntegrationSyslog>): Observable<IntegrationSyslog> {
        return this.apiService
            .post<CreateIntegrationRequest<IntegrationSyslog>, CreateIntegrationResponse>('integration', request)
            .pipe(
                map((response) => response.id),
                filter((id) => !!id),
                switchMap((id) => this.getSyslogIntegration(id).pipe(take(1)))
            );
    }

    updateSyslogIntegration(
        id: string,
        request: UpdateIntegrationRequest<IntegrationSyslog>
    ): Observable<IntegrationSyslog> {
        return this.apiService
            .patch<UpdateIntegrationRequest<IntegrationSyslog>, EmptyIntegrationResponse>(`integration/${id}`, request)
            .pipe(
                filter((response) => response && !Object.keys(response).length),
                switchMap(() => this.getSyslogIntegration(id).pipe(take(1)))
            );
    }

    deleteSyslogIntegration(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyIntegrationResponse>(`integration/syslog/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }
}
