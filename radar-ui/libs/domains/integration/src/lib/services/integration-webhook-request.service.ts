import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    CreateIntegrationRequest,
    CreateIntegrationResponse,
    EmptyIntegrationResponse,
    GetIntegrationsResponse,
    IntegrationWebhook,
    UpdateIntegrationRequest
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class IntegrationWebhookRequestService {
    private readonly apiService = inject(ApiService);

    getWebhookIntegrations(): Observable<IntegrationWebhook[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetIntegrationsResponse<IntegrationWebhook>>('integration/webhook/list')
            .pipe(map((response) => response.integrations));
    }

    getWebhookIntegration(id: string): Observable<IntegrationWebhook> {
        return this.apiService.get<ApiEmptyRequest, IntegrationWebhook>(`integration/webhook/${id}`);
    }

    createWebhookIntegration(request: CreateIntegrationRequest<IntegrationWebhook>): Observable<IntegrationWebhook> {
        return this.apiService
            .post<CreateIntegrationRequest<IntegrationWebhook>, CreateIntegrationResponse>('integration', request)
            .pipe(
                map((response) => response.id),
                filter((id) => !!id),
                switchMap((id) => this.getWebhookIntegration(id).pipe(take(1)))
            );
    }

    updateWebhookIntegration(
        id: string,
        request: UpdateIntegrationRequest<IntegrationWebhook>
    ): Observable<IntegrationWebhook> {
        return this.apiService
            .patch<UpdateIntegrationRequest<IntegrationWebhook>, EmptyIntegrationResponse>(`integration/${id}`, request)
            .pipe(
                filter((response) => response && !Object.keys(response).length),
                switchMap(() => this.getWebhookIntegration(id).pipe(take(1)))
            );
    }

    deleteWebhookIntegration(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyIntegrationResponse>(`integration/webhook/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }
}
