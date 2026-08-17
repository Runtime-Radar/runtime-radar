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
    Output,
    inject
} from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Observable, debounceTime, distinctUntilChanged, distinctUntilKeyChanged, map, startWith, tap } from 'rxjs';

import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { ClusterFormType } from '../../interfaces/cluster-form-state.interface';
import { ClusterCreateFormOutputs, ClusterDataBaseForm } from '../../interfaces/cluster-form.interface';

@Component({
    selector: 'cs-cluster-feature-database-form-component',
    styleUrl: '../cluster-abstract-form.component.scss',
    templateUrl: './cluster-database-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureDataBaseFormComponent implements AfterViewInit, OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Input({ required: true }) type!: ClusterFormType;

    @Input() title?: string;

    @Input() values?: ClusterDataBaseForm | null;

    @Output() formChange = new EventEmitter<ClusterCreateFormOutputs<ClusterDataBaseForm>>();

    readonly form: FormGroup<FormScheme<ClusterDataBaseForm>> = this.formBuilder.group({
        isInternalCluster: [true],
        user: ['', Validators.required],
        password: ['', Validators.required],
        database: [''],
        isTls: [false],
        hasCheckCert: [false],
        ca: [''],
        isPersistence: [false],
        storageClass: [''],
        address: ['', [Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN)]]
    });

    readonly isInternalCluster$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isInternalCluster'),
        map(() => utils.getFormValues<ClusterDataBaseForm>(this.form.controls).isInternalCluster),
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

    readonly isTlsEnabled$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isTls'),
        map(() => utils.getFormValues<ClusterDataBaseForm>(this.form.controls).isTls),
        tap((isTls) => {
            if (!isTls) {
                this.form.get('hasCheckCert')?.setValue(false, { onlySelf: true });
                utils.toggleControlEnable(this.form.get('ca'), false);
            }
        })
    );

    readonly hasCheckCert$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        map(() => utils.getFormValues<ClusterDataBaseForm>(this.form.controls)),
        distinctUntilChanged(
            (a, b) => a.isInternalCluster === b.isInternalCluster && a.hasCheckCert === b.hasCheckCert
        ),
        map((form) => !form.isInternalCluster && form.hasCheckCert),
        tap((isCaEnable) => {
            utils.toggleControlEnable(this.form.get('ca'), isCaEnable);
        })
    );

    readonly isPersistence$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isPersistence'),
        map(() => utils.getFormValues<ClusterDataBaseForm>(this.form.controls).isPersistence),
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
            const formValues = utils.getFormValues<ClusterDataBaseForm>(this.form.controls);
            this.formChange.emit({
                form: utils.getTrimmedFormValues<ClusterDataBaseForm>(formValues),
                isValid
            });
        })
    );

    readonly tooltipPlacements = PopUpPlacements;

    isDatabaseControlEnable = false;

    ngOnInit() {
        this.isDatabaseControlEnable = this.type === 'clickhouse' || this.type === 'postgres';

        this.onFormChanges$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.isDatabaseControlEnable) {
            this.form.get('database')?.addValidators(Validators.required);
        }

        if (this.values) {
            this.form.patchValue({
                isInternalCluster: this.values.isInternalCluster,
                user: this.values.user,
                password: this.values.password,
                database: this.isDatabaseControlEnable ? this.values.database : '',
                isTls: this.values.isTls,
                hasCheckCert: this.values.isTls ? this.values.hasCheckCert : false,
                ca: !this.values.isInternalCluster && this.values.hasCheckCert ? this.values.ca : '',
                isPersistence: this.values.isInternalCluster ? this.values.isPersistence : false,
                storageClass: this.values.isInternalCluster ? this.values.storageClass : '',
                address: !this.values.isInternalCluster ? this.values.address : ''
            });
        }
    }
}
