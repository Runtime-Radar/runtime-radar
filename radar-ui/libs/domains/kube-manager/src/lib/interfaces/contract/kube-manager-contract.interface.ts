import { RuleSeverity } from '@cs/domains/rule';

import {
    KubeManagerNodeSpec,
    KubeManagerNodeStatus,
    KubeManagerPodSpec,
    KubeManagerPodStatus
} from './kube-manager-core-contract.interface';

export enum KubeManagerPodPhase {
    PENDING = 'Pending',
    RUNNING = 'Running',
    SUCCEEDED = 'Succeeded',
    FAILED = 'Failed',
    UNKNOWN = 'Unknown'
}

export interface KubeManagerOwner {
    kind: string;
    name: string;
    scannable: boolean;
}

export interface KubeManagerContainer {
    name: string;
    image_url: string;
    image_digest: string;
}

export interface KubeManagerPod {
    uid: string;
    name: string;
    namespace: string;
    node_name: string;
    phase: KubeManagerPodPhase;
    containers: KubeManagerContainer[];
    owners: KubeManagerOwner[];
}

export interface KubeManagerPodMeta {
    metadata: unknown;
    spec: KubeManagerPodSpec;
    status: KubeManagerPodStatus;
}

export interface KubeManagerNodeMeta {
    metadata: unknown;
    spec: KubeManagerNodeSpec;
    status: KubeManagerNodeStatus;
}

export interface KubeManagerPodDetectorRating {
    detector_id: string;
    severity: RuleSeverity;
    count: number;
}
