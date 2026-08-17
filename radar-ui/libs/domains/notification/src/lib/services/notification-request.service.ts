import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { IntegrationType } from '@cs/domains/integration';
import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    CreateNotificationRequest,
    CreateNotificationResponse,
    EmptyNotificationResponse,
    GetNotificationResponse,
    GetNotificationTemplateResponse,
    GetNotificationsRequest,
    GetNotificationsResponse,
    Notification,
    UpdateNotificationRequest
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class NotificationRequestService {
    private readonly apiService = inject(ApiService);

    getNotifications(request: Partial<GetNotificationsRequest> = {}): Observable<Notification[]> {
        return this.apiService
            .get<Partial<GetNotificationsRequest>, GetNotificationsResponse>('notification/list', request)
            .pipe(map((response) => response.notifications));
    }

    /** @external */
    getNotificationTemplate(eventType: string, integrationType: IntegrationType): Observable<string> {
        return this.apiService
            .get<ApiEmptyRequest, GetNotificationTemplateResponse>('notification/default-template', {
                event_type: eventType,
                integration_type: integrationType
            })
            .pipe(map((response) => response.template));
    }

    /** @external */
    getNotification(id: string): Observable<GetNotificationResponse> {
        return this.apiService.get<ApiEmptyRequest, GetNotificationResponse>(`notification/${id}`);
    }

    createNotification(request: CreateNotificationRequest): Observable<Notification> {
        return this.apiService
            .post<CreateNotificationRequest, CreateNotificationResponse>('notification', request)
            .pipe(
                map((response) => response.id),
                filter((id) => !!id),
                switchMap((id) =>
                    this.getNotification(id).pipe(
                        take(1),
                        map((response) => response.notification)
                    )
                )
            );
    }

    updateNotification(id: string, request: UpdateNotificationRequest): Observable<Notification> {
        return this.apiService
            .patch<UpdateNotificationRequest, EmptyNotificationResponse>(`notification/${id}`, request)
            .pipe(
                filter((response) => response && !Object.keys(response).length),
                switchMap(() =>
                    this.getNotification(id).pipe(
                        take(1),
                        map((response) => response.notification)
                    )
                )
            );
    }

    deleteNotification(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyNotificationResponse>(`notification/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }
}
