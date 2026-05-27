import { ChangeDetectionStrategy, Component, Input, booleanAttribute, model } from '@angular/core';
import {
    ControlValueAccessor,
    FormBuilder,
    FormControl,
    NG_VALIDATORS,
    NG_VALUE_ACCESSOR,
    ValidationErrors,
    ValidatorFn,
    Validators
} from '@angular/forms';
import { KbqTagEditChange, KbqTagInputEvent } from '@koobiq/components/tags';

import { FORM_SEPARATOR_KEY_CODES } from '@cs/core';

@Component({
    selector: 'cs-tag-input-component',
    templateUrl: './shared-tag-input.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    providers: [
        {
            provide: NG_VALUE_ACCESSOR,
            useExisting: SharedTagInputComponent,
            multi: true
        },
        {
            provide: NG_VALIDATORS,
            useExisting: SharedTagInputComponent,
            multi: true
        }
    ]
})
export class SharedTagInputComponent implements ControlValueAccessor {
    @Input({ transform: booleanAttribute }) editable = false;

    @Input() validators?: ValidatorFn[];

    @Input() placeholder?: string;

    @Input() hintText?: string;

    @Input() testLocator?: string;

    @Input() id?: string;

    isTouched = false;

    isDisabled = false;

    control = this.formBuilder.array<FormControl<string>>([]);

    readonly editInputModel = model<string>('');

    readonly separatorKeyCodes = FORM_SEPARATOR_KEY_CODES;

    constructor(private readonly formBuilder: FormBuilder) {}

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onChange = (tags: string[]) => {};

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

    validate(): ValidationErrors | null {
        if (this.control.invalid) {
            return { invalid: true };
        }

        if (this.control.hasValidator(Validators.required) && !this.control.length) {
            return { required: true };
        }

        return null;
    }

    setDisabledState(isDisabled: boolean) {
        this.isDisabled = isDisabled;
    }

    writeValue(tags: string[] | null) {
        if (tags?.length) {
            tags.forEach((tag) => {
                this.control.push(
                    this.formBuilder.nonNullable.control(tag, this.validators?.length ? this.validators : null)
                );
            });
        }
    }

    addTag(event: KbqTagInputEvent) {
        const value = event.value.trim();
        if (value) {
            this.control.push(
                this.formBuilder.nonNullable.control(value, this.validators?.length ? this.validators : null)
            );
            event.input.value = '';
            this.onChange(this.control.value);
        }

        this.markAsTouched();
    }

    removeTag(index: number) {
        this.control.removeAt(index);
        this.onChange(this.control.value);
        this.markAsTouched();
    }

    editTag({ type, tag }: KbqTagEditChange, index: number) {
        switch (type) {
            case 'start': {
                this.editInputModel.set(tag.value);
                break;
            }
            case 'cancel': {
                this.editInputModel.set('');
                break;
            }
            case 'submit': {
                if (!this.editInputModel()) {
                    tag.remove();
                } else {
                    this.control.controls[index].patchValue(this.editInputModel());
                    this.editInputModel.set('');
                }
                this.onChange(this.control.value);
                this.markAsTouched();
                break;
            }
        }
    }
}
