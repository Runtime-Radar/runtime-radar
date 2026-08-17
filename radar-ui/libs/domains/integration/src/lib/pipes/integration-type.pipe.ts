import { Pipe, PipeTransform } from '@angular/core';

import { INTEGRATION_TYPE } from '../constants/integration.constant';
import { IntegrationType } from '../interfaces/contract/integration-contract.interface';

@Pipe({
    name: 'integrationTypeLocalization',
    pure: false,
    standalone: false
})
export class IntegrationTypeLocalizationPipe implements PipeTransform {
    transform(type?: IntegrationType): string {
        const value = INTEGRATION_TYPE.find((item) => item.id === type);

        return value ? value.localizationKey : '';
    }
}
