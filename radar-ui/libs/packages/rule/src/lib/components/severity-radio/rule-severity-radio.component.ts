import {
    AbstractControl,
    ControlValueAccessor,
    NG_VALIDATORS,
    NG_VALUE_ACCESSOR,
    ValidationErrors,
    Validators
} from '@angular/forms';
import { Component, Input } from '@angular/core';

import { RULE_SEVERITIES, RuleSeverity } from '@cs/domains/rule';

@Component({
    selector: 'cs-rule-package-severity-radio-component',
    templateUrl: './rule-severity-radio.component.html',
    styleUrls: ['./rule-severity-radio.component.scss', '../rule-abstract-radio.component.scss'],
    standalone: false,
    providers: [
        {
            provide: NG_VALUE_ACCESSOR,
            useExisting: RulePackageSeverityRadioComponent,
            multi: true
        },
        {
            provide: NG_VALIDATORS,
            useExisting: RulePackageSeverityRadioComponent,
            multi: true
        }
    ]
})
export class RulePackageSeverityRadioComponent implements ControlValueAccessor {
    @Input() id?: string;

    @Input() testLocator?: string;

    @Input() noneLabelLocalizationKey?: string;

    isTouched = false;

    isDisabled = false;

    severity = RuleSeverity.NONE;

    readonly ruleSeverityOptions = RULE_SEVERITIES;

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onChange = (severity: RuleSeverity) => {};

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

    writeValue(severity?: RuleSeverity | null) {
        if (severity) {
            this.severity = severity;
        }
    }

    validate(control: AbstractControl): ValidationErrors | null {
        return control.hasValidator(Validators.required) && !control.value ? { required: true } : null;
    }

    changeSeverity(severity: RuleSeverity) {
        if (!this.isDisabled) {
            this.severity = severity;
            this.onChange(this.severity);
        }

        this.markAsTouched();
    }
}
