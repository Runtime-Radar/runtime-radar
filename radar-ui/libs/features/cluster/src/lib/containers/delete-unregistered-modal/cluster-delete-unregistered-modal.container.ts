import { KbqCodeBlockFile } from '@koobiq/components/code-block';
import { KbqModalRef } from '@koobiq/components/modal';
import { catchError } from 'rxjs/operators';
import { AfterViewInit, ChangeDetectionStrategy, Component, Input, inject } from '@angular/core';
import { BehaviorSubject, Observable, map, of, switchMap } from 'rxjs';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';

import { ClusterRequestService } from '@cs/domains/cluster';
import { I18nService } from '@cs/i18n';
import { SharedClipboardService } from '@cs/shared';

@Component({
    templateUrl: './cluster-delete-unregistered-modal.container.html',
    styleUrl: './cluster-delete-unregistered-modal.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureDeleteUnregisteredModalContainer implements AfterViewInit {
    private readonly modal = inject(KbqModalRef);
    private readonly toastService = inject(KbqToastService);

    private readonly clipboardService = inject(SharedClipboardService);
    private readonly clusterRequestService = inject(ClusterRequestService);
    private readonly i18nService = inject(I18nService);

    @Input() id?: string;

    private id$ = new BehaviorSubject<string | undefined>(undefined);

    readonly commandFiles$: Observable<KbqCodeBlockFile[]> = this.id$.pipe(
        switchMap((id) => {
            if (!id) {
                return of([]);
            }

            return this.clusterRequestService.getUninstallClusterCommand(id).pipe(
                map((content) => [
                    {
                        content,
                        language: 'bash'
                    }
                ]),
                catchError(() => {
                    this.toastService.show({
                        style: KbqToastStyle.Error,
                        title: this.i18nService.translate('Cluster.DeleteModal.Notification.Text.CommandFailed')
                    });

                    // @todo: refactor this logic since there is modal's blink
                    this.modal.destroy();

                    return of([]);
                })
            );
        })
    );

    ngAfterViewInit() {
        if (this.id) {
            this.id$.next(this.id);
        }
    }

    dispatch(cmd?: string) {
        if (cmd) {
            this.clipboardService.copyToClipboard(cmd);
        }

        this.modal.destroy();
    }
}
