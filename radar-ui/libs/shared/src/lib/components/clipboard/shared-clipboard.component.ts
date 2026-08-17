import { ChangeDetectionStrategy, Component, Input, booleanAttribute, inject } from '@angular/core';

import { SharedClipboardService } from '@cs/shared';

@Component({
    selector: 'cs-clipboard-component',
    templateUrl: './shared-clipboard.component.html',
    styleUrl: './shared-clipboard.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedClipboardComponent {
    private readonly clipboardService = inject(SharedClipboardService);

    @Input({ required: true }) value!: string | null;

    @Input({ transform: booleanAttribute }) isButtonTextVisible = false;

    @Input() isDisabled? = false;

    copyToClipboard() {
        if (this.value) {
            this.clipboardService.copyToClipboard(this.value);
        }
    }
}
