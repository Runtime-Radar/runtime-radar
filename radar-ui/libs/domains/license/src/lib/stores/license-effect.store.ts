import { Action } from '@ngrx/store';
import { HttpErrorResponse } from '@angular/common/http';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, of } from 'rxjs';
import { catchError, map, switchMap, take } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { ROLE_LOAD_DONE_EVENT_ACTION } from '@cs/domains/role';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { ApiErrorCode, ApiPathService, ApiUtilsService as apiUtils } from '@cs/api';

import { LicenseRequestService } from '../services/license-request.service';
import { UPDATE_LICENSE_STATE_DOC_ACTION } from './license-action.store';

@Injectable({
    providedIn: 'root'
})
export class LicenseEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly toastService = inject(KbqToastService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly i18nService = inject(I18nService);
    private readonly licenseRequestService = inject(LicenseRequestService);

    readonly loadAppVersion$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(ROLE_LOAD_DONE_EVENT_ACTION),
            switchMap(() =>
                this.licenseRequestService.getAppVersion().pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.STATE_CHILD_UNREGISTERED) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Common.Pseudo.Notification.ChildClusterUnregistered')
                            });
                        }

                        return of('');
                    })
                )
            ),
            map((version) =>
                UPDATE_LICENSE_STATE_DOC_ACTION({
                    appVersion: version,
                    hostAppVersion: version
                })
            )
        )
    );

    readonly loadAppVersionByClusterSwitch$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.licenseRequestService.getAppVersion().pipe(take(1))),
            switchMap((version) =>
                this.apiPathService.host$.pipe(
                    take(1),
                    map((host) => ({
                        version,
                        isHost: host === ''
                    }))
                )
            ),
            map(({ version, isHost }) => {
                if (isHost) {
                    return UPDATE_LICENSE_STATE_DOC_ACTION({
                        appVersion: version,
                        hostAppVersion: version
                    });
                }

                return UPDATE_LICENSE_STATE_DOC_ACTION({ appVersion: version });
            })
        )
    );

    readonly loadCentralUrl$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(ROLE_LOAD_DONE_EVENT_ACTION, SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.licenseRequestService.getCentralUrl().pipe(take(1))),
            map((centralUrl) => UPDATE_LICENSE_STATE_DOC_ACTION({ centralUrl }))
        )
    );
}
