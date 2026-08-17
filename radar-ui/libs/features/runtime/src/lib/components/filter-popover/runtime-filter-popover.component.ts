import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { BehaviorSubject, Observable, distinctUntilChanged, map, tap } from 'rxjs';
import {
    ChangeDetectionStrategy,
    Component,
    DestroyRef,
    EventEmitter,
    Input,
    OnInit,
    Output,
    inject
} from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';

import { DetectorExtended } from '@cs/domains/detector';
import { RUNTIME_EVENT_TYPE } from '@cs/domains/runtime';
import { FormScheme, CoreUtilsService as utils } from '@cs/core';

import { RUNTIME_FILTER_INITIAL_STATE } from '../../constants/runtime-filter.constant';
import {
    RUNTIME_FILTER_POPOVER_CONTROL_LABELS,
    RUNTIME_FILTER_POPOVER_CONTROL_PLACEHOLDERS,
    RUNTIME_FILTER_POPOVER_VISIBLE_CONTROL_KEYS
} from './runtime-filter-popover.constant';
import {
    RuntimeEventFilterKey,
    RuntimeEventFilterRuleNode,
    RuntimeEventFilters
} from '../../interfaces/runtime-filter.interface';

@Component({
    selector: 'cs-runtime-feature-filter-popover-component',
    templateUrl: './runtime-filter-popover.component.html',
    styleUrl: './runtime-filter-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureFilterPopoverComponent implements OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    readonly ruleNodes$ = new BehaviorSubject<RuntimeEventFilterRuleNode[] | null>(null);
    @Input() set ruleNodes(values: RuntimeEventFilterRuleNode[] | null) {
        this.ruleNodes$.next(values);
    }

    @Input() detectors?: DetectorExtended[] | null;

    @Input({ required: true }) set filters(values: RuntimeEventFilters | null) {
        const filters: Partial<RuntimeEventFilters> =
            values && Object.keys(values).length
                ? values
                : utils.getFormValues<RuntimeEventFilters>(this.filterForm.controls);
        let hiddenKeys: RuntimeEventFilterKey[] = this.formControlKeys.filter(
            (key) => !RUNTIME_FILTER_POPOVER_VISIBLE_CONTROL_KEYS.includes(key)
        );

        Object.entries(filters)
            .filter(([key, _]) => this.formControlKeys.includes(key as RuntimeEventFilterKey))
            .forEach(([key, value]) => {
                const symbol = key as RuntimeEventFilterKey; // @todo: create type guard to remove 'as' syntax
                if (value) {
                    hiddenKeys = hiddenKeys.filter((item) => item !== symbol);
                }
            });

        this.filterForm.patchValue(filters);
        this.selectedFilters = { ...filters };
        this.hiddenDropdownMenuFilterKey = [...hiddenKeys];
    }

    @Output() filterChange = new EventEmitter<RuntimeEventFilters>();

    readonly filterForm: FormGroup<FormScheme<RuntimeEventFilters>> = this.formBuilder.group({
        type: RUNTIME_FILTER_INITIAL_STATE.type,
        hasThreats: RUNTIME_FILTER_INITIAL_STATE.hasThreats,
        hasIncident: RUNTIME_FILTER_INITIAL_STATE.hasIncident,
        argument: RUNTIME_FILTER_INITIAL_STATE.argument,
        binary: RUNTIME_FILTER_INITIAL_STATE.binary,
        container: RUNTIME_FILTER_INITIAL_STATE.container,
        function: RUNTIME_FILTER_INITIAL_STATE.function,
        image: RUNTIME_FILTER_INITIAL_STATE.image,
        namespace: RUNTIME_FILTER_INITIAL_STATE.namespace,
        pod: RUNTIME_FILTER_INITIAL_STATE.pod,
        period: RUNTIME_FILTER_INITIAL_STATE.period,
        detectors: [RUNTIME_FILTER_INITIAL_STATE.detectors],
        rules: [RUNTIME_FILTER_INITIAL_STATE.rules]
    });

    readonly defaultFilterValues: RuntimeEventFilters = {
        type: RUNTIME_FILTER_INITIAL_STATE.type,
        hasThreats: RUNTIME_FILTER_INITIAL_STATE.hasThreats,
        hasIncident: RUNTIME_FILTER_INITIAL_STATE.hasIncident,
        argument: RUNTIME_FILTER_INITIAL_STATE.argument,
        binary: RUNTIME_FILTER_INITIAL_STATE.binary,
        container: RUNTIME_FILTER_INITIAL_STATE.container,
        function: RUNTIME_FILTER_INITIAL_STATE.function,
        image: RUNTIME_FILTER_INITIAL_STATE.image,
        namespace: RUNTIME_FILTER_INITIAL_STATE.namespace,
        pod: RUNTIME_FILTER_INITIAL_STATE.pod,
        period: RUNTIME_FILTER_INITIAL_STATE.period,
        detectors: RUNTIME_FILTER_INITIAL_STATE.detectors,
        rules: RUNTIME_FILTER_INITIAL_STATE.rules
    };

    private readonly areToggleSelected$: Observable<boolean> = this.filterForm.valueChanges.pipe(
        map(() => utils.getFormValues<RuntimeEventFilters>(this.filterForm.controls)),
        distinctUntilChanged((a, b) => a.hasThreats === b.hasThreats && a.hasIncident === b.hasIncident),
        tap((values) => {
            if (!values.hasThreats) {
                this.filterForm.get('detectors')?.setValue([], { onlySelf: true });
            }
            if (!values.hasIncident) {
                this.filterForm.get('rules')?.setValue([], { onlySelf: true });
            }
        }),
        map((values) => values.hasThreats && values.hasIncident)
    );

    private readonly isRulesControlDisabled$: Observable<boolean> = this.ruleNodes$.pipe(
        map((rules) => !!rules?.length),
        tap((hasRules) => {
            if (hasRules) {
                this.filterForm.get('rules')?.enable({ onlySelf: true });
            } else {
                this.filterForm.get('rules')?.disable({ onlySelf: true });
            }
        })
    );

    hiddenDropdownMenuFilterKey: RuntimeEventFilterKey[] = [];

    selectedFilters: Partial<RuntimeEventFilters> = utils.getFormValues<RuntimeEventFilters>(this.filterForm.controls);

    readonly formControlKeys: RuntimeEventFilterKey[] = Object.values(RuntimeEventFilterKey);

    readonly runtimeEventFilterKey = RuntimeEventFilterKey;

    readonly runtimeEventTypeOptions = RUNTIME_EVENT_TYPE;

    readonly runtimeFilterPopoverControlLabelCollection = RUNTIME_FILTER_POPOVER_CONTROL_LABELS;

    readonly runtimeFilterPopoverControlPlaceholderCollection = RUNTIME_FILTER_POPOVER_CONTROL_PLACEHOLDERS;

    ngOnInit() {
        this.areToggleSelected$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
        this.isRulesControlDisabled$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    hideControl(key: RuntimeEventFilterKey) {
        this.hiddenDropdownMenuFilterKey.push(key);
        this.filterForm.patchValue({
            [key]: RUNTIME_FILTER_INITIAL_STATE[key]
        });
    }

    showControl(key: RuntimeEventFilterKey) {
        this.hiddenDropdownMenuFilterKey.splice(this.hiddenDropdownMenuFilterKey.indexOf(key), 1);
    }

    changeFilter(filters: RuntimeEventFilters) {
        this.filterForm.patchValue(filters);
        this.filterChange.emit(filters);
    }
}
