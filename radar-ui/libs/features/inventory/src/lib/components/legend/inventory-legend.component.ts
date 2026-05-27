import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, Input } from '@angular/core';

import { KubeManagerNode } from '@cs/domains/kube-manager';

import { INVENTORY_NODE_BG_COLORS } from '../../constants/inventory-scheme.constant';

@Component({
    selector: 'cs-inventory-feature-legend-component',
    templateUrl: './inventory-legend.component.html',
    styleUrls: ['./inventory-legend.component.scss', '../inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class InventoryFeatureLegendComponent {
    @Input({ required: true }) nodes!: KubeManagerNode[];

    readonly nodeBgColors = INVENTORY_NODE_BG_COLORS;

    readonly tooltipPlacements = PopUpPlacements;

    isCollapsed = false;

    collapse() {
        this.isCollapsed = !this.isCollapsed;
    }
}
