import { take } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import { ApiPathService } from '@cs/api';
import { AuthStoreService } from '@cs/domains/auth';
import { I18N_LOCAL_STORAGE_KEY, I18nService } from '@cs/i18n';

import { CoreMetaService } from './core-meta.service';
import { CoreWindowService } from './core-window.service';
import { DEFAULT_LOCALE } from '../tokens/core-locale.token';
import { DEFAULT_TRANSLATION_DICTS } from '../constants';
import { IS_CHILD_CLUSTER } from '../tokens/core-child-cluster.token';

@Injectable({
    providedIn: 'root'
})
export class CoreInitService {
    private readonly apiPathService = inject(ApiPathService);
    private readonly authStoreService = inject(AuthStoreService);
    private readonly coreMetaService = inject(CoreMetaService);
    private readonly coreWindowService = inject(CoreWindowService);
    private readonly defaultLocale = inject<string>(DEFAULT_LOCALE);
    private readonly i18nService = inject(I18nService);
    private readonly isChildCluster = inject<boolean>(IS_CHILD_CLUSTER);

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
        const value = this.coreWindowService.localStorage.getItem(I18N_LOCAL_STORAGE_KEY);

        return value || this.defaultLocale;
    }
}
