import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';

import { InventorySidepanelContainerProps } from '../../interfaces/inventory-sidepanel.interface';

@Component({
    templateUrl: './inventory-sidepanel-container.component.html',
    styleUrls: ['./inventory-sidepanel-container.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeatureSidepanelContainerComponent {
    readonly props = inject<InventorySidepanelContainerProps>(KBQ_SIDEPANEL_DATA);
}
