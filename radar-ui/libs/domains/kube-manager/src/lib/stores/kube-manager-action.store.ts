import { createAction, props } from '@ngrx/store';

import { GetKubeManagerPodsRequest, KubeManagerPod, KubeManagerState } from '../interfaces';

export const LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION = createAction('[Kube Manager] Load');

export const POLLING_LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION = createAction('[Kube Manager] Polling Load');

export const LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION = createAction(
    '[Kube Manager] Load With Filters',
    props<{ filters: Partial<GetKubeManagerPodsRequest> }>()
);

export const UPDATE_KUBE_MANAGER_STATE_DOC_ACTION = createAction(
    '[Kube Manager] (Doc) Update State',
    props<Partial<Omit<KubeManagerState, 'list'>>>()
);

export const SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION = createAction(
    '[Kube Manager] (Doc) Set All',
    props<{ list: KubeManagerPod[] }>()
);
