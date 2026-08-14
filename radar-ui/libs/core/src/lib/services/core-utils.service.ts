import { Injectable } from '@angular/core';
import { AbstractControl, FormArray, FormControl, FormGroup, FormRecord, Validators } from '@angular/forms';

interface FormControls {
    [key: string]: FormControl | FormRecord | FormArray;
}

@Injectable({
    providedIn: 'root'
})
export class CoreUtilsService {
    static generateUuid(prefix = ''): string {
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        const template = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (slice) => {
            /* eslint no-bitwise: "off" */
            const part = (Math.random() * 16) | 0;
            const value = slice == 'x' ? part : (part & 0x3) | 0x8;

            return value.toString(16);
        });

        return `${prefix}${template}`;
    }

    static isDefined<T>(
        value: T | string | null | undefined
    ): value is T extends string | null | undefined ? never : T {
        return value !== null && value !== undefined && value !== '';
    }

    static isFormValid(controls: FormControls): boolean {
        return Object.keys(controls).every((key) => {
            const control = controls[key];
            const childControls = (control as FormRecord | FormArray).controls;

            if (childControls && !Array.isArray(childControls) && Object.keys(childControls).length) {
                return Object.keys(childControls).every((id) => {
                    const recordControls = (childControls[id] as FormGroup).controls as FormControls;

                    return this.isFormValid(recordControls);
                });
            }

            let isChildValid = true;
            if (childControls && Array.isArray(childControls)) {
                isChildValid = childControls.every((item) => this.isFormControlValid(item));
            }

            if (!isChildValid) {
                return false;
            }

            return this.isFormControlValid(control);
        });
    }

    static isFormControlValid(control: AbstractControl | null): boolean {
        if (!control || control.status === 'DISABLED') {
            return true;
        }

        const errors = control.validator ? control.validator(control) : null;
        if (errors !== null) {
            return !Object.keys(errors).length;
        }

        const asyncErrors = control.asyncValidator ? control.errors : null;
        if (asyncErrors !== null) {
            return !Object.keys(asyncErrors).length;
        }

        return true;
    }

    static getFormValues<T>(controls: FormControls): T {
        return Object.keys(controls).reduce((acc, key) => {
            acc[key as keyof T] = controls[key]?.value;

            return acc;
        }, {} as T);
    }

    // @todo: refactor the method to support native types
    static getTrimmedFormValues<T extends object>(values: T): T {
        return Object.keys(values).reduce((acc, key) => {
            let value = values[key as keyof T];

            if (typeof value === 'string') {
                (value as string) = (value as string).trim();
            } else if (value === null) {
                (value as null) = null;
            } else if (Object.getPrototypeOf(value) === Object.prototype && !!Object.keys(value as object).length) {
                (value as object) = this.getTrimmedFormValues(value as object);
            } else if (Array.isArray(value) && !!(value as unknown[]).length) {
                (value as unknown[]) = (value as unknown[]).map((item) =>
                    typeof item === 'string' ? item.trim() : item
                );
            }

            return {
                ...acc,
                [key]: value
            };
        }, {} as T);
    }

    /**
     * Enable/disable provided control and turn on/off validation.
     *
     * @param control - It is a form's link from parent entity.
     */
    static toggleControlEnable(control: AbstractControl | null, isEnable: boolean, value: unknown = '') {
        if (isEnable) {
            control?.enable({ onlySelf: true });
            control?.addValidators(Validators.required);
        } else {
            control?.setValue(value, { onlySelf: true });
            control?.disable({ onlySelf: true });
            control?.removeValidators(Validators.required);
        }
    }

    /**
     * Excludes the specified properties from the object.
     *
     * @param obj - Original object.
     * @param keys - Keys to select.
     */
    static omit<T extends object, K extends keyof T>(obj: T, keys: K | K[]): Omit<T, K> {
        const excludedKeys = Array.isArray(keys) ? keys : [keys];
        const result = { ...obj };

        excludedKeys.forEach((key) => {
            delete result[key];
        });

        return result;
    }

    /**
     * Compares for deep equality
     *
     * @param a - First value.
     * @param b - Second value.
     */
    static isEqual<T>(a: T, b: T): boolean {
        if (a === b) {
            return true;
        }

        if (a == null || b == null) {
            return false;
        }

        if (typeof a !== 'object' || typeof b !== 'object') {
            return false;
        }

        if (Array.isArray(a) && Array.isArray(b)) {
            if (a.length !== b.length) {
                return false;
            }

            return a.every((item, index) => CoreUtilsService.isEqual(item, b[index]));
        }

        const keysA = Object.keys(a) as (keyof T)[];
        const keysB = Object.keys(b) as (keyof T)[];

        if (keysA.length !== keysB.length) {
            return false;
        }

        return keysA.every((key) => {
            const valueA = a[key];
            const valueB = b[key];

            if (typeof valueA === 'object' && typeof valueB === 'object') {
                return CoreUtilsService.isEqual(valueA, valueB);
            }

            return valueA === valueB;
        });
    }
}
