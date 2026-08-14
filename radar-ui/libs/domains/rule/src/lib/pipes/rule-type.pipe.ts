import { Pipe, PipeTransform } from '@angular/core';

import { RULE_TYPE } from '../constants/rule.constant';
import { RuleType } from '../interfaces/contract/rule-contract.interface';

@Pipe({
    name: 'ruleTypeLocalization',
    pure: false,
    standalone: false
})
export class RuleTypeLocalizationPipe implements PipeTransform {
    transform(type?: RuleType): string {
        const value = RULE_TYPE.find((item) => item.id === type);

        return value ? value.localizationKey : '';
    }
}
