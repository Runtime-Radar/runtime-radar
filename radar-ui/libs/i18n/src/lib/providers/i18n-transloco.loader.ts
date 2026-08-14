import { HttpClient } from '@angular/common/http';
import { EMPTY, Observable, catchError, of } from 'rxjs';
import { Injectable, inject } from '@angular/core';
import { Translation, TranslocoLoader } from '@jsverse/transloco';

type TranslocoLoaderData = Parameters<TranslocoLoader['getTranslation']>[1];

@Injectable({
    providedIn: 'root'
})
export class I18nTranslocoLoader implements TranslocoLoader {
    private readonly http = inject(HttpClient);

    getTranslation(path: string, data?: TranslocoLoaderData): Observable<Translation> {
        if (!data) {
            return of(EMPTY);
        }

        const locale = path.split('/').reverse()[0];
        const translationPath = `/assets/i18n/${locale}/${data?.scope}.json`;

        return this.http.get<Translation>(translationPath).pipe(catchError(() => of({})));
    }
}
