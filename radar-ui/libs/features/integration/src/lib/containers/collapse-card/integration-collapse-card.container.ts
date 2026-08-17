import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, Input, OnInit, inject } from '@angular/core';
import { KbqSidepanelConfig, KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Observable, filter, map, of, switchMap, take } from 'rxjs';

import { ApiPathService } from '@cs/api';
import { DetectorStoreService } from '@cs/domains/detector';
import { I18nService } from '@cs/i18n';
import { LicenseStoreService } from '@cs/domains/license';
import { SharedModalService } from '@cs/shared';
import { CoreUtilsService as utils } from '@cs/core';
import { ClusterStoreService, RegisteredCluster } from '@cs/domains/cluster';
import { Integration, IntegrationStoreService, IntegrationType } from '@cs/domains/integration';
import { Notification, NotificationStoreService } from '@cs/domains/notification';
import { PermissionName, PermissionType, RolePermissionMap } from '@cs/domains/role';
import { Rule, RuleStoreService } from '@cs/domains/rule';
import { RulePackageSidepanelInfoComponent, RuleSidepanelInfoProps } from '@cs/packages/rule';

import { IntegrationFeatureSidepanelFormComponent } from '../../components/sidepanel-form/integration-sidepanel-form.component';
import { IntegrationFeatureSidepanelRecipientFormComponent } from '../../components/sidepanel-recipient-form/integration-sidepanel-recipient-form.component';
import { IntegrationProtocolType } from '../../interfaces/integration-form.interface';
import { IntegrationRecipientForm } from '../../interfaces/integration-recipient-form.interace';
import { IntegrationFeatureHelperService as integrationHelper } from '../../services/integration-helper.service';
import {
    IntegrationSidepanelFormOutputs,
    IntegrationSidepanelFormProps,
    IntegrationSidepanelRecipientFormProps
} from '../../interfaces/integration-sidepanel.interface';

@Component({
    selector: 'cs-integration-feature-collapse-card-container',
    templateUrl: './integration-collapse-card.container.html',
    styleUrl: './integration-collapse-card.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureCollapseCardContainer implements OnInit {
    private readonly sidepanelService = inject(KbqSidepanelService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly detectorStoreService = inject(DetectorStoreService);
    private readonly i18nService = inject(I18nService);
    private readonly integrationStoreService = inject(IntegrationStoreService);
    private readonly licenseStoreService = inject(LicenseStoreService);
    private readonly notificationStoreService = inject(NotificationStoreService);
    private readonly ruleStoreService = inject(RuleStoreService);
    private readonly sharedModalService = inject(SharedModalService);

    @Input({ required: true }) type!: IntegrationType;

    @Input({ required: true }) permissions!: RolePermissionMap;

    @Input() list?: Integration[] | null;

    readonly notifications$ = (id: string): Observable<Notification[]> =>
        this.notificationStoreService.notificationsByIntegrationId$(id);

    readonly rules$ = (notificationId: string): Observable<Rule[]> =>
        this.ruleStoreService.rulesByNotificationId$(notificationId);

    private readonly activeRegisteredCluster$: Observable<RegisteredCluster | undefined> =
        this.apiPathService.host$.pipe(
            switchMap((host) =>
                this.clusterStoreService.registeredClusters$.pipe(
                    take(1),
                    map((clusters) => clusters.find((item) => item.own_cs_url === host))
                )
            )
        );

    readonly expandedCollection = new Map<string, boolean>();

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    readonly integrationType = IntegrationType;

    readonly tooltipPlacements = PopUpPlacements;

    ngOnInit() {
        this.list?.forEach((item) => {
            this.expandedCollection.set(item.id, true);
        });
    }

    toogleCardExpand(id: string) {
        this.expandedCollection.set(id, !this.expandedCollection.get(id));
    }

    openCreateRecipientSidepanel(integration: Integration) {
        const config: KbqSidepanelConfig<IntegrationSidepanelRecipientFormProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                centralUrl$: this.licenseStoreService.centralUrl$,
                activeRegisteredCluster$: this.activeRegisteredCluster$,
                integration
            }
        };

        this.sidepanelService
            .open(IntegrationFeatureSidepanelRecipientFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((form: IntegrationRecipientForm) => {
                switch (this.type) {
                    case IntegrationType.EMAIL:
                        this.notificationStoreService.createNotification({
                            integration_id: integration.id,
                            integration_type: IntegrationType.EMAIL,
                            name: form.name,
                            recipients: form.recipients,
                            event_type: form.eventType,
                            template: !form.isTemplateDefault ? form.template : '',
                            cs_cluster_id: form.clusterId,
                            cs_cluster_name: form.clusterName,
                            central_cs_url: form.centralUrl,
                            own_cs_url: form.clusterUrl,
                            email: {
                                subject_template: form.subjectTemplate
                            }
                        });
                        break;
                    case IntegrationType.SYSLOG:
                        this.notificationStoreService.createNotification({
                            integration_id: integration.id,
                            integration_type: IntegrationType.SYSLOG,
                            name: form.name,
                            recipients: form.recipients,
                            event_type: form.eventType,
                            template: !form.isTemplateDefault ? form.template : '',
                            cs_cluster_id: form.clusterId,
                            cs_cluster_name: form.clusterName,
                            central_cs_url: form.centralUrl,
                            own_cs_url: form.clusterUrl,
                            syslog: {}
                        });
                        break;
                    case IntegrationType.WEBHOOK:
                        this.notificationStoreService.createNotification({
                            integration_id: integration.id,
                            integration_type: IntegrationType.WEBHOOK,
                            name: form.name,
                            recipients: form.recipients,
                            event_type: form.eventType,
                            template: !form.isTemplateDefault ? form.template : '',
                            cs_cluster_id: form.clusterId,
                            cs_cluster_name: form.clusterName,
                            central_cs_url: form.centralUrl,
                            own_cs_url: form.clusterUrl,
                            webhook: {
                                path: form.path,
                                headers: integrationHelper.convertHeadersToRequestNode(form.header)
                            }
                        });
                        break;
                }
            });
    }

    openEditRecipientSidepanel(integration: Integration, notification: Notification) {
        const config: KbqSidepanelConfig<IntegrationSidepanelRecipientFormProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                centralUrl$: this.licenseStoreService.centralUrl$,
                activeRegisteredCluster$: this.activeRegisteredCluster$,
                integration,
                notification,
                isEdit: true
            }
        };

        this.sidepanelService
            .open(IntegrationFeatureSidepanelRecipientFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((form: IntegrationRecipientForm) => {
                switch (this.type) {
                    case IntegrationType.EMAIL:
                        this.notificationStoreService.updateNotification(notification.id, {
                            integration_id: integration.id,
                            integration_type: IntegrationType.EMAIL,
                            name: form.name,
                            recipients: form.recipients,
                            event_type: form.eventType,
                            template: !form.isTemplateDefault ? form.template : '',
                            cs_cluster_id: form.clusterId,
                            cs_cluster_name: form.clusterName,
                            central_cs_url: form.centralUrl,
                            own_cs_url: form.clusterUrl,
                            email: {
                                subject_template: form.subjectTemplate
                            }
                        });
                        break;
                    case IntegrationType.SYSLOG:
                        this.notificationStoreService.updateNotification(notification.id, {
                            integration_id: integration.id,
                            integration_type: IntegrationType.SYSLOG,
                            name: form.name,
                            recipients: form.recipients,
                            event_type: form.eventType,
                            template: !form.isTemplateDefault ? form.template : '',
                            cs_cluster_id: form.clusterId,
                            cs_cluster_name: form.clusterName,
                            central_cs_url: form.centralUrl,
                            own_cs_url: form.clusterUrl,
                            syslog: {}
                        });
                        break;
                    case IntegrationType.WEBHOOK:
                        this.notificationStoreService.updateNotification(notification.id, {
                            integration_id: integration.id,
                            integration_type: IntegrationType.WEBHOOK,
                            name: form.name,
                            recipients: form.recipients,
                            event_type: form.eventType,
                            template: !form.isTemplateDefault ? form.template : '',
                            cs_cluster_id: form.clusterId,
                            cs_cluster_name: form.clusterName,
                            central_cs_url: form.centralUrl,
                            own_cs_url: form.clusterUrl,
                            webhook: {
                                path: form.path,
                                headers: integrationHelper.convertHeadersToRequestNode(form.header)
                            }
                        });
                        break;
                }
            });
    }

    openDeleteRecipientModal(id: string) {
        this.sharedModalService.delete({
            title: this.i18nService.translate('Integration.DeleteRecipientModal.Content.Title'),
            content: this.i18nService.translate('Integration.DeleteRecipientModal.Content.Text'),
            confirmText: this.i18nService.translate('Integration.DeleteRecipientModal.Button.Confirm'),
            cancelText: this.i18nService.translate('Integration.DeleteRecipientModal.Button.Cancel'),
            confirmHandler: () => {
                this.notificationStoreService.deleteNotification(id);
            }
        });
    }

    openEditSidepanel(item: Integration) {
        const config: KbqSidepanelConfig<Partial<IntegrationSidepanelFormProps>> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                type: item.type,
                email: item.type === IntegrationType.EMAIL ? item : undefined,
                syslog: item.type === IntegrationType.SYSLOG ? item : undefined,
                webhook: item.type === IntegrationType.WEBHOOK ? item : undefined,
                isEdit: true
            }
        };

        this.sidepanelService
            .open(IntegrationFeatureSidepanelFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((outputs: IntegrationSidepanelFormOutputs) => {
                switch (outputs.type) {
                    case IntegrationType.EMAIL:
                        this.integrationStoreService.updateEmailIntegration(item.id, {
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
                        this.integrationStoreService.updateSyslogIntegration(item.id, {
                            type: outputs.type,
                            name: outputs.syslog.name,
                            skip_check: outputs.hasSkipCheck,
                            syslog: {
                                address: outputs.syslog.address
                            }
                        });
                        break;
                    case IntegrationType.WEBHOOK:
                        this.integrationStoreService.updateWebhookIntegration(item.id, {
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

    openDeleteModal(id: string) {
        this.sharedModalService.delete({
            title: this.i18nService.translate('Integration.DeleteModal.Content.Title'),
            content: this.i18nService.translate('Integration.DeleteModal.Content.Text'),
            confirmText: this.i18nService.translate('Integration.DeleteModal.Button.Confirm'),
            cancelText: this.i18nService.translate('Integration.DeleteModal.Button.Cancel'),
            confirmHandler: () => {
                switch (this.type) {
                    case IntegrationType.EMAIL:
                        this.integrationStoreService.deleteEmailIntegration(id);
                        break;
                    case IntegrationType.SYSLOG:
                        this.integrationStoreService.deleteSyslogIntegration(id);
                        break;
                    case IntegrationType.WEBHOOK:
                        this.integrationStoreService.deleteWebhookIntegration(id);
                        break;
                }
            }
        });
    }

    openRuleSidepanel(rule: Rule) {
        const notifications$: Observable<Notification[]> = this.notificationStoreService.notifications$.pipe(
            map((list) => list.filter((item) => rule.rule.notify?.targets.includes(item.id)))
        );

        const config: KbqSidepanelConfig<RuleSidepanelInfoProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                isDeleted: false,
                rule$: of(rule),
                detectors$: this.detectorStoreService.detectors$(),
                notifications$
            }
        };

        this.sidepanelService.open(RulePackageSidepanelInfoComponent, config).afterClosed().pipe(take(1)).subscribe();
    }
}
