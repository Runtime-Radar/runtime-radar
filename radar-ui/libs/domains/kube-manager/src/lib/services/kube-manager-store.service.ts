import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Store } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import {
    GetKubeManagerPodsRequest,
    KubeManagerNamespace,
    KubeManagerNamespaceGroup,
    KubeManagerNode,
    KubeManagerPodExtended,
    KubeManagerState
} from '../interfaces';
import {
    LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION,
    LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION
} from '../stores/kube-manager-action.store';
import {
    getKubeManagerGroupNamespaces,
    getKubeManagerLoadStatus,
    getKubeManagerNamespaces,
    getKubeManagerNodes,
    getKubeManagerPodsByNamespace
} from '../stores/kube-manager-selector.store';

@Injectable({
    providedIn: 'root'
})
export class KubeManagerStoreService {
    readonly loadStatus$: Observable<LoadStatus> = this.store.select(getKubeManagerLoadStatus);

    readonly nodes$: Observable<KubeManagerNode[]> = this.store.select(getKubeManagerNodes);

    readonly namespaces$ = (node?: string): Observable<KubeManagerNamespace[]> =>
        this.store.select(getKubeManagerNamespaces(node));

    readonly groupNamespaces$ = (node?: string): Observable<KubeManagerNamespaceGroup[]> =>
        this.store.select(getKubeManagerGroupNamespaces(node));

    readonly pods$ = (namespace: string, node?: string): Observable<KubeManagerPodExtended[]> =>
        this.store.select(getKubeManagerPodsByNamespace(namespace, node));

    constructor(private readonly store: Store<KubeManagerState>) {}

    initPods() {
        this.store.dispatch(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION());
    }

    loadPods(filters: Partial<GetKubeManagerPodsRequest>) {
        this.store.dispatch(LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION({ filters }));
    }
}
