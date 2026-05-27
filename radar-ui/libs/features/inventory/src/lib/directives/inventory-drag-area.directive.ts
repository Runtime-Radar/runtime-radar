import { AfterViewInit, Directive, ElementRef, EventEmitter, HostListener, Input, OnInit, Output } from '@angular/core';

import { CoreWindowService } from '@cs/core';

@Directive({
    selector: '[dragArea]'
})
export class InventoryFeatureDragAreaDirective implements OnInit, AfterViewInit {
    @Input({ required: true }) dragContent!: HTMLElement;

    @Input() dragIgnoreClassName?: string;

    @Output() scaleChange = new EventEmitter<number>();

    el!: HTMLElement;

    startX = 0;
    startY = 0;

    startTop = 0;
    startLeft = 0;

    scale = 1;
    minScale = 0.3;
    maxScale = 1;

    isDraggable = false;

    constructor(
        private readonly elRef: ElementRef<HTMLElement>,
        private readonly coreWindowService: CoreWindowService
    ) {}

    ngOnInit() {
        this.el = this.elRef.nativeElement;
    }

    ngAfterViewInit() {
        this.dragContent.style.left = '0';
        this.dragContent.style.top = '0';
        this.dragContent.style.transformOrigin = '0 0';
    }

    @HostListener('pointerdown', ['$event'])
    onPointerDown(event: PointerEvent) {
        if (event.pointerType === 'mouse' && event.button !== 0) {
            return;
        }

        if (this.dragIgnoreClassName && (event.target as HTMLElement).closest(this.dragIgnoreClassName)) {
            return;
        }

        this.el.setPointerCapture(event.pointerId);
        this.isDraggable = true;
        this.startX = event.clientX;
        this.startY = event.clientY;
        this.startLeft = parseFloat(this.dragContent.style.left) || 0;
        this.startTop = parseFloat(this.dragContent.style.top) || 0;
    }

    @HostListener('pointermove', ['$event'])
    onPointerMove(event: PointerEvent) {
        if (!this.isDraggable) {
            return;
        }

        this.dragContent.style.left = this.startLeft + (event.clientX - this.startX) + 'px';
        this.dragContent.style.top = this.startTop + (event.clientY - this.startY) + 'px';
    }

    @HostListener('pointerup', ['$event'])
    @HostListener('pointercancel', ['$event'])
    onPointerEnd(event: PointerEvent) {
        this.isDraggable = false;

        this.el.releasePointerCapture(event.pointerId);
    }

    // @todo: add debounce for wheel
    @HostListener('wheel', ['$event'])
    onWheel(event: WheelEvent) {
        event.preventDefault();

        const delta = event.deltaY < 0 ? 1.1 : 0.9;
        this.scaleArea(delta, event.clientY, event.clientX);
    }

    /** @external */
    zoom(scale: number) {
        const screen = this.coreWindowService.innerScreen;
        this.scaleArea(scale, screen.height / 2, screen.width / 2);
    }

    /** @external */
    fit() {
        this.applyScaleParams(1, 0, 0);
    }

    private scaleArea(delta: number, clientY: number, clientX: number) {
        const scale = this.scale * delta;
        if (scale < this.minScale || scale > this.maxScale) {
            return;
        }

        const rect = this.dragContent.getBoundingClientRect();
        const top = parseFloat(this.dragContent.style.top || '0') - (clientY - rect.top) * (delta - 1);
        const left = parseFloat(this.dragContent.style.left || '0') - (clientX - rect.left) * (delta - 1);

        this.applyScaleParams(scale, top, left);
    }

    private applyScaleParams(scale: number, top: number, left: number) {
        this.scale = scale;
        this.scaleChange.emit(scale);

        this.dragContent.style.transform = `scale(${this.scale})`;
        this.dragContent.style.top = `${top}px`;
        this.dragContent.style.left = `${left}px`;
    }
}
