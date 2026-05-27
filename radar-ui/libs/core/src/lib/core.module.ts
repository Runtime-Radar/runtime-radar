import { EffectsModule } from '@ngrx/effects';
import { StoreModule } from '@ngrx/store';
import { NgModule, Optional, SkipSelf } from '@angular/core';

import { AuthDomainModule } from '@cs/domains/auth';
import { ClusterDomainModule } from '@cs/domains/cluster';
import { LicenseDomainModule } from '@cs/domains/license';
import { RoleDomainModule } from '@cs/domains/role';
import { UserDomainModule } from '@cs/domains/user';

import { CoreNavigationEffectStore } from './stores/navigation/core-navigation-effect.store';
import { coreNavigationReducer } from './stores/navigation/core-navigation-selector.store';

@NgModule({
    imports: [
        AuthDomainModule,
        LicenseDomainModule,
        RoleDomainModule,
        ClusterDomainModule,
        UserDomainModule,
        StoreModule.forRoot(coreNavigationReducer),
        EffectsModule.forRoot([CoreNavigationEffectStore])
    ]
})
export class CoreModule {
    constructor(@Optional() @SkipSelf() parentModule: CoreModule) {
        if (parentModule) {
            throw new Error('core module is already loaded');
        }
    }
}
