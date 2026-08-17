import { ElementRef, Renderer2 } from '@angular/core';
import { createEnvironmentInjector, runInInjectionContext } from '@angular/core';

import { RuleSeverity, RuleVerdict } from '@cs/domains/rule';

import { RulePackageSeverityColorDirective } from './rule-severity-color.directive';

describe('RulePackageSeverityColorDirective', () => {
    let elementRef: ElementRef<HTMLElement>;
    let directive: RulePackageSeverityColorDirective;

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

        directive = runInInjectionContext(injector, () => new RulePackageSeverityColorDirective());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should call setStyle when severity has been provided', () => {
        directive.severity = RuleSeverity.HIGH;

        directive.ngOnChanges();

        expect(renderer.setStyle).toHaveBeenCalledWith(elementRef.nativeElement, 'color', '#e8612d');
    });

    it('should call setStyle when verdict has been provided', () => {
        directive.verdict = RuleVerdict.UNWANTED;

        directive.ngOnChanges();

        expect(renderer.setStyle).toHaveBeenCalledWith(elementRef.nativeElement, 'color', '#e19f12');
    });
});
