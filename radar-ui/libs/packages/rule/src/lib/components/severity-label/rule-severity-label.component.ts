import { ChangeDetectionStrategy, Component, Input, booleanAttribute } from '@angular/core';

import { RuleSeverity, RuleVerdict } from '@cs/domains/rule';

import { RULE_SEVERITY_UNDEFINED_LOCALIZATION_KEY } from '../../constants/rule.constant';

@Component({
    selector: 'cs-rule-package-severity-label-component',
    templateUrl: './rule-severity-label.component.html',
    styleUrl: './rule-severity-label.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RulePackageSeverityLabelComponent {
    @Input() severity?: RuleSeverity | null;

    @Input() verdict?: RuleVerdict | null;

    @Input() noneLabelLocalizationKey?: string;

    @Input({ transform: booleanAttribute }) isWidthAuto = false;

    @Input({ transform: booleanAttribute }) isLabelColored = false;

    readonly undefinedLocalizationKey = RULE_SEVERITY_UNDEFINED_LOCALIZATION_KEY;
}
