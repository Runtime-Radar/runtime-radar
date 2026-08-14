import {
    DateAdapter,
    DateFormatter,
    KBQ_DATE_FORMATS,
    KBQ_DATE_LOCALE,
    KBQ_LOCALE_SERVICE,
    KbqFormattersModule,
    KbqLocaleService
} from '@koobiq/components/core';
import {
    KBQ_LUXON_DATE_ADAPTER_OPTIONS,
    KBQ_LUXON_DATE_FORMATS,
    LuxonDateAdapter
} from '@koobiq/angular-luxon-adapter/adapter';
import { LOCALE_ID, ModuleWithProviders, NgModule, inject } from '@angular/core';
import { TRANSLOCO_MESSAGE_FORMAT_CONFIG, provideTranslocoMessageformat } from '@jsverse/transloco-messageformat';
import { TranslocoConfig, TranslocoModule, provideTransloco } from '@jsverse/transloco';
import {
    TranslocoMarkupModule,
    defaultTranslocoMarkupTranspilers,
    provideTranslationMarkupTranspiler
} from 'ngx-transloco-markup';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

import { AVAILABLE_LOCALES, DEFAULT_LOCALE } from '@cs/core';

import { I18nTemplateTranspiler } from './transpilers/i18n-template.transpiler';
import { I18nTranslocoLoader } from './providers/i18n-transloco.loader';

const DEFAULT_TRANSLOCO_CONFIG: Partial<TranslocoConfig> = {
    availableLangs: [],
    defaultLang: '',
    fallbackLang: '',
    interpolation: ['<<<', '>>>'],
    reRenderOnLangChange: true
};

const translocoMarkupTranspilers = [
    defaultTranslocoMarkupTranspilers(),
    provideTranslationMarkupTranspiler(I18nTemplateTranspiler)
];

@NgModule({
    imports: [KbqFormattersModule, TranslocoMarkupModule, TranslocoModule],
    providers: [
        {
            provide: LOCALE_ID,
            useFactory: () => inject<string>(DEFAULT_LOCALE)
        },
        {
            provide: DateFormatter,
            deps: [DateAdapter, KBQ_DATE_LOCALE]
        },
        ...translocoMarkupTranspilers
    ],
    exports: [TranslocoMarkupModule, TranslocoModule]
})
export class I18nModule {
    static forRoot(config: Partial<TranslocoConfig> = {}): ModuleWithProviders<I18nModule> {
        return {
            ngModule: I18nModule,
            providers: [
                provideTransloco({
                    config: {
                        ...DEFAULT_TRANSLOCO_CONFIG,
                        ...config
                    },
                    loader: I18nTranslocoLoader
                }),
                {
                    provide: TRANSLOCO_MESSAGE_FORMAT_CONFIG,
                    useFactory: () => {
                        const locales = inject<string[]>(AVAILABLE_LOCALES);
                        return { locales };
                    }
                },
                {
                    provide: KBQ_DATE_FORMATS,
                    useValue: KBQ_LUXON_DATE_FORMATS
                },
                {
                    provide: DateAdapter,
                    useClass: LuxonDateAdapter,
                    deps: [KBQ_DATE_LOCALE, KBQ_LUXON_DATE_ADAPTER_OPTIONS, KBQ_LOCALE_SERVICE]
                },
                {
                    provide: KBQ_LOCALE_SERVICE,
                    useClass: KbqLocaleService
                },
                provideTranslocoMessageformat(),
                provideHttpClient(withInterceptorsFromDi())
            ]
        };
    }
}
