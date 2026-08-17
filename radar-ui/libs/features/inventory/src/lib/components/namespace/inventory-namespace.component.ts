import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output } from '@angular/core';

import { KubeManagerNamespace, KubeManagerPodExtended } from '@cs/domains/kube-manager';

import {
    InventorySidepanelContext,
    InventorySidepanelContextType,
    InventorySidepanelNamespaceOutputs
} from '../../interfaces/inventory-sidepanel.interface';

@Component({
    selector: 'cs-inventory-feature-namespace-component',
    templateUrl: './inventory-namespace.component.html',
    styleUrls: ['./inventory-namespace.component.scss', '../inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeatureNamespaceComponent {
    @Input({ required: true }) namespace!: KubeManagerNamespace;

    @Input({ required: true }) pods!: KubeManagerPodExtended[];

    @Input({ required: true }) context!: InventorySidepanelContext;

    @Output() headerClick = new EventEmitter<InventorySidepanelNamespaceOutputs>();

    readonly tooltipPlacements = PopUpPlacements;

    readonly sidepanelContextType = InventorySidepanelContextType;
}
