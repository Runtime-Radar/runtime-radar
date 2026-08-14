import { Pipe, PipeTransform } from '@angular/core';

const HOURS_IN_DAY = 24;

@Pipe({
    name: 'hoursFormatter',
    standalone: false
})
export class SharedHoursFormatterPipe implements PipeTransform {
    transform(hours: number | null): number {
        if (!hours) {
            console.warn('hours must be provided');

            return 0;
        }

        return Math.round(hours / HOURS_IN_DAY);
    }
}
