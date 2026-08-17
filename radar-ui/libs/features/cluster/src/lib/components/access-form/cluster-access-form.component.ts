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

import { ClusterStoreService } from '@cs/domains/cluster';
import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { ClusterFeatureUrlValidator } from '../../validators/cluster-url.validator';
import { ClusterAccessForm, ClusterCreateFormOutputs } from '../../interfaces/cluster-form.interface';

@Component({
    selector: 'cs-cluster-feature-access-form-component',
    styleUrl: '../cluster-abstract-form.component.scss',
    templateUrl: './cluster-access-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureAccessFormComponent implements AfterViewInit, OnInit {
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Input() values?: ClusterAccessForm | null;

    @Input() centralUrl?: string | null;

    @Output() formChange = new EventEmitter<ClusterCreateFormOutputs<ClusterAccessForm>>();

    readonly form: FormGroup<FormScheme<ClusterAccessForm>> = this.formBuilder.group({
        proxyUrl: ['', [Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME)]],
        ownCsUrl: [
            '',
            [Validators.required, Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME)],
            [ClusterFeatureUrlValidator.isUrlUnique(this.clusterStoreService)]
        ], // the value is stores into apiPathService.host
        centralCsUrl: ['', [Validators.required, Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME)]],
        namespace: ['', Validators.required],
        name: ['', Validators.required]
    });

    private readonly onFormChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(500),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls)),
        tap((isValid) => {
            const formValues = utils.getFormValues<ClusterAccessForm>(this.form.controls);
            this.formChange.emit({
                form: utils.getTrimmedFormValues<ClusterAccessForm>(formValues),
                isValid
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
                proxyUrl: this.values.proxyUrl,
                ownCsUrl: this.values.ownCsUrl,
                centralCsUrl: this.values.centralCsUrl || this.centralUrl,
                namespace: this.values.namespace,
                name: this.values.name
            });
        }
    }
}
