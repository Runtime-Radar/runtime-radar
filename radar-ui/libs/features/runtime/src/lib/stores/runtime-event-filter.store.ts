import { ComponentStore } from '@ngrx/component-store';
import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Observable } from 'rxjs';
import { ActivatedRoute, Params } from '@angular/router';
import { Injectable, inject } from '@angular/core';

import { RuntimeContext } from '@cs/domains/runtime';
import { CoreWindowService, CoreUtilsService as utils } from '@cs/core';

import { RuntimeEventFilterEntity } from '../interfaces/runtime-state.interface';
import { RuntimeFeatureConfigUtilsService as runtimeConfigUtils } from '../services/runtime-utils.service';
import {
    RUNTIME_FILTER_INITIAL_CONTEXT_STATE,
    RUNTIME_FILTER_INITIAL_STATE
} from '../constants/runtime-filter.constant';
import { RuntimeEventContext, RuntimeEventFilters } from '../interfaces/runtime-filter.interface';

const RUNTIME_FILTER_STATE_ENTITIES_LIMIT = 50;

const RUNTIME_FILTER_HISTORY_LIMIT = 10;

const RUNTIME_FILTER_HISTORY_SESSION_KEY = 'rntmfltrhstr';

type RuntimeEventContextWithId = RuntimeEventContext & Pick<RuntimeEventFilterEntity, 'id'>;

@Injectable()
export class RuntimeFeatureEventFilterComponentStore extends ComponentStore<RuntimeEventFilterEntity[]> {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly route = inject(ActivatedRoute);

    private readonly coreWindowService = inject(CoreWindowService);

    readonly history$: Observable<RuntimeEventFilterEntity[]> = this.select((entities) =>
        this.getPurifiedEntities(entities).slice(0, RUNTIME_FILTER_HISTORY_LIMIT)
    );

    readonly filters$: Observable<RuntimeEventFilters> = this.select(
        (entities) =>
            utils.omit<RuntimeEventFilterEntity, keyof RuntimeEventContextWithId>(entities[0], [
                'id',
                'activeContextId',
                'context',
                'execId',
                'parentExecId'
            ]) as RuntimeEventFilters
    );

    readonly context$: Observable<RuntimeEventContext> = this.select((entities) => {
        if (!entities[0]) {
            return {} as RuntimeEventContext;
        }
        const { activeContextId, context, execId, parentExecId } = entities[0];

        return { activeContextId, context, execId, parentExecId };
    });

    readonly queryParams$: Observable<Params> = this.select((entities) =>
        Object.entries(
            utils.omit<RuntimeEventFilterEntity, keyof Pick<RuntimeEventFilterEntity, 'id'>>(entities[0], ['id'])
        ).reduce<Params>((acc, [key, value]) => (value ? { ...acc, [key]: value } : acc), {})
    );

    readonly update = this.updater(
        (entities: RuntimeEventFilterEntity[], values: Partial<RuntimeEventFilterEntity>) => {
            entities.unshift({
                ...entities[0],
                ...values,
                id: this.dateAdapter.today().toJSDate().toISOString() // RFC3339
            });

            if (entities.length > RUNTIME_FILTER_STATE_ENTITIES_LIMIT) {
                entities.pop();
            }

            this.coreWindowService.sessionStorage.setItem(RUNTIME_FILTER_HISTORY_SESSION_KEY, JSON.stringify(entities));

            return entities;
        }
    );

    constructor() {
        super([]);

        const state = this.getHistoryFromStorage();
        const queryParams = this.getRouteQueryParams();
        const hasValues = Object.values(
            utils.omit<RuntimeEventFilterEntity, keyof Pick<RuntimeEventFilterEntity, 'id'>>(queryParams, ['id'])
        ).some((value) => runtimeConfigUtils.isEventFilterValueValid(value));
        if (hasValues) {
            state.unshift(queryParams);
        }

        this.setState(state);
    }

    setContextIds(execId: string, parentExecId: string, context?: RuntimeContext, activeContextId?: string) {
        this.update({ activeContextId, context, execId, parentExecId });
    }

    private getPurifiedEntities(entities: RuntimeEventFilterEntity[]): RuntimeEventFilterEntity[] {
        return entities.filter((curr, index) => {
            const currFilters = utils.omit<RuntimeEventFilterEntity, keyof RuntimeEventContextWithId>(curr, [
                'id',
                'activeContextId',
                'context',
                'execId',
                'parentExecId'
            ]);
            const currValue = JSON.stringify(currFilters);

            const isEmpty = Object.values(currFilters).some((value) =>
                runtimeConfigUtils.isEventFilterValueValid(value)
            );
            const isDuplicate =
                index ===
                entities.findIndex((prev) => {
                    const prevFilters = utils.omit<RuntimeEventFilterEntity, keyof RuntimeEventContextWithId>(prev, [
                        'id',
                        'activeContextId',
                        'context',
                        'execId',
                        'parentExecId'
                    ]);

                    return JSON.stringify(prevFilters) === currValue;
                });

            return isEmpty && isDuplicate;
        });
    }

    private getHistoryFromStorage(): RuntimeEventFilterEntity[] {
        const storage: string | null = this.coreWindowService.sessionStorage.getItem(
            RUNTIME_FILTER_HISTORY_SESSION_KEY
        );
        const state: RuntimeEventFilterEntity[] | null = storage ? JSON.parse(storage) : null;
        const isValid = Array.isArray(state) && state.every((item) => typeof item === 'object');

        return isValid ? state : [];
    }

    private getRouteQueryParams(): RuntimeEventFilterEntity {
        const filterKeys = Object.keys({ ...RUNTIME_FILTER_INITIAL_STATE, ...RUNTIME_FILTER_INITIAL_CONTEXT_STATE });
        const queryParams = Object.entries(this.route.snapshot.queryParams).reduce<Params>((acc, [key, value]) => {
            if (filterKeys.includes(key) && runtimeConfigUtils.isEventFilterValueValid(value as string)) {
                let data = value;
                switch (key) {
                    case 'context':
                        data = Number(value) || undefined;
                        break;
                    case 'hasThreats':
                        data = value === 'true';
                        break;
                    case 'hasIncident':
                        data = value === 'true';
                        break;
                    case 'detectors':
                    case 'rules':
                        data = String(value).split(',');
                        break;
                }

                return { ...acc, [key]: data };
            }

            return acc;
        }, {});

        return {
            ...RUNTIME_FILTER_INITIAL_CONTEXT_STATE,
            ...RUNTIME_FILTER_INITIAL_STATE,
            ...queryParams,
            id: this.dateAdapter.today().toJSDate().toISOString() // RFC3339
        };
    }
}
