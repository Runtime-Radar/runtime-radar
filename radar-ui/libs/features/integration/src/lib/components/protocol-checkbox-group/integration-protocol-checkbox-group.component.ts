import {
    AbstractControl,
    ControlValueAccessor,
    NG_VALIDATORS,
    NG_VALUE_ACCESSOR,
    ValidationErrors,
    Validators
} from '@angular/forms';
import { ChangeDetectionStrategy, Component } from '@angular/core';

import { IntegrationProtocolType } from '../../interfaces/integration-form.interface';

@Component({
    selector: 'cs-integration-feature-protocol-checkbox-group-component',
    templateUrl: './integration-protocol-checkbox-group.component.html',
    styleUrl: './integration-protocol-checkbox-group.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false,
    providers: [
        {
            provide: NG_VALUE_ACCESSOR,
            useExisting: IntegrationFeatureProtocolCheckboxGroupComponent,
            multi: true
        },
        {
            provide: NG_VALIDATORS,
            useExisting: IntegrationFeatureProtocolCheckboxGroupComponent,
            multi: true
        }
    ]
})
export class IntegrationFeatureProtocolCheckboxGroupComponent implements ControlValueAccessor {
    protocol = IntegrationProtocolType.NONE;

    isTouched = false;

    isDisabled = false;

    isStartTls = false;

    isTls = false;

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onChange = (protocol: IntegrationProtocolType) => {};

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onTouched = () => {};

    registerOnChange(fn: any) {
        this.onChange = fn;
    }

    registerOnTouched(fn: any) {
        this.onTouched = fn;
    }

    markAsTouched() {
        if (!this.isTouched) {
            this.isTouched = true;
            this.onTouched();
        }
    }

    setDisabledState(isDisabled: boolean) {
        this.isDisabled = isDisabled;
    }

    writeValue(protocol: IntegrationProtocolType | null) {
        if (protocol) {
            this.protocol = protocol;
            this.isStartTls = protocol === IntegrationProtocolType.START_TLS;
            this.isTls = protocol === IntegrationProtocolType.TLS;
        }
    }

    validate(control: AbstractControl): ValidationErrors | null {
        return control.hasValidator(Validators.required) && control.value === IntegrationProtocolType.NONE
            ? { required: true }
            : null;
    }

    changeStartTls() {
        if (!this.isDisabled) {
            this.isStartTls = !this.isStartTls;
            this.isTls = false;
            this.onChange(this.isStartTls ? IntegrationProtocolType.START_TLS : IntegrationProtocolType.NONE);
        }

        this.markAsTouched();
    }

    changeTls() {
        if (!this.isDisabled) {
            this.isStartTls = false;
            this.isTls = !this.isTls;
            this.onChange(this.isTls ? IntegrationProtocolType.TLS : IntegrationProtocolType.NONE);
        }

        this.markAsTouched();
    }
}
