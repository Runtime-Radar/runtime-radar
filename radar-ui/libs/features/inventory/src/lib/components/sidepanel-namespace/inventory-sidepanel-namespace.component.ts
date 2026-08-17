import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';

import { KubeManagerPodPhase } from '@cs/domains/kube-manager';

import { InventorySidepanelNamespaceProps } from '../../interfaces/inventory-sidepanel.interface';

@Component({
    templateUrl: './inventory-sidepanel-namespace.component.html',
    styleUrls: ['./inventory-sidepanel-namespace.component.scss', '../inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeatureSidepanelNamespaceComponent {
    readonly props = inject<InventorySidepanelNamespaceProps>(KBQ_SIDEPANEL_DATA);

    readonly podPhase = KubeManagerPodPhase;
}
