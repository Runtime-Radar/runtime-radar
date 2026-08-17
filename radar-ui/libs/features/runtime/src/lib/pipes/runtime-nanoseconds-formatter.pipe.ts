import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Pipe, PipeTransform, inject } from '@angular/core';

import { I18nService } from '@cs/i18n';

const DEFAULT_FRACTION = 9;

@Pipe({
    name: 'runtimeNanosecondsFormatter',
    pure: false,
    standalone: false
})
export class RuntimeFeatureNanosecondsFormatterPipe implements PipeTransform {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly i18nService = inject(I18nService);

    transform(date: string, fraction = DEFAULT_FRACTION): string {
        const deserializedDate = this.dateAdapter.deserialize(date);
        if (!deserializedDate) {
            console.warn('date must be valid');

            return '';
        }

        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        return (
            deserializedDate.setLocale(this.i18nService.getLocale()).toFormat('D, TT.') +
            date.substring(20, fraction >= 9 ? 29 : 20 + fraction)
        );
    }
}
