import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { provideAutoSpy } from 'jest-auto-spies';
import { provideMockActions } from '@ngrx/effects/testing';
import { render } from '@testing-library/angular';
import { Action, Store } from '@ngrx/store';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { ReplaySubject, of } from 'rxjs';

import { I18nService } from '@cs/i18n';
import { CoreWindowService, LoadStatus } from '@cs/core';

import { RuntimeEffectStore } from './runtime-effect.store';
import { RuntimeRequestService } from '../services/runtime-request.service';
import {
    CHECK_RUNTIME_CHANGES_TODO_ACTION,
    CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION,
    CREATE_RUNTIME_CONFIG_TODO_ACTION,
    GET_RUNTIME_CONFIG_STATUS_TODO_ACTION,
    LOAD_RUNTIME_CONFIG_TODO_ACTION,
    RESET_RUNTIME_CONFIG_TODO_ACTION,
    SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION,
    UPDATE_RUNTIME_STATE_DOC_ACTION
} from './runtime-action.store';
import { RUNTIME_EVENT_PROCESSOR, RUNTIME_MONITOR_CONFIG } from '../mocks/runtime.mock';
import {
    RuntimeConfigStatus,
    RuntimeEventProcessorHistoryControl,
    RuntimeMonitorConfig,
    RuntimeState
} from '../interfaces';

const dateProviders = [
    {
        provide: DateAdapter,
        useValue: {
            today: () => DateTime.fromISO('2025-09-18T12:00:00Z')
        }
    }
];

describe('RuntimeEffectStore', () => {
    let effects: RuntimeEffectStore;
    let actions$: ReplaySubject<Action>;
    let i18nService: jest.Mocked<I18nService>;
    let coreWindowService: jest.Mocked<CoreWindowService>;
    let runtimeRequestService: jest.Mocked<RuntimeRequestService>;
    let store: jest.Mocked<Store<RuntimeState>>;
    let toastService: jest.Mocked<KbqToastService>;

    beforeEach(async () => {
        actions$ = new ReplaySubject(0);

        coreWindowService = {
            localStorage: {
                getItem: jest.fn(() => 'true'),
                setItem: jest.fn()
            } as unknown as Storage
        } as jest.Mocked<CoreWindowService>;
        const coreWindowProviders = [
            {
                provide: CoreWindowService,
                useValue: coreWindowService
            }
        ];
        const { fixture } = await render('<div></div>', {
            providers: [
                ...dateProviders,
                ...coreWindowProviders,
                RuntimeEffectStore,
                provideMockActions(() => actions$),
                provideAutoSpy(I18nService),
                provideAutoSpy(RuntimeRequestService),
                provideAutoSpy(Store<RuntimeState>),
                provideAutoSpy(KbqToastService)
            ]
        });

        effects = fixture.debugElement.injector.get(RuntimeEffectStore);
        runtimeRequestService = fixture.debugElement.injector.get(
            RuntimeRequestService
        ) as jest.Mocked<RuntimeRequestService>;
        i18nService = fixture.debugElement.injector.get(I18nService) as jest.Mocked<I18nService>;
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<RuntimeState>>;
        toastService = fixture.debugElement.injector.get(KbqToastService) as jest.Mocked<KbqToastService>;

        i18nService.translate.mockReturnValue('message');
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('checkExpertMode$', () => {
        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION with correct value', (done) => {
            effects.checkExpertMode$.subscribe((action) => {
                expect(action).toEqual(
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        isExpertMode: true
                    })
                );
                done();
            });

            actions$.next(LOAD_RUNTIME_CONFIG_TODO_ACTION());
        });
    });

    describe('loadConfig$', () => {
        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION with correct config and historyControl', (done) => {
            runtimeRequestService.getEventProcessor.mockReturnValue(of(RUNTIME_EVENT_PROCESSOR));
            runtimeRequestService.getRuntimeMonitor.mockReturnValue(
                of({
                    id: 'rm1',
                    config: RUNTIME_MONITOR_CONFIG
                })
            );

            effects.loadConfig$.subscribe((action) => {
                expect(action).toEqual(
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        config: RUNTIME_MONITOR_CONFIG,
                        historyControl: RuntimeEventProcessorHistoryControl.ALL
                    })
                );
                done();
            });

            actions$.next(LOAD_RUNTIME_CONFIG_TODO_ACTION());
        });
    });

    describe('resetConfig$', () => {
        beforeEach(() => {
            runtimeRequestService.resetConfigToDefault.mockReturnValue(of(true));
            runtimeRequestService.getRuntimeMonitor.mockReturnValue(
                of({
                    id: 'rm1',
                    config: RUNTIME_MONITOR_CONFIG
                })
            );
        });

        it('should call getRuntimeMonitor', () => {
            const actions: Action[] = [];
            effects.resetConfig$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(RESET_RUNTIME_CONFIG_TODO_ACTION());

            expect(runtimeRequestService.getRuntimeMonitor).toHaveBeenCalled();
        });

        it('should show success toast', () => {
            const actions: Action[] = [];
            effects.resetConfig$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(RESET_RUNTIME_CONFIG_TODO_ACTION());

            expect(toastService.show).toHaveBeenCalledWith({
                style: KbqToastStyle.Success,
                title: 'message'
            });
        });

        it('should show warning toast', () => {
            const actions: Action[] = [];
            runtimeRequestService.resetConfigToDefault.mockReturnValue(of(false));
            effects.resetConfig$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(RESET_RUNTIME_CONFIG_TODO_ACTION());

            expect(toastService.show).toHaveBeenCalledWith({
                style: KbqToastStyle.Warning,
                title: 'message'
            });
        });

        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION and GET_RUNTIME_CONFIG_STATUS_TODO_ACTION', (done) => {
            const actions: Action[] = [];
            effects.resetConfig$.subscribe((action) => {
                actions.push(action);
                expect(actions).toEqual([
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        config: RUNTIME_MONITOR_CONFIG,
                        hasChanges: false,
                        hasPoliciesChanges: false,
                        configStatus: RuntimeConfigStatus.STAY
                    }),
                    GET_RUNTIME_CONFIG_STATUS_TODO_ACTION()
                ]);
                done();
            });

            actions$.next(RESET_RUNTIME_CONFIG_TODO_ACTION());
        });
    });

    describe('createRuntimeMonitor$', () => {
        beforeEach(() => {
            const config: RuntimeMonitorConfig = {
                ...RUNTIME_MONITOR_CONFIG,
                tracing_policies: {
                    uuid1: {
                        name: 'name',
                        enabled: true
                    }
                }
            };
            store.select.mockReturnValue(of(config));
            runtimeRequestService.createRuntimeMonitor.mockReturnValue(
                of({
                    id: 'rm1',
                    config: RUNTIME_MONITOR_CONFIG
                })
            );
        });

        it('should call createRuntimeMonitor with correct arguments', () => {
            const actions: Action[] = [];
            effects.createRuntimeMonitor$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(
                CREATE_RUNTIME_CONFIG_TODO_ACTION({
                    config: RUNTIME_MONITOR_CONFIG,
                    historyControl: RuntimeEventProcessorHistoryControl.ALL
                })
            );

            expect(runtimeRequestService.createRuntimeMonitor).toHaveBeenCalledWith(RUNTIME_MONITOR_CONFIG);
        });

        it('should not call createRuntimeMonitor', () => {
            const actions: Action[] = [];
            store.select.mockReturnValue(of(RUNTIME_MONITOR_CONFIG));
            effects.createRuntimeMonitor$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(
                CREATE_RUNTIME_CONFIG_TODO_ACTION({
                    config: RUNTIME_MONITOR_CONFIG,
                    historyControl: RuntimeEventProcessorHistoryControl.ALL
                })
            );

            expect(runtimeRequestService.createRuntimeMonitor).not.toHaveBeenCalled();
        });

        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION', (done) => {
            const actions: Action[] = [];
            effects.createRuntimeMonitor$.subscribe((action) => {
                actions.push(action);
                expect(actions).toEqual([
                    GET_RUNTIME_CONFIG_STATUS_TODO_ACTION(),
                    CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION(),
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        config: RUNTIME_MONITOR_CONFIG,
                        configStatus: RuntimeConfigStatus.STAY
                    })
                ]);
                done();
            });

            actions$.next(
                CREATE_RUNTIME_CONFIG_TODO_ACTION({
                    config: RUNTIME_MONITOR_CONFIG,
                    historyControl: RuntimeEventProcessorHistoryControl.ALL
                })
            );
        });
    });

    describe('createEventProcessor', () => {
        beforeEach(() => {
            store.select.mockReturnValue(of(RuntimeEventProcessorHistoryControl.NONE));
            runtimeRequestService.createEventProcessor.mockReturnValue(of(RUNTIME_EVENT_PROCESSOR));
        });

        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION', (done) => {
            const actions: Action[] = [];
            effects.createEventProcessor$.subscribe((action) => {
                actions.push(action);
                expect(actions).toEqual([
                    CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION(),
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        historyControl: RuntimeEventProcessorHistoryControl.ALL
                    })
                ]);
                done();
            });

            actions$.next(
                CREATE_RUNTIME_CONFIG_TODO_ACTION({
                    config: RUNTIME_MONITOR_CONFIG,
                    historyControl: RuntimeEventProcessorHistoryControl.ALL
                })
            );
        });
    });

    describe('checkChanges$', () => {
        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION with MODIFY status when config was modified', (done) => {
            store.select.mockReturnValueOnce(of(RuntimeEventProcessorHistoryControl.ALL));
            store.select.mockReturnValueOnce(of(RUNTIME_MONITOR_CONFIG));

            effects.checkChanges$.subscribe((action) => {
                expect(action).toEqual(
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        hasChanges: true,
                        hasPoliciesChanges: false,
                        configStatus: RuntimeConfigStatus.MODIFY
                    })
                );
                done();
            });

            actions$.next(
                CHECK_RUNTIME_CHANGES_TODO_ACTION({
                    config: {
                        ...RUNTIME_MONITOR_CONFIG,
                        historyControl: RuntimeEventProcessorHistoryControl.NONE
                    }
                })
            );
        });
    });

    describe('switchExpertMode$', () => {
        beforeEach(() => {
            store.select.mockReturnValue(of(true));
        });

        it('should call localStorage.setItem', (done) => {
            effects.switchExpertMode$.subscribe(() => {
                expect(coreWindowService.localStorage.setItem).toHaveBeenCalledWith('xprtmd', 'false');
                done();
            });

            actions$.next(SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION());
        });

        it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION', (done) => {
            effects.switchExpertMode$.subscribe((action) => {
                expect(action).toEqual(UPDATE_RUNTIME_STATE_DOC_ACTION({ isExpertMode: false }));
                done();
            });

            actions$.next(SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION());
        });
    });
});
