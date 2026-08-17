import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { HttpErrorResponse } from '@angular/common/http';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, of } from 'rxjs';
import { catchError, filter, map, switchMap, take, tap } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';

import { Rule } from '../interfaces';
import { RuleRequestService } from '../services/rule-request.service';
import { RuleState } from '../interfaces/state/rule-state.interface';
import { getRuleLoadStatus } from './rule-selector.store';
import {
    CREATE_RULE_ENTITY_TODO_ACTION,
    DELETE_RULE_ENTITY_DOC_ACTION,
    DELETE_RULE_ENTITY_TODO_ACTION,
    LOAD_RULE_ENTITIES_TODO_ACTION,
    POLLING_LOAD_RULE_ENTITIES_TODO_ACTION,
    SET_ALL_RULE_ENTITIES_DOC_ACTION,
    SET_RULE_ENTITY_DOC_ACTION,
    UPDATE_RULE_ENTITY_DOC_ACTION,
    UPDATE_RULE_ENTITY_TODO_ACTION,
    UPDATE_RULE_STATE_DOC_ACTION
} from './rule-action.store';

@Injectable({
    providedIn: 'root'
})
export class RuleEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly i18nService = inject(I18nService);
    private readonly ruleRequestService = inject(RuleRequestService);
    private readonly store = inject<Store<RuleState>>(Store);

    readonly loadRules$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_RULE_ENTITIES_TODO_ACTION),
            switchMap(() =>
                this.ruleRequestService.getRules().pipe(
                    take(1),
                    catchError(() => of(undefined))
                )
            ),
            switchMap((list) => {
                if (list === undefined) {
                    return [
                        UPDATE_RULE_STATE_DOC_ACTION({
                            loadStatus: LoadStatus.ERROR
                        })
                    ];
                }

                return [
                    SET_ALL_RULE_ENTITIES_DOC_ACTION({ list }),
                    UPDATE_RULE_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];
            })
        )
    );

    readonly reloadRules$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.store.select(getRuleLoadStatus).pipe(take(1))),
            filter((status) => status !== LoadStatus.INIT),
            map(() => LOAD_RULE_ENTITIES_TODO_ACTION())
        )
    );

    readonly pollingLoadRules$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_RULE_ENTITIES_TODO_ACTION),
            switchMap(() => this.ruleRequestService.getRules().pipe(take(1))),
            switchMap((list) => [
                SET_ALL_RULE_ENTITIES_DOC_ACTION({ list }),
                UPDATE_RULE_STATE_DOC_ACTION({ lastUpdate: this.dateAdapter.today().toMillis() })
            ])
        )
    );

    readonly createRule$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_RULE_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.ruleRequestService.createRule(action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.NAME_MUST_BE_UNIQUE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Rule.Pseudo.Notification.NameMustBeUnique')
                            });
                        }

                        return of({} as Rule);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) => SET_RULE_ENTITY_DOC_ACTION({ item })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Rule.Pseudo.Notification.Created')
                });
            })
        )
    );

    readonly updateRule$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(UPDATE_RULE_ENTITY_TODO_ACTION),
            switchMap((action) =>
                this.ruleRequestService.updateRule(action.id, action.item).pipe(
                    take(1),
                    catchError((error: HttpErrorResponse) => {
                        if (apiUtils.getReasonCode(error) === ApiErrorCode.NAME_MUST_BE_UNIQUE) {
                            this.toastService.show({
                                style: KbqToastStyle.Warning,
                                title: this.i18nService.translate('Rule.Pseudo.Notification.NameMustBeUnique')
                            });
                        }

                        return of({} as Rule);
                    })
                )
            ),
            filter((item) => !!item.id),
            map((item) =>
                UPDATE_RULE_ENTITY_DOC_ACTION({
                    item: {
                        id: item.id,
                        changes: item
                    }
                })
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Rule.Pseudo.Notification.Updated')
                });
            })
        )
    );

    readonly deleteRule$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_RULE_ENTITY_TODO_ACTION),
            switchMap((item) => this.ruleRequestService.deleteRule(item.id).pipe(take(1))),
            filter((id) => !!id),
            map((id) => DELETE_RULE_ENTITY_DOC_ACTION({ id })),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Rule.Pseudo.Notification.Deleted')
                });
            })
        )
    );
}
