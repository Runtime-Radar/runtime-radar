import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import {
    AbstractControl,
    ControlValueAccessor,
    NG_VALIDATORS,
    NG_VALUE_ACCESSOR,
    ValidationErrors,
    Validators
} from '@angular/forms';
import { ChangeDetectionStrategy, Component, Input, inject } from '@angular/core';

import { RuntimeEventDateTimePeriod } from '../../interfaces/runtime-filter.interface';
import {
    RUNTIME_FILTER_DATETIME_PERIOD,
    RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR
} from '../../constants/runtime-filter.constant';

@Component({
    selector: 'cs-runtime-feature-datetime-period-picker-component',
    templateUrl: './runtime-datetime-period-picker.component.html',
    styleUrl: './runtime-datetime-period-picker.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false,
    providers: [
        {
            provide: NG_VALUE_ACCESSOR,
            useExisting: RuntimeFeatureDateTimePeriodPickerComponent,
            multi: true
        },
        {
            provide: NG_VALIDATORS,
            useExisting: RuntimeFeatureDateTimePeriodPickerComponent,
            multi: true
        }
    ]
})
export class RuntimeFeatureDateTimePeriodPickerComponent implements ControlValueAccessor {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);

    @Input() id?: string;

    @Input() testLocator?: string;

    readonly runtimeEventDateTimePeriodPresetOptions = RUNTIME_FILTER_DATETIME_PERIOD;

    readonly runtimeEventDateTimePeriod = RuntimeEventDateTimePeriod;

    period = ''; // RFC3339

    dateTimeFrom: DateTime | null = null;

    dateTimeTo: DateTime | null = null;

    isTouched = false;

    isDisabled = false;

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onChange = (period: string) => {};

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onTouched = () => {};

    registerOnChange(fn: any) {
        this.onChange = fn;
    }

    registerOnTouched(fn: any) {
        this.onTouched = fn;
    }

    markAsTouched() {
        if (!this.isTouched) {
            this.isTouched = true;
            this.onTouched();
        }
    }

    setDisabledState(isDisabled: boolean) {
        this.isDisabled = isDisabled;
    }

    writeValue(period: string | null) {
        if (period) {
            this.period = period;
            this.parsePeriod(period);
        } else {
            this.setDateTime(null, null);
        }
    }

    validate(control: AbstractControl): ValidationErrors | null {
        return control.hasValidator(Validators.required) && !control.value ? { required: true } : null;
    }

    changeDateFrom(from: DateTime | null) {
        this.dateTimeFrom = from;
        if (this.dateTimeFrom && this.dateTimeFrom.toMillis() > (this.dateTimeTo || this.dateTimeFrom).toMillis()) {
            this.dateTimeTo = null;
        }

        this.setPeriod();
        this.markAsTouched();
    }

    changeDateTo(to: DateTime | null) {
        this.dateTimeTo = to;
        this.setPeriod();
        this.markAsTouched();
    }

    selectPeriod(period: RuntimeEventDateTimePeriod) {
        const today = this.dateAdapter.today().startOf('second');

        switch (period) {
            case RuntimeEventDateTimePeriod.ONE_MINUTE:
                this.setDateTime(today.minus({ minute: 1 }), today);
                break;
            case RuntimeEventDateTimePeriod.TEN_MINUTES:
                this.setDateTime(today.minus({ minute: 10 }), today);
                break;
            case RuntimeEventDateTimePeriod.ONE_HOUR:
                this.setDateTime(today.minus({ hour: 1 }), today);
                break;
            case RuntimeEventDateTimePeriod.ONE_DAY:
                this.setDateTime(today.minus({ day: 1 }), today);
                break;
            case RuntimeEventDateTimePeriod.CUSTOM:
                this.setDateTime(today, today);
                break;
        }
        this.setPeriod();
    }

    private setDateTime(from: DateTime | null, to: DateTime | null) {
        this.dateTimeFrom = from;
        this.dateTimeTo = to;
    }

    private setPeriod() {
        let period = '';
        if (this.dateTimeTo || this.dateTimeFrom) {
            period = [
                this.dateTimeFrom ? this.dateTimeFrom.toJSDate().toISOString() : '', // RFC3339
                this.dateTimeTo ? this.dateTimeTo.toJSDate().toISOString() : '' // RFC3339
            ].join(RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR);
        }

        this.onChange(period);
    }

    private parsePeriod(period: string) {
        const [from, to] = period.split(RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR);
        this.setDateTime(from ? DateTime.fromISO(from) : null, to ? DateTime.fromISO(to) : null);
    }
}
