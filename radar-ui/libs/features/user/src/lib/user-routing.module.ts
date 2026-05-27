import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { userActivateGuard } from '@cs/domains/user';
import { PermissionName, rolePermissionsResolver } from '@cs/domains/role';

import { UserFeatureListContainer } from './containers/list/user-list.container';

const routes: Routes = [
    {
        path: '',
        children: [
            {
                path: '',
                component: UserFeatureListContainer,
                canActivate: [userActivateGuard],
                resolve: {
                    permissions: rolePermissionsResolver
                },
                data: {
                    localizationTitleKey: 'User.ListPage.Header.Title',
                    permissions: [PermissionName.USERS]
                }
            }
        ]
    }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class UserFeatureRoutingModule {}
