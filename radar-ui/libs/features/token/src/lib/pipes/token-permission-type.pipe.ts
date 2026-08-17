import { Pipe, PipeTransform, inject } from '@angular/core';

import { I18nService } from '@cs/i18n';
import { PermissionType } from '@cs/domains/role';
import { TokenPermissionName } from '@cs/domains/token';

@Pipe({
    name: 'tokenPermissionType',
    pure: false,
    standalone: false
})
export class TokenFeaturePermissionTypePipe implements PipeTransform {
    private readonly i18nService = inject(I18nService);

    transform(type?: PermissionType, permissionName?: TokenPermissionName): string {
        switch (type) {
            case PermissionType.CREATE:
                if (permissionName === TokenPermissionName.RULES) {
                    return this.i18nService.translate('Token.CreateForm.RulePermissions.Label.CanCreate');
                } else {
                    return '—';
                }
            case PermissionType.READ:
                if (permissionName === TokenPermissionName.RULES) {
                    return this.i18nService.translate('Token.CreateForm.RulePermissions.Label.CanRead');
                } else if (permissionName === TokenPermissionName.EVENTS) {
                    return this.i18nService.translate('Token.CreateForm.EventPermissions.Label.CanRead');
                } else if (permissionName === TokenPermissionName.CLUSTERS) {
                    return this.i18nService.translate('Token.CreateForm.ClusterPermissions.Label.CanRead');
                } else {
                    return '—';
                }
            case PermissionType.UPDATE:
                if (permissionName === TokenPermissionName.RULES) {
                    return this.i18nService.translate('Token.CreateForm.RulePermissions.Label.CanUpdate');
                } else {
                    return '—';
                }
            case PermissionType.DELETE:
                if (permissionName === TokenPermissionName.RULES) {
                    return this.i18nService.translate('Token.CreateForm.RulePermissions.Label.CanDelete');
                } else {
                    return '—';
                }
            default:
                return '—';
        }
    }
}
