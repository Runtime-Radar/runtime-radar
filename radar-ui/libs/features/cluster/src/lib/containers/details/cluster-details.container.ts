import { ActivatedRoute } from '@angular/router';
import { DateTime } from 'luxon';
import { KbqBadgeColors } from '@koobiq/components/badge';
import { KbqCodeBlockFile } from '@koobiq/components/code-block';
import { ChangeDetectionStrategy, Component, OnInit } from '@angular/core';
import { Observable, map, of, switchMap } from 'rxjs';

import { RouterName } from '@cs/core';
import { CLUSTER_CREATE_FRAGMENT, ClusterRequestService, ClusterStatus, GetClusterResponse } from '@cs/domains/cluster';

@Component({
    templateUrl: './cluster-details.container.html',
    styleUrl: './cluster-details.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ClusterFeatureDetailsContainer implements OnInit {
    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly clusterResponse = this.route.snapshot.data['cluster'] as GetClusterResponse;

    readonly installCommandFiles$: Observable<KbqCodeBlockFile[]> = this.route.params.pipe(
        map((params) => params['clusterId']),
        switchMap((id: string | undefined) => this.getInstallCommand(id, false))
    );

    readonly yamlCommandFiles$: Observable<KbqCodeBlockFile[]> = this.route.params.pipe(
        map((params) => params['clusterId']),
        switchMap((id: string | undefined) => this.getInstallCommand(id, true))
    );

    readonly yamlValues$: Observable<string> = this.route.params.pipe(
        map((params) => params['clusterId']),
        switchMap((id: string | undefined) => (id ? this.clusterRequestService.getClusterYaml(id) : '')),
        map((yaml) => (yaml ? `data:text/yaml,${encodeURI(yaml)}` : ''))
    );

    readonly yamlFileName$: Observable<string> = of(this.clusterResponse).pipe(
        map((response) => this.getClusterYamlFileName(response.cluster.name))
    );

    readonly routerName = RouterName;

    readonly clusterStatus = ClusterStatus;

    readonly badgeColors = KbqBadgeColors;

    readonly dateShortFormat = DateTime.DATE_SHORT;

    isCurrentlyCreated = false;

    constructor(
        private readonly clusterRequestService: ClusterRequestService,
        private readonly route: ActivatedRoute
    ) {}

    ngOnInit() {
        if (this.route.snapshot.fragment === CLUSTER_CREATE_FRAGMENT) {
            this.isCurrentlyCreated = true;
        }
    }

    private getInstallCommand(id: string | undefined, isYaml: boolean): Observable<KbqCodeBlockFile[]> {
        if (!id) {
            return of([]);
        }

        return this.clusterRequestService.getInstallClusterCommand(id, isYaml).pipe(
            map((cmd) => [
                {
                    content: cmd.replace('values.yaml', this.getClusterYamlFileName(this.clusterResponse.cluster.name)),
                    language: 'bash'
                }
            ])
        );
    }

    private getClusterYamlFileName(name: string): string {
        return `${name.toLowerCase().replaceAll(' ', '_')}_values.yaml`;
    }
}
