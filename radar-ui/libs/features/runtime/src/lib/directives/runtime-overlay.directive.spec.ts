import { Renderer2 } from '@angular/core';
import { createEnvironmentInjector, runInInjectionContext } from '@angular/core';

import { CoreWindowService } from '@cs/core';

import { RuntimeFeatureOverlayDirective } from './runtime-overlay.directive';

describe('RuntimeFeatureOverlayDirective', () => {
    let coreWindowService: jest.Mocked<CoreWindowService>;
    let directive: RuntimeFeatureOverlayDirective;

    const renderer = { setStyle: jest.fn() } as unknown as Renderer2;
    const overlay = {
        remove: jest.fn(),
        classList: {
            add: jest.fn()
        } as unknown as DOMTokenList
    } as unknown as HTMLDivElement;

    beforeEach(() => {
        coreWindowService = {
            document: {
                createElement: jest.fn(() => overlay),
                body: {
                    appendChild: jest.fn()
                } as unknown as HTMLElement
            } as unknown as Document
        } as jest.Mocked<CoreWindowService>;

        const injector = createEnvironmentInjector(
            [
                {
                    provide: CoreWindowService,
                    useValue: coreWindowService
                },
                {
                    provide: Renderer2,
                    useValue: renderer
                }
            ],
            null as any
        );

        directive = runInInjectionContext(injector, () => new RuntimeFeatureOverlayDirective());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should call setStyle with correct display attr', () => {
        directive.isOverlayed = true;

        directive.ngOnChanges();

        expect(renderer.setStyle).toHaveBeenCalledWith(overlay, 'display', 'block');
    });

    it('should call appendChild', () => {
        directive.ngOnInit();

        expect(coreWindowService.document.body.appendChild).toHaveBeenCalledWith(overlay);
    });

    it('should call remove', () => {
        directive.ngOnDestroy();

        expect(overlay.remove).toHaveBeenCalled();
    });
});
