import { ElementRef, Renderer2 } from '@angular/core';
import { createEnvironmentInjector, runInInjectionContext } from '@angular/core';

import { RuntimeFeatureHighlightReasonFragmentDirective } from './runtime-highlight-reason-fragment.directive';

describe('RuntimeFeatureHighlightReasonFragmentDirective', () => {
    let elementRef: ElementRef<HTMLElement>;
    let directive: RuntimeFeatureHighlightReasonFragmentDirective;

    const renderer = {
        createElement: jest.fn((tag: string) => document.createElement(tag)),
        createText: jest.fn((text: string) => document.createTextNode(text)),
        appendChild: jest.fn((parent: Node, child: Node) => parent.appendChild(child)),
        addClass: jest.fn((el: HTMLElement, className: string) => el.classList.add(className))
    } as unknown as Renderer2;

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

        directive = runInInjectionContext(injector, () => new RuntimeFeatureHighlightReasonFragmentDirective());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should render plain text without highlights', () => {
        directive.reason = 'simple text';

        directive.ngAfterViewInit();

        expect(elementRef.nativeElement.innerHTML).toEqual('simple text');
    });

    it('should highlight single fragment', () => {
        directive.reason = 'This is `important` reason';

        directive.ngAfterViewInit();

        expect(elementRef.nativeElement.innerHTML).toEqual(
            'This is <span class="runtime-reason-fragment">important</span> reason'
        );
    });

    it('should highlight multiple fragments', () => {
        directive.reason = '`first` and `second`';

        directive.ngAfterViewInit();

        expect(elementRef.nativeElement.innerHTML).toEqual(
            '<span class="runtime-reason-fragment">first</span> and <span class="runtime-reason-fragment">second</span>'
        );
    });

    it('should render empty string', () => {
        directive.reason = '';

        directive.ngAfterViewInit();

        expect(elementRef.nativeElement.innerHTML).toEqual('');
    });
});
