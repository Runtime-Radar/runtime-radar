import { KbqSidepanelService } from '@koobiq/components/sidepanel';
import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, Input, inject } from '@angular/core';

import { KubeManagerPodDetectorRating } from '@cs/domains/kube-manager';
import { RouterName } from '@cs/core';

@Component({
    selector: 'cs-inventory-feature-detector-table-component',
    templateUrl: './inventory-detector-table.component.html',
    styleUrl: './inventory-detector-table.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeatureDetectorTableComponent {
    private readonly sidepanelService = inject(KbqSidepanelService);

    @Input() detectors?: KubeManagerPodDetectorRating[] | null;

    readonly routerName = RouterName;

    readonly tooltipPlacements = PopUpPlacements;

    closeSidepanel() {
        this.sidepanelService.closeAll();
    }
}
