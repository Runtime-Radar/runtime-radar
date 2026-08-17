import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { BrowserModule } from '@angular/platform-browser';
import { StoreDevtoolsModule } from '@ngrx/store-devtools';
import { NavigationActionTiming, RouterState, StoreRouterConnectingModule } from '@ngrx/router-store';
import { NgModule, inject, provideAppInitializer } from '@angular/core';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';
import { API_PATH, API_SINGLE_TENANT_PATHS } from '@cs/api';
import {
    AVAILABLE_LOCALES,
    CoreInitService,
    CoreModule,
    DEFAULT_LOCALE,
    IS_CHILD_CLUSTER,
    POLLING_INTERVAL,
    REFRESH_INTERVAL
} from '@cs/core';

import { AppContainer } from './app.container';
import { AppRoutingModule } from './app-routing.module';
import { NavbarComponent } from './components/navbar/navbar.component';
import { environment } from '../environments/environment';

@NgModule({
    imports: [
        AppRoutingModule,
        BrowserAnimationsModule,
        BrowserModule,
        CoreModule,
        SharedModule,
        I18nModule.forRoot({
            availableLangs: environment.availableLocales,
            defaultLang: environment.defaultLocale,
            fallbackLang: environment.defaultLocale,
            prodMode: environment.production
        }),
        StoreRouterConnectingModule.forRoot({
            navigationActionTiming: NavigationActionTiming.PostActivation,
            routerState: RouterState.Full
        }),
        StoreDevtoolsModule.instrument({
            logOnly: environment.production
        })
    ],
    providers: [
        {
            provide: API_PATH,
            useValue: environment.api
        },
        {
            provide: API_SINGLE_TENANT_PATHS,
            useValue: environment.singleTenant
        },
        {
            provide: AVAILABLE_LOCALES,
            useValue: environment.availableLocales
        },
        {
            provide: DEFAULT_LOCALE,
            useValue: environment.defaultLocale
        },
        {
            provide: POLLING_INTERVAL,
            useValue: environment.pollingInterval
        },
        {
            provide: REFRESH_INTERVAL,
            useValue: environment.refreshInterval
        },
        {
            provide: IS_CHILD_CLUSTER,
            useValue: environment.childCluster
        },
        provideHttpClient(withInterceptorsFromDi()),
        provideAppInitializer(() => inject(CoreInitService).initialize())
    ],
    declarations: [AppContainer, NavbarComponent],
    bootstrap: [AppContainer]
})
export class AppModule {}
