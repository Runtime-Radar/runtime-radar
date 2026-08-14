import { Observable } from 'rxjs';
import { provideAutoSpy } from 'jest-auto-spies';
import { render } from '@testing-library/angular';
import { AsyncValidatorFn, FormControl, ValidationErrors } from '@angular/forms';

import { RuntimeFeaturePolicyNameService } from '../services/runtime-policy-name.service';
import { RuntimeFeaturePolicyNameValidator } from './runtime-policy-name.validator';

describe('RuntimeFeaturePolicyNameValidator', () => {
    let runtimeFeaturePolicyNameService: jest.Mocked<RuntimeFeaturePolicyNameService>;
    let nameUniqueValidator: AsyncValidatorFn;

    beforeEach(async () => {
        const { fixture } = await render('<div></div>', {
            providers: [provideAutoSpy(RuntimeFeaturePolicyNameService)]
        });

        runtimeFeaturePolicyNameService = fixture.debugElement.injector.get(
            RuntimeFeaturePolicyNameService
        ) as jest.Mocked<RuntimeFeaturePolicyNameService>;
        nameUniqueValidator = RuntimeFeaturePolicyNameValidator.isNameUnique(
            runtimeFeaturePolicyNameService,
            'originName'
        );

        runtimeFeaturePolicyNameService.get.mockReturnValue(['originName', 'controlName']);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('isNameUnique', () => {
        it('should return name validation error', (done) => {
            (nameUniqueValidator(new FormControl('controlName')) as Observable<ValidationErrors | null>).subscribe(
                (result) => {
                    expect(result).toEqual({ nameExists: true });
                    done();
                }
            );
        });

        it('should return nullable value when name is already exist', (done) => {
            (nameUniqueValidator(new FormControl('name')) as Observable<ValidationErrors | null>).subscribe(
                (result) => {
                    expect(result).toBeNull();
                    done();
                }
            );
        });

        it('should return nullable value', (done) => {
            (nameUniqueValidator(new FormControl('')) as Observable<ValidationErrors | null>).subscribe((result) => {
                expect(result).toBeNull();
                done();
            });
        });
    });
});
