import { Observable } from 'rxjs';
import { Store } from '@ngrx/store';
import { Injectable, inject } from '@angular/core';

import { LoadStatus } from '@cs/core';

import { RuleState } from '../interfaces/state/rule-state.interface';
import {
    CREATE_RULE_ENTITY_TODO_ACTION,
    DELETE_RULE_ENTITY_TODO_ACTION,
    UPDATE_RULE_ENTITY_TODO_ACTION
} from '../stores/rule-action.store';
import { CreateRuleRequest, Rule, UpdateRuleRequest } from '../interfaces';
import { getRule, getRuleLoadStatus, getRules, getRulesByNotificationId } from '../stores/rule-selector.store';

@Injectable({
    providedIn: 'root'
})
export class RuleStoreService {
    private readonly store = inject<Store<RuleState>>(Store);

    readonly rules$: Observable<Rule[]> = this.store.select(getRules);

    readonly rulesByNotificationId$ = (id: string): Observable<Rule[]> =>
        this.store.select(getRulesByNotificationId(id));

    readonly rule$ = (id: string): Observable<Rule | undefined> => this.store.select(getRule(id));

    readonly loadStatus$: Observable<LoadStatus> = this.store.select(getRuleLoadStatus);

    createRule(item: CreateRuleRequest) {
        this.store.dispatch(CREATE_RULE_ENTITY_TODO_ACTION({ item }));
    }

    updateRule(id: string, item: UpdateRuleRequest) {
        this.store.dispatch(UPDATE_RULE_ENTITY_TODO_ACTION({ id, item }));
    }

    deleteRule(id: string) {
        this.store.dispatch(DELETE_RULE_ENTITY_TODO_ACTION({ id }));
    }
}
