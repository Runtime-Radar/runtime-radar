import { KbqModalRef } from '@koobiq/components/modal';
import { PasswordRules } from '@koobiq/components/form-field';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Observable, debounceTime, distinctUntilChanged, map, startWith } from 'rxjs';

import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

interface PasswordModalForm {
    password: string;
}

@Component({
    templateUrl: './shared-password-modal.component.html',
    styles: [
        `
            .shared-password-modal-form {
                padding-top: 1px;
            }
        `
    ],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedPasswordModalComponent {
    private readonly formBuilder = inject(FormBuilder);
    private readonly modal = inject(KbqModalRef);

    readonly form: FormGroup<FormScheme<PasswordModalForm>> = this.formBuilder.group({
        password: [
            '',
            [
                Validators.required,
                Validators.minLength(8),
                Validators.maxLength(16),
                Validators.pattern(FORM_VALIDATION_REG_EXP.PASSWORD)
            ]
        ]
    });

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls))
    );

    readonly passwordRules = PasswordRules;

    dispatch(isSuccessful: boolean) {
        this.modal.destroy(isSuccessful ? utils.getTrimmedFormValues(this.form.value).password : undefined);
    }
}
