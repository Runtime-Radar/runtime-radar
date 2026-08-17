import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { AVAILABLE_LOCALES } from '@cs/core';
import { ApiModule } from '@cs/api';

import { MitreEffectStore } from './stores/mitre-effect.store';
import { MITRE_DOMAIN_KEY, MITRE_DOMAIN_REDUCER, mitreDomainReducerFactory } from './stores/mitre-selector.store';

@NgModule({
    imports: [
        ApiModule,
        EffectsModule.forFeature([MitreEffectStore]),
        StoreModule.forFeature(MITRE_DOMAIN_KEY, MITRE_DOMAIN_REDUCER)
    ],
    providers: [
        {
            provide: MITRE_DOMAIN_REDUCER,
            useFactory: mitreDomainReducerFactory,
            deps: [AVAILABLE_LOCALES]
        }
    ]
})
export class MitreDomainModule {}
