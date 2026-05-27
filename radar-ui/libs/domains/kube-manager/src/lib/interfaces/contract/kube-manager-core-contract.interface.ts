// kube-manager/vendor/k8s.io/api/core/v1/generated.proto
export interface KubeManagerPodSpec {
    volumes: unknown[];
    initContainers: unknown[];
    containers: unknown[];
    ephemeralContainers: unknown[];
    restartPolicy: string;
    terminationGracePeriodSeconds: number;
    activeDeadlineSeconds: number;
    dnsPolicy: string;
    nodeSelector: Map<string, string>;
    serviceAccountName: string;
    serviceAccount: string;
    automountServiceAccountToken: boolean;
    nodeName: string;
    hostNetwork: boolean;
    hostPID: boolean;
    hostIPC: boolean;
    shareProcessNamespace: boolean;
    securityContext: unknown;
    imagePullSecrets: unknown[];
    hostname: string;
    subdomain: string;
    affinity: unknown;
    schedulerName: string;
    tolerations: unknown[];
    hostAliases: unknown[];
    priorityClassName: string;
    priority: number;
    dnsConfig: unknown;
    readinessGates: unknown[];
    runtimeClassName: string;
    enableServiceLinks: boolean;
    preemptionPolicy: string;
    overhead: Map<string, unknown>;
    topologySpreadConstraints: unknown[];
    setHostnameAsFQDN: boolean;
    os: unknown;
    hostUsers: boolean;
    schedulingGates: unknown[];
    resourceClaims: unknown[];
    resources: unknown;
    hostnameOverride: string;
}

export interface KubeManagerPodStatus {
    observedGeneration: number;
    phase: string;
    conditions: unknown[];
    message: string;
    reason: string;
    nominatedNodeName: string;
    hostIP: string;
    hostIPs: unknown[];
    podIP: string;
    podIPs: unknown[];
    startTime: string;
    initContainerStatuses: unknown[];
    containerStatuses: unknown[];
    qosClass: string;
    ephemeralContainerStatuses: unknown[];
    resize: string;
    resourceClaimStatuses: unknown[];
    extendedResourceClaimStatus: unknown;
}

export interface KubeManagerNodeSpec {
    podCIDR: string;
    podCIDRs: string[];
    providerID: string;
    unschedulable: boolean;
    taints: unknown[];
    configSource: unknown;
    externalID: string;
}

export interface KubeManagerNodeStatus {
    capacity: Map<string, unknown>;
    allocatable: Map<string, unknown>;
    phase: string;
    conditions: unknown[];
    addresses: unknown[];
    daemonEndpoints: unknown;
    nodeInfo: unknown;
    images: unknown[];
    volumesInUse: string[];
    volumesAttached: unknown[];
    config: unknown;
    runtimeHandlers: unknown[];
    features: unknown;
}
