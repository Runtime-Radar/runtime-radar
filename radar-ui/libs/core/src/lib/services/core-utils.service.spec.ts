import { FormArray, FormControl, FormGroup, Validators } from '@angular/forms';

import { CoreUtilsService } from './core-utils.service';
import { FormScheme } from '../interfaces/core-form-scheme.interface';

interface MockForm {
    node: string;
    size: number;
    entities: string[];
    isLoaded: boolean;
}

describe('CoreUtilsService', () => {
    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('generateUuid', () => {
        it('should generate UUID with prefix', () => {
            jest.spyOn(Math, 'random').mockReturnValue(0.1);
            const uuid = CoreUtilsService.generateUuid('pre-');

            expect(uuid).toEqual('pre-11111111-1111-4111-9111-111111111111');
        });
    });

    describe('isDefined', () => {
        it('should return true for defined non-empty values', () => {
            expect(CoreUtilsService.isDefined('abc')).toBe(true);
        });

        it('should return false for null', () => {
            expect(CoreUtilsService.isDefined(null)).toBe(false);
        });

        it('should return false for undefined', () => {
            expect(CoreUtilsService.isDefined(undefined)).toBe(false);
        });

        it('should return false for empty string', () => {
            expect(CoreUtilsService.isDefined('')).toBe(false);
        });
    });

    describe('isFormControlValid', () => {
        it('should return true for valid control', () => {
            const control = new FormControl('value', Validators.required);

            expect(CoreUtilsService.isFormControlValid(control)).toBe(true);
        });

        it('should return false for invalid control', () => {
            const control = new FormControl('', Validators.required);

            expect(CoreUtilsService.isFormControlValid(control)).toBe(false);
        });

        it('should return true for disabled control', () => {
            const control = new FormControl('', Validators.required);
            control.disable();

            expect(CoreUtilsService.isFormControlValid(control)).toBe(true);
        });
    });

    describe('isFormValid', () => {
        it('should return true for valid form', () => {
            const form = new FormGroup({
                name: new FormControl('Name', [Validators.required, Validators.minLength(2)]),
                count: new FormControl(30, [Validators.required, Validators.min(18)]),
                tags: new FormArray([
                    new FormControl('tag1', [Validators.required, Validators.minLength(3)]),
                    new FormControl('tag2', [Validators.required, Validators.minLength(3)])
                ])
            });

            expect(CoreUtilsService.isFormValid(form.controls)).toBe(true);
        });

        it('should return false for invalid FormArray', () => {
            const form = new FormGroup({
                list: new FormArray([new FormControl('', Validators.required)])
            });

            expect(CoreUtilsService.isFormValid(form.controls)).toBe(false);
        });
    });

    describe('getFormValues', () => {
        it('should return form values', () => {
            const controls: FormScheme<MockForm> = {
                node: new FormControl('node'),
                size: new FormControl(30),
                entities: new FormControl(['sub0, sub1']),
                isLoaded: new FormControl(false)
            };

            expect(CoreUtilsService.getFormValues<MockForm>(controls)).toEqual({
                node: 'node',
                size: 30,
                entities: ['sub0, sub1'],
                isLoaded: false
            });
        });
    });

    describe('getTrimmedFormValues', () => {
        it('should trim object values', () => {
            const values = {
                value: ' value ',
                array: [' array'],
                entity: {
                    node: 'sub '
                }
            };

            expect(CoreUtilsService.getTrimmedFormValues(values)).toEqual({
                value: 'value',
                array: ['array'],
                entity: {
                    node: 'sub'
                }
            });
        });
    });

    describe('toggleControlEnable', () => {
        it('should enable control and add validator', () => {
            const control = new FormControl('', Validators.required);

            CoreUtilsService.toggleControlEnable(control, true);

            expect(control.enabled).toBe(true);
            expect(control.hasValidator(Validators.required)).toBe(true);
        });

        it('should disable control and remove validator', () => {
            const control = new FormControl('text', Validators.required);

            CoreUtilsService.toggleControlEnable(control, false);

            expect(control.disabled).toBe(true);
            expect(control.validator).toBe(null);
        });
    });

    describe('omit', () => {
        it('should omit a single key', () => {
            expect(CoreUtilsService.omit({ a: 1, b: 2 }, 'a')).toEqual({ b: 2 });
        });

        it('should omit multiple keys', () => {
            expect(CoreUtilsService.omit({ a: 1, b: 2, c: 3 }, ['a', 'b'])).toEqual({ c: 3 });
        });
    });

    describe('isEqual', () => {
        it('should return true for equal primitives', () => {
            expect(CoreUtilsService.isEqual(10, 10)).toBe(true);
        });

        it('should return false for non-equal primitives', () => {
            expect(CoreUtilsService.isEqual(10, 20)).toBe(false);
        });

        it('should return true for equal objects', () => {
            expect(CoreUtilsService.isEqual({ a: 1 }, { a: 1 })).toBe(true);
        });

        it('should return false for non-equal nested structures', () => {
            expect(CoreUtilsService.isEqual({ a: { b: 1 } }, { a: { b: 2 } })).toBe(false);
        });

        it('should return false for non-equal objects', () => {
            expect(CoreUtilsService.isEqual({ a: 1 }, { a: 2 })).toBe(false);
        });

        it('should return true for equal arrays', () => {
            expect(CoreUtilsService.isEqual([1, 2], [1, 2])).toBe(true);
        });

        it('should return false for non-equal arrays', () => {
            expect(CoreUtilsService.isEqual([1, 2], [2, 1])).toBe(false);
        });
    });
});
