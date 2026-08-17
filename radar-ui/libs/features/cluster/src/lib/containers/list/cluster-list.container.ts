import { ActivatedRoute } from '@angular/router';
import { DateTime } from 'luxon';
import { KbqAlertColors } from '@koobiq/components/alert';
import { KbqBadgeColors } from '@koobiq/components/badge';
import { PopUpPlacements } from '@koobiq/components/core';
import { BehaviorSubject, Observable, map, switchMap } from 'rxjs';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { IModalOptionsForService, KbqModalService } from '@koobiq/components/modal';

import { ApiPathService } from '@cs/api';
import { I18nService } from '@cs/i18n';
import { ClusterStatus, ClusterStoreService, GetClustersResponse, RegisteredCluster } from '@cs/domains/cluster';
import { DEFAULT_PAGINATOR_PAGE_INDEX, SharedModalService } from '@cs/shared';
import { PermissionName, PermissionType, RolePermissionMap } from '@cs/domains/role';

import { ClusterEditPopoverOutputs } from '../../interfaces/cluster-popover.interface';
import { ClusterFeatureDeleteUnregisteredModalContainer } from '../delete-unregistered-modal/cluster-delete-unregistered-modal.container';
import { ClusterRouterName } from '../../interfaces/cluster-navigation.interface';

const CLUSTER_PAGINATOR_PAGE_SIZE = 10;

@Component({
    templateUrl: './cluster-list.container.html',
    styleUrl: './cluster-list.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureListContainer {
    private readonly modalService = inject(KbqModalService);
    private readonly route = inject(ActivatedRoute);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly i18nService = inject(I18nService);
    private readonly sharedModalService = inject(SharedModalService);

    readonly pageIndex$ = new BehaviorSubject(DEFAULT_PAGINATOR_PAGE_INDEX);

    readonly clustersResponse$: Observable<GetClustersResponse> = this.pageIndex$.pipe(
        switchMap((index) =>
            this.clusterStoreService.clusters$.pipe(
                map((list) => {
                    const start = (index - 1) * CLUSTER_PAGINATOR_PAGE_SIZE;

                    // @todo: replace to server-side pagination
                    return {
                        clusters: list.slice(start, start + CLUSTER_PAGINATOR_PAGE_SIZE),
                        total: list.length
                    };
                })
            )
        )
    );

    readonly clusters$: Observable<RegisteredCluster[]> = this.clusterStoreService.registeredClusters$;

    readonly activeClusterHost$ = this.apiPathService.host$;

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    readonly clusterStatus = ClusterStatus;

    readonly clusterRouterName = ClusterRouterName;

    readonly dateShortFormat = DateTime.DATE_SHORT;

    readonly paginatorPageSize = CLUSTER_PAGINATOR_PAGE_SIZE;

    readonly alertColors = KbqAlertColors;

    readonly badgeColors = KbqBadgeColors;

    readonly tooltipPlacements = PopUpPlacements;

    changeName(outputs: ClusterEditPopoverOutputs) {
        this.clusterStoreService.updateCluster(outputs.id, {
            name: outputs.name,
            config: outputs.config
        });
    }

    openDeleteModal(id: string, status: ClusterStatus) {
        switch (status) {
            case ClusterStatus.REGISTERED:
                this.openRegisteredDeleteModal(id);
                break;
            case ClusterStatus.UNREGISTERED:
                this.openUnregisteredDeleteModal(id);
                break;
        }
    }

    changePage(pageIndex: number) {
        this.pageIndex$.next(pageIndex);
    }

    switchCluster(id: string) {
        this.clusterStoreService.switchCluster(id);
    }

    private openUnregisteredDeleteModal(id: string) {
        this.sharedModalService.delete({
            title: this.i18nService.translate('Cluster.DeleteModal.Title.Unregistered'),
            content: this.i18nService.translate('Cluster.DeleteModal.Text.Unregistered'),
            confirmText: this.i18nService.translate('Cluster.DeleteModal.Button.Confirm'),
            cancelText: this.i18nService.translate('Cluster.DeleteModal.Button.Cancel'),
            confirmHandler: () => {
                this.clusterStoreService.deleteCluster(id);
            }
        });
    }

    private openRegisteredDeleteModal(id: string) {
        const config: IModalOptionsForService = {
            kbqComponent: ClusterFeatureDeleteUnregisteredModalContainer,
            kbqComponentParams: { id },
            kbqClosable: false
        };

        this.modalService.open(config);
    }
}
