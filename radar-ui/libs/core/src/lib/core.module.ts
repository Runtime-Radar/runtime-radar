import { EffectsModule } from '@ngrx/effects';
import { StoreModule } from '@ngrx/store';
import { NgModule, inject } from '@angular/core';

import { AuthDomainModule } from '@cs/domains/auth';
import { ClusterDomainModule } from '@cs/domains/cluster';
import { LicenseDomainModule } from '@cs/domains/license';
import { MitreDomainModule } from '@cs/domains/mitre';
import { RoleDomainModule } from '@cs/domains/role';
import { UserDomainModule } from '@cs/domains/user';

import { CoreNavigationEffectStore } from './stores/navigation/core-navigation-effect.store';
import { coreNavigationReducer } from './stores/navigation/core-navigation-selector.store';

@NgModule({
    imports: [
        AuthDomainModule,
        ClusterDomainModule,
        LicenseDomainModule,
        MitreDomainModule,
        RoleDomainModule,
        UserDomainModule,
        EffectsModule.forRoot([CoreNavigationEffectStore]),
        StoreModule.forRoot(coreNavigationReducer)
    ]
})
export class CoreModule {
    constructor() {
        const parentModule = inject(CoreModule, { optional: true, skipSelf: true });

        if (parentModule) {
            throw new Error('core module is already loaded');
        }
    }
}
