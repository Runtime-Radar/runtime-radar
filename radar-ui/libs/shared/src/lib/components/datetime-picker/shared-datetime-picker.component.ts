import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { TimeFormats } from '@koobiq/components/timepicker';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output, inject } from '@angular/core';

@Component({
    selector: 'cs-datetime-picker-component',
    templateUrl: './shared-datetime-picker.component.html',
    styleUrl: './shared-datetime-picker.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedDatetimePickerComponent {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);

    @Input() id?: string;

    @Input() testLocator?: string;

    @Input() label?: string;

    @Input() minDate?: DateTime | null;

    @Input() set date(dateTime: DateTime | null) {
        this.dateTime = dateTime;
    }

    @Output() dateTimeChange = new EventEmitter<DateTime | null>();

    readonly timeFormats = TimeFormats;

    dateTime: DateTime | null = null;

    clearDateTime() {
        this.dateTime = null;
        this.dateTimeChange.emit(null);
    }

    changeDate(dateTime: DateTime | null) {
        const value = dateTime ? dateTime.set({ hour: 0, minute: 0, second: 0, millisecond: 0 }) : null;
        this.setDateTime(value);
    }

    changeTime(dateTime: DateTime | null) {
        this.setDateTime(dateTime);
    }

    private setDateTime(dateTime: DateTime | null) {
        const diffMillis = dateTime ? dateTime.diff(this.dateTime || this.dateAdapter.today()).milliseconds : 0;
        if (diffMillis !== 0) {
            this.dateTime = dateTime;
            this.dateTimeChange.emit(dateTime);
        }
    }
}
