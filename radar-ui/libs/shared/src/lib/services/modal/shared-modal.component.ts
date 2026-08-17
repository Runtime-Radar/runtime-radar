import { CommonModule } from '@angular/common';
import { KbqButtonModule } from '@koobiq/components/button';
import { ChangeDetectionStrategy, Component, Input, inject } from '@angular/core';
import { KbqModalModule, KbqModalRef } from '@koobiq/components/modal';

@Component({
    templateUrl: './shared-modal.component.html',
    styleUrl: './shared-modal.component.scss',
    imports: [KbqModalModule, KbqButtonModule, CommonModule],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: true
})
export class SharedModalComponent {
    private readonly modal = inject(KbqModalRef);

    @Input() title?: string;

    @Input({ required: true }) content!: string;

    @Input({ required: true }) confirmText!: string;

    @Input({ required: true }) cancelText!: string;

    close(isSuccessful: boolean) {
        this.modal.destroy(isSuccessful || undefined);
    }
}
