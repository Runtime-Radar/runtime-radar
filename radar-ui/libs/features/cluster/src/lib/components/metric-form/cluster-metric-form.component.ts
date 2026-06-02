import { PopUpPlacements } from '@koobiq/components/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import {
    AfterViewInit,
    ChangeDetectionStrategy,
    Component,
    DestroyRef,
    EventEmitter,
    Input,
    OnInit,
    Output
} from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Observable, debounceTime, distinctUntilChanged, distinctUntilKeyChanged, map, startWith, tap } from 'rxjs';

import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { ClusterCreateFormOutputs, ClusterMetricForm } from '../../interfaces/cluster-form.interface';

@Component({
    selector: 'cs-cluster-feature-metric-form-component',
    styleUrl: '../cluster-abstract-form.component.scss',
    templateUrl: './cluster-metric-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ClusterFeatureMetricFormComponent implements AfterViewInit, OnInit {
    @Input() values?: ClusterMetricForm | null;

    @Output() formChange = new EventEmitter<ClusterCreateFormOutputs<ClusterMetricForm>>();

    readonly form: FormGroup<FormScheme<ClusterMetricForm>> = this.formBuilder.group({
        isMetricEnabled: [false],
        isInternalCluster: [true],
        user: ['', Validators.required],
        password: ['', Validators.required],
        isPersistence: [true],
        storageClass: [''],
        address: ['']
    });

    readonly isMetricEnabled$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isMetricEnabled'),
        map(() => utils.getFormValues<ClusterMetricForm>(this.form.controls).isMetricEnabled),
        tap((isIngressEnabled) => {
            utils.toggleControlEnable(this.form.get('user'), isIngressEnabled);
            utils.toggleControlEnable(this.form.get('password'), isIngressEnabled);
            utils.toggleControlEnable(this.form.get('address'), isIngressEnabled);
        })
    );

    readonly isInternalCluster$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isInternalCluster'),
        map(() => utils.getFormValues<ClusterMetricForm>(this.form.controls).isInternalCluster),
        tap((isInternal) => {
            utils.toggleControlEnable(this.form.get('user'), isInternal);
            utils.toggleControlEnable(this.form.get('password'), isInternal);
            utils.toggleControlEnable(this.form.get('address'), !isInternal);
            if (!isInternal) {
                this.form.get('isPersistence')?.setValue(false, { onlySelf: true });
                this.form.get('address')?.addValidators(Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME));

                const control = this.form.get('storageClass');
                control?.setValue('', { onlySelf: true });
                control?.disable({ onlySelf: true });
            }
        })
    );

    readonly isPersistence$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isPersistence'),
        map(() => utils.getFormValues<ClusterMetricForm>(this.form.controls).isPersistence),
        tap((isPersistence) => {
            const control = this.form.get('storageClass');
            if (isPersistence) {
                control?.enable({ onlySelf: true });
            } else {
                control?.disable({ onlySelf: true });
                control?.setValue('', { onlySelf: true });
            }
        })
    );

    private readonly onFormChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(500),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls)),
        tap((isValid) => {
            const formValues = utils.getFormValues<ClusterMetricForm>(this.form.controls);
            this.formChange.emit({
                form: utils.getTrimmedFormValues<ClusterMetricForm>(formValues),
                isValid
            });
        })
    );

    readonly tooltipPlacements = PopUpPlacements;

    constructor(
        private readonly destroyRef: DestroyRef,
        private readonly formBuilder: FormBuilder
    ) {}

    ngOnInit() {
        this.onFormChanges$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.values) {
            this.form.patchValue({
                isMetricEnabled: this.values.isMetricEnabled,
                isInternalCluster: this.values.isInternalCluster,
                user: this.values.user,
                password: this.values.password,
                isPersistence: this.values.isPersistence,
                storageClass: this.values.storageClass,
                address: this.values.address
            });
        }
    }
}
