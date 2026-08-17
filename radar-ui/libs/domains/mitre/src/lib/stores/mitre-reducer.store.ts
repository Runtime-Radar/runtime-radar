import { createEntityAdapter } from '@ngrx/entity';
import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import { MitreListState, MitreState, MitreTactic } from '../interfaces';
import { SET_ALL_MITRE_ENTITIES_DOC_ACTION, UPDATE_MITRE_STATE_DOC_ACTION } from './mitre-action.store';

export const adapter = createEntityAdapter<MitreTactic>();

function getInitialState(availableLocales: string[]): MitreState {
    return {
        loadStatus: LoadStatus.INIT,
        lastUpdate: 0,
        list: availableLocales.reduce<MitreListState>(
            (acc, locale) => ({
                ...acc,
                [locale]: adapter.getInitialState()
            }),
            {}
        )
    };
}

export const mitreEntitySelector = adapter.getSelectors();

export function mitreReducerFactory(availableLocales: string[]) {
    const reducer: ActionReducer<MitreState, Action> = createReducer(
        getInitialState(availableLocales),
        on(UPDATE_MITRE_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values })),
        on(SET_ALL_MITRE_ENTITIES_DOC_ACTION, (state, { list, locale }) => ({
            ...state,
            list: {
                ...state.list,
                [locale]: adapter.setAll(list, state.list[locale])
            }
        }))
    );

    return function (state: MitreState | undefined, action: Action): MitreState {
        return reducer(state, action);
    };
}
