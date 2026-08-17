import { ChangeDetectionStrategy, Component } from '@angular/core';

import { RouterName } from '@cs/core';

@Component({
    templateUrl: './forbidden-permission.container.html',
    styleUrl: './forbidden-permission.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ForbiddenFeaturePermissionContainer {
    readonly routerName = RouterName;
}
