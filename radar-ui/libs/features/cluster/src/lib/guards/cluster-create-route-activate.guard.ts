import { inject } from '@angular/core';
import { LoadStatus, RouterName } from '@cs/core';
import { Observable, filter, map, switchMap, tap } from 'rxjs';
import { Router, UrlTree } from '@angular/router';

import { AuthStoreService } from '@cs/domains/auth';
import { ClusterStoreService } from '@cs/domains/cluster';
import { PermissionName, PermissionType, Role, RoleStoreService } from '@cs/domains/role';

const clusterFeatureCreateRouteActivate = (): Observable<boolean | UrlTree> => {
    const router = inject(Router);

    const authStoreService = inject(AuthStoreService);
    const clusterStoreService = inject(ClusterStoreService);
    const roleStoreService = inject(RoleStoreService);

    return clusterStoreService.loadStatus$.pipe(
        filter((status) => [LoadStatus.LOADED, LoadStatus.ERROR].includes(status)),
        map((status) => status === LoadStatus.LOADED),
        tap((isLoaded) => {
            if (!isLoaded) {
                router.navigate([RouterName.ERROR]);
            }
        }),
        switchMap(() => {
            return authStoreService.credentials$.pipe(
                switchMap((credentials) => roleStoreService.role$(credentials.roleId)),
                map((role: Role | undefined) =>
                    role
                        ? role.role_permissions[PermissionName.CLUSTERS].actions.includes(PermissionType.CREATE)
                        : false
                )
            );
        }),
        tap((isActivate) => {
            if (!isActivate) {
                router.navigate([RouterName.CLUSTERS]);
            }
        })
    );
};

export const clusterFeatureCreateRouteActivateGuard = () => clusterFeatureCreateRouteActivate();
