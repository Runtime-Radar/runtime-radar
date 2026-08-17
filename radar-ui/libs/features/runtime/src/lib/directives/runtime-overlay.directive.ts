import { Directive, Input, OnChanges, OnDestroy, OnInit, Renderer2, booleanAttribute, inject } from '@angular/core';

import { CoreWindowService } from '@cs/core';

@Directive({
    selector: '[overlay]',
    standalone: false
})
export class RuntimeFeatureOverlayDirective implements OnChanges, OnInit, OnDestroy {
    private readonly coreWindowService = inject(CoreWindowService);
    private readonly renderer = inject(Renderer2);

    @Input({ transform: booleanAttribute }) isOverlayed = false;

    private readonly overlay: HTMLDivElement;

    constructor() {
        this.overlay = this.coreWindowService.document.createElement('div');
        this.overlay.classList.add('overlay');
        this.renderer.setStyle(this.overlay, 'display', 'none');
        this.renderer.setStyle(this.overlay, 'position', 'fixed');
        this.renderer.setStyle(this.overlay, 'top', '0');
        this.renderer.setStyle(this.overlay, 'left', '0');
        this.renderer.setStyle(this.overlay, 'width', '100%');
        this.renderer.setStyle(this.overlay, 'height', '100%');
        this.renderer.setStyle(this.overlay, 'z-index', '2');
        this.renderer.setStyle(this.overlay, 'background-color', 'rgba(0, 0, 0, 0.5)');
    }

    ngOnChanges() {
        this.renderer.setStyle(this.overlay, 'display', this.isOverlayed ? 'block' : 'none');
    }

    ngOnInit() {
        this.coreWindowService.document.body.appendChild(this.overlay);
    }

    ngOnDestroy() {
        this.overlay.remove();
    }
}
