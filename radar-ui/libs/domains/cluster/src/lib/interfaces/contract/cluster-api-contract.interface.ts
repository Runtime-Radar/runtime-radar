import { Cluster, RegisteredCluster } from './cluster-contract.interface';

export interface GetClustersResponse {
    clusters: Cluster[];
    total: number;
}

export interface GetRegisteredClustersResponse {
    clusters: RegisteredCluster[];
}

export interface GetClusterCommandResponse {
    cmd: string;
}

export interface GetClusterYamlResponse {
    yaml: string;
}

export interface GetClusterResponse {
    cluster: Cluster;
    deleted: boolean;
}

export type CreateClusterRequest = Pick<Cluster, 'name' | 'config'>;

export interface CreateClusterResponse {
    id: string;
}

export type UpdateClusterRequest = CreateClusterRequest;

export type EmptyClusterResponse = Record<string, unknown>;
