import { KubeManagerPod } from '../contract/kube-manager-contract.interface';

export interface KubeManagerNode {
    id: string;
    node_name: string;
}

export interface KubeManagerNamespaceGroup {
    id: string;
    namespaces: KubeManagerNamespace[];
}

export interface KubeManagerNamespace {
    id: string;
    namespace: string;
    nodes: string[];
    podCount: number;
    isVisible: boolean | null;
}

export interface KubeManagerPodExtended extends KubeManagerPod {
    isVisible: boolean | null;
    bgColor?: string;
    textColor?: string;
}
