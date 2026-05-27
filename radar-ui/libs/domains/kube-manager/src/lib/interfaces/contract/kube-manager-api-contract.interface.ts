import {
    KubeManagerNodeMeta,
    KubeManagerPod,
    KubeManagerPodDetectorRating,
    KubeManagerPodMeta
} from './kube-manager-contract.interface';

interface KubeManagerDetectorRatingFilter {
    pod_name: string;
    pod_namespace: string;
    container_name?: string;
}

export interface KubeManagerDetectorRatingPeriod {
    from: string; // RFC3339
    to: string; // RFC3339
}

export interface GetKubeManagerPodsRequest {
    pods: string[];
    nodes: string[];
    namespaces: string[];
    containers: string[];
}

export interface GetKubeManagerPodsResponse {
    total: number;
    pods: KubeManagerPod[];
}

export interface GetKubeManagerPodRequest {
    name: string;
    namespace: string;
}

export interface GetKubeManagerPodResponse {
    pod: KubeManagerPodMeta;
}

export interface GetKubeManagerNodeRequest {
    name: string;
}

export interface GetKubeManagerNodeResponse {
    node: KubeManagerNodeMeta;
}

export interface GetKubeManagerDetectorRatingRequest {
    filter: KubeManagerDetectorRatingFilter;
    period: KubeManagerDetectorRatingPeriod;
    count: number;
}

export interface GetKubeManagerDetectorRatingResponse {
    counters: KubeManagerPodDetectorRating[];
}
