import { Injectable } from '@angular/core';
import { Observable, map } from 'rxjs';

import { ApiService } from '@cs/api';

import {
    GetKubeManagerDetectorRatingRequest,
    GetKubeManagerDetectorRatingResponse,
    GetKubeManagerNodeRequest,
    GetKubeManagerNodeResponse,
    GetKubeManagerPodRequest,
    GetKubeManagerPodResponse,
    GetKubeManagerPodsRequest,
    GetKubeManagerPodsResponse,
    KubeManagerNodeMeta,
    KubeManagerPod,
    KubeManagerPodDetectorRating,
    KubeManagerPodMeta
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class KubeManagerRequestService {
    constructor(private readonly apiService: ApiService) {}

    getPods(request: Partial<GetKubeManagerPodsRequest> = {}): Observable<KubeManagerPod[]> {
        return this.apiService
            .get<Partial<GetKubeManagerPodsRequest>, GetKubeManagerPodsResponse>('pod/list', request)
            .pipe(map((response) => response.pods));
    }

    /** @external */
    getPod(name: string, namespace: string): Observable<KubeManagerPodMeta> {
        return this.apiService
            .get<GetKubeManagerPodRequest, GetKubeManagerPodResponse>('pod', { name, namespace })
            .pipe(map((response) => response.pod));
    }

    /** @external */
    getNode(name: string): Observable<KubeManagerNodeMeta> {
        return this.apiService
            .get<GetKubeManagerNodeRequest, GetKubeManagerNodeResponse>('node', { name })
            .pipe(map((response) => response.node));
    }

    /** @external */
    getDetectorRating(request: GetKubeManagerDetectorRatingRequest): Observable<KubeManagerPodDetectorRating[]> {
        return this.apiService
            .post<
                GetKubeManagerDetectorRatingRequest,
                GetKubeManagerDetectorRatingResponse
            >('stats/detector/rating', request)
            .pipe(map((response) => response.counters));
    }
}
