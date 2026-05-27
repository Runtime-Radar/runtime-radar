import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { KbqAlertColors } from '@koobiq/components/alert';
import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';

import { RuntimeCapabilityType } from '@cs/domains/runtime';

import { RuntimeSidepanelThreatsProps } from '../../interfaces/runtime-sidepanel.interface';

@Component({
    templateUrl: './runtime-sidepanel-threats.component.html',
    styleUrl: './runtime-sidepanel-threats.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RuntimeFeatureSidepanelThreatsComponent {
    readonly isCapSysAdmin = this.props.effectives.includes(RuntimeCapabilityType.CAP_SYS_ADMIN);

    readonly alertColors = KbqAlertColors;

    readonly tooltipPlacements = PopUpPlacements;

    constructor(@Inject(KBQ_SIDEPANEL_DATA) public readonly props: RuntimeSidepanelThreatsProps) {}
}
