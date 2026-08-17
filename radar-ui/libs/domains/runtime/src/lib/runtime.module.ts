import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { ApiModule } from '@cs/api';

import { RuntimeContextLocalizationPipe } from './pipes/runtime-context.pipe';
import { RuntimeEffectStore } from './stores/runtime-effect.store';
import { RuntimeEventTypeLocalizationPipe } from './pipes/runtime-event-type.pipe';
import { RUNTIME_DOMAIN_KEY, runtimeDomainReducer } from './stores/runtime-selector.store';

@NgModule({
    imports: [
        ApiModule,
        EffectsModule.forFeature([RuntimeEffectStore]),
        StoreModule.forFeature(RUNTIME_DOMAIN_KEY, runtimeDomainReducer)
    ],
    declarations: [RuntimeContextLocalizationPipe, RuntimeEventTypeLocalizationPipe],
    exports: [RuntimeContextLocalizationPipe, RuntimeEventTypeLocalizationPipe]
})
export class RuntimeDomainModule {}
