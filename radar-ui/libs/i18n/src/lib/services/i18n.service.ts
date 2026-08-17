import { Params } from '@angular/router';
import { BehaviorSubject, Observable, bufferCount, concatMap, from, map, tap } from 'rxjs';
import { DateAdapter, DateFormatter, KBQ_DATE_FORMATS, KBQ_LOCALE_SERVICE } from '@koobiq/components/core';
import { DateTime, Settings } from 'luxon';
import { Injectable, inject } from '@angular/core';
import { LangDefinition, TranslocoService } from '@jsverse/transloco';

import { AVAILABLE_LOCALES, CoreWindowService, DEFAULT_LOCALE } from '@cs/core';

import { I18nLocale } from '../interfaces/i18n.interface';
import { I18N_DATE_LOCALE_FORMATS, I18N_LOCAL_STORAGE_KEY } from '../constants/i18n.constant';

@Injectable({
    providedIn: 'root'
})
export class I18nService {
    private readonly availableLocales = inject<string[]>(AVAILABLE_LOCALES);
    private readonly defaultLocale = inject<string>(DEFAULT_LOCALE);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly dateFormatter = inject<DateFormatter<DateTime>>(DateFormatter);
    private readonly dateFormats = inject(KBQ_DATE_FORMATS);
    private readonly kbqLocaleService = inject(KBQ_LOCALE_SERVICE);
    private readonly translocoService = inject(TranslocoService);

    private readonly coreWindowService = inject(CoreWindowService);

    private locale = this.translocoService.getActiveLang();

    readonly locale$ = new BehaviorSubject<string>(this.locale);

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

    setLocale(locale: string) {
        if (this.locale === locale) {
            return;
        }

        this.initLocale(locale);
    }

    initLocale(locale: string) {
        this.locale = this.isLocaleAvailable(locale) ? locale : this.defaultLocale;
        this.locale$.next(this.locale);
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
            this.availableLocales.find((availableLocale) => availableLocale === (locale as I18nLocale)) !== undefined
        );
    }
}
