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

import { ClusterCreateFormOutputs, ClusterRabbitForm } from '../../interfaces/cluster-form.interface';

@Component({
    selector: 'cs-cluster-feature-rabbit-form-component',
    styleUrl: '../cluster-abstract-form.component.scss',
    templateUrl: './cluster-rabbit-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ClusterFeatureRabbitFormComponent implements AfterViewInit, OnInit {
    @Input() values?: ClusterRabbitForm | null;

    @Output() formChange = new EventEmitter<ClusterCreateFormOutputs<ClusterRabbitForm>>();

    readonly form: FormGroup<FormScheme<ClusterRabbitForm>> = this.formBuilder.group({
        isInternalCluster: [true],
        user: ['', Validators.required],
        password: ['', Validators.required],
        isPersistence: [false],
        storageClass: [''],
        address: ['', [Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN)]]
    });

    readonly isInternalCluster$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isInternalCluster'),
        map(() => utils.getFormValues<ClusterRabbitForm>(this.form.controls).isInternalCluster),
        tap((isInternal) => {
            utils.toggleControlEnable(this.form.get('address'), !isInternal);
            if (!isInternal) {
                this.form.get('isPersistence')?.setValue(false, { onlySelf: true });

                const control = this.form.get('storageClass');
                control?.setValue('', { onlySelf: true });
                control?.disable({ onlySelf: true });
            }
        })
    );

    readonly isPersistence$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isPersistence'),
        map(() => utils.getFormValues<ClusterRabbitForm>(this.form.controls).isPersistence),
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
            const formValues = utils.getFormValues<ClusterRabbitForm>(this.form.controls);
            this.formChange.emit({
                form: utils.getTrimmedFormValues<ClusterRabbitForm>(formValues),
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
                isInternalCluster: this.values.isInternalCluster,
                user: this.values.user,
                password: this.values.password,
                isPersistence: this.values.isInternalCluster ? this.values.isPersistence : false,
                storageClass: this.values.isInternalCluster ? this.values.storageClass : '',
                address: !this.values.isInternalCluster ? this.values.address : ''
            });
        }
    }
}
