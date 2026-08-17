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

import { IntegrationSyslog } from '@cs/domains/integration';
import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { IntegrationSyslogForm } from '../../interfaces/integration-form.interface';

@Component({
    selector: 'cs-integration-feature-syslog-form-component',
    templateUrl: './integration-syslog-form.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureSyslogFormComponent implements AfterViewInit, OnInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);

    @Input() values?: IntegrationSyslog;

    @Output() formChange = new EventEmitter<IntegrationSyslogForm | undefined>();

    readonly form: FormGroup<FormScheme<IntegrationSyslogForm>> = this.formBuilder.group({
        name: ['', Validators.required],
        address: ['', [Validators.required, Validators.pattern(FORM_VALIDATION_REG_EXP.TCP_UPD_DOMAIN_SCHEME)]]
    });

    private readonly onFormValidChanges$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls)),
        tap((isValid) => {
            if (isValid) {
                const formValues = utils.getFormValues<IntegrationSyslogForm>(this.form.controls);
                this.formChange.emit(utils.getTrimmedFormValues<IntegrationSyslogForm>(formValues));
            } else {
                this.formChange.emit(undefined);
            }
        })
    );

    ngOnInit() {
        this.onFormValidChanges$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.values) {
            this.form.patchValue({
                name: this.values.name,
                address: this.values.syslog.address
            });
        }
    }
}
