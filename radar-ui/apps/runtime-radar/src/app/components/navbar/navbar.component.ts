import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { Observable } from 'rxjs';
import { take } from 'rxjs/operators';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output, ViewChild } from '@angular/core';
import { IModalOptionsForService, KbqModalService, ModalSize } from '@koobiq/components/modal';

import { AuthCredentials } from '@cs/domains/auth';
import { LicenseStoreService } from '@cs/domains/license';
import { SharedPasswordModalComponent } from '@cs/shared';
import { UserStoreService } from '@cs/domains/user';
import { CoreNavigationStoreService, LoadStatus, NAVIGATION, NavigationMenu, RouterName } from '@cs/core';

@Component({
    selector: 'cs-navbar-component',
    templateUrl: './navbar.component.html',
    styleUrl: './navbar.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class NavbarComponent {
    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input({ required: true }) credentials!: AuthCredentials;

    @Input() role?: string;

    @Output() signOut = new EventEmitter<void>();

    readonly appVersion$: Observable<string> = this.licenseStoreService.appVersion$;

    readonly loadStatus$ = this.coreNavigationStoreService.loadStatus$;

    readonly navigationItems: NavigationMenu[] = NAVIGATION;

    readonly loadStatus = LoadStatus;

    readonly routerName = RouterName;

    constructor(
        private readonly coreNavigationStoreService: CoreNavigationStoreService,
        private readonly modalService: KbqModalService,
        private readonly userStoreService: UserStoreService,
        private readonly licenseStoreService: LicenseStoreService
    ) {}

    hidePopover() {
        this.popover.hide();
    }

    changePassword() {
        const config: IModalOptionsForService = {
            kbqComponent: SharedPasswordModalComponent,
            kbqSize: ModalSize.Small,
            kbqClosable: false
        };

        this.hidePopover();
        this.modalService
            .open<SharedPasswordModalComponent, string | undefined>(config)
            .afterClose.pipe(take(1))
            .subscribe((password?: string) => {
                if (password) {
                    this.userStoreService.changePassword(password);
                }
            });
    }

    onSignOut() {
        this.signOut.emit();
    }
}
