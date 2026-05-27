import { KbqBadgeColors } from '@koobiq/components/badge';
import { KbqCodeBlockFile } from '@koobiq/components/code-block';
import { Router } from '@angular/router';
import { stringify } from 'yaml';
import { ChangeDetectionStrategy, Component, Inject } from '@angular/core';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Observable, map } from 'rxjs';

import { RouterName } from '@cs/core';
import { KubeManagerPodPhase, KubeManagerRequestService } from '@cs/domains/kube-manager';

import { InventorySidepanelPodProps } from '../../interfaces/inventory-sidepanel.interface';

@Component({
    templateUrl: './inventory-sidepanel-pod.component.html',
    styleUrls: ['./inventory-sidepanel-pod.component.scss', '../inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class InventoryFeatureSidepanelPodComponent {
    readonly codeConfigFiles$: Observable<KbqCodeBlockFile[]> = this.kubeManagerRequestService
        .getPod(this.props.pod.name, this.props.pod.namespace)
        .pipe(
            map((meta) => [
                {
                    filename: `${this.props.pod.name}.yaml`,
                    content: stringify(meta),
                    language: 'yaml'
                }
            ])
        );

    readonly badgeColors = KbqBadgeColors;

    readonly podPhase = KubeManagerPodPhase;

    constructor(
        private readonly router: Router,
        private readonly kubeManagerRequestService: KubeManagerRequestService,
        private readonly sidepanelService: KbqSidepanelService,
        @Inject(KBQ_SIDEPANEL_DATA) public readonly props: InventorySidepanelPodProps
    ) {}

    goToRuntimeEventPage() {
        this.sidepanelService.closeAll();
        // @todo: replace static route to variable from runtime module
        this.router.navigate([RouterName.RUNTIME, 'events'], {
            queryParams: {
                namespace: this.props.pod.namespace,
                pod: this.props.pod.name
            }
        });
    }
}
