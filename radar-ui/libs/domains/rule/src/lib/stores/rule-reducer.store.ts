import { createEntityAdapter } from '@ngrx/entity';
import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import { Rule } from '../interfaces';
import { RuleState } from '../interfaces/state/rule-state.interface';
import {
    DELETE_RULE_ENTITY_DOC_ACTION,
    SET_ALL_RULE_ENTITIES_DOC_ACTION,
    SET_RULE_ENTITY_DOC_ACTION,
    UPDATE_RULE_ENTITY_DOC_ACTION,
    UPDATE_RULE_STATE_DOC_ACTION
} from './rule-action.store';

export const adapter = createEntityAdapter<Rule>();

const INITIAL_STATE: RuleState = {
    loadStatus: LoadStatus.INIT,
    lastUpdate: 0,
    list: adapter.getInitialState()
};

const reducer: ActionReducer<RuleState, Action> = createReducer(
    INITIAL_STATE,
    on(UPDATE_RULE_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values })),
    on(SET_ALL_RULE_ENTITIES_DOC_ACTION, (state, { list }) => ({
        ...state,
        list: adapter.setAll(list, state.list)
    })),
    on(SET_RULE_ENTITY_DOC_ACTION, (state, { item }) => ({
        ...state,
        list: adapter.setOne(item, state.list)
    })),
    on(UPDATE_RULE_ENTITY_DOC_ACTION, (state, { item }) => ({
        ...state,
        list: adapter.updateOne(item, state.list)
    })),
    on(DELETE_RULE_ENTITY_DOC_ACTION, (state, { id }) => ({
        ...state,
        list: adapter.removeOne(id, state.list)
    }))
);

export const ruleEntitySelector = adapter.getSelectors();

export function ruleReducer(state: RuleState | undefined, action: Action): RuleState {
    return reducer(state, action);
}
