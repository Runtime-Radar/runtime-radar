import { Observable } from 'rxjs';

import {
    KubeManagerContainer,
    KubeManagerNamespace,
    KubeManagerNode,
    KubeManagerPodDetectorRating,
    KubeManagerPodExtended
} from '@cs/domains/kube-manager';

export enum InventorySidepanelContextType {
    NONE,
    NODE,
    NAMESPACE,
    POD,
    CONTAINER
}

export interface InventorySidepanelContext {
    id: string;
    sidepanelId: string;
    path: string;
    type: InventorySidepanelContextType;
}

export interface InventorySidepanelNodeProps {
    node: KubeManagerNode;
    nodeColorPairs: Map<string, number>;
    namespaceHandler: (outputs: InventorySidepanelNamespaceOutputs) => void;
    podHandler: (outputs: InventorySidepanelPodOutputs) => void;
}

export interface InventorySidepanelNamespaceProps {
    namespace: KubeManagerNamespace;
    pods: KubeManagerPodExtended[];
    podHandler: (outputs: InventorySidepanelPodOutputs) => void;
}

export interface InventorySidepanelNamespaceOutputs {
    namespace: KubeManagerNamespace;
    pods: KubeManagerPodExtended[];
}

export interface InventorySidepanelPodProps {
    pod: KubeManagerPodExtended;
    detectors$: Observable<KubeManagerPodDetectorRating[]>;
    containerHandler: (outputs: InventorySidepanelContainerOutputs) => void;
}

export type InventorySidepanelPodOutputs = {
    pod: KubeManagerPodExtended;
};

export interface InventorySidepanelContainerProps {
    container: KubeManagerContainer;
    pod: KubeManagerPodExtended;
    detectors$: Observable<KubeManagerPodDetectorRating[]>;
}

export interface InventorySidepanelContainerOutputs {
    container: KubeManagerContainer;
    pod: KubeManagerPodExtended;
}
