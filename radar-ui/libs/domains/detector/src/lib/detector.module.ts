import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { ApiModule } from '@cs/api';

import { DetectorEffectStore } from './stores/detector-effect.store';
import { DETECTOR_DOMAIN_KEY, detectorDomainReducer } from './stores/detector-selector.store';

@NgModule({
    imports: [
        ApiModule,
        EffectsModule.forFeature([DetectorEffectStore]),
        StoreModule.forFeature(DETECTOR_DOMAIN_KEY, detectorDomainReducer)
    ]
})
export class DetectorDomainModule {}
