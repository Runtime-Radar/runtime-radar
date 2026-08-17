import { ElementRef, Renderer2 } from '@angular/core';
import { createEnvironmentInjector, runInInjectionContext } from '@angular/core';

import { RuleSeverity, RuleVerdict } from '@cs/domains/rule';

import { RulePackageSeverityBgColorDirective } from './rule-severity-bgcolor.directive';

describe('RulePackageSeverityBgColorDirective', () => {
    let elementRef: ElementRef<HTMLElement>;
    let directive: RulePackageSeverityBgColorDirective;

    const renderer = { setStyle: jest.fn() } as unknown as Renderer2;

    beforeEach(() => {
        elementRef = new ElementRef(document.createElement('span'));

        const injector = createEnvironmentInjector(
            [
                {
                    provide: ElementRef,
                    useValue: elementRef
                },
                {
                    provide: Renderer2,
                    useValue: renderer
                }
            ],
            null as any
        );

        directive = runInInjectionContext(injector, () => new RulePackageSeverityBgColorDirective());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should call setStyle when severity has been provided', () => {
        directive.severity = RuleSeverity.LOW;

        directive.ngOnChanges();

        expect(renderer.setStyle).toHaveBeenCalledWith(elementRef.nativeElement, 'background-color', '#eff6ff');
    });

    it('should call setStyle when verdict has been provided', () => {
        directive.verdict = RuleVerdict.DANGEROUS;

        directive.ngOnChanges();

        expect(renderer.setStyle).toHaveBeenCalledWith(elementRef.nativeElement, 'background-color', '#fef2ef');
    });
});
