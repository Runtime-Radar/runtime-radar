import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { provideAutoSpy } from 'jest-auto-spies';
import { provideMockActions } from '@ngrx/effects/testing';
import { render } from '@testing-library/angular';
import { Action, Store } from '@ngrx/store';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { ReplaySubject, of, throwError } from 'rxjs';

import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { ApiErrorCode, ApiUtilsService as apiUtils } from '@cs/api';

import { RuleEffectStore } from './rule-effect.store';
import { RuleRequestService } from '../services/rule-request.service';
import { RuleState } from '../interfaces/state/rule-state.interface';
import {
    CREATE_RULE_ENTITY_TODO_ACTION,
    DELETE_RULE_ENTITY_DOC_ACTION,
    DELETE_RULE_ENTITY_TODO_ACTION,
    LOAD_RULE_ENTITIES_TODO_ACTION,
    SET_ALL_RULE_ENTITIES_DOC_ACTION,
    SET_RULE_ENTITY_DOC_ACTION,
    UPDATE_RULE_ENTITY_DOC_ACTION,
    UPDATE_RULE_ENTITY_TODO_ACTION,
    UPDATE_RULE_STATE_DOC_ACTION
} from './rule-action.store';
import { RULE, RULE_DATE_TIME } from '../mocks/rule.mock';

const dateProviders = [
    {
        provide: DateAdapter,
        useValue: {
            today: () => DateTime.fromISO('2025-09-18T12:00:00Z')
        }
    }
];

describe('RuleEffectStore', () => {
    let actions$: ReplaySubject<Action>;
    let effects: RuleEffectStore;
    let i18nService: jest.Mocked<I18nService>;
    let ruleRequestService: jest.Mocked<RuleRequestService>;
    let store: jest.Mocked<Store<RuleState>>;
    let toastService: jest.Mocked<KbqToastService>;

    beforeEach(async () => {
        actions$ = new ReplaySubject(0);

        const { fixture } = await render('<div></div>', {
            providers: [
                ...dateProviders,
                RuleEffectStore,
                provideMockActions(() => actions$),
                provideAutoSpy(I18nService),
                provideAutoSpy(RuleRequestService),
                provideAutoSpy(Store<RuleState>),
                provideAutoSpy(KbqToastService)
            ]
        });

        effects = fixture.debugElement.injector.get(RuleEffectStore);
        i18nService = fixture.debugElement.injector.get(I18nService) as jest.Mocked<I18nService>;
        i18nService.translate.mockReturnValue('message');
        ruleRequestService = fixture.debugElement.injector.get(RuleRequestService) as jest.Mocked<RuleRequestService>;
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<RuleState>>;
        toastService = fixture.debugElement.injector.get(KbqToastService) as jest.Mocked<KbqToastService>;
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('loadRules$', () => {
        it('should call getRules', () => {
            const actions: Action[] = [];
            ruleRequestService.getRules.mockReturnValue(of([RULE]));
            effects.loadRules$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(LOAD_RULE_ENTITIES_TODO_ACTION());

            expect(ruleRequestService.getRules).toHaveBeenCalled();
        });

        it('should emit SET_ALL_RULE_ENTITIES_DOC_ACTION and UPDATE_RULE_STATE_DOC_ACTION when rules are loaded', (done) => {
            const actions: Action[] = [];
            ruleRequestService.getRules.mockReturnValue(of([RULE]));
            effects.loadRules$.subscribe((action) => {
                actions.push(action);
                expect(actions).toEqual([
                    SET_ALL_RULE_ENTITIES_DOC_ACTION({ list: [RULE] }),
                    UPDATE_RULE_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: RULE_DATE_TIME.toMillis()
                    })
                ]);
                done();
            });

            actions$.next(LOAD_RULE_ENTITIES_TODO_ACTION());
        });

        it('should emit UPDATE_RULE_STATE_DOC_ACTION when rules are failed', (done) => {
            ruleRequestService.getRules.mockReturnValue(throwError(() => new Error('fail')));

            effects.loadRules$.subscribe((action) => {
                expect(action).toEqual(
                    UPDATE_RULE_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.ERROR
                    })
                );
                done();
            });

            actions$.next(LOAD_RULE_ENTITIES_TODO_ACTION());
        });
    });

    describe('reloadRules$', () => {
        it('should emit LOAD_RULE_ENTITIES_TODO_ACTION', (done) => {
            store.select.mockReturnValue(of(LoadStatus.LOADED));

            effects.reloadRules$.subscribe((action) => {
                expect(action).toEqual(LOAD_RULE_ENTITIES_TODO_ACTION());
                done();
            });

            actions$.next(SWITCH_CLUSTER_EVENT_ACTION());
        });
    });

    describe('createRule$', () => {
        beforeEach(() => {
            ruleRequestService.createRule.mockReturnValue(of(RULE));
        });

        it('should call createRule with correct arguments', (done) => {
            effects.createRule$.subscribe(() => {
                expect(ruleRequestService.createRule).toHaveBeenCalledWith(RULE);
                done();
            });

            actions$.next(CREATE_RULE_ENTITY_TODO_ACTION({ item: RULE }));
        });

        it('should emit SET_RULE_ENTITY_DOC_ACTION', (done) => {
            effects.createRule$.subscribe((action) => {
                expect(action).toEqual(SET_RULE_ENTITY_DOC_ACTION({ item: RULE }));
                done();
            });

            actions$.next(CREATE_RULE_ENTITY_TODO_ACTION({ item: RULE }));
        });

        it('should show success toast', (done) => {
            effects.createRule$.subscribe(() => {
                expect(toastService.show).toHaveBeenCalled();
                done();
            });

            actions$.next(CREATE_RULE_ENTITY_TODO_ACTION({ item: RULE }));
        });

        it('should show warning toast with NAME_MUST_BE_UNIQUE error', () => {
            const actions: Action[] = [];
            ruleRequestService.createRule.mockReturnValue(throwError(() => new Error('fail')));
            jest.spyOn(apiUtils, 'getReasonCode').mockReturnValue(ApiErrorCode.NAME_MUST_BE_UNIQUE);
            effects.createRule$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(CREATE_RULE_ENTITY_TODO_ACTION({ item: RULE }));

            expect(actions).toEqual([]);
            expect(toastService.show).toHaveBeenCalledWith({
                style: KbqToastStyle.Warning,
                title: 'message'
            });
        });
    });

    describe('updateRule$', () => {
        beforeEach(() => {
            ruleRequestService.updateRule.mockReturnValue(
                of({
                    ...RULE,
                    name: 'updated'
                })
            );
        });

        it('should call updateRule with correct arguments', (done) => {
            effects.updateRule$.subscribe(() => {
                expect(ruleRequestService.updateRule).toHaveBeenCalledWith(RULE.id, RULE);
                done();
            });

            actions$.next(UPDATE_RULE_ENTITY_TODO_ACTION({ id: RULE.id, item: RULE }));
        });

        it('should emit UPDATE_RULE_ENTITY_DOC_ACTION', (done) => {
            effects.updateRule$.subscribe((action) => {
                expect(action).toEqual(
                    UPDATE_RULE_ENTITY_DOC_ACTION({
                        item: {
                            id: RULE.id,
                            changes: { ...RULE, name: 'updated' }
                        }
                    })
                );
                done();
            });

            actions$.next(UPDATE_RULE_ENTITY_TODO_ACTION({ id: RULE.id, item: RULE }));
        });

        it('should show success toast', (done) => {
            effects.updateRule$.subscribe(() => {
                expect(toastService.show).toHaveBeenCalled();
                done();
            });

            actions$.next(UPDATE_RULE_ENTITY_TODO_ACTION({ id: RULE.id, item: RULE }));
        });
    });

    describe('deleteRule$', () => {
        const ruleId = 'id2';

        beforeEach(() => {
            ruleRequestService.deleteRule.mockReturnValue(of(ruleId));
        });

        it('should call deleteRule with correct arguments', (done) => {
            effects.deleteRule$.subscribe(() => {
                expect(ruleRequestService.deleteRule).toHaveBeenCalledWith(ruleId);
                done();
            });

            actions$.next(DELETE_RULE_ENTITY_TODO_ACTION({ id: ruleId }));
        });

        it('should emit DELETE_RULE_ENTITY_DOC_ACTION', (done) => {
            effects.deleteRule$.subscribe((action) => {
                expect(action).toEqual(DELETE_RULE_ENTITY_DOC_ACTION({ id: ruleId }));
                done();
            });

            actions$.next(DELETE_RULE_ENTITY_TODO_ACTION({ id: ruleId }));
        });

        it('should show success toast', (done) => {
            effects.deleteRule$.subscribe(() => {
                expect(toastService.show).toHaveBeenCalled();
                done();
            });

            actions$.next(DELETE_RULE_ENTITY_TODO_ACTION({ id: ruleId }));
        });
    });
});
