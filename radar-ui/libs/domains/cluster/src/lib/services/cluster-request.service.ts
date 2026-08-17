import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    Cluster,
    CreateClusterRequest,
    CreateClusterResponse,
    EmptyClusterResponse,
    GetClusterCommandResponse,
    GetClusterResponse,
    GetClusterYamlResponse,
    GetClustersResponse,
    GetRegisteredClustersResponse,
    RegisteredCluster,
    UpdateClusterRequest
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class ClusterRequestService {
    private readonly apiService = inject(ApiService);

    getClusters(): Observable<Cluster[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetClustersResponse>('cluster/page/1?page_size=1000')
            .pipe(map((response) => response.clusters));
    }

    getRegisteredClusters(): Observable<RegisteredCluster[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetRegisteredClustersResponse>('cluster/registered')
            .pipe(map((response) => response.clusters));
    }

    /** @external */
    getInstallClusterCommand(id: string, isYamlCommand = false): Observable<string> {
        return this.apiService
            .get<ApiEmptyRequest, GetClusterCommandResponse>(`cluster/${id}/cmd/install`, {
                use_values_file: isYamlCommand
            })
            .pipe(map((response) => response.cmd));
    }

    /** @external */
    getUninstallClusterCommand(id: string): Observable<string> {
        return this.apiService
            .get<ApiEmptyRequest, GetClusterCommandResponse>(`cluster/${id}/cmd/uninstall`)
            .pipe(map((response) => response.cmd));
    }

    /** @external */
    getClusterYaml(id: string): Observable<string> {
        return this.apiService
            .get<ApiEmptyRequest, GetClusterYamlResponse>(`cluster/${id}/cmd/values`)
            .pipe(map((response) => response.yaml));
    }

    getCluster(id: string): Observable<GetClusterResponse> {
        return this.apiService.get<ApiEmptyRequest, GetClusterResponse>(`cluster/${id}`);
    }

    createCluster(request: CreateClusterRequest): Observable<Cluster> {
        return this.apiService.post<CreateClusterRequest, CreateClusterResponse>('cluster', request).pipe(
            map((response) => response.id),
            filter((id) => !!id),
            switchMap((id) =>
                this.getCluster(id).pipe(
                    take(1),
                    map((response) => response.cluster)
                )
            )
        );
    }

    updateCluster(id: string, request: UpdateClusterRequest): Observable<Cluster> {
        return this.apiService.patch<UpdateClusterRequest, EmptyClusterResponse>(`cluster/${id}`, request).pipe(
            filter((response) => response && !Object.keys(response).length),
            switchMap(() =>
                this.getCluster(id).pipe(
                    take(1),
                    map((response) => response.cluster)
                )
            )
        );
    }

    deleteCluster(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyClusterResponse>(`cluster/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }
}
