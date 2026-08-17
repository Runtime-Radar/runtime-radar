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

import {
    CoreValidators,
    FORM_VALIDATION_DENIED_IP,
    FORM_VALIDATION_REG_EXP,
    FormScheme,
    CoreUtilsService as utils
} from '@cs/core';
import { INTEGRATION_EMAIL_AUTH_TYPE, IntegrationEmail, IntegrationEmailAuthType } from '@cs/domains/integration';

import { IntegrationEmailForm, IntegrationProtocolType } from '../../interfaces/integration-form.interface';

@Component({
    selector: 'cs-integration-feature-email-form-component',
    templateUrl: './integration-email-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureEmailFormComponent implements AfterViewInit, OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Input() values?: IntegrationEmail;

    @Output() formChange = new EventEmitter<IntegrationEmailForm | undefined>();

    readonly form: FormGroup<FormScheme<IntegrationEmailForm>> = this.formBuilder.group({
        name: ['', Validators.required],
        server: [
            '',
            [
                Validators.required,
                Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN),
                CoreValidators.isIpSegmentAllowed(FORM_VALIDATION_DENIED_IP.LOCALHOST)
            ]
        ],
        username: ['', Validators.required],
        password: ['', Validators.required],
        ca: [''],
        from: ['', [Validators.required, Validators.email]],
        authType: [IntegrationEmailAuthType.PLAIN],
        protocol: [IntegrationProtocolType.NONE],
        isInsecure: [false]
    });

    private readonly onFormValidChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls)),
        tap((isValid) => {
            if (isValid) {
                const formValues = utils.getFormValues<IntegrationEmailForm>(this.form.controls);
                this.formChange.emit(utils.getTrimmedFormValues<IntegrationEmailForm>(formValues));
            } else {
                this.formChange.emit(undefined);
            }
        })
    );

    private readonly isInsecureControlEnable$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('protocol'),
        map((values) => values.protocol !== IntegrationProtocolType.NONE),
        tap((hasProtocol) => {
            const control = this.form.get('isInsecure');
            if (hasProtocol) {
                control?.enable({ onlySelf: true });
            } else {
                control?.disable({ onlySelf: true });
                control?.setValue(false, { onlySelf: true });
            }
        })
    );

    private readonly isAuthTypeEnable$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('authType'),
        map((values) => values.authType !== IntegrationEmailAuthType.NONE),
        distinctUntilChanged(),
        tap((isNoneType) => {
            utils.toggleControlEnable(this.form.get('username'), isNoneType);
            utils.toggleControlEnable(this.form.get('password'), isNoneType);
        })
    );

    readonly integrationEmailAuthTypeOptions = INTEGRATION_EMAIL_AUTH_TYPE;

    ngOnInit() {
        this.onFormValidChanges$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
        this.isAuthTypeEnable$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
        this.isInsecureControlEnable$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.values) {
            let protocol = IntegrationProtocolType.NONE;
            if (this.values.email.use_tls) {
                protocol = IntegrationProtocolType.TLS;
            } else if (this.values.email.use_start_tls) {
                protocol = IntegrationProtocolType.START_TLS;
            }

            this.form.patchValue({
                name: this.values.name,
                server: this.values.email.server,
                username: this.values.email.username,
                ca: this.values.email.ca,
                from: this.values.email.from,
                authType: this.values.email.auth_type,
                isInsecure: !this.values.email.insecure,
                protocol
            });
        }
    }
}
