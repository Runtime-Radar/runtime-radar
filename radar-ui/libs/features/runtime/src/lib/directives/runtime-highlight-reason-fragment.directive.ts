import { AfterViewInit, Directive, ElementRef, Input, Renderer2, inject } from '@angular/core';

const HIGHLIGHT_FRAGMENT_REG_EXP = /(`[^`]+`)/g;
const HIGHLIGHT_MARKER = '`';
const HIGHLIGHT_CLASS = 'runtime-reason-fragment';

@Directive({
    selector: '[runtimeHighlightReasonFragment]',
    standalone: false
})
export class RuntimeFeatureHighlightReasonFragmentDirective implements AfterViewInit {
    private readonly elementRef = inject(ElementRef);
    private readonly renderer = inject(Renderer2);

    @Input({ required: true }) reason!: string;

    ngAfterViewInit() {
        this.render();
    }

    private render() {
        const host = this.elementRef.nativeElement;
        host.textContent = '';

        this.reason
            .split(HIGHLIGHT_FRAGMENT_REG_EXP)
            .filter((fragment) => !!fragment)
            .forEach((fragment) => {
                if (fragment.startsWith(HIGHLIGHT_MARKER) && fragment.endsWith(HIGHLIGHT_MARKER)) {
                    const span = this.renderer.createElement('span');
                    this.renderer.addClass(span, HIGHLIGHT_CLASS);
                    this.renderer.appendChild(span, this.renderer.createText(fragment.slice(1, -1)));
                    this.renderer.appendChild(host, span);
                } else {
                    this.renderer.appendChild(host, this.renderer.createText(fragment));
                }
            });
    }
}
