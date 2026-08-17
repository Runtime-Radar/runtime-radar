import { ActivatedRoute } from '@angular/router';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Observable, filter, take } from 'rxjs';

import { ApiPathService } from '@cs/api';
import { ClusterStoreService, RegisteredCluster } from '@cs/domains/cluster';
import {
    IntegrationEmail,
    IntegrationStoreService,
    IntegrationSyslog,
    IntegrationType,
    IntegrationWebhook
} from '@cs/domains/integration';
import { LoadStatus, CoreUtilsService as utils } from '@cs/core';
import { PermissionName, PermissionType, RolePermissionMap } from '@cs/domains/role';

import { IntegrationFeatureSidepanelFormComponent } from '../../components/sidepanel-form/integration-sidepanel-form.component';
import { IntegrationProtocolType } from '../../interfaces/integration-form.interface';
import { IntegrationSidepanelFormOutputs } from '../../interfaces/integration-sidepanel.interface';

@Component({
    templateUrl: './integration-list.container.html',
    styleUrl: './integration-list.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureListContainer {
    private readonly route = inject(ActivatedRoute);
    private readonly sidepanelService = inject(KbqSidepanelService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly integrationStoreService = inject(IntegrationStoreService);

    readonly loadStatus$: Observable<LoadStatus> = this.integrationStoreService.loadStatus$;

    readonly emailIntegrations$: Observable<IntegrationEmail[]> = this.integrationStoreService.emailIntegrations$;

    readonly syslogIntegrations$: Observable<IntegrationSyslog[]> = this.integrationStoreService.syslogIntegrations$;

    readonly webhookIntegrations$: Observable<IntegrationWebhook[]> = this.integrationStoreService.webhookIntegrations$;

    readonly clusters$: Observable<RegisteredCluster[]> = this.clusterStoreService.registeredClusters$;

    readonly activeClusterHost$ = this.apiPathService.host$;

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    readonly integrationType = IntegrationType;

    readonly loadStatus = LoadStatus;

    openCreateSidepanel() {
        this.sidepanelService
            .open(IntegrationFeatureSidepanelFormComponent, {
                position: KbqSidepanelPosition.Right,
                hasBackdrop: true,
                data: {}
            })
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((outputs: IntegrationSidepanelFormOutputs) => {
                switch (outputs.type) {
                    case IntegrationType.EMAIL:
                        this.integrationStoreService.createEmailIntegration({
                            type: outputs.type,
                            name: outputs.email.name,
                            skip_check: outputs.hasSkipCheck,
                            email: {
                                auth_type: outputs.email.authType,
                                from: outputs.email.from,
                                server: outputs.email.server,
                                username: outputs.email.username,
                                password: outputs.email.password,
                                ca: outputs.email.ca,
                                use_tls: outputs.email.protocol === IntegrationProtocolType.TLS,
                                use_start_tls: outputs.email.protocol === IntegrationProtocolType.START_TLS,
                                insecure: !outputs.email.isInsecure
                            }
                        });
                        break;
                    case IntegrationType.SYSLOG:
                        this.integrationStoreService.createSyslogIntegration({
                            type: outputs.type,
                            name: outputs.syslog.name,
                            skip_check: outputs.hasSkipCheck,
                            syslog: {
                                address: outputs.syslog.address
                            }
                        });
                        break;
                    case IntegrationType.WEBHOOK:
                        this.integrationStoreService.createWebhookIntegration({
                            type: outputs.type,
                            name: outputs.webhook.name,
                            skip_check: outputs.hasSkipCheck,
                            webhook: {
                                url: outputs.webhook.url,
                                login: outputs.webhook.login,
                                password: outputs.webhook.password,
                                ca: outputs.webhook.ca,
                                insecure: !outputs.webhook.isInsecure
                            }
                        });
                        break;
                }
            });
    }

    switchCluster(id: string) {
        this.clusterStoreService.switchCluster(id);
    }
}
