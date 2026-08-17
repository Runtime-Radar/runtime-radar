import { KeyValue } from '@angular/common';
import { Pipe, PipeTransform } from '@angular/core';

import { RuntimeSettingPermissionForm } from '../interfaces/runtime-form.interface';

@Pipe({
    name: 'runtimePermissionsFilter',
    standalone: false
})
export class RuntimeFeaturePermissionsFilterPipe implements PipeTransform {
    transform(
        values?: KeyValue<string, RuntimeSettingPermissionForm>[],
        isFilterByAllowedType?: boolean
    ): KeyValue<string, RuntimeSettingPermissionForm>[] {
        return values ? values.filter((item) => item.value.isAllowedType === isFilterByAllowedType) : [];
    }
}
