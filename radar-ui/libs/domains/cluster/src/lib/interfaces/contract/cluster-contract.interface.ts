export enum ClusterStatus {
    REGISTERED = 'STATUS_REGISTERED',
    UNREGISTERED = 'STATUS_UNREGISTERED'
}

interface AbstractClusterDataBase {
    address?: string;
    user: string;
    password: string;
    use_tls: boolean;
    check_cert?: boolean;
    ca?: string;
    persistence?: boolean;
    storage_class?: string;
}

export interface ClusterPostgres extends AbstractClusterDataBase {
    database: string;
}

export type ClusterRedis = AbstractClusterDataBase;
export type ClusterClickhouse = ClusterPostgres;

export interface ClusterRegistry {
    address: string;
    user: string;
    password: string;
    image_short_names: boolean;
    chart_version?: string;
}

export interface ClusterGrafana {
    address?: string;
    user?: string;
    password?: string;
}

export interface ClusterPrometheus {
    deploy: boolean;
    persistence: boolean;
    storage_class?: string;
}

export interface ClusterIngress {
    ingress_class: string;
    hostname: string;
    cert?: string;
    cert_key?: string;
}

export interface ClusterNodePort {
    port: string;
}

export interface ClusterRabbit {
    address?: string;
    user: string;
    password: string;
    persistence?: boolean; // if internal, it is true
    storage_class?: string;
}

export interface ClusterConfig {
    version: string;
    proxy_url?: string;
    own_cs_url: string;
    central_cs_url: string;
    namespace: string;
    postgres: ClusterPostgres;
    clickhouse: ClusterClickhouse;
    redis: ClusterRedis;
    rabbit: ClusterRabbit;
    registry: ClusterRegistry;
    enable_metrics: boolean;
    grafana?: ClusterGrafana;
    prometheus?: ClusterPrometheus;
    ingress?: ClusterIngress;
    node_port?: ClusterNodePort;
}

export interface Cluster {
    id: string;
    name: string;
    created_at: string; // RFC3339
    status: ClusterStatus;
    config: ClusterConfig;
}

export interface RegisteredCluster {
    id: string;
    name: string;
    own_cs_url: string;
}
