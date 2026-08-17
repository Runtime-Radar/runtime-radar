import { Directive, ElementRef, Input, OnChanges, Renderer2, inject } from '@angular/core';

import { RuleSeverity, RuleVerdict } from '@cs/domains/rule';

const SEVERITY_BACKGROUND_COLORS = new Map<RuleSeverity, string>([
    [RuleSeverity.CRITICAL, '#fff3f3'],
    [RuleSeverity.HIGH, '#fef2ef'],
    [RuleSeverity.MEDIUM, '#fff4dd'],
    [RuleSeverity.LOW, '#eff6ff'],
    [RuleSeverity.NONE, 'inherit']
]);

const VERDICT_BACKGROUND_COLORS = new Map<RuleVerdict, string>([
    [RuleVerdict.DANGEROUS, '#fef2ef'],
    [RuleVerdict.UNWANTED, '#fff4dd'],
    [RuleVerdict.NONE, 'inherit']
]);

@Directive({
    selector: '[severityBgColor]',
    standalone: false
})
export class RulePackageSeverityBgColorDirective implements OnChanges {
    private readonly el = inject(ElementRef);
    private readonly renderer = inject(Renderer2);

    @Input() severity?: RuleSeverity | null;

    @Input() verdict?: RuleVerdict | null;

    ngOnChanges() {
        if (this.severity) {
            this.renderer.setStyle(
                this.el.nativeElement,
                'background-color',
                SEVERITY_BACKGROUND_COLORS.get(this.severity)
            );
        }

        if (this.verdict) {
            this.renderer.setStyle(
                this.el.nativeElement,
                'background-color',
                VERDICT_BACKGROUND_COLORS.get(this.verdict)
            );
        }
    }
}
