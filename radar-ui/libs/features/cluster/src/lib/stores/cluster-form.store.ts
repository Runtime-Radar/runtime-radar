import { ComponentStore } from '@ngrx/component-store';
import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Injectable } from '@angular/core';
import { Observable, distinctUntilChanged } from 'rxjs';

import { CoreWindowService, CoreUtilsService as utils } from '@cs/core';

import { ClusterFormState } from '../interfaces/cluster-form-state.interface';
import { ClusterStepName } from '../interfaces/cluster-stepper.interface';
import {
    CLUSTER_ACCESS_FORM_INITIAL_STATE,
    CLUSTER_DATABASE_FORM_INITIAL_STATE,
    CLUSTER_INGRESS_FORM_INITIAL_STATE,
    CLUSTER_METRIC_FORM_INITIAL_STATE,
    CLUSTER_RABBIT_FORM_INITIAL_STATE,
    CLUSTER_REDIS_FORM_INITIAL_STATE,
    CLUSTER_REGISTRY_FORM_INITIAL_STATE
} from '../constants/cluster-form.constant';

const CLUSTER_FORM_SESSION_KEY = 'clstrfrm';

const CLUSTER_FORM_INITIAL_STATE: Omit<ClusterFormState, 'id'> = {
    step: ClusterStepName.REGISTRY,
    registry: CLUSTER_REGISTRY_FORM_INITIAL_STATE,
    clickhouse: CLUSTER_DATABASE_FORM_INITIAL_STATE,
    postgres: CLUSTER_DATABASE_FORM_INITIAL_STATE,
    redis: CLUSTER_REDIS_FORM_INITIAL_STATE,
    rabbit: CLUSTER_RABBIT_FORM_INITIAL_STATE,
    metric: CLUSTER_METRIC_FORM_INITIAL_STATE,
    ingress: CLUSTER_INGRESS_FORM_INITIAL_STATE,
    access: CLUSTER_ACCESS_FORM_INITIAL_STATE
};

@Injectable()
export class ClusterFeatureFormComponentStore extends ComponentStore<ClusterFormState> {
    form: ClusterFormState | null = null;

    readonly form$: Observable<ClusterFormState> = this.select((state) => state);

    readonly partialForm$ = <T>(key: keyof Omit<ClusterFormState, 'id'>): Observable<T> =>
        this.select((state) => state[key] as T).pipe(distinctUntilChanged((a, b) => utils.isEqual(a, b)));

    readonly update = this.updater((state: ClusterFormState, values: Partial<ClusterFormState>) => {
        const object = { ...state, ...values };
        if (!utils.isEqual(state, object)) {
            this.coreWindowService.sessionStorage.setItem(CLUSTER_FORM_SESSION_KEY, JSON.stringify(object));
            this.form = object;
        }

        return object;
    });

    constructor(
        private readonly dateAdapter: DateAdapter<DateTime>,
        private readonly coreWindowService: CoreWindowService
    ) {
        super();

        const initialState = this.getInitialState();
        this.form = initialState;
        this.setState(initialState);
    }

    resetState() {
        this.update({
            ...CLUSTER_FORM_INITIAL_STATE,
            id: this.dateAdapter.today().toMillis()
        });
    }

    private getInitialState(): ClusterFormState {
        return (
            this.getFormStateFromStorage() || {
                ...CLUSTER_FORM_INITIAL_STATE,
                id: this.dateAdapter.today().toMillis()
            }
        );
    }

    private getFormStateFromStorage(): ClusterFormState | undefined {
        const storage: string | null = this.coreWindowService.sessionStorage.getItem(CLUSTER_FORM_SESSION_KEY);
        const values: ClusterFormState | undefined = storage ? JSON.parse(storage) : undefined;
        if (!values) {
            return undefined;
        }

        const originKeys = Object.keys(CLUSTER_FORM_INITIAL_STATE);
        const isValid = Object.keys(values)
            .filter((key) => key !== 'id')
            .every((key) => originKeys.includes(key));
        if (!isValid) {
            return undefined;
        }

        return values;
    }
}
