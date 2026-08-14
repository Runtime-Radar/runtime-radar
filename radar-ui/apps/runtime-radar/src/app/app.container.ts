import { ChangeDetectionStrategy, Component, OnInit, TemplateRef, ViewChild, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, bufferWhen, delay, distinctUntilChanged, map, switchMap } from 'rxjs';

import { ApiPathService } from '@cs/api';
import { I18nService } from '@cs/i18n';
import { LicenseStoreService } from '@cs/domains/license';
import { AuthCredentials, AuthStoreService } from '@cs/domains/auth';
import { CoreNavigationStoreService, CoreWindowService, LoadStatus, RouterName } from '@cs/core';
import { Role, RoleStoreService } from '@cs/domains/role';

@Component({
    selector: 'cs-app-container',
    templateUrl: './app.container.html',
    styleUrl: './app.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class AppContainer implements OnInit {
    private readonly toastService = inject(KbqToastService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly authStoreService = inject(AuthStoreService);
    private readonly coreNavigationStoreService = inject(CoreNavigationStoreService);
    private readonly coreWindowService = inject(CoreWindowService);
    private readonly i18nService = inject(I18nService);
    private readonly licenseStoreService = inject(LicenseStoreService);
    private readonly roleStoreService = inject(RoleStoreService);

    @ViewChild('clusterToastActionTemplate') clusterToastActionTemplate!: TemplateRef<any>;

    readonly activeClusterHost$ = this.apiPathService.host$;

    readonly credentials$: Observable<AuthCredentials> = this.authStoreService.credentials$;

    readonly loadStatus$: Observable<LoadStatus> = this.authStoreService.loadStatus$;

    readonly routerName$: Observable<RouterName> = this.coreNavigationStoreService.routerName$;

    readonly appVersion$: Observable<string> = this.licenseStoreService.appVersion$;

    readonly hostAppVersion$: Observable<string> = this.licenseStoreService.hostAppVersion$;

    readonly isVersionDiff$: Observable<boolean> = this.licenseStoreService.appVersion$.pipe(
        switchMap((appVersion) =>
            this.licenseStoreService.hostAppVersion$.pipe(map((hostVersion) => hostVersion !== appVersion))
        )
    );

    readonly role$: Observable<Role | undefined> = this.credentials$.pipe(
        map((credentials) => credentials.roleId),
        distinctUntilChanged(),
        switchMap((roleId) => this.roleStoreService.role$(roleId))
    );

    readonly loadStatus = LoadStatus;

    readonly routerName = RouterName;

    ngOnInit() {
        this.apiPathService.error$
            .pipe(
                bufferWhen(() =>
                    this.apiPathService.error$.pipe(
                        /* eslint @typescript-eslint/no-magic-numbers: "off" */
                        delay(500)
                    )
                ),
                switchMap((list) =>
                    this.apiPathService.host$.pipe(
                        map((host) => ({
                            host,
                            error: list.at(-1)
                        }))
                    )
                )
            )
            .subscribe(({ host }) => {
                if (host) {
                    this.toastService.show(
                        {
                            style: KbqToastStyle.Warning,
                            title: this.i18nService.translate('Common.Pseudo.Notification.CertAuthorityInvalid.Title'),
                            caption: this.i18nService.translate(
                                'Common.Pseudo.Notification.CertAuthorityInvalid.Caption'
                            ),
                            actions: this.clusterToastActionTemplate
                        },
                        0
                    );
                }
            });
    }

    signOut() {
        this.authStoreService.signOut();
    }

    reloadPage() {
        this.coreWindowService.location.reload();
    }
}
