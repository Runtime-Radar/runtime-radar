import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Directive, ElementRef, Input, OnChanges, Renderer2, inject } from '@angular/core';

const EXPIRATION_DAYS_LIMIT = 3;

@Directive({
    selector: '[tokenExpirationColor]',
    standalone: false
})
export class TokenFeatureExpirationColorDirective implements OnChanges {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly el = inject(ElementRef);
    private readonly renderer = inject(Renderer2);

    @Input() expiresAt?: string | null; // RFC3339

    @Input() invalidatedAt?: string;

    ngOnChanges() {
        if (this.invalidatedAt) {
            this.renderer.setStyle(this.el.nativeElement, 'background-color', '#fed3cd');
            this.renderer.setStyle(this.el.nativeElement, 'color', '#c91a03');
        } else if (this.expiresAt) {
            const dateTime = this.dateAdapter.parse(this.expiresAt, undefined);
            const value = dateTime?.diffNow('days').toObject().days || 0;

            if (value <= 0) {
                this.renderer.setStyle(this.el.nativeElement, 'background-color', '#e8e9ee');
                this.renderer.setStyle(this.el.nativeElement, 'color', '#21222c');
            } else if (value <= EXPIRATION_DAYS_LIMIT) {
                this.renderer.setStyle(this.el.nativeElement, 'background-color', '#ffe3b2');
                this.renderer.setStyle(this.el.nativeElement, 'color', '#a86b00');
            } else {
                this.renderer.setStyle(this.el.nativeElement, 'background-color', '#bdecac');
                this.renderer.setStyle(this.el.nativeElement, 'color', '#367d1c');
            }
        } else {
            this.renderer.setStyle(this.el.nativeElement, 'background-color', '#bdecac');
            this.renderer.setStyle(this.el.nativeElement, 'color', '#367d1c');
        }
    }
}
