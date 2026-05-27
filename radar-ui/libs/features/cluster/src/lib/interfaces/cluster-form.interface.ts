export interface ClusterRegistryForm {
    address: string;
    user: string;
    password: string;
    isImageShortName: boolean;
}

export interface ClusterDataBaseForm {
    isInternalCluster: boolean;
    user: string;
    password: string;
    database: string;
    isTls: boolean;
    hasCheckCert: boolean; // isTls
    ca: string; // isTls
    isPersistence: boolean; // internal
    storageClass: string; // internal && isPersistence
    address: string; // external
}

export interface ClusterRabbitForm {
    isInternalCluster: boolean;
    user: string;
    password: string;
    isPersistence: boolean; // internal
    storageClass: string; // internal && isPersistence
    address: string; // external
}

export interface ClusterMetricForm {
    isMetricEnabled: boolean;
    isInternalCluster: boolean;
    user: string; // internal
    password: string; // internal
    isPersistence: boolean; // internal
    storageClass: string; // internal && isPersistence
    address: string; // external
}

export interface ClusterIngressForm {
    isIngressEnabled: boolean;
    ingressClass: string;
    hostname: string;
    cert: string; // optional
    certKey: string; // optional
    isNodePortEnabled: boolean;
    port: string; // optional
}

export interface ClusterAccessForm {
    proxyUrl: string; // optional
    ownCsUrl: string;
    centralCsUrl: string;
    namespace: string;
    name: string;
}

export interface ClusterCreateFormOutputs<T> {
    form: T;
    isValid: boolean;
}
