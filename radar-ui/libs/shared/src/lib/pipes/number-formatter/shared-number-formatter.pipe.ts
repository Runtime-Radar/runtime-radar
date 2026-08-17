import { Pipe, PipeTransform, inject } from '@angular/core';

import { I18nService } from '@cs/i18n';

@Pipe({
    name: 'numberFormatter',
    pure: false,
    standalone: false
})
export class SharedNumberFormatterPipe implements PipeTransform {
    private readonly i18nService = inject(I18nService);

    transform(number: number | null | undefined): string {
        if (number === null || number === undefined) {
            console.warn('number must be provided');

            return '';
        }

        return Intl.NumberFormat(this.i18nService.getLocale()).format(number);
    }
}
