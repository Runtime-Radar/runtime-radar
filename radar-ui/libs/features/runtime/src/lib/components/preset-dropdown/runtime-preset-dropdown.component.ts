import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { ChangeDetectionStrategy, Component, EventEmitter, OnInit, Output, inject } from '@angular/core';

import { RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR } from '../../constants/runtime-filter.constant';
import { RuntimeEventFilters } from '../../interfaces/runtime-filter.interface';

@Component({
    selector: 'cs-runtime-feature-preset-dropdown-component',
    templateUrl: './runtime-preset-dropdown.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeaturePresetDropdownComponent implements OnInit {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);

    @Output() presetChange = new EventEmitter<RuntimeEventFilters>();

    readonly presetDropdownFiltersCollection = new Map<RuntimeEventFilters, string>();

    select(filters: RuntimeEventFilters) {
        this.presetChange.emit(filters);
    }

    ngOnInit() {
        const today = this.dateAdapter.today().startOf('second');

        this.presetDropdownFiltersCollection
            .set(
                this.getFilterSettings(today.minus({ minute: 1 }), today, true),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneMinuteWithThreats'
            )
            .set(
                this.getFilterSettings(today.minus({ hour: 1 }), today, true),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneHourWithThreats'
            )
            .set(
                this.getFilterSettings(today.minus({ day: 1 }), today, true),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneDayWithThreats'
            )
            .set(
                this.getFilterSettings(today.minus({ week: 1 }), today, true),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneWeekWithThreats'
            )
            .set(
                this.getFilterSettings(today.minus({ minute: 1 }), today),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneMinute'
            )
            .set(
                this.getFilterSettings(today.minus({ hour: 1 }), today),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneHour'
            )
            .set(
                this.getFilterSettings(today.minus({ day: 1 }), today),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneDay'
            )
            .set(
                this.getFilterSettings(today.minus({ week: 1 }), today),
                'Runtime.EventsPage.Filter.PresetDrowdown.Option.OneWeek'
            );
    }

    private getFilterSettings(from: DateTime, to: DateTime, hasThreats = false): RuntimeEventFilters {
        return {
            type: null,
            argument: '',
            binary: '',
            container: '',
            function: '',
            image: '',
            namespace: '',
            pod: '',
            period: from
                .toJSDate()
                .toISOString()
                .concat(RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR, to.toJSDate().toISOString()), // RFC3339
            hasThreats,
            hasIncident: false,
            detectors: [],
            rules: []
        };
    }
}
