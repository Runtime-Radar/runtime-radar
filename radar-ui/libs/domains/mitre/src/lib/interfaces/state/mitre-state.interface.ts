import { EntityState } from '@ngrx/entity';

import { LoadStatus } from '@cs/core';

import { MitreTactic } from '../contract/mitre-contract.interface';

export type MitreEntityState = EntityState<MitreTactic>;

export type MitreListState = {
    [key: string]: MitreEntityState;
};

export interface MitreState {
    loadStatus: LoadStatus;
    lastUpdate: number;
    list: MitreListState;
}
