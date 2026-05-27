import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';

import { InventorySidepanelContainerProps } from '../../interfaces/inventory-sidepanel.interface';

@Component({
    templateUrl: './inventory-sidepanel-container.component.html',
    styleUrls: ['./inventory-sidepanel-container.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class InventoryFeatureSidepanelContainerComponent {
    constructor(@Inject(KBQ_SIDEPANEL_DATA) public readonly props: InventorySidepanelContainerProps) {}
}
