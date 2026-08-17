import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Pipe, PipeTransform, inject } from '@angular/core';

import { I18nService } from '@cs/i18n';

@Pipe({
    name: 'tokenExpirationLabel',
    pure: false,
    standalone: false
})
export class TokenFeatureExpirationLabelPipe implements PipeTransform {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly i18nService = inject(I18nService);

    transform(dateStr: string, expiresAt?: string | null, invalidatedAt?: string): string {
        if (invalidatedAt) {
            return `${this.i18nService.translate('Token.Pseudo.Label.Revoked')} ${dateStr}`;
        }

        const dateTime = this.dateAdapter.parse(expiresAt, undefined);
        const value = dateTime?.diffNow('days').toObject().days || 0;

        return value <= 0
            ? `${this.i18nService.translate('Token.Pseudo.Label.Expired')} ${dateStr}`
            : `${this.i18nService.translate('Token.Pseudo.Label.Expirating')} ${dateStr}`;
    }
}
