import { ElementRef, Renderer2 } from '@angular/core';
import { createEnvironmentInjector, runInInjectionContext } from '@angular/core';

import { RuntimeEventType } from '@cs/domains/runtime';

import { RuntimeFeatureEventTypeIconDirective } from './runtime-event-type-icon.directive';

describe('RuntimeFeatureEventTypeIconDirective', () => {
    let elementRef: ElementRef<HTMLElement>;
    let directive: RuntimeFeatureEventTypeIconDirective;

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

        directive = runInInjectionContext(injector, () => new RuntimeFeatureEventTypeIconDirective());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should call setStyle with correct background', () => {
        directive.type = RuntimeEventType.KPROBE;

        directive.ngOnInit();

        expect(renderer.setStyle).toHaveBeenCalledWith(
            elementRef.nativeElement,
            'background',
            `center / cover no-repeat url('/assets/images/runtime/icon-kprobe.svg')`
        );
    });

    it('should not call setStyle when type is unsupported', () => {
        directive.type = 'unsupported' as RuntimeEventType;

        directive.ngOnInit();

        expect(renderer.setStyle).not.toHaveBeenCalled();
    });
});
