import { KbqBadgeColors } from '@koobiq/components/badge';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { LicenseRequestService } from '@cs/domains/license';
import { Cluster, ClusterStatus } from '@cs/domains/cluster';

@Component({
    templateUrl: './switch-page.container.html',
    styleUrls: ['./switch-page.container.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SwitchFeaturePageContainer {
    private readonly licenseRequestService = inject(LicenseRequestService);

    readonly url$: Observable<string> = this.licenseRequestService.getCentralUrl().pipe(
        tap((url) => {
            if (!url) {
                console.warn('url must be provided');
            }
        })
    );

    readonly cluster: Omit<Cluster, 'created_at' | 'config'> | null = null;

    readonly clusterStatus = ClusterStatus;

    readonly badgeColors = KbqBadgeColors;
}
