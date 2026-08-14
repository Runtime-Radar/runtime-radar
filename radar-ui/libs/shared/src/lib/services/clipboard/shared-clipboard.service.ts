import { KbqAlertColors } from '@koobiq/components/alert';
import { KbqToastService } from '@koobiq/components/toast';
import { Injectable, inject } from '@angular/core';

import { CoreWindowService } from '@cs/core';
import { I18nService } from '@cs/i18n';

@Injectable({
    providedIn: 'root'
})
export class SharedClipboardService {
    private readonly toastService = inject(KbqToastService);

    private readonly coreWindowService = inject(CoreWindowService);
    private readonly i18nService = inject(I18nService);

    copyToClipboard(value: string) {
        if (!this.coreWindowService.isSecureContext && !this.coreWindowService.navigator.clipboard) {
            this.toastService.show({
                style: KbqAlertColors.Warning,
                title: this.i18nService.translate('Common.Clipboard.Notification.ClipboardDenied')
            });

            return;
        }

        this.coreWindowService.navigator.clipboard.writeText(value).then(() => {
            this.toastService.show({
                style: KbqAlertColors.Success,
                title: this.i18nService.translate('Common.Clipboard.Notification.CopiedToClipboard')
            });
        });
    }
}
