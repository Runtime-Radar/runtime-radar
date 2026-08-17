import { AfterViewInit, ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, debounceTime, distinctUntilChanged, map, startWith } from 'rxjs';

import { FormScheme, CoreUtilsService as utils } from '@cs/core';

import { RuntimeFeaturePolicyNameService } from '../../services/runtime-policy-name.service';
import { RuntimeFeaturePolicyNameValidator } from '../../validators/runtime-policy-name.validator';
import { RuntimeFeatureYamlValidator } from '../../validators/runtime-yaml.validator';
import { RuntimeSettingPolicyForm } from '../../interfaces/runtime-form.interface';
import { RuntimeSidepanelPolicyFormProps } from '../../interfaces/runtime-sidepanel.interface';

const DEFAULT_RUNTIME_POLICY_YAML = `apiVersion: cilium.io/v1alpha1
kind: TracingPolicy
metadata:
  name: "policy-name"`;

@Component({
    templateUrl: './runtime-sidepanel-policy-form.component.html',
    styleUrl: './runtime-sidepanel-policy-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureSidepanelPolicyFormComponent implements AfterViewInit {
    private readonly formBuilder = inject(FormBuilder);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    private readonly runtimeFeaturePolicyNameService = inject(RuntimeFeaturePolicyNameService);

    readonly props = inject<Partial<RuntimeSidepanelPolicyFormProps>>(KBQ_SIDEPANEL_DATA);

    readonly form: FormGroup<FormScheme<RuntimeSettingPolicyForm>> = this.formBuilder.group({
        isEnabled: [false],
        name: [
            '',
            [Validators.required],
            [RuntimeFeaturePolicyNameValidator.isNameUnique(this.runtimeFeaturePolicyNameService, this.props.name)]
        ],
        description: ['', Validators.required],
        yaml: [DEFAULT_RUNTIME_POLICY_YAML, [Validators.required, RuntimeFeatureYamlValidator.isYamlCodeValid()]]
    });

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls))
    );

    ngAfterViewInit() {
        if (this.props.isEdit) {
            this.form.patchValue({
                name: this.props.name,
                description: this.props.description,
                yaml: this.props.yaml
            });
        }
    }

    confirm() {
        if (this.form.valid) {
            const formValues = utils.getFormValues<RuntimeSettingPolicyForm>(this.form.controls);
            this.sidepanelRef.close(utils.getTrimmedFormValues<RuntimeSettingPolicyForm>(formValues));
        }
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }
}
