import { Pipe, PipeTransform } from '@angular/core';

import { RUNTIME_EVENT_TYPE } from '../constants/runtime.constant';
import { RuntimeEventType } from '../interfaces/contract/runtime-event-contract.interface';

@Pipe({
    name: 'runtimeEventTypeLocalization',
    pure: false,
    standalone: false
})
export class RuntimeEventTypeLocalizationPipe implements PipeTransform {
    transform(type?: RuntimeEventType): string {
        const value = RUNTIME_EVENT_TYPE.find((item) => item.id === type);

        return value ? value.localizationKey : '';
    }
}
