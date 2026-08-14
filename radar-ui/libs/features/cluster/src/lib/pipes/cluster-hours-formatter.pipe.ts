import { Pipe, PipeTransform, inject } from '@angular/core';

import { I18nService } from '@cs/i18n';

const HOURS_IN_DAY = 24;

@Pipe({
    name: 'clusterHoursFormatter',
    pure: false,
    standalone: false
})
export class ClusterFeatureHoursFormatterPipe implements PipeTransform {
    private readonly i18nService = inject(I18nService);

    transform(hours: number): string {
        if (hours <= 0) {
            return '';
        }

        const days = Math.floor(hours / HOURS_IN_DAY);
        const remainingHours = Math.floor(hours % HOURS_IN_DAY);
        const parts: string[] = [];

        if (days > 0) {
            parts.push(`${days} ${this.i18nService.translate('Cluster.Pseudo.Duration.Days', { count: days })}`);
        }

        if (remainingHours > 0) {
            parts.push(
                `${remainingHours} ${this.i18nService.translate('Cluster.Pseudo.Duration.Hours', { count: remainingHours })}`
            );
        }

        return parts.join(' ');
    }
}
