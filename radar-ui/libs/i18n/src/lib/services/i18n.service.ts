import { Params } from '@angular/router';
import {
    DateAdapter,
    DateFormatter,
    KBQ_DATE_FORMATS,
    KBQ_LOCALE_SERVICE,
    KbqDateFormats,
    KbqLocaleService
} from '@koobiq/components/core';
import { DateTime, Settings } from 'luxon';
import { Inject, Injectable } from '@angular/core';
import { LangDefinition, TranslocoService } from '@jsverse/transloco';
import { Observable, bufferCount, concatMap, from, map, tap } from 'rxjs';

import { CoreWindowService } from '@cs/core';

import { I18nLocale } from '../interfaces/i18n.interface';
import {
    I18N_AVAILABLE_LOCALES,
    I18N_DATE_LOCALE_FORMATS,
    I18N_DEFAULT_LOCALE,
    I18N_LOCAL_STORAGE_KEY
} from '../constants/i18n.constant';

@Injectable({
    providedIn: 'root'
})
export class I18nService {
    private locale = this.translocoService.getActiveLang();

    constructor(
        private readonly dateAdapter: DateAdapter<DateTime>,
        private readonly dateFormatter: DateFormatter<DateTime>,
        private readonly translocoService: TranslocoService,
        private readonly coreWindowService: CoreWindowService,
        @Inject(KBQ_DATE_FORMATS) private readonly dateFormats: KbqDateFormats,
        @Inject(KBQ_LOCALE_SERVICE) private readonly kbqLocaleService: KbqLocaleService
    ) {}

    loadTranslation(names: string[]): Observable<boolean> {
        const dicts = this.translocoService
            .getAvailableLangs()
            .map((lang: string | LangDefinition) => names.map((dict) => `${dict}/${String(lang)}`))
            .flat();

        return from(dicts).pipe(
            concatMap((dict) =>
                this.translocoService
                    .load(dict)
                    .pipe(map((translation) => !(translation && !Object.keys(translation).length)))
            ),
            bufferCount(dicts.length),
            map((statuses) => statuses.every(Boolean)),
            tap((isLoaded) => {
                if (!isLoaded) {
                    console.warn('dicts must be loaded');
                }
            })
        );
    }

    getLocale(): string {
        return this.locale;
    }

    initLocale(locale: string) {
        this.locale = this.isLocaleAvailable(locale) ? locale : I18N_DEFAULT_LOCALE;
        this.kbqLocaleService.setLocale(this.locale);
        this.dateAdapter.setLocale(this.locale);
        this.dateFormatter.setLocale(this.locale);
        this.dateFormats.dateInput = I18N_DATE_LOCALE_FORMATS[this.locale as I18nLocale];
        this.translocoService.setActiveLang(this.locale);
        this.coreWindowService.localStorage.setItem(I18N_LOCAL_STORAGE_KEY, this.locale);
        Settings.defaultLocale = this.locale.split('-')[0];
    }

    translate(key: string, params: Params = {}): string {
        const scope = key.split('.')[0];

        return scope ? this.translocoService.translate(key, params, scope) : '';
    }

    private isLocaleAvailable(locale: string): boolean {
        return (
            I18N_AVAILABLE_LOCALES.find((availableLocale) => availableLocale === (locale as I18nLocale)) !== undefined
        );
    }
}
