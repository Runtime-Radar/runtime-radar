import { ChangeDetectionStrategy, Component, Input } from '@angular/core';

import { RuleSeverity } from '@cs/domains/rule';

@Component({
    selector: 'cs-rule-package-severity-grade-component',
    templateUrl: './rule-severity-grade.component.html',
    styleUrl: './rule-severity-grade.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RulePackageSeverityGradeComponent {
    @Input() low?: string;

    @Input() medium?: string;

    @Input() high?: string;

    @Input() critical?: string;

    readonly ruleSeverity = RuleSeverity;
}
