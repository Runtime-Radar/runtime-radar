import { Pipe, PipeTransform } from '@angular/core';

import { I18nService } from '@cs/i18n';

import { RuntimeEventFilterEntity } from '../interfaces/runtime-state.interface';
import { RuntimeEventFilters } from '../interfaces/runtime-filter.interface';
import { RuntimeFeatureConfigUtilsService as runtimeConfigUtils } from '../services/runtime-utils.service';

const RUNTIME_HISTORY_SEPARATOR = ', ';

const RUNTIME_HISTORY_LABELS = new Map<string, string>([
    ['type', 'Runtime.EventsPage.Filter.FilterPopover.Label.Type'],
    ['argument', 'Runtime.EventsPage.Filter.FilterPopover.Label.Argument'],
    ['binary', 'Runtime.EventsPage.Filter.FilterPopover.Label.Binary'],
    ['container', 'Runtime.EventsPage.Filter.FilterPopover.Label.Container'],
    ['function', 'Runtime.EventsPage.Filter.FilterPopover.Label.Function'],
    ['image', 'Runtime.EventsPage.Filter.FilterPopover.Label.Image'],
    ['namespace', 'Runtime.EventsPage.Filter.FilterPopover.Label.Namespace'],
    ['pod', 'Runtime.EventsPage.Filter.FilterPopover.Label.Pod'],
    ['period', 'Runtime.EventsPage.Filter.FilterPopover.Label.Period'],
    ['hasThreats', 'Runtime.EventsPage.Filter.FilterPopover.Abbr.HasThreats'],
    ['hasIncident', 'Runtime.EventsPage.Filter.FilterPopover.Abbr.HasIncident'],
    ['detectors', 'Runtime.EventsPage.Filter.FilterPopover.Label.Detectors'],
    ['rules', 'Runtime.EventsPage.Filter.FilterPopover.Label.Rules']
]);

@Pipe({
    name: 'runtimeHistoryLabel',
    pure: false
})
export class RuntimeFeatureHistoryLabelPipe implements PipeTransform {
    constructor(private readonly i18nService: I18nService) {}

    transform(entity: RuntimeEventFilterEntity | Partial<RuntimeEventFilters> | null, limit?: number): string {
        const labels = Object.entries(entity || {}).reduce((acc, [key, value]) => {
            const localizationKey = RUNTIME_HISTORY_LABELS.get(key);
            if (!!localizationKey && runtimeConfigUtils.isEventFilterValueValid(value)) {
                acc.push(this.i18nService.translate(localizationKey));
            }

            return acc;
        }, [] as string[]);

        if (limit && limit < labels.length) {
            return `${labels.slice(0, limit).join(RUNTIME_HISTORY_SEPARATOR) + RUNTIME_HISTORY_SEPARATOR}+${
                labels.length - limit
            }`;
        }

        return labels.join(RUNTIME_HISTORY_SEPARATOR);
    }
}
