import { take } from 'rxjs';
import { Inject, Injectable } from '@angular/core';

import { ApiPathService } from '@cs/api';
import { AuthStoreService } from '@cs/domains/auth';
import { CoreWindowService, IS_CHILD_CLUSTER } from '@cs/core';
import { I18N_DEFAULT_LOCALE, I18N_LOCAL_STORAGE_KEY, I18nService } from '@cs/i18n';

import { CoreMetaService } from './core-meta.service';
import { DEFAULT_TRANSLATION_DICTS } from '../constants';

@Injectable({
    providedIn: 'root'
})
export class CoreInitService {
    constructor(
        private readonly apiPathService: ApiPathService,
        private readonly authStoreService: AuthStoreService,
        private readonly coreMetaService: CoreMetaService,
        private readonly i18nService: I18nService,
        private readonly coreWindowService: CoreWindowService,
        @Inject(IS_CHILD_CLUSTER) private readonly isChildCluster: boolean
    ) {}

    initialize(): Promise<void> {
        return new Promise((resolve) => {
            this.apiPathService.initialize();
            this.coreMetaService.initPageMetaTags();
            this.authStoreService.getLocationPath(this.coreWindowService.location);
            this.i18nService.initLocale(this.getLocaleFromStorage());
            this.i18nService
                .loadTranslation(DEFAULT_TRANSLATION_DICTS)
                .pipe(take(1))
                .subscribe(() => {
                    if (this.isChildCluster) {
                        this.authStoreService.redirectToSwitchRoute();
                    } else {
                        this.authStoreService.applyTokens();
                    }
                });

            resolve();
        });
    }

    private getLocaleFromStorage(): string {
        return this.coreWindowService.localStorage.getItem(I18N_LOCAL_STORAGE_KEY) || I18N_DEFAULT_LOCALE;
    }
}
