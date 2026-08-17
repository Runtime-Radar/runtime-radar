import { DateTime } from 'luxon';
import { TranslocoService } from '@jsverse/transloco';
import { of } from 'rxjs';
import { render } from '@testing-library/angular';
import {
    DateAdapter,
    DateFormatter,
    KBQ_DATE_FORMATS,
    KBQ_LOCALE_SERVICE,
    KbqLocaleService
} from '@koobiq/components/core';
import { createSpyFromClass, provideAutoSpy } from 'jest-auto-spies';

import { AVAILABLE_LOCALES, CoreWindowService, DEFAULT_LOCALE, TranslationDict } from '@cs/core';

import { I18nLocale } from '../interfaces/i18n.interface';
import { I18nService } from './i18n.service';

const dateProviders = [
    {
        provide: DateAdapter,
        useValue: {
            today: () => DateTime.fromISO('2025-09-18T12:00:00Z'),
            setLocale: jest.fn(),
            getLocale: jest.fn()
        }
    },
    { provide: KBQ_DATE_FORMATS, useValue: {} },
    provideAutoSpy(DateFormatter)
];

const kbqProviders = [
    { provide: KBQ_DATE_FORMATS, useValue: {} },
    {
        provide: KBQ_LOCALE_SERVICE,
        useValue: {
            setLocale: jest.fn(),
            getLocale: jest.fn(() => I18nLocale.EN)
        }
    }
];

describe('I18nService', () => {
    let coreWindowService: jest.Mocked<CoreWindowService>;
    let dateAdapter: DateAdapter<unknown>;
    let dateFormatter: DateFormatter<unknown>;
    let i18nService: I18nService;
    let localeService: jest.Mocked<KbqLocaleService>;
    let translocoService: jest.Mocked<TranslocoService>;

    const availableLocales = ['en-US'];
    const defaultLocale = 'en-US';

    beforeEach(async () => {
        translocoService = createSpyFromClass(TranslocoService);
        translocoService.getActiveLang.mockReturnValue(I18nLocale.EN);
        translocoService.getAvailableLangs.mockReturnValue([I18nLocale.EN]);
        translocoService.translate.mockReturnValue('');

        coreWindowService = {
            localStorage: {
                getItem: jest.fn(() => 'true'),
                setItem: jest.fn()
            } as unknown as Storage
        } as jest.Mocked<CoreWindowService>;
        const coreWindowProviders = [
            {
                provide: CoreWindowService,
                useValue: coreWindowService
            }
        ];
        const { fixture } = await render('<div></div>', {
            providers: [
                ...coreWindowProviders,
                ...dateProviders,
                ...kbqProviders,
                {
                    provide: TranslocoService,
                    useValue: translocoService
                },
                {
                    provide: AVAILABLE_LOCALES,
                    useValue: availableLocales
                },
                {
                    provide: DEFAULT_LOCALE,
                    useValue: defaultLocale
                }
            ]
        });

        translocoService = fixture.debugElement.injector.get(TranslocoService) as jest.Mocked<TranslocoService>;
        localeService = fixture.debugElement.injector.get(KBQ_LOCALE_SERVICE) as jest.Mocked<KbqLocaleService>;
        i18nService = fixture.debugElement.injector.get(I18nService);
        dateAdapter = fixture.debugElement.injector.get(DateAdapter);
        dateFormatter = fixture.debugElement.injector.get(DateFormatter);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('getLocale', () => {
        it('should return current locale', () => {
            expect(i18nService.getLocale()).toEqual(I18nLocale.EN);
        });
    });

    describe('setLocale', () => {
        let spyInitLocale: jest.SpyInstance;

        beforeEach(() => {
            spyInitLocale = jest.spyOn(i18nService, 'initLocale');
        });

        it('should not call initLocale when locale is not changed', () => {
            i18nService.setLocale(I18nLocale.EN);

            expect(spyInitLocale).not.toHaveBeenCalled();
        });

        it('should call initLocale', () => {
            i18nService.setLocale(I18nLocale.EN);

            expect(spyInitLocale).toHaveBeenCalledWith(I18nLocale.EN);
        });
    });

    describe('initLocale', () => {
        it('should return default locale when provided is not available', () => {
            i18nService['isLocaleAvailable'] = () => false;
            i18nService.initLocale('xx-XX');

            expect(localeService.setLocale).toHaveBeenCalledWith(defaultLocale);
        });

        it('should update locale', () => {
            i18nService['isLocaleAvailable'] = () => true;
            i18nService.initLocale(I18nLocale.EN);

            expect(localeService.setLocale).toHaveBeenCalledWith(I18nLocale.EN);
            expect(dateAdapter.setLocale).toHaveBeenCalledWith(I18nLocale.EN);
            expect(dateFormatter.setLocale).toHaveBeenCalledWith(I18nLocale.EN);
            expect(translocoService.setActiveLang).toHaveBeenCalledWith(I18nLocale.EN);
        });

        it('should set correct local to localStorage', () => {
            i18nService['isLocaleAvailable'] = () => true;
            i18nService.initLocale(I18nLocale.EN);

            expect(coreWindowService.localStorage.setItem).toHaveBeenCalledWith('locale', I18nLocale.EN);
        });
    });

    describe('loadTranslation', () => {
        it('should return true if all dicts loaded', (done) => {
            translocoService.getAvailableLangs.mockReturnValue(availableLocales);
            translocoService.load.mockReturnValue(of({ ok: true }));

            i18nService.loadTranslation([TranslationDict.COMMON, TranslationDict.AUTH]).subscribe((result) => {
                expect(result).toBe(true);
                done();
            });
        });

        it('should call transloco load with args', () => {
            translocoService.getAvailableLangs.mockReturnValue(availableLocales);
            translocoService.load.mockReturnValue(of({ ok: true }));

            i18nService.loadTranslation([TranslationDict.COMMON]).subscribe();

            expect(translocoService.load).toHaveBeenCalledWith(`${TranslationDict.COMMON}/${I18nLocale.EN}`);
        });

        it('should return false if any dict is empty', (done) => {
            // eslint-disable-next-line @typescript-eslint/no-empty-function
            jest.spyOn(console, 'warn').mockImplementation(() => {});

            translocoService.getAvailableLangs.mockReturnValue(availableLocales);
            translocoService.load.mockReturnValue(of({}));

            i18nService.loadTranslation([TranslationDict.COMMON]).subscribe((result) => {
                expect(result).toBe(false);
                done();
            });
        });
    });

    describe('translate', () => {
        it('should return empty string if scope does not exist', () => {
            expect(i18nService.translate('', { value: 1 })).toEqual('');
            expect(translocoService.translate).not.toHaveBeenCalled();
        });

        it('should call translate with correct args', () => {
            i18nService.translate('Common.Pseudo.Text.AppName', { value: 1 });

            expect(translocoService.translate).toHaveBeenCalledWith(
                'Common.Pseudo.Text.AppName',
                { value: 1 },
                'Common'
            );
        });
    });
});
