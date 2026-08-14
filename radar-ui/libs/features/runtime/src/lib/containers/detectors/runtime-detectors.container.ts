import { KbqBadgeColors } from '@koobiq/components/badge';
import { ActivatedRoute, Router } from '@angular/router';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { IModalOptionsForService, KbqModalService, ModalSize } from '@koobiq/components/modal';
import { KbqSidepanelConfig, KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Observable, map, take } from 'rxjs';

import { ApiPathService } from '@cs/api';
import { I18nService } from '@cs/i18n';
import { RouterName } from '@cs/core';
import { SharedModalService } from '@cs/shared';
import { ClusterStoreService, RegisteredCluster } from '@cs/domains/cluster';
import { DetectorExtended, DetectorStoreService, DetectorType } from '@cs/domains/detector';
import { PermissionName, PermissionType, RolePermissionMap } from '@cs/domains/role';

import { RUNTIME_NAVIGATION_TABS } from '../../constants/runtime-navigation.constant';
import { RuntimeFeatureSidepanelDetectorComponent } from '../../components/sidepanel-detector/runtime-sidepanel-detector.component';
import { RuntimeFeatureUploadDetectorModalComponent } from '../../components/upload-detector-modal/runtime-upload-detector-modal.component';
import { RuntimeSidepanelDetectorProps } from '../../interfaces/runtime-sidepanel.interface';

@Component({
    templateUrl: './runtime-detectors.container.html',
    styleUrl: './runtime-detectors.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureDetectorsContainer {
    private readonly modalService = inject(KbqModalService);
    private readonly route = inject(ActivatedRoute);
    private readonly router = inject(Router);
    private readonly sidepanelService = inject(KbqSidepanelService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly detectorStoreService = inject(DetectorStoreService);
    private readonly i18nService = inject(I18nService);
    private readonly sharedModalService = inject(SharedModalService);

    readonly detectors$: Observable<DetectorExtended[]> = this.detectorStoreService
        .detectors$([DetectorType.RUNTIME])
        .pipe(map((detectors) => detectors.sort((a, b) => a.id.localeCompare(b.id))));

    readonly clusters$: Observable<RegisteredCluster[]> = this.clusterStoreService.registeredClusters$;

    readonly activeClusterHost$ = this.apiPathService.host$;

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    readonly badgeColors = KbqBadgeColors;

    readonly runtimeNavigationTabs = RUNTIME_NAVIGATION_TABS;

    tabChange(path?: string) {
        this.router.navigate([RouterName.RUNTIME, path]);
    }

    openCreateModal() {
        const config: IModalOptionsForService = {
            kbqComponent: RuntimeFeatureUploadDetectorModalComponent,
            kbqSize: ModalSize.Medium,
            kbqClosable: false
        };

        this.modalService
            .open<RuntimeFeatureUploadDetectorModalComponent, string[] | undefined>(config)
            .afterClose.pipe(take(1))
            .subscribe((base64list?: string[]) => {
                if (base64list && base64list.length) {
                    this.detectorStoreService.createRuntimeDetectors(base64list);
                }
            });
    }

    openDeleteModal(key: string, version: number) {
        this.sharedModalService.delete({
            title: this.i18nService.translate('Runtime.DeleteModal.Content.Title'),
            content: this.i18nService.translate('Runtime.DeleteModal.Content.Text'),
            confirmText: this.i18nService.translate('Runtime.DeleteModal.Button.Confirm'),
            cancelText: this.i18nService.translate('Runtime.DeleteModal.Button.Cancel'),
            confirmHandler: () => {
                this.detectorStoreService.deleteRuntimeDetector(key, version);
            }
        });
    }

    openViewDetectorSidepanel(detector: DetectorExtended) {
        const config: KbqSidepanelConfig<RuntimeSidepanelDetectorProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                detector,
                permissions: this.permissions[PermissionName.SYSTEM],
                deleteHandler: this.openDeleteModal.bind(this)
            }
        };

        this.sidepanelService
            .open(RuntimeFeatureSidepanelDetectorComponent, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe();
    }

    switchCluster(id: string) {
        this.clusterStoreService.switchCluster(id);
    }
}
