import { KbqModalService } from '@koobiq/components/modal';
import { take } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import { SharedModalComponent } from './shared-modal.component';
import { SharedModalParams } from './shared-modal.interface';

@Injectable({
    providedIn: 'root'
})
export class SharedModalService {
    private readonly modalService = inject(KbqModalService);

    delete(params: SharedModalParams) {
        this.modalService
            .open<SharedModalComponent, boolean | undefined>({
                kbqContent: SharedModalComponent,
                // @todo: prop is deprecated, replace after koobiq@18 migration
                kbqComponentParams: params,
                kbqClosable: false
            })
            .afterClose.pipe(take(1))
            .subscribe((isSuccessful?: boolean) => {
                if (isSuccessful) {
                    params.confirmHandler();
                } else if (params.cancelHandler) {
                    params.cancelHandler();
                }
            });
    }
}
