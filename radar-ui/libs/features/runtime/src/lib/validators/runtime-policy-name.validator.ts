import { AbstractControl, AsyncValidatorFn, ValidationErrors } from '@angular/forms';
import { Observable, map, of } from 'rxjs';

import { RuntimeFeaturePolicyNameService } from '../services/runtime-policy-name.service';

export class RuntimeFeaturePolicyNameValidator {
    static isNameUnique(
        runtimeFeaturePolicyNameService: RuntimeFeaturePolicyNameService,
        originName: string | undefined
    ): AsyncValidatorFn {
        return (control: AbstractControl): Observable<ValidationErrors | null> => {
            if (!control.value) {
                return of(null);
            }

            return of(control.value).pipe(
                map((value) => {
                    const names = runtimeFeaturePolicyNameService.get().filter((name) => name !== originName);

                    return names.includes(value) ? { nameExists: true } : null;
                })
            );
        };
    }
}
