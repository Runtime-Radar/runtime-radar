import { inject } from '@angular/core';
import { ActivatedRouteSnapshot, ResolveFn, Router } from '@angular/router';
import { EMPTY, catchError, of, switchMap } from 'rxjs';

import { RouterName } from '@cs/core';
import { Cluster, ClusterRequestService, ClusterStoreService, GetClusterResponse } from '@cs/domains/cluster';

export const clusterFeatureDetailsResolver: ResolveFn<GetClusterResponse> = (route: ActivatedRouteSnapshot) => {
    const clusterRequestService = inject(ClusterRequestService);
    const clusterStoreService = inject(ClusterStoreService);
    const router = inject(Router);
    const clusterId = route.paramMap.get('clusterId');

    const catchErrorHandler = () => {
        router.navigate([RouterName.ERROR]);

        return EMPTY;
    };

    if (!clusterId) {
        return catchErrorHandler();
    }

    return clusterStoreService.cluster$(clusterId).pipe(
        switchMap((cluster: Cluster | undefined) => {
            if (cluster && cluster.id) {
                return of({
                    cluster,
                    deleted: false
                });
            }

            return clusterRequestService.getCluster(clusterId).pipe(catchError(() => catchErrorHandler()));
        })
    );
};
