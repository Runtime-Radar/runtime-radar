import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { KbqCodeBlockFile } from '@koobiq/components/code-block';
import { stringify } from 'yaml';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import {
    KubeManagerNamespaceGroup,
    KubeManagerPodExtended,
    KubeManagerPodPhase,
    KubeManagerRequestService,
    KubeManagerStoreService
} from '@cs/domains/kube-manager';

import { InventorySidepanelNodeProps } from '../../interfaces/inventory-sidepanel.interface';
import { INVENTORY_NODE_BG_COLORS, INVENTORY_NODE_TEXT_COLORS } from '../../constants/inventory-scheme.constant';

@Component({
    templateUrl: './inventory-sidepanel-node.container.html',
    styleUrls: ['./inventory-sidepanel-node.container.scss', '../../components/inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeatureSidepanelNodeContainer {
    private readonly kubeManagerRequestService = inject(KubeManagerRequestService);
    private readonly kubeManagerStoreService = inject(KubeManagerStoreService);

    readonly props = inject<InventorySidepanelNodeProps>(KBQ_SIDEPANEL_DATA);

    readonly groupNamespaces$ = (node?: string): Observable<KubeManagerNamespaceGroup[]> =>
        this.kubeManagerStoreService.groupNamespaces$(node);

    readonly pods$ = (namespace: string, node?: string): Observable<KubeManagerPodExtended[]> =>
        this.kubeManagerStoreService.pods$(namespace, node).pipe(
            map((pods) =>
                pods
                    .filter((item) => item.isVisible)
                    .map((item) => {
                        const i = this.props.nodeColorPairs.get(item.node_name);
                        return {
                            ...item,
                            bgColor: i ? INVENTORY_NODE_BG_COLORS[i] : undefined,
                            textColor: i ? INVENTORY_NODE_TEXT_COLORS[i] : undefined
                        };
                    })
            )
        );

    readonly codeConfigFiles$: Observable<KbqCodeBlockFile[]> = this.kubeManagerRequestService
        .getNode(this.props.node.node_name)
        .pipe(
            map((meta) => [
                {
                    filename: `${this.props.node.node_name}.yaml`,
                    content: stringify(meta),
                    language: 'yaml'
                }
            ])
        );

    readonly podPhase = KubeManagerPodPhase;
}
