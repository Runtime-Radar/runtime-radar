import { Pipe, PipeTransform } from '@angular/core';

import { RULE_SEVERITIES, RuleSeverity } from '@cs/domains/rule';

import { RULE_SEVERITY_UNDEFINED_LOCALIZATION_KEY } from '../constants/rule.constant';

@Pipe({
    name: 'ruleSeverityLocalization',
    pure: false,
    standalone: false
})
export class RulePackageSeverityLocalizationPipe implements PipeTransform {
    transform(severity?: RuleSeverity | null, noneLabelLocalizationKey?: string): string {
        const value = RULE_SEVERITIES.find((item) => item.id === severity);

        if (severity === RuleSeverity.NONE && noneLabelLocalizationKey) {
            return noneLabelLocalizationKey;
        }

        return value ? value.localizationKey : RULE_SEVERITY_UNDEFINED_LOCALIZATION_KEY;
    }
}
