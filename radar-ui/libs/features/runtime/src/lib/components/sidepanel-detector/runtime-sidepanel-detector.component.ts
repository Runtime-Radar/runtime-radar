import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';

import { PermissionType } from '@cs/domains/role';
import { Router } from '@angular/router';
import { RouterName } from '@cs/core';

import { RuntimeRouterName } from '../../interfaces/runtime-navigation.interface';
import { RuntimeSidepanelDetectorProps } from '../../interfaces/runtime-sidepanel.interface';

@Component({
    templateUrl: './runtime-sidepanel-detector.component.html',
    styleUrls: ['./runtime-sidepanel-detector.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureSidepanelDetectorComponent {
    private readonly router = inject(Router);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    readonly props = inject<RuntimeSidepanelDetectorProps>(KBQ_SIDEPANEL_DATA);

    readonly permissionType = PermissionType;

    goToDetectorFilter() {
        this.sidepanelRef.close(undefined);
        this.router.navigate([RouterName.RUNTIME, RuntimeRouterName.EVENTS], {
            queryParams: {
                hasThreats: true,
                detectors: [this.props.detector.key]
            }
        });
    }

    deleteDetector() {
        if (this.props.deleteHandler) {
            this.props.deleteHandler(this.props.detector.key, this.props.detector.version);
            this.sidepanelRef.close(undefined);
        }
    }
}
