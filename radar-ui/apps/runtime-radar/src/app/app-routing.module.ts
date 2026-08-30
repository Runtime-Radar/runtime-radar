import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { i18nTranslationActivateGuard } from '@cs/i18n';
import { DEFAULT_ROUTER_NAME, RouterName, TranslationDict } from '@cs/core';
import { PermissionName, rolePermissionActivateGuard, rolePermissionsResolver } from '@cs/domains/role';
import {
    authErrorRouteActivateGuard,
    authSuccessRouteActivateChildGuard,
    authSuccessRouteActivateGuard
} from '@cs/domains/auth';

const routes: Routes = [
    {
        path: RouterName.DEFAULT,
        redirectTo: DEFAULT_ROUTER_NAME,
        pathMatch: 'full'
    },
    {
        path: RouterName.SIGN_IN,
        loadChildren: () => import('@cs/features/sign-in').then((m) => m.SignInFeatureModule),
        canActivate: [authErrorRouteActivateGuard, i18nTranslationActivateGuard],
        data: {
            translateDicts: [TranslationDict.AUTH]
        }
    },
    {
        path: RouterName.INVENTORY,
        loadChildren: () => import('@cs/features/inventory').then((m) => m.InventoryFeatureModule),
        canActivate: [authSuccessRouteActivateGuard, i18nTranslationActivateGuard],
        data: {
            translateDicts: [TranslationDict.INVENTORY]
        }
    },
    {
        path: RouterName.CLUSTERS,
        loadChildren: () => import('@cs/features/cluster').then((m) => m.ClusterFeatureModule),
        canActivate: [authSuccessRouteActivateGuard, i18nTranslationActivateGuard, rolePermissionActivateGuard],
        resolve: {
            permissions: rolePermissionsResolver
        },
        data: {
            translateDicts: [TranslationDict.CLUSTER],
            permissions: [PermissionName.CLUSTERS],
            guards: [PermissionName.CLUSTERS]
        }
    },
    {
        path: RouterName.RUNTIME,
        loadChildren: () => import('@cs/features/runtime').then((m) => m.RuntimeFeatureModule),
        canActivate: [authSuccessRouteActivateGuard, i18nTranslationActivateGuard, rolePermissionActivateGuard],
        data: {
            translateDicts: [TranslationDict.RUNTIME, TranslationDict.RULE],
            guards: [PermissionName.SYSTEM]
        }
    },
    {
        path: RouterName.SETTINGS,
        canActivate: [authSuccessRouteActivateGuard],
        canActivateChild: [authSuccessRouteActivateChildGuard],
        children: [
            {
                path: RouterName.INTEGRATIONS,
                loadChildren: () => import('@cs/features/integration').then((m) => m.IntegrationFeatureModule),
                canActivate: [i18nTranslationActivateGuard, rolePermissionActivateGuard],
                resolve: {
                    permissions: rolePermissionsResolver
                },
                data: {
                    translateDicts: [TranslationDict.INTEGRATION],
                    permissions: [PermissionName.INTEGRATIONS, PermissionName.NOTIFICATIONS],
                    guards: [PermissionName.INTEGRATIONS]
                }
            },
            {
                path: RouterName.RULES,
                loadChildren: () => import('@cs/features/rule').then((m) => m.RuleFeatureModule),
                canActivate: [i18nTranslationActivateGuard, rolePermissionActivateGuard],
                resolve: {
                    permissions: rolePermissionsResolver
                },
                data: {
                    translateDicts: [TranslationDict.RULE],
                    permissions: [PermissionName.RULES],
                    guards: [PermissionName.RULES]
                }
            },
            {
                path: RouterName.USERS,
                loadChildren: () => import('@cs/features/user').then((m) => m.UserFeatureModule),
                canActivate: [i18nTranslationActivateGuard, rolePermissionActivateGuard],
                data: {
                    translateDicts: [TranslationDict.USER],
                    permissions: [PermissionName.USERS, PermissionName.SYSTEM],
                    guards: [PermissionName.USERS]
                }
            },
            {
                path: '**',
                redirectTo: RouterName.USERS
            }
        ]
    },
    {
        path: RouterName.TOKENS,
        loadChildren: () => import('@cs/features/token').then((m) => m.TokenFeatureModule),
        canActivate: [authSuccessRouteActivateGuard, i18nTranslationActivateGuard, rolePermissionActivateGuard],
        resolve: {
            permissions: rolePermissionsResolver
        },
        data: {
            translateDicts: [TranslationDict.TOKEN],
            permissions: [
                PermissionName.CLUSTERS,
                PermissionName.RULES,
                PermissionName.EVENTS,
                PermissionName.TOKENS,
                PermissionName.INVALIDATE_TOKENS,
                PermissionName.SYSTEM
            ],
            guards: [PermissionName.TOKENS]
        }
    },
    {
        path: RouterName.SWITCH,
        loadChildren: () => import('@cs/features/switch').then((m) => m.SwitchFeatureModule),
        canActivate: [i18nTranslationActivateGuard],
        data: {
            translateDicts: [TranslationDict.CLUSTER]
        }
    },
    {
        path: RouterName.FORBIDDEN,
        loadChildren: () => import('@cs/features/forbidden').then((m) => m.ForbiddenFeatureModule),
        canActivate: [authSuccessRouteActivateGuard]
    },
    {
        path: RouterName.ERROR,
        loadChildren: () => import('@cs/features/error').then((m) => m.ErrorFeatureModule),
        canActivate: [i18nTranslationActivateGuard],
        data: {
            translateDicts: [TranslationDict.COMMON]
        }
    },
    {
        path: '**',
        redirectTo: RouterName.ERROR,
        pathMatch: 'full'
    }
];

@NgModule({
    imports: [
        RouterModule.forRoot(routes, {
            bindToComponentInputs: true,
            enableTracing: false,
            scrollPositionRestoration: 'enabled',
            anchorScrolling: 'enabled',
            onSameUrlNavigation: 'reload'
        })
    ],
    exports: [RouterModule]
})
export class AppRoutingModule {}
