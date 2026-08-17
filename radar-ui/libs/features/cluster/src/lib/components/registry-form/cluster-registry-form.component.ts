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
import { Observable, debounceTime, distinctUntilChanged, map, startWith, tap } from 'rxjs';

import { FormScheme, CoreUtilsService as utils } from '@cs/core';

import { ClusterCreateFormOutputs, ClusterRegistryForm } from '../../interfaces/cluster-form.interface';

const CLUSTER_ADDRESS_VALIDATION_REG_EXP =
    /^((([a-zA-Z0-9]+([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]+)|(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})|([a-zA-Z0-9-]+[a-zA-Z0-9]))(:([1-9]\d{0,4}|[1-5]\d{4}|6[0-4]\d{3}|65[0-4]\d{2}|655[0-2]\d|6553[0-5]))?(\/[a-zA-Z0-9-]+)?$/;

@Component({
    selector: 'cs-cluster-feature-registry-form-component',
    styleUrl: '../cluster-abstract-form.component.scss',
    templateUrl: './cluster-registry-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureRegistryFormComponent implements AfterViewInit, OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Input() values?: ClusterRegistryForm | null;

    @Output() formChange = new EventEmitter<ClusterCreateFormOutputs<ClusterRegistryForm>>();

    readonly form: FormGroup<FormScheme<ClusterRegistryForm>> = this.formBuilder.group({
        address: ['', [Validators.pattern(CLUSTER_ADDRESS_VALIDATION_REG_EXP)]],
        user: [''],
        password: [''],
        isImageShortName: [true]
    });

    private readonly onFormChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(500),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls)),
        tap((isValid) => {
            const formValues = utils.getFormValues<ClusterRegistryForm>(this.form.controls);
            this.formChange.emit({
                form: formValues,
                isValid:
                    isValid &&
                    ((!formValues.user && !formValues.password) || (!!formValues.user && !!formValues.password))
            });
        })
    );

    readonly tooltipPlacements = PopUpPlacements;

    ngOnInit() {
        this.onFormChanges$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.values) {
            this.form.patchValue({
                address: this.values.address,
                user: this.values.user,
                password: this.values.password,
                isImageShortName: this.values.isImageShortName
            });
        }
    }
}
