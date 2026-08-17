import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import {
    BehaviorSubject,
    Observable,
    catchError,
    combineLatest,
    interval,
    map,
    of,
    startWith,
    switchMap,
    tap
} from 'rxjs';
import { ChangeDetectionStrategy, Component, Input, inject } from '@angular/core';

import { RuntimeRequestService } from '@cs/domains/runtime';

import { RuntimeCounterDateRange } from './runtime-event-counter.interface';

const RUNTIME_COUNTER_INTERVAL_MS = 10000;

const RUNTIME_COUNTER_THOUSAND_PLACE_LIMIT = 10000;

@Component({
    selector: 'cs-runtime-feature-event-counter-component',
    templateUrl: './runtime-event-counter.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureEventCounterComponent {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);

    private readonly runtimeRequestService = inject(RuntimeRequestService);

    @Input() set updateCounter(value: string | null) {
        if (value) {
            this.click$.next(value);
        }
    }

    private readonly click$ = new BehaviorSubject('');

    readonly count$: Observable<number> = combineLatest([
        this.click$,
        interval(RUNTIME_COUNTER_INTERVAL_MS).pipe(startWith(0))
    ]).pipe(
        map(() => this.getCounterDateRange(1, 10)),
        switchMap(({ from, to }) =>
            this.runtimeRequestService.getEventCount(from, to).pipe(
                switchMap((value) => {
                    /* eslint @typescript-eslint/no-magic-numbers: "off" */
                    if (value > 5) {
                        this.isPerSeconds = true;

                        return of(value);
                    }

                    const range = this.getCounterDateRange(3600);

                    return this.runtimeRequestService
                        .getEventCount(range.from, range.to)
                        .pipe(tap(() => (this.isPerSeconds = false)));
                }),
                catchError(() => of(0))
            )
        )
    );

    readonly thousandPlaceLimit = RUNTIME_COUNTER_THOUSAND_PLACE_LIMIT;

    isPerSeconds = true;

    private getCounterDateRange(durationInSec: number, todayShiftInSec = 0): RuntimeCounterDateRange {
        const today = this.dateAdapter.today().minus({ seconds: todayShiftInSec });

        return {
            from: today.minus({ seconds: durationInSec }).toJSDate().toISOString(), // RFC3339
            to: today.toJSDate().toISOString() // RFC3339
        };
    }
}
