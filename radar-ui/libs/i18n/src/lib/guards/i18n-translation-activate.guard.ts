import { ActivatedRouteSnapshot } from '@angular/router';
import { inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { LangDefinition, TranslocoService } from '@jsverse/transloco';
import { Observable, bufferCount, concatMap, from, map, of, tap } from 'rxjs';

import { I18nService } from '../services/i18n.service';

const translateDictionaryCollection = new Set<string>();

const i18nTranslationActivate = (route: ActivatedRouteSnapshot): Observable<boolean> => {
    const i18nService = inject(I18nService);
    const toastService = inject(KbqToastService);
    const translocoService = inject(TranslocoService);

    /* eslint @typescript-eslint/dot-notation: "off" */
    const translateDicts: string[] | undefined = route.data['translateDicts'];
    if (!translateDicts) {
        return of(true);
    }

    // @todo: divide languages and load them lazily during page transition
    const dicts = translocoService
        .getAvailableLangs()
        .map((lang: string | LangDefinition) => translateDicts.map((dict) => `${dict}/${String(lang)}`))
        .flat();

    return from(dicts).pipe(
        concatMap((dict) => {
            if (translateDictionaryCollection.has(dict)) {
                return of(true);
            }

            return translocoService.load(dict).pipe(
                map((translation) => !(translation && !Object.keys(translation).length)),
                tap((isLoaded) => {
                    if (isLoaded) {
                        translateDictionaryCollection.add(dict);
                    } else {
                        toastService.show({
                            style: KbqToastStyle.Error,
                            title: i18nService.translate('Common.Pseudo.Notification.TranslationLoadError')
                        });
                    }
                })
            );
        }),
        bufferCount(dicts.length),
        map((statuses) => statuses.every(Boolean))
    );
};

export const i18nTranslationActivateGuard = (route: ActivatedRouteSnapshot) => i18nTranslationActivate(route);

export const i18nTranslationActivateChildGuard = (route: ActivatedRouteSnapshot) => i18nTranslationActivate(route);
