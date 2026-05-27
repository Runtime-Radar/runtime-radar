import { ElementRef, Renderer2 } from '@angular/core';

import { RuntimeEventType } from '@cs/domains/runtime';

import { RuntimeFeatureEventTypeIconDirective } from './runtime-event-type-icon.directive';

describe('RuntimeFeatureEventTypeIconDirective', () => {
    let elRef: ElementRef;
    let directive: RuntimeFeatureEventTypeIconDirective;

    const renderer = { setStyle: jest.fn() } as unknown as Renderer2;

    beforeEach(() => {
        elRef = new ElementRef(document.createElement('span'));
        directive = new RuntimeFeatureEventTypeIconDirective(elRef, renderer);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should call setStyle with correct background', () => {
        directive.type = RuntimeEventType.KPROBE;

        directive.ngOnInit();

        expect(renderer.setStyle).toHaveBeenCalledWith(
            elRef.nativeElement,
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
