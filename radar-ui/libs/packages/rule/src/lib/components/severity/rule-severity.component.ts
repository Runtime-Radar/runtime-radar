import { ChangeDetectionStrategy, Component, Input, OnChanges, booleanAttribute } from '@angular/core';

import { RuleSeverity, RuleVerdict } from '@cs/domains/rule';

const SEVERITY_ITERATION = {
    [RuleSeverity.NONE]: 1,
    [RuleSeverity.LOW]: 1,
    [RuleSeverity.MEDIUM]: 2,
    [RuleSeverity.HIGH]: 3,
    [RuleSeverity.CRITICAL]: 3
};

const VERDICT_SEVERITY_RELATIONS = new Map<RuleVerdict, RuleSeverity>([
    [RuleVerdict.NONE, RuleSeverity.NONE],
    [RuleVerdict.UNWANTED, RuleSeverity.MEDIUM],
    [RuleVerdict.DANGEROUS, RuleSeverity.HIGH]
]);

@Component({
    selector: 'cs-rule-package-severity-component',
    templateUrl: './rule-severity.component.html',
    styleUrl: './rule-severity.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RulePackageSeverityComponent implements OnChanges {
    @Input() severity?: RuleSeverity | null;

    @Input() verdict?: RuleVerdict | null;

    @Input() size = 'small'; // small | large

    @Input() direction = 'row-reverse'; // row | row-reverse

    @Input({ transform: booleanAttribute }) isWidthAuto = false;

    severityIteration: string[] = [];

    ngOnChanges() {
        if (this.verdict) {
            this.severity = VERDICT_SEVERITY_RELATIONS.get(this.verdict);
        }

        this.severityIteration = this.severity ? Array(SEVERITY_ITERATION[this.severity]).fill('') : [];
    }
}
