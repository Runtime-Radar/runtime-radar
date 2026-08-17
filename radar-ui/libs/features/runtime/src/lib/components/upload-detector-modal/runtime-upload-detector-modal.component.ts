import { KbqModalRef } from '@koobiq/components/modal';
import { BehaviorSubject, Observable, map } from 'rxjs';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';

import { I18nService } from '@cs/i18n';
import { CoreUtilsService as utils } from '@cs/core';

interface RuntimeDetectorFile {
    id: string;
    name: string;
    size: number;
    base64: string;
    isDeleted: boolean;
}

@Component({
    templateUrl: './runtime-upload-detector-modal.component.html',
    styleUrl: './runtime-upload-detector-modal.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureUploadDetectorModalComponent {
    private readonly modal = inject(KbqModalRef);
    private readonly toastService = inject(KbqToastService);

    private readonly i18nService = inject(I18nService);

    private readonly files$ = new BehaviorSubject<RuntimeDetectorFile[]>([]);

    readonly isUploadInProgress$ = new BehaviorSubject(false);

    readonly uploadedFiles$: Observable<RuntimeDetectorFile[]> = this.files$.pipe(
        map((list) => list.filter((item) => !item.isDeleted))
    );

    upload(event: Event) {
        const files: FileList = (event.target as HTMLInputElement).files || ({} as FileList);
        Object.keys(files).forEach((key) => {
            const file = files.item(key as unknown as number);
            if (file && file.type === 'application/wasm') {
                const reader = new FileReader();
                reader.readAsDataURL(file);
                reader.onloadstart = () => {
                    this.isUploadInProgress$.next(true);
                };
                reader.onloadend = () => {
                    this.isUploadInProgress$.next(false);
                };
                reader.onload = () => {
                    const result = reader.result;
                    if (typeof result === 'string') {
                        /* eslint @typescript-eslint/no-magic-numbers: "off" */
                        const values = this.getPatchedFiles(file, result.substring(29));
                        this.files$.next(values);
                    }
                };
                reader.onerror = () => {
                    console.warn('file must be valid');
                };
            }
        });

        (event.target as HTMLInputElement).value = '';
    }

    delete(id: string) {
        this.files$.next(
            this.files$.value.map((item) => {
                if (item.id === id) {
                    item.isDeleted = true;
                }

                return item;
            })
        );
    }

    dispatch(isSuccessful: boolean) {
        const base64values = this.files$.value.filter((item) => !item.isDeleted).map((item) => item.base64);
        this.modal.destroy(isSuccessful ? base64values : undefined);
    }

    private getPatchedFiles(file: File, base64: string): RuntimeDetectorFile[] {
        const files = this.files$.value;
        const item = files.find((obj) => obj.base64 === base64 && obj.name === file.name);
        const value: RuntimeDetectorFile = {
            id: utils.generateUuid(),
            name: file.name,
            size: file.size,
            isDeleted: false,
            base64
        };

        if (item) {
            if (!item.isDeleted) {
                this.toastService.show({
                    style: KbqToastStyle.Warning,
                    title: this.i18nService.translate('Runtime.Pseudo.Notification.DetectorDuplicate')
                });
            }

            return files.map((obj) => {
                if (obj.id === item.id) {
                    obj.isDeleted = false;
                }

                return obj;
            });
        }

        return [...files, value];
    }
}
