import { AfterViewInit, ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, debounceTime, distinctUntilChanged, map, startWith } from 'rxjs';

import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';

import { RuntimeSettingPermissionForm } from '../../interfaces/runtime-form.interface';
import { RuntimeSidepanelPermissionFormProps } from '../../interfaces/runtime-sidepanel.interface';

@Component({
    templateUrl: './runtime-sidepanel-permission-form.component.html',
    styleUrl: './runtime-sidepanel-permission-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureSidepanelPermissionFormComponent implements AfterViewInit {
    private readonly formBuilder = inject(FormBuilder);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    readonly props = inject<Partial<RuntimeSidepanelPermissionFormProps>>(KBQ_SIDEPANEL_DATA);

    readonly form: FormGroup<FormScheme<RuntimeSettingPermissionForm>> = this.formBuilder.group({
        isAllowedType: [true],
        namespaces: [[] as string[]],
        pods: [[] as string[]],
        labels: [[] as string[]]
    });

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => {
            const values = utils.getFormValues<RuntimeSettingPermissionForm>(this.form.controls);

            return (
                utils.isFormValid(this.form.controls) &&
                !!(values.namespaces.length || values.pods.length || values.labels.length)
            );
        })
    );

    readonly formValidationRegExp = FORM_VALIDATION_REG_EXP;

    ngAfterViewInit() {
        if (this.props.isEdit) {
            this.form.patchValue({
                namespaces: this.props.namespaces || [],
                pods: this.props.pods || [],
                labels: this.props.labels || [],
                isAllowedType: this.props.isAllowedType
            });
        }
    }

    confirm() {
        if (this.form.valid) {
            const formValues = utils.getFormValues<RuntimeSettingPermissionForm>(this.form.controls);
            this.sidepanelRef.close(utils.getTrimmedFormValues<RuntimeSettingPermissionForm>(formValues));
        }
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }
}
