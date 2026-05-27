import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';

import { KubeManagerPodPhase } from '@cs/domains/kube-manager';

import { InventorySidepanelNamespaceProps } from '../../interfaces/inventory-sidepanel.interface';

@Component({
    templateUrl: './inventory-sidepanel-namespace.component.html',
    styleUrls: ['./inventory-sidepanel-namespace.component.scss', '../inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class InventoryFeatureSidepanelNamespaceComponent {
    readonly podPhase = KubeManagerPodPhase;

    constructor(@Inject(KBQ_SIDEPANEL_DATA) public readonly props: InventorySidepanelNamespaceProps) {}
}
