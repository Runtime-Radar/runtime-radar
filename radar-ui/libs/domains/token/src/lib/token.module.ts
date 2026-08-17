import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { ApiModule } from '@cs/api';

import { TokenEffectStore } from './stores/token-effect.store';
import { TOKEN_DOMAIN_KEY, tokenDomainReducer } from './stores/token-selector.store';

@NgModule({
    imports: [
        ApiModule,
        EffectsModule.forFeature([TokenEffectStore]),
        StoreModule.forFeature(TOKEN_DOMAIN_KEY, tokenDomainReducer)
    ]
})
export class TokenDomainModule {}
