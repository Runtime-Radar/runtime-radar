import { DateTime } from 'luxon';
import { BehaviorSubject, Observable, map, take, tap } from 'rxjs';
import { ChangeDetectionStrategy, Component, OnDestroy, OnInit, ViewChild, inject } from '@angular/core';
import { DateAdapter, PopUpPlacements } from '@koobiq/components/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { KbqSidepanelConfig, KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';

import { ApiPathService } from '@cs/api';
import { ClusterStoreService, RegisteredCluster } from '@cs/domains/cluster';
import { FormScheme, LoadStatus } from '@cs/core';
import {
    KubeManagerDetectorRatingPeriod,
    KubeManagerNamespaceGroup,
    KubeManagerNode,
    KubeManagerPodDetectorRating,
    KubeManagerPodExtended,
    KubeManagerRequestService,
    KubeManagerStoreService
} from '@cs/domains/kube-manager';

import { InventoryFeatureDragAreaDirective } from '../../directives/inventory-drag-area.directive';
import { InventoryFeatureSidepanelContainerComponent } from '../../components/sidepanel-container/inventory-sidepanel-container.component';
import { InventoryFeatureSidepanelContextService } from '../../services/inventory-sidepanel-context.service';
import { InventoryFeatureSidepanelNamespaceComponent } from '../../components/sidepanel-namespace/inventory-sidepanel-namespace.component';
import { InventoryFeatureSidepanelNodeContainer } from '../sidepanel-node/inventory-sidepanel-node.container';
import { InventoryFeatureSidepanelPodComponent } from '../../components/sidepanel-pod/inventory-sidepanel-pod.component';
import { InventoryFilters } from '../../interfaces/inventory-filter.interface';
import { INVENTORY_NODE_BG_COLORS, INVENTORY_NODE_TEXT_COLORS } from '../../constants/inventory-scheme.constant';
import {
    InventorySidepanelContainerOutputs,
    InventorySidepanelContainerProps,
    InventorySidepanelContextType,
    InventorySidepanelNamespaceOutputs,
    InventorySidepanelNamespaceProps,
    InventorySidepanelNodeProps,
    InventorySidepanelPodOutputs,
    InventorySidepanelPodProps
} from '../../interfaces/inventory-sidepanel.interface';

@Component({
    templateUrl: './inventory-map.container.html',
    styleUrls: ['./inventory-map.container.scss', '../../components/inventory-abstract.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class InventoryFeatureMapContainer implements OnInit, OnDestroy {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly formBuilder = inject(FormBuilder);
    private readonly sidepanelService = inject(KbqSidepanelService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly inventoryFeatureSidepanelContextService = inject(InventoryFeatureSidepanelContextService);
    private readonly kubeManagerRequestService = inject(KubeManagerRequestService);
    private readonly kubeManagerStoreService = inject(KubeManagerStoreService);

    @ViewChild(InventoryFeatureDragAreaDirective) dragAreaDirective!: InventoryFeatureDragAreaDirective;

    readonly filterForm: FormGroup<FormScheme<InventoryFilters>> = this.formBuilder.group({
        nodes: [[] as string[]],
        namespaces: [[] as string[]],
        pods: [[] as string[]],
        containers: [[] as string[]]
    });

    readonly defaultFilterValues: InventoryFilters = {
        nodes: [],
        namespaces: [],
        pods: [],
        containers: []
    };

    readonly selectedNode$ = new BehaviorSubject<string | null>(null);

    readonly sidepanelContext$ = this.inventoryFeatureSidepanelContextService.context$;

    readonly scaleValue$ = new BehaviorSubject(1);

    readonly loadStatus$: Observable<LoadStatus> = this.kubeManagerStoreService.loadStatus$;

    readonly nodes$: Observable<KubeManagerNode[]> = this.kubeManagerStoreService.nodes$.pipe(
        tap((nodes) =>
            nodes.forEach((item, i) => {
                this.nodeColorPairCollection.set(item.node_name, i);
            })
        )
    );

    readonly groupNamespaces$ = (node?: string): Observable<KubeManagerNamespaceGroup[]> =>
        this.kubeManagerStoreService.groupNamespaces$(node);

    readonly pods$ = (namespace: string, node?: string): Observable<KubeManagerPodExtended[]> =>
        this.kubeManagerStoreService.pods$(namespace, node).pipe(
            map((pods) =>
                pods.map((item) => {
                    const i = this.nodeColorPairCollection.get(item.node_name);
                    return {
                        ...item,
                        bgColor: i ? INVENTORY_NODE_BG_COLORS[i] : undefined,
                        textColor: i ? INVENTORY_NODE_TEXT_COLORS[i] : undefined
                    };
                })
            )
        );

    readonly detectorRating$ = (
        namespace: string,
        pod: string,
        container?: string
    ): Observable<KubeManagerPodDetectorRating[]> =>
        this.kubeManagerRequestService.getDetectorRating({
            count: 15,
            period: this.getTimestampPeriod(),
            filter: {
                pod_name: pod,
                pod_namespace: namespace,
                container_name: container || undefined
            }
        });

    readonly clusters$: Observable<RegisteredCluster[]> = this.clusterStoreService.registeredClusters$;

    readonly activeClusterHost$ = this.apiPathService.host$;

    readonly sidepanelContextType = InventorySidepanelContextType;

    readonly tooltipPlacements = PopUpPlacements;

    readonly loadStatus = LoadStatus;

    readonly nodeColorPairCollection = new Map<string, number>();

    readonly nodeBgColors = INVENTORY_NODE_BG_COLORS;

    readonly nodeTextColors = INVENTORY_NODE_TEXT_COLORS;

    ngOnInit() {
        this.kubeManagerStoreService.initPods();
    }

    ngOnDestroy() {
        if (this.inventoryFeatureSidepanelContextService.get().length) {
            this.sidepanelService.closeAll();
        }
    }

    switchCluster(id: string) {
        this.clusterStoreService.switchCluster(id);
    }

    zoomIn() {
        this.dragAreaDirective.zoom(0.8);
    }

    zoomOut() {
        this.dragAreaDirective.zoom(1.2);
    }

    zoomFit() {
        this.dragAreaDirective.fit();
    }

    zoomChange(scale: number) {
        this.scaleValue$.next(scale);
    }

    filterChange(filters: InventoryFilters) {
        this.kubeManagerStoreService.loadPods(filters);
    }

    openViewNodeSidepanel(node: KubeManagerNode) {
        const uuid = `node:${node.node_name}`;
        const config: KbqSidepanelConfig<InventorySidepanelNodeProps> = {
            id: uuid,
            position: KbqSidepanelPosition.Right,
            overlayPanelClass: 'inventory-feature-sidepanel-node',
            hasBackdrop: false,
            data: {
                node,
                nodeColorPairs: this.nodeColorPairCollection,
                namespaceHandler: this.openViewNamespaceSidepanel.bind(this),
                podHandler: this.openViewPodSidepanel.bind(this)
            }
        };

        this.selectedNode$.next(node.node_name);
        this.inventoryFeatureSidepanelContextService.slice({
            id: node.node_name,
            sidepanelId: uuid,
            path: node.node_name,
            type: InventorySidepanelContextType.NODE
        });
        this.sidepanelService
            .open(InventoryFeatureSidepanelNodeContainer, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe(() => {
                this.inventoryFeatureSidepanelContextService.remove(node.node_name);
                if (!this.inventoryFeatureSidepanelContextService.get().length) {
                    this.selectedNode$.next(null);
                }
            });
    }

    openViewNamespaceSidepanel(outputs: InventorySidepanelNamespaceOutputs) {
        const uuid = `namespace:${outputs.namespace.namespace}`;
        const config: KbqSidepanelConfig<InventorySidepanelNamespaceProps> = {
            id: uuid,
            position: KbqSidepanelPosition.Right,
            overlayPanelClass: 'inventory-feature-sidepanel-namespace',
            hasBackdrop: false,
            data: {
                namespace: outputs.namespace,
                pods: outputs.pods,
                podHandler: this.openViewPodSidepanel.bind(this)
            }
        };

        // @todo: make multiple nodes selection
        this.selectedNode$.next(outputs.namespace.nodes[0]);
        this.inventoryFeatureSidepanelContextService.slice({
            id: outputs.namespace.namespace,
            sidepanelId: uuid,
            path: `${outputs.namespace.nodes[0]}:${outputs.namespace.namespace}`,
            type: InventorySidepanelContextType.NAMESPACE
        });
        this.sidepanelService
            .open(InventoryFeatureSidepanelNamespaceComponent, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe(() => {
                this.inventoryFeatureSidepanelContextService.remove(outputs.namespace.namespace);
                if (!this.inventoryFeatureSidepanelContextService.get().length) {
                    this.selectedNode$.next(null);
                }
            });
    }

    openViewPodSidepanel(outputs: InventorySidepanelPodOutputs) {
        const uuid = `pod:${outputs.pod.uid}`;
        const config: KbqSidepanelConfig<InventorySidepanelPodProps> = {
            id: uuid,
            position: KbqSidepanelPosition.Right,
            overlayPanelClass: 'inventory-feature-sidepanel-pod',
            hasBackdrop: false,
            data: {
                pod: outputs.pod,
                detectors$: this.detectorRating$(outputs.pod.namespace, outputs.pod.name),
                containerHandler: this.openViewContainerSidepanel.bind(this)
            }
        };

        this.selectedNode$.next(outputs.pod.node_name);
        this.inventoryFeatureSidepanelContextService.slice({
            id: outputs.pod.uid,
            sidepanelId: uuid,
            path: `${outputs.pod.node_name}:${outputs.pod.namespace}:${outputs.pod.uid}`,
            type: InventorySidepanelContextType.POD
        });
        this.sidepanelService
            .open(InventoryFeatureSidepanelPodComponent, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe(() => {
                this.inventoryFeatureSidepanelContextService.remove(outputs.pod.uid);
                if (!this.inventoryFeatureSidepanelContextService.get().length) {
                    this.selectedNode$.next(null);
                }
            });
    }

    openViewContainerSidepanel(outputs: InventorySidepanelContainerOutputs) {
        const uuid = `container:${outputs.pod.name.split('').reverse().join('')}${outputs.container.name}`;
        const config: KbqSidepanelConfig<InventorySidepanelContainerProps> = {
            id: uuid,
            position: KbqSidepanelPosition.Right,
            overlayPanelClass: 'inventory-feature-sidepanel-container',
            hasBackdrop: false,
            data: {
                container: outputs.container,
                pod: outputs.pod,
                detectors$: this.detectorRating$(outputs.pod.namespace, outputs.pod.name, outputs.container.name)
            }
        };

        this.selectedNode$.next(outputs.pod.node_name);
        this.inventoryFeatureSidepanelContextService.slice({
            id: outputs.container.name,
            sidepanelId: uuid,
            path: `${outputs.pod.node_name}:${outputs.pod.namespace}:${outputs.pod.uid}:${outputs.container.name}`,
            type: InventorySidepanelContextType.CONTAINER
        });
        this.sidepanelService
            .open(InventoryFeatureSidepanelContainerComponent, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe(() => {
                this.inventoryFeatureSidepanelContextService.remove(outputs.container.name);
                if (!this.inventoryFeatureSidepanelContextService.get().length) {
                    this.selectedNode$.next(null);
                }
            });
    }

    private getTimestampPeriod(): KubeManagerDetectorRatingPeriod {
        const to = this.dateAdapter.today();

        return {
            to: to.toJSDate().toISOString(), // RFC3339
            from: to.minus({ week: 1 }).toJSDate().toISOString() // RFC3339
        };
    }
}
