import { BehaviorSubject, Subject } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import { CoreWindowService, FORM_VALIDATION_REG_EXP } from '@cs/core';

import { API_PATH } from '../tokens/api-path.token';
import { API_SINGLE_TENANT_PATHS } from '../tokens/api-single-tenant-paths.token';

const API_CLUSTER_PATH_SESSION_KEY = 'pclstrpth';

const API_CLUSTER_URL_QUERY_PARAM_KEY = 'clusterUrl';

@Injectable({
    providedIn: 'root'
})
export class ApiPathService {
    private readonly apiPath = inject<string>(API_PATH);
    private readonly apiSingleTenantPaths = inject<string[]>(API_SINGLE_TENANT_PATHS);
    private readonly coreWindowService = inject(CoreWindowService);

    private host = '';

    private requestedHost = '';

    readonly host$ = new BehaviorSubject(this.host);

    readonly error$ = new Subject<string>();

    initialize() {
        // there are links which should switch cluster based on query params into email templates
        const requested =
            new URLSearchParams(this.coreWindowService.location.search).get(API_CLUSTER_URL_QUERY_PARAM_KEY) || '';
        const stored = this.coreWindowService.sessionStorage.getItem(API_CLUSTER_PATH_SESSION_KEY) || '';

        // such a link can carry any host, so it is applied only once matched against the registered clusters
        if (FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME.test(requested)) {
            this.requestedHost = requested;
        }

        if (FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME.test(stored)) {
            this.setHost(stored);
        }
    }

    /** Returns the yet unvalidated host from the query param and forgets it. */
    takeRequestedHost(): string {
        const value = this.requestedHost;
        this.requestedHost = '';

        return value;
    }

    get(path: string): string {
        const segment = path.substring(0, path.indexOf('/')) || path;
        const host = this.apiSingleTenantPaths.includes(segment) ? '' : this.host;

        return `${host}${this.apiPath}${path}`;
    }

    /** True when url targets this app or the cluster the user switched to, and may carry the token. */
    isOwnRequest(url: string): boolean {
        return url.startsWith(this.apiPath) || (!!this.host && url.startsWith(`${this.host}${this.apiPath}`));
    }

    setHost(value: string) {
        this.host = value;
        this.host$.next(value);
        this.coreWindowService.sessionStorage.setItem(API_CLUSTER_PATH_SESSION_KEY, value);
    }

    setError(value: string) {
        this.error$.next(value);
    }
}
