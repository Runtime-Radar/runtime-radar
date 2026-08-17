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

import { IntegrationWebhook } from '@cs/domains/integration';
import {
    CoreValidators,
    FORM_VALIDATION_DENIED_IP,
    FORM_VALIDATION_REG_EXP,
    FormScheme,
    CoreUtilsService as utils
} from '@cs/core';

import { IntegrationWebhookForm } from '../../interfaces/integration-form.interface';

@Component({
    selector: 'cs-integration-feature-webhook-form-component',
    templateUrl: './integration-webhook-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureWebhookFormComponent implements AfterViewInit, OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Input() values?: IntegrationWebhook;

    @Output() formChange = new EventEmitter<IntegrationWebhookForm | undefined>();

    readonly form: FormGroup<FormScheme<IntegrationWebhookForm>> = this.formBuilder.group({
        name: ['', Validators.required],
        url: [
            '',
            [
                Validators.required,
                Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME),
                CoreValidators.isIpSegmentAllowed(FORM_VALIDATION_DENIED_IP.LOCALHOST)
            ]
        ],
        login: [''],
        password: [''],
        ca: [''],
        isInsecure: [false]
    });

    private readonly onFormValidChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => {
            const isValid = utils.isFormValid(this.form.controls);
            const formValues = utils.getFormValues<IntegrationWebhookForm>(this.form.controls);

            return {
                formValues,
                isValid:
                    isValid &&
                    ((!formValues.login && !formValues.password) || (!!formValues.login && !!formValues.password))
            };
        }),
        tap(({ isValid, formValues }) => {
            if (isValid) {
                this.formChange.emit(utils.getTrimmedFormValues<IntegrationWebhookForm>(formValues));
            } else {
                this.formChange.emit(undefined);
            }
        }),
        map(({ isValid }) => isValid)
    );

    ngOnInit() {
        this.onFormValidChanges$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.values) {
            this.form.patchValue({
                name: this.values.name,
                url: this.values.webhook.url,
                login: this.values.webhook.login,
                ca: this.values.webhook.ca,
                isInsecure: !this.values.webhook.insecure
            });
        }
    }
}
