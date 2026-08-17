import { Pipe, PipeTransform } from '@angular/core';

import { RULE_VERDICTS, RuleVerdict } from '@cs/domains/rule';

import { RULE_SEVERITY_UNDEFINED_LOCALIZATION_KEY } from '../constants/rule.constant';

@Pipe({
    name: 'ruleVerdictLocalization',
    pure: false,
    standalone: false
})
export class RulePackageVerdictLocalizationPipe implements PipeTransform {
    transform(verdict?: RuleVerdict | null, noneLabelLocalizationKey?: string): string {
        const value = RULE_VERDICTS.find((item) => item.id === verdict);

        if (verdict === RuleVerdict.NONE && noneLabelLocalizationKey) {
            return noneLabelLocalizationKey;
        }

        return value ? value.localizationKey : RULE_SEVERITY_UNDEFINED_LOCALIZATION_KEY;
    }
}
