import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ChangeDetectionStrategy, Component, DestroyRef, EventEmitter, OnInit, Output, inject } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { Observable, debounceTime, map, tap } from 'rxjs';

import { FormScheme, CoreUtilsService as utils } from '@cs/core';
import { RULE_SEVERITIES, RULE_TYPE, RuleSeverity, RuleType } from '@cs/domains/rule';

import { RuleFilters, RuleFiltersPopover } from '../../interfaces/rule-filter.interface';

@Component({
    selector: 'cs-rule-feature-panel-filter-component',
    templateUrl: './rule-panel-filter.component.html',
    styleUrl: './rule-panel-filter.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuleFeaturePanelFilterComponent implements OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Output() filterChange = new EventEmitter<RuleFilters>();

    readonly queryForm: FormGroup<FormScheme<Pick<RuleFilters, 'name'>>> = this.formBuilder.group({
        name: ['']
    });

    readonly filterForm: FormGroup<FormScheme<RuleFiltersPopover>> = this.formBuilder.group({
        type: [[] as RuleType[]],
        notifySeverity: [[] as RuleSeverity[]]
    });

    readonly defaultFilterValues: RuleFiltersPopover = {
        type: [],
        notifySeverity: []
    };

    private readonly queryHandler$: Observable<string> = this.queryForm.valueChanges.pipe(
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(500),
        map((values) => values.name ?? ''),
        tap((name) => {
            this.filterChange.emit({
                ...utils.getFormValues<RuleFiltersPopover>(this.filterForm.controls),
                name: name.trim()
            });
        })
    );

    readonly ruleTypeOptions = RULE_TYPE;

    readonly ruleSeverityOptions = RULE_SEVERITIES;

    ngOnInit() {
        this.queryHandler$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    changeFilter(filters: RuleFiltersPopover) {
        this.filterForm.patchValue(filters);
        this.filterChange.emit({
            ...filters,
            name: utils.getFormValues<Pick<RuleFilters, 'name'>>(this.queryForm.controls).name
        });
    }
}
