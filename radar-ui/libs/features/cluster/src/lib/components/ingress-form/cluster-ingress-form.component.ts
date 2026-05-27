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

import { ClusterCreateFormOutputs, ClusterIngressForm } from '../../interfaces/cluster-form.interface';

@Component({
    selector: 'cs-cluster-feature-ingress-form-component',
    styleUrl: '../cluster-abstract-form.component.scss',
    templateUrl: './cluster-ingress-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ClusterFeatureIngressFormComponent implements AfterViewInit, OnInit {
    @Input() values?: ClusterIngressForm | null;

    @Output() formChange = new EventEmitter<ClusterCreateFormOutputs<ClusterIngressForm>>();

    readonly form: FormGroup<FormScheme<ClusterIngressForm>> = this.formBuilder.group({
        isIngressEnabled: [true],
        ingressClass: [''],
        hostname: ['', [Validators.required, Validators.pattern(FORM_VALIDATION_REG_EXP.DOMAIN)]],
        cert: [''],
        certKey: [''],
        isNodePortEnabled: [true],
        port: ['', [Validators.min(30000), Validators.max(32767), Validators.pattern(FORM_VALIDATION_REG_EXP.NUMBER)]]
    });

    readonly isIngressEnabled$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isIngressEnabled'),
        map(() => utils.getFormValues<ClusterIngressForm>(this.form.controls).isIngressEnabled),
        tap((isIngressEnabled) => {
            utils.toggleControlEnable(this.form.get('hostname'), isIngressEnabled);
            if (!isIngressEnabled) {
                this.form.get('cert')?.setValue('', { onlySelf: true });
                this.form.get('certKey')?.setValue('', { onlySelf: true });
            }
        })
    );

    readonly isNodePortEnabled$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('isNodePortEnabled'),
        map(() => utils.getFormValues<ClusterIngressForm>(this.form.controls).isNodePortEnabled),
        tap((isNodePortEnabled) => {
            utils.toggleControlEnable(this.form.get('port'), isNodePortEnabled);
        })
    );

    private readonly onFormChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(500),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls)),
        tap((isValid) => {
            const formValues = utils.getFormValues<ClusterIngressForm>(this.form.controls);
            this.formChange.emit({
                form: utils.getTrimmedFormValues<ClusterIngressForm>(formValues),
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
                isIngressEnabled: this.values.isIngressEnabled,
                ingressClass: this.values.ingressClass,
                hostname: this.values.hostname,
                cert: this.values.cert,
                certKey: this.values.certKey,
                isNodePortEnabled: this.values.isNodePortEnabled,
                port: this.values.port
            });
        }
    }
}
