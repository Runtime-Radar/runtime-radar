import { Injectable, inject } from '@angular/core';
import { Observable, filter, map, switchMap, take } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    CreateRuleRequest,
    CreateRuleResponse,
    EmptyRuleResponse,
    GetRuleResponse,
    GetRulesResponse,
    Rule,
    UpdateRuleRequest
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class RuleRequestService {
    private readonly apiService = inject(ApiService);

    getRules(): Observable<Rule[]> {
        return this.apiService
            .get<ApiEmptyRequest, GetRulesResponse>('rule/page/1?page_size=1000')
            .pipe(map((response) => response.rules));
    }

    /** @external */
    getRule(id: string): Observable<GetRuleResponse> {
        return this.apiService.get<ApiEmptyRequest, GetRuleResponse>(`rule/${id}`);
    }

    createRule(request: CreateRuleRequest): Observable<Rule> {
        return this.apiService.post<CreateRuleRequest, CreateRuleResponse>('rule', request).pipe(
            map((response) => response.id),
            filter((id) => !!id),
            switchMap((id) =>
                this.getRule(id).pipe(
                    take(1),
                    map((response) => response.rule)
                )
            )
        );
    }

    updateRule(id: string, request: UpdateRuleRequest): Observable<Rule> {
        return this.apiService.patch<UpdateRuleRequest, EmptyRuleResponse>(`rule/${id}`, request).pipe(
            filter((response) => response && !Object.keys(response).length),
            switchMap(() =>
                this.getRule(id).pipe(
                    take(1),
                    map((response) => response.rule)
                )
            )
        );
    }

    deleteRule(id: string): Observable<string> {
        return this.apiService
            .delete<EmptyRuleResponse>(`rule/${id}`)
            .pipe(map((response) => (response && !Object.keys(response).length ? id : '')));
    }
}
