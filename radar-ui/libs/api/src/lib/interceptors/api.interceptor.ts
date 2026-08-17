import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { catchError } from 'rxjs/operators';
import {
    HttpErrorResponse,
    HttpEvent,
    HttpHandler,
    HttpInterceptor,
    HttpRequest,
    HttpStatusCode
} from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, throwError } from 'rxjs';

import { CoreWindowService } from '@cs/core';
import { I18nService } from '@cs/i18n';

import { ApiErrorCode } from '../interfaces/contract/api-error-contract.interface';
import { ApiPathService } from '../services/api-path.service';
import { ApiUtilsService as apiUtils } from '../services/api-utils.service';

const API_NOTIFICATION_DEBOUNCE_INTERVAL = 3000;

@Injectable({
    providedIn: 'root'
})
export class ApiErrorInterceptor implements HttpInterceptor {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly coreWindowService = inject(CoreWindowService);
    private readonly i18nService = inject(I18nService);

    private readonly notificationTimestamps = new Map<string, number>();

    intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
        return next.handle(request).pipe(
            catchError((error: HttpErrorResponse) => {
                if (apiUtils.getReasonCode(error) !== ApiErrorCode.UNKNOWN_ISSUE) {
                    return throwError(() => error);
                }

                if (error.status === 0) {
                    this.apiPathService.setError('ERR_NETWORK_UNKNOWN');
                }

                if (this.shouldToastShown(error)) {
                    switch (error.status as HttpStatusCode) {
                        case HttpStatusCode.BadRequest:
                            this.toastService.show({
                                style: KbqToastStyle.Error,
                                title: this.i18nService.translate('Common.Pseudo.Notification.BadRequest')
                            });
                            break;
                        case HttpStatusCode.Forbidden:
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Common.Pseudo.Notification.Forbidden')
                            });
                            break;
                        case HttpStatusCode.NotFound:
                            this.toastService.show({
                                style: KbqToastStyle.Error,
                                title: this.i18nService.translate('Common.Pseudo.Notification.NotFound')
                            });
                            break;
                        case HttpStatusCode.UnprocessableEntity:
                            this.toastService.show({
                                style: KbqToastStyle.Error,
                                title: this.i18nService.translate('Common.Pseudo.Notification.UnprocessableEntity')
                            });
                            break;
                        case HttpStatusCode.ServiceUnavailable:
                            this.toastService.show({
                                style: KbqToastStyle.Error,
                                title: this.i18nService.translate('Common.Pseudo.Notification.ServiceUnavailable')
                            });
                            break;
                    }
                }

                return throwError(() => error);
            })
        );
    }

    private shouldToastShown(error: HttpErrorResponse): boolean {
        const notificationKey = this.coreWindowService.btoa(`${error.status}_${error.statusText}`);
        const currentTime = this.dateAdapter.today().toMillis();
        const lastShownTime = this.notificationTimestamps.get(notificationKey) || 0;

        if (currentTime - lastShownTime > API_NOTIFICATION_DEBOUNCE_INTERVAL) {
            this.notificationTimestamps.set(notificationKey, currentTime);

            return true;
        }

        return false;
    }
}
