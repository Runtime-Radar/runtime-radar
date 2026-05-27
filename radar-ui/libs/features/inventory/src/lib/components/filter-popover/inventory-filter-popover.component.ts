import { KbqBadgeColors } from '@koobiq/components/badge';
import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { PopUpSizes } from '@koobiq/components/core';
import { BehaviorSubject, Observable, map, startWith, switchMap, take } from 'rxjs';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { FORM_SEPARATOR_KEY_CODES, FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { InventoryFilters } from '../../interfaces/inventory-filter.interface';

const INVENTORY_FILTER_INITIAL_STATE: InventoryFilters = {
    nodes: [],
    namespaces: [],
    pods: [],
    containers: []
};

@Component({
    selector: 'cs-inventory-feature-filter-popover-component',
    templateUrl: './inventory-filter-popover.component.html',
    styleUrl: './inventory-filter-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class InventoryFeatureFilterPopoverComponent implements OnInit {
    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input() filters?: Partial<InventoryFilters> | null;

    @Output() filterChange = new EventEmitter<InventoryFilters>();

    readonly form: FormGroup<FormScheme<InventoryFilters>> = this.formBuilder.group({
        nodes: [[] as string[]],
        namespaces: [[] as string[]],
        pods: [[] as string[]],
        containers: [[] as string[]]
    });

    readonly filtersCounter$ = new BehaviorSubject(0);

    readonly selectedFilters$ = new BehaviorSubject<Partial<InventoryFilters>>(
        utils.getFormValues<InventoryFilters>(this.form.controls)
    );

    readonly hasChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        map(() => utils.getFormValues<InventoryFilters>(this.form.controls)),
        map((values) => !utils.isEqual(values, INVENTORY_FILTER_INITIAL_STATE))
    );

    readonly hasFiltersChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        map(() => utils.isFormValid(this.form.controls)),
        switchMap((isValid) => {
            const formValues = utils.getFormValues<InventoryFilters>(this.form.controls);
            return this.selectedFilters$.pipe(
                map((selectedFilters) => isValid && !utils.isEqual(formValues, selectedFilters)),
                take(1)
            );
        })
    );

    readonly isResetIconVisible$: Observable<boolean> = this.selectedFilters$.pipe(
        map((selectedFilters) => !utils.isEqual(INVENTORY_FILTER_INITIAL_STATE, selectedFilters))
    );

    readonly tooltipSizes = PopUpSizes;

    readonly badgeColors = KbqBadgeColors;

    readonly separatorKeyCodes = FORM_SEPARATOR_KEY_CODES;

    readonly validators = Validators;

    readonly formValidationRegExp = FORM_VALIDATION_REG_EXP;

    constructor(private readonly formBuilder: FormBuilder) {}

    ngOnInit() {
        if (this.filters) {
            this.form.patchValue(this.filters);
            this.selectedFilters$.next({ ...this.filters });
            this.filtersCounter$.next(Object.values(this.filters).filter((value) => value?.length).length);
        }
    }

    confirm() {
        const values = utils.getFormValues<InventoryFilters>(this.form.controls);
        const trimmedValues = utils.getTrimmedFormValues<InventoryFilters>(values);
        this.filtersCounter$.next(Object.values(values).filter((value) => value?.length).length);
        this.selectedFilters$.next(trimmedValues);
        this.filterChange.emit(trimmedValues);
        this.popover.hide();
    }

    reset() {
        this.form.patchValue(INVENTORY_FILTER_INITIAL_STATE);
        this.confirm();
    }

    cancel() {
        this.popover.hide();
    }
}
