import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    CreateIntegrationRequest,
    CreateIntegrationResponse,
    EmptyIntegrationResponse,
    GetIntegrationsResponse,
    IntegrationEmail,
    UpdateIntegrationRequest
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class IntegrationEmailRequestService {
    private readonly apiService = inject(ApiService);

    getEmailIntegrations(): Observable<IntegrationEmail[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetIntegrationsResponse<IntegrationEmail>>('integration/email/list')
            .pipe(map((response) => response.integrations));
    }

    getEmailIntegration(id: string): Observable<IntegrationEmail> {
        return this.apiService.get<ApiEmptyRequest, IntegrationEmail>(`integration/email/${id}`);
    }

    createEmailIntegration(request: CreateIntegrationRequest<IntegrationEmail>): Observable<IntegrationEmail> {
        return this.apiService
            .post<CreateIntegrationRequest<IntegrationEmail>, CreateIntegrationResponse>('integration', request)
            .pipe(
                map((response) => response.id),
                filter((id) => !!id),
                switchMap((id) => this.getEmailIntegration(id).pipe(take(1)))
            );
    }

    updateEmailIntegration(
        id: string,
        request: UpdateIntegrationRequest<IntegrationEmail>
    ): Observable<IntegrationEmail> {
        return this.apiService
            .patch<UpdateIntegrationRequest<IntegrationEmail>, EmptyIntegrationResponse>(`integration/${id}`, request)
            .pipe(
                filter((response) => response && !Object.keys(response).length),
                switchMap(() => this.getEmailIntegration(id).pipe(take(1)))
            );
    }

    deleteEmailIntegration(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyIntegrationResponse>(`integration/email/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }
}
