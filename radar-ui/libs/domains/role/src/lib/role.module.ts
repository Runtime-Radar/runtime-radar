import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { ApiModule } from '@cs/api';

import { RoleEffectStore } from './stores/role-effect.store';
import { ROLE_DOMAIN_KEY, roleDomainReducer } from './stores/role-selector.store';

@NgModule({
    imports: [
        ApiModule,
        EffectsModule.forFeature([RoleEffectStore]),
        StoreModule.forFeature(ROLE_DOMAIN_KEY, roleDomainReducer)
    ]
})
export class RoleDomainModule {}
