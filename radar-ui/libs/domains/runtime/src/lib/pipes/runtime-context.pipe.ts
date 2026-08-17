import { Pipe, PipeTransform } from '@angular/core';

import { RUNTIME_CONTEXT } from '../constants/runtime.constant';
import { RuntimeContext } from '../interfaces/runtime-form.interface';

@Pipe({
    name: 'runtimeContextLocalization',
    pure: false,
    standalone: false
})
export class RuntimeContextLocalizationPipe implements PipeTransform {
    transform(context?: RuntimeContext): string {
        const value = RUNTIME_CONTEXT.find((item) => item.id === context);

        return value ? value.localizationKey : '';
    }
}
