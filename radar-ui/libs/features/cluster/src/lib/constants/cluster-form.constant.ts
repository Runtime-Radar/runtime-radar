import {
    ClusterAccessForm,
    ClusterDataBaseForm,
    ClusterIngressForm,
    ClusterMetricForm,
    ClusterRabbitForm,
    ClusterRegistryForm
} from '../interfaces/cluster-form.interface';

export const CLUSTER_REGISTRY_FORM_INITIAL_STATE: ClusterRegistryForm = {
    address: '',
    user: '',
    password: '',
    isImageShortName: false
};

export const CLUSTER_DATABASE_FORM_INITIAL_STATE: ClusterDataBaseForm = {
    isInternalCluster: true,
    user: 'runtime-radar',
    password: '',
    database: 'runtime-radar',
    isTls: true,
    hasCheckCert: true,
    ca: '',
    isPersistence: true,
    storageClass: '',
    address: ''
};

export const CLUSTER_REDIS_FORM_INITIAL_STATE: ClusterDataBaseForm = {
    ...CLUSTER_DATABASE_FORM_INITIAL_STATE,
    isPersistence: false,
    storageClass: ''
};

export const CLUSTER_RABBIT_FORM_INITIAL_STATE: ClusterRabbitForm = {
    isInternalCluster: true,
    user: 'runtime-radar',
    password: '',
    isPersistence: true,
    storageClass: '',
    address: ''
};

export const CLUSTER_METRIC_FORM_INITIAL_STATE: ClusterMetricForm = {
    isMetricEnabled: false,
    isInternalCluster: true,
    user: 'runtime-radar',
    password: '',
    isPersistence: true,
    storageClass: '',
    address: ''
};

export const CLUSTER_INGRESS_FORM_INITIAL_STATE: ClusterIngressForm = {
    isIngressEnabled: true,
    ingressClass: '',
    hostname: 'default.example.com',
    cert: '',
    certKey: '',
    isNodePortEnabled: true,
    port: ''
};

export const CLUSTER_ACCESS_FORM_INITIAL_STATE: ClusterAccessForm = {
    proxyUrl: '',
    ownCsUrl: '',
    centralCsUrl: '',
    namespace: 'runtime-radar',
    name: ''
};
