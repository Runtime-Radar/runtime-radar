import { PasswordRules } from '@koobiq/components/form-field';
import { AfterViewInit, ChangeDetectionStrategy, Component, OnInit, inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, debounceTime, distinctUntilChanged, map, take } from 'rxjs';

import { DEFAULT_ADMIN_ROLE_ID } from '@cs/domains/role';
import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { UserForm } from '../../interfaces/user-form.interface';
import { UserSidepanelFormProps } from '../../interfaces/user-sidepanel.interface';

const USER_FORM_USERNAME_VALIDATION_REG_EXP = /^[\w.-]+$/;

@Component({
    templateUrl: './user-sidepanel-user-form.component.html',
    styleUrl: './user-sidepanel-user-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class UserFeatureSidepanelUserFormComponent implements AfterViewInit, OnInit {
    private readonly formBuilder = inject(FormBuilder);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    readonly props = inject<UserSidepanelFormProps>(KBQ_SIDEPANEL_DATA);

    // @todo: replace observable to separate permission
    readonly isAdminRole$: Observable<boolean> = this.props.credentials$.pipe(
        take(1),
        map((credentials) => credentials.roleId === DEFAULT_ADMIN_ROLE_ID)
    );

    readonly form: FormGroup<FormScheme<UserForm>> = this.formBuilder.group({
        username: ['', [Validators.pattern(USER_FORM_USERNAME_VALIDATION_REG_EXP)]],
        email: ['', Validators.email],
        password: [
            '',
            [Validators.minLength(8), Validators.maxLength(16), Validators.pattern(FORM_VALIDATION_REG_EXP.PASSWORD)]
        ],
        roleId: ['', Validators.required]
    });

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls))
    );

    readonly passwordRules = PasswordRules;

    ngOnInit() {
        if (!this.props.isEdit) {
            this.form.get('username')?.addValidators(Validators.required);
            this.form.get('password')?.addValidators(Validators.required);
        }
    }

    ngAfterViewInit() {
        this.form.patchValue({
            email: this.props.user?.email || '',
            roleId: this.props.user?.role_id || ''
        });
    }

    confirm() {
        const formValues = utils.getFormValues<UserForm>(this.form.controls);
        this.sidepanelRef.close(utils.getTrimmedFormValues<UserForm>(formValues));
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }
}
