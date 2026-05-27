import { EntityState } from '@ngrx/entity';

import { LoadStatus } from '@cs/core';

import { KubeManagerPod } from '../contract/kube-manager-contract.interface';

export type KubeManagerEntityState = EntityState<KubeManagerPod>;

export interface KubeManagerState {
    loadStatus: LoadStatus;
    lastUpdate: number;
    list: KubeManagerEntityState;
}
