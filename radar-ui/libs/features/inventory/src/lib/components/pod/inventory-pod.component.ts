import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output } from '@angular/core';

import { KubeManagerPodExtended, KubeManagerPodPhase } from '@cs/domains/kube-manager';

import {
    InventorySidepanelContainerOutputs,
    InventorySidepanelContext,
    InventorySidepanelContextType,
    InventorySidepanelPodOutputs
} from '../../interfaces/inventory-sidepanel.interface';

@Component({
    selector: 'cs-inventory-feature-pod-component',
    templateUrl: './inventory-pod.component.html',
    styleUrls: ['./inventory-pod.component.scss', '../inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeaturePodComponent {
    @Input({ required: true }) pod!: KubeManagerPodExtended;

    @Input({ required: true }) context!: InventorySidepanelContext;

    @Output() headerClick = new EventEmitter<InventorySidepanelPodOutputs>();

    @Output() bucketClick = new EventEmitter<InventorySidepanelContainerOutputs>();

    readonly podPhase = KubeManagerPodPhase;

    readonly tooltipPlacements = PopUpPlacements;

    readonly sidepanelContextType = InventorySidepanelContextType;
}
