import { Dictionary } from '@ngrx/entity';
import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { kubeManagerReducer } from './kube-manager-reducer.store';
import {
    KubeManagerEntityState,
    KubeManagerNamespace,
    KubeManagerNamespaceGroup,
    KubeManagerNode,
    KubeManagerPod,
    KubeManagerPodExtended,
    KubeManagerState
} from '../interfaces';

export const KUBE_MANAGER_DOMAIN_KEY = 'kube-manager';

export interface KubeManagerDomainState {
    readonly domain: KubeManagerState;
}

const getPodsByNamespace = (
    entities: Dictionary<KubeManagerPod>,
    namespace: string,
    node?: string
): KubeManagerPodExtended[] =>
    Object.values(entities)
        .filter((item) => !!item)
        .filter((item) => item.namespace === namespace)
        .map((item) => ({
            ...item,
            isVisible: node ? item.node_name === node : null
        }));

const selectKubeManagerDomainState = createFeatureSelector<KubeManagerDomainState>(KUBE_MANAGER_DOMAIN_KEY);
const selectKubeManagerState = createSelector(
    selectKubeManagerDomainState,
    (state: KubeManagerDomainState) => state.domain
);
const selectKubeManagerEntityState = createSelector(selectKubeManagerState, (state: KubeManagerState) => state.list);

export const getKubeManagerLoadStatus = createSelector(
    selectKubeManagerState,
    (state: KubeManagerState) => state.loadStatus
);

export const getKubeManagerLastUpdate = createSelector(
    selectKubeManagerState,
    (state: KubeManagerState) => state.lastUpdate
);

export const getKubeManagerNodes = createSelector(
    selectKubeManagerEntityState,
    (state: KubeManagerEntityState): KubeManagerNode[] => {
        const nodes = Object.values(state.entities).reduce(
            (acc, item) => (item && item.node_name ? [...acc, item.node_name] : acc),
            [] as string[]
        );

        return [...new Set(nodes)].map((node, i) => ({
            id: `nodeId${i}`,
            node_name: node
        }));
    }
);

export const getKubeManagerNamespaces = (node?: string) =>
    createSelector(selectKubeManagerEntityState, (state: KubeManagerEntityState): KubeManagerNamespace[] => {
        const namespaces = Object.values(state.entities).reduce(
            (acc, item) => (item ? [...acc, item.namespace] : acc),
            [] as string[]
        );

        return [...new Set(namespaces)].map((namespace, i) => {
            const pods = getPodsByNamespace(state.entities, namespace, node);
            const nodes = pods.reduce((acc, item) => {
                if (item.node_name && !acc.includes(item.node_name)) {
                    acc.push(item.node_name);
                }

                return acc;
            }, [] as string[]);

            return {
                id: `namespaceId${i}`,
                namespace,
                nodes,
                podCount: pods.length,
                isVisible: node ? pods.some((item) => item.isVisible) : null
            };
        });
    });

export const getKubeManagerGroupNamespaces = (node?: string) =>
    createSelector(getKubeManagerNamespaces(node), (state: KubeManagerNamespace[]): KubeManagerNamespaceGroup[] => {
        return state.reduce((acc, item, i) => {
            if (item.nodes.length === 1) {
                const gi = acc.findIndex(
                    (g) => g.namespaces[0].nodes.length === 1 && g.namespaces[0].nodes[0] === item.nodes[0]
                );

                if (gi !== -1) {
                    acc[gi].namespaces.push(item);
                } else {
                    acc.push({
                        id: `groupId${i}`,
                        namespaces: [item]
                    });
                }
            } else {
                acc.push({
                    id: `groupId${i}`,
                    namespaces: [item]
                });
            }

            return acc;
        }, [] as KubeManagerNamespaceGroup[]);
    });

export const getKubeManagerPodsByNamespace = (namespace: string, node?: string) =>
    createSelector(selectKubeManagerEntityState, (state: KubeManagerEntityState) =>
        getPodsByNamespace(state.entities, namespace, node)
    );

export const kubeManagerDomainReducer: ActionReducerMap<KubeManagerDomainState> = {
    domain: kubeManagerReducer
};
