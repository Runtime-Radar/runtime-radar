import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { Rule } from '../interfaces/contract/rule-contract.interface';
import { RuleEntityState, RuleState } from '../interfaces/state/rule-state.interface';
import { ruleEntitySelector, ruleReducer } from './rule-reducer.store';

export const RULE_DOMAIN_KEY = 'rule';

export interface RuleDomainState {
    readonly domain: RuleState;
}

const selectRuleDomainState = createFeatureSelector<RuleDomainState>(RULE_DOMAIN_KEY);
const selectRuleState = createSelector(selectRuleDomainState, (state: RuleDomainState) => state.domain);
const selectRuleEntityState = createSelector(selectRuleState, (state: RuleState) => state.list);

export const getRuleLoadStatus = createSelector(selectRuleState, (state: RuleState) => state.loadStatus);

export const getRuleLastUpdate = createSelector(selectRuleState, (state: RuleState) => state.lastUpdate);

export const getRules = createSelector(selectRuleEntityState, (state: RuleEntityState) =>
    ruleEntitySelector.selectAll(state)
);

export const getRule = (id: string) =>
    createSelector(selectRuleEntityState, (state: RuleEntityState) => state.entities[id]);

/* eslint @typescript-eslint/no-unnecessary-type-assertion: "off" */
export const getRulesByNotificationId = (notificationId: string) =>
    createSelector(
        selectRuleEntityState,
        (state: RuleEntityState) =>
            Object.values(state.entities).filter(
                (item) => item !== undefined && item.rule.notify?.targets.includes(notificationId)
            ) as Rule[]
    );

export const ruleDomainReducer: ActionReducerMap<RuleDomainState> = {
    domain: ruleReducer
};
