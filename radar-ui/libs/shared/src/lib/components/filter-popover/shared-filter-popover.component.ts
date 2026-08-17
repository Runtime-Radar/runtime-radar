import { KbqBadgeColors } from '@koobiq/components/badge';
import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import {
    AfterViewInit,
    ChangeDetectionStrategy,
    Component,
    DestroyRef,
    EventEmitter,
    Input,
    OnChanges,
    OnInit,
    Output,
    ViewChild,
    inject
} from '@angular/core';
import { BehaviorSubject, Subject } from 'rxjs';
import { FormControl, FormGroup, FormGroupDirective } from '@angular/forms';
import { PopUpPlacements, PopUpSizes } from '@koobiq/components/core';

import { AbstractFilters, FormScheme, CoreUtilsService as utils } from '@cs/core';

type AbstractFilterFormControls<C> = {
    [K in keyof C]: FormControl<C[K] | null>;
};

@Component({
    selector: 'cs-filter-popover-component',
    templateUrl: './shared-filter-popover.component.html',
    styleUrl: './shared-filter-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedFilterPopoverComponent<T extends AbstractFilters> implements OnChanges, OnInit, AfterViewInit {
    private readonly destroyRef = inject(DestroyRef);

    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input() id?: string;

    @Input({ required: true }) form!: FormGroup<FormScheme<T>>;

    @Input({ required: true }) formRef!: FormGroupDirective;

    @Input({ required: true }) defaultValues!: T;

    @Input() selected?: Partial<T>;

    @Output() formChange = new EventEmitter<T>();

    readonly filtersCounter$ = new Subject();

    readonly hasChanges$ = new BehaviorSubject(false);

    readonly hasFiltersChanges$ = new BehaviorSubject(false);

    readonly isResetIconVisible$ = new BehaviorSubject(false);

    readonly tooltipSizes = PopUpSizes;

    readonly tooltipPlacements = PopUpPlacements;

    readonly badgeColors = KbqBadgeColors;

    private selectedFilters: T | null = null;

    testLocator = this.id ? `${this.id}-` : '';

    ngOnChanges() {
        this.applySelectedFilters();
    }

    ngOnInit() {
        this.form.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
            const controls = this.form.controls as AbstractFilterFormControls<T>;
            const values = utils.getTrimmedFormValues<T>(utils.getFormValues<T>(controls));
            const isValid = utils.isFormValid(controls);

            this.hasFiltersChanges$.next(isValid && !utils.isEqual(values, this.selectedFilters));
            this.hasChanges$.next(!utils.isEqual(values, this.defaultValues));
        });
    }

    ngAfterViewInit() {
        this.applySelectedFilters();
    }

    confirm() {
        const values = utils.getFormValues<T>(this.form.controls as AbstractFilterFormControls<T>);
        this.formChange.emit(utils.getTrimmedFormValues<T>(values));
        this.applySelectedFilters(values);
        this.popover.hide();
    }

    reset() {
        this.formRef.form.patchValue(this.defaultValues);
        this.formChange.emit(this.defaultValues);
        this.applySelectedFilters();
        this.popover.hide();
    }

    cancel() {
        this.popover.hide();
    }

    private applySelectedFilters(filters?: T) {
        const values = filters || utils.getFormValues<T>(this.form.controls as AbstractFilterFormControls<T>);
        const controls = Object.values(values).filter(
            (value) => (Array.isArray(value) && !!value.length) || (!Array.isArray(value) && !!value)
        );

        this.selectedFilters = utils.getTrimmedFormValues<T>(values);
        this.filtersCounter$.next(controls.length);
        this.hasFiltersChanges$.next(!utils.isEqual(values, this.selectedFilters));
        this.isResetIconVisible$.next(!utils.isEqual(this.defaultValues, this.selectedFilters));
    }
}
