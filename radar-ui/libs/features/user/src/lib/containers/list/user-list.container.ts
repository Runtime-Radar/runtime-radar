import { ActivatedRoute } from '@angular/router';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KbqSidepanelConfig, KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Observable, filter, take } from 'rxjs';

import { AuthStoreService } from '@cs/domains/auth';
import { I18nService } from '@cs/i18n';
import { SharedModalService } from '@cs/shared';
import { CoreUtilsService as utils } from '@cs/core';
import { PermissionName, PermissionType, Role, RolePermissionMap, RoleStoreService } from '@cs/domains/role';
import { User, UserStoreService } from '@cs/domains/user';

import { UserFeatureSidepanelUserFormComponent } from '../../components/sidepanel-user-form/user-sidepanel-user-form.component';
import { UserForm } from '../../interfaces/user-form.interface';
import { UserSidepanelFormProps } from '../../interfaces/user-sidepanel.interface';

@Component({
    templateUrl: './user-list.container.html',
    styleUrl: './user-list.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class UserFeatureListContainer {
    private readonly route = inject(ActivatedRoute);
    private readonly sidepanelService = inject(KbqSidepanelService);

    private readonly authStoreService = inject(AuthStoreService);
    private readonly i18nService = inject(I18nService);
    private readonly roleStoreService = inject(RoleStoreService);
    private readonly sharedModalService = inject(SharedModalService);
    private readonly userStoreService = inject(UserStoreService);

    readonly users$: Observable<User[]> = this.userStoreService.users$;

    readonly roles$: Observable<Role[]> = this.roleStoreService.roles$;

    readonly role$ = (id: string): Observable<Role | undefined> => this.roleStoreService.role$(id);

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    openCreateSidepanel(roleId?: string) {
        const config: KbqSidepanelConfig<UserSidepanelFormProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                roles$: this.roles$,
                credentials$: this.authStoreService.credentials$,
                isEdit: false,
                user: {
                    role_id: roleId
                }
            }
        };

        this.sidepanelService
            .open(UserFeatureSidepanelUserFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((form: UserForm) => {
                this.userStoreService.createUser({
                    username: form.username,
                    email: form.email,
                    password: form.password,
                    role_id: form.roleId
                });
            });
    }

    openEditSidepanel(user: User) {
        const config: KbqSidepanelConfig<UserSidepanelFormProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                roles$: this.roles$,
                credentials$: this.authStoreService.credentials$,
                isEdit: true,
                user
            }
        };

        this.sidepanelService
            .open(UserFeatureSidepanelUserFormComponent, config)
            .afterClosed()
            .pipe(take(1), filter(utils.isDefined))
            .subscribe((form: UserForm) => {
                this.userStoreService.updateUser(user.id, {
                    email: form.email,
                    password: form.password,
                    roleId: form.roleId
                });
            });
    }

    openDeleteModal(id: string) {
        this.sharedModalService.delete({
            content: this.i18nService.translate('User.DeleteModal.Content.Text'),
            confirmText: this.i18nService.translate('User.DeleteModal.Button.Confirm'),
            cancelText: this.i18nService.translate('User.DeleteModal.Button.Cancel'),
            confirmHandler: () => {
                this.userStoreService.deleteUser(id);
            }
        });
    }
}
