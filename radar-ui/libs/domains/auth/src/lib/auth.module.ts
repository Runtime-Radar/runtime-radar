import { EffectsModule } from '@ngrx/effects';
import { StoreModule } from '@ngrx/store';
import { ErrorHandler, NgModule } from '@angular/core';

import { ApiModule } from '@cs/api';
import { RoleDomainModule } from '@cs/domains/role';

import { AUTH_HEADERS_PROVIDER } from './providers/auth-headers.provider';
import { AuthChunksLoadHandler } from './handlers/auth-chunks-load.handler';
import { AuthEffectStore } from './stores/auth-effect.store';
import { AUTH_DOMAIN_KEY, authDomainReducer } from './stores/auth-selector.store';

@NgModule({
    imports: [
        ApiModule,
        RoleDomainModule,
        EffectsModule.forFeature([AuthEffectStore]),
        StoreModule.forFeature(AUTH_DOMAIN_KEY, authDomainReducer)
    ],
    providers: [
        AUTH_HEADERS_PROVIDER,
        {
            provide: ErrorHandler,
            useClass: AuthChunksLoadHandler
        }
    ]
})
export class AuthDomainModule {}
