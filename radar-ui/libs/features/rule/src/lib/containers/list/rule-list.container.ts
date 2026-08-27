import { ActivatedRoute } from '@angular/router';
import { KbqBadgeColors } from '@koobiq/components/badge';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KbqSidepanelConfig, KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Observable, filter, forkJoin, map, mergeMap, of, take } from 'rxjs';

import { ApiPathService } from '@cs/api';
import { DetectorStoreService } from '@cs/domains/detector';
import { I18nService } from '@cs/i18n';
import { SharedModalService } from '@cs/shared';
import { ClusterStoreService, RegisteredCluster } from '@cs/domains/cluster';
import { CreateRuleRequest, Rule, RuleStoreService, RuleType, UpdateRuleRequest } from '@cs/domains/rule';
import { LoadStatus, CoreUtilsService as utils } from '@cs/core';
import { Notification, NotificationRequestService } from '@cs/domains/notification';
import { PermissionName, PermissionType, RolePermissionMap } from '@cs/domains/role';
import {
    RuleForm,
    RulePackageSidepanelFormComponent,
    RulePackageSidepanelInfoComponent,
    RuleSidepanelFormProps,
    RuleSidepanelInfoProps
} from '@cs/packages/rule';

import { RuleFilters } from '../../interfaces/rule-filter.interface';
import { RuleFeatureHelperService as ruleHelper } from '../../services/rule-helper.service';

@Component({
    templateUrl: './rule-list.container.html',
    styleUrl: './rule-list.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuleFeatureListContainer {
    private readonly route = inject(ActivatedRoute);
    private readonly sidepanelService = inject(KbqSidepanelService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly detectorStoreService = inject(DetectorStoreService);
    private readonly i18nService = inject(I18nService);
    private readonly notificationRequestService = inject(NotificationRequestService);
    private readonly ruleStoreService = inject(RuleStoreService);
    private readonly sharedModalService = inject(SharedModalService);

    readonly rules$: Observable<Rule[]> = this.ruleStoreService.rules$;

    readonly loadStatus$: Observable<LoadStatus> = this.ruleStoreService.loadStatus$;

    readonly clusters$: Observable<RegisteredCluster[]> = this.clusterStoreService.registeredClusters$;

    readonly activeClusterHost$ = this.apiPathService.host$;

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    readonly badgeColors = KbqBadgeColors;

    readonly loadStatus = LoadStatus;

    filters?: RuleFilters;

    changeFilter(values: RuleFilters) {
        this.filters = values;
    }

    openViewSidepanel(rule: Rule) {
        const notifications$: Observable<Notification[]> = this.ruleStoreService.rule$(rule.id).pipe(
            map((item) => item?.rule.notify?.targets),
            mergeMap((ids) => {
                if (!ids?.length) {
                    return of([] as Notification[]);
                }

                return forkJoin(
                    ids.map((id) =>
                        this.notificationRequestService
                            .getNotification(id)
                            .pipe(map((response) => response.notification))
                    )
                );
            })
        );

        const config: KbqSidepanelConfig<RuleSidepanelInfoProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                isDeleted: false,
                permissions: this.permissions[PermissionName.RULES],
                rule$: this.ruleStoreService.rule$(rule.id),
                detectors$: this.detectorStoreService.detectors$(),
                notifications$,
                updateHandler: this.openEditSidepanel.bind(this),
                deleteHandler: this.openDeleteModal.bind(this, rule.id)
            }
        };

        this.sidepanelService.open(RulePackageSidepanelInfoComponent, config).afterClosed().pipe(take(1)).subscribe();
    }

    openCreateSidepanel() {
        const config: KbqSidepanelConfig<Partial<RuleSidepanelFormProps>> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {}
        };

        this.sidepanelService
            .open(RulePackageSidepanelFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((form: RuleForm) => {
                const request: CreateRuleRequest = {
                    name: form.name,
                    type: RuleType.TYPE_RUNTIME,
                    rule: {
                        version: '1', // @todo: create environment constant
                        block: ruleHelper.convertFormValuesToBlockEntity(form),
                        notify: ruleHelper.convertFormValuesToNotifyEntity(form),
                        whitelist: ruleHelper.convertWhiteListToRequestNode(form)
                    },
                    scope: {
                        version: '1', // @todo: create environment constant
                        image_names: form.imageNames,
                        registries: form.registries,
                        namespaces: form.namespaces,
                        clusters: [],
                        pods: form.pods,
                        containers: form.containers,
                        nodes: form.nodes
                    }
                };

                this.ruleStoreService.createRule(request);
            });
    }

    openEditSidepanel(rule: Rule) {
        const config: KbqSidepanelConfig<Partial<RuleSidepanelFormProps>> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                isEdit: true,
                rule
            }
        };

        this.sidepanelService
            .open(RulePackageSidepanelFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((form: RuleForm) => {
                const request: UpdateRuleRequest = {
                    name: form.name,
                    type: RuleType.TYPE_RUNTIME,
                    rule: {
                        version: '1', // @todo: create environment constant
                        block: ruleHelper.convertFormValuesToBlockEntity(form),
                        notify: ruleHelper.convertFormValuesToNotifyEntity(form),
                        whitelist: ruleHelper.convertWhiteListToRequestNode(form)
                    },
                    scope: {
                        version: '1', // @todo: create environment constant
                        image_names: form.imageNames,
                        registries: form.registries,
                        namespaces: form.namespaces,
                        clusters: [],
                        pods: form.pods,
                        containers: form.containers,
                        nodes: form.nodes
                    }
                };

                this.ruleStoreService.updateRule(rule.id, request);
                this.sidepanelService.closeAll();
            });
    }

    openDeleteModal(id: string) {
        this.sharedModalService.delete({
            content: this.i18nService.translate('Rule.DeleteModal.Title.Text'),
            confirmText: this.i18nService.translate('Rule.DeleteModal.Button.Confirm'),
            cancelText: this.i18nService.translate('Rule.DeleteModal.Button.Cancel'),
            confirmHandler: () => {
                this.ruleStoreService.deleteRule(id);
            }
        });
    }

    switchCluster(id: string) {
        this.clusterStoreService.switchCluster(id);
    }
}
