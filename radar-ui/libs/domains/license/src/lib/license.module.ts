import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { ApiModule } from '@cs/api';

import { LicenseEffectStore } from './stores/license-effect.store';
import { LICENSE_DOMAIN_KEY, licenseDomainReducer } from './stores/license-selector.store';

@NgModule({
    imports: [
        ApiModule,
        EffectsModule.forFeature([LicenseEffectStore]),
        StoreModule.forFeature(LICENSE_DOMAIN_KEY, licenseDomainReducer)
    ]
})
export class LicenseDomainModule {}
