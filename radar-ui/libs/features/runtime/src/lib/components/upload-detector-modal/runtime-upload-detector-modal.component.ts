import { KbqModalRef } from '@koobiq/components/modal';
import { BehaviorSubject, Observable, map } from 'rxjs';
import { ChangeDetectionStrategy, Component } from '@angular/core';
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
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RuntimeFeatureUploadDetectorModalComponent {
    private readonly files$ = new BehaviorSubject<RuntimeDetectorFile[]>([]);

    readonly isUploadInProgress$ = new BehaviorSubject(false);

    readonly uploadedFiles$: Observable<RuntimeDetectorFile[]> = this.files$.pipe(
        map((list) => list.filter((item) => !item.isDeleted))
    );

    constructor(
        private readonly i18nService: I18nService,
        private readonly modal: KbqModalRef,
        private readonly toastService: KbqToastService
    ) {}

    upload(event: Event): void {
        const input = event.target as HTMLInputElement;

        if (!input) {
            return;
        }

        const fileList = input.files;
        const selectedFiles = fileList ? Array.from(fileList) : [];

        input.value = '';

        const wasmFiles = selectedFiles.filter((file) => this.isWasmFile(file));

        if (wasmFiles.length === 0) {
            this.warnInvalidFile();
            return;
        }

        this.isUploadInProgress$.next(true);

        let completed = 0;

        wasmFiles.forEach((file) => {
            const reader = new FileReader();

            reader.onload = () => {
                const result = reader.result;

                if (typeof result === 'string') {
                    const commaIndex = result.indexOf(',');
                    const base64 = commaIndex >= 0 ? result.slice(commaIndex + 1) : '';

                    if (this.isWasmBase64(base64)) {
                        const values = this.getPatchedFiles(file, base64);
                        this.files$.next(values);
                    } else {
                        this.warnInvalidFile(file.name);
                    }
                }
            };

            reader.onerror = () => {
                this.warnInvalidFile(file.name);
            };

            reader.onloadend = () => {
                completed += 1;

                if (completed === wasmFiles.length) {
                    this.isUploadInProgress$.next(false);
                }
            };

            reader.readAsDataURL(file);
        });
    }

    delete(id: string): void {
        this.files$.next(this.files$.value.map((item) => (item.id === id ? { ...item, isDeleted: true } : item)));
    }

    dispatch(isSuccessful: boolean): void {
        const base64values = this.files$.value.filter((item) => !item.isDeleted).map((item) => item.base64);

        this.modal.destroy(isSuccessful ? base64values : undefined);
    }

    private warnInvalidFile(fileName?: string): void {
        const message = fileName ? `File "${fileName}" is not a valid WASM` : 'No valid WASM files selected';

        console.warn(message);
    }

    private isWasmFile(file: File): boolean {
        return file.type === 'application/wasm' || file.name.toLowerCase().endsWith('.wasm');
    }

    // WASM magic bytes: 0x0061736d -> Base64: AGFzbQ
    private isWasmBase64(base64: string): boolean {
        return base64.startsWith('AGFzbQ');
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

            return files.map((obj) => (obj.id === item.id ? { ...obj, isDeleted: false } : obj));
        }

        return [...files, value];
    }
}
