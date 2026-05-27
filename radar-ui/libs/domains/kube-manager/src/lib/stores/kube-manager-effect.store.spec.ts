import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { provideAutoSpy } from 'jest-auto-spies';
import { provideMockActions } from '@ngrx/effects/testing';
import { render } from '@testing-library/angular';
import { Action, Store } from '@ngrx/store';
import { ReplaySubject, of, throwError } from 'rxjs';

import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';

import { KUBE_MANAGER_PODS } from '../mocks/kube-manager.mock';
import { KubeManagerEffectStore } from './kube-manager-effect.store';
import { KubeManagerRequestService } from '../services/kube-manager-request.service';
import { GetKubeManagerPodsRequest, KubeManagerState } from '../interfaces';
import {
    LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION,
    LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION,
    SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION,
    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION
} from './kube-manager-action.store';

const dateTime = DateTime.fromISO('2025-09-18T12:00:00Z');
const dateProviders = [
    {
        provide: DateAdapter,
        useValue: {
            today: () => dateTime
        }
    }
];

describe('KubeManagerEffectStore', () => {
    let effects: KubeManagerEffectStore;
    let actions$: ReplaySubject<Action>;
    let kubeManagerRequestService: jest.Mocked<KubeManagerRequestService>;
    let store: jest.Mocked<Store<KubeManagerState>>;

    beforeEach(async () => {
        actions$ = new ReplaySubject(0);

        const { fixture } = await render('<div></div>', {
            providers: [
                ...dateProviders,
                KubeManagerEffectStore,
                provideMockActions(() => actions$),
                provideAutoSpy(KubeManagerRequestService),
                provideAutoSpy(Store<KubeManagerState>)
            ]
        });

        effects = fixture.debugElement.injector.get(KubeManagerEffectStore);
        kubeManagerRequestService = fixture.debugElement.injector.get(
            KubeManagerRequestService
        ) as jest.Mocked<KubeManagerRequestService>;
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<KubeManagerState>>;
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('loadPods$', () => {
        it('should call getPods', () => {
            const actions: Action[] = [];
            effects.loadPods$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION());

            expect(kubeManagerRequestService.getPods).toHaveBeenCalled();
        });

        it('should emit SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION when pods are loaded', (done) => {
            kubeManagerRequestService.getPods.mockReturnValue(of(KUBE_MANAGER_PODS));

            const actions: Action[] = [];
            effects.loadPods$.subscribe((action) => {
                actions.push(action);
                expect(actions).toEqual([
                    SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION({ list: KUBE_MANAGER_PODS }),
                    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: dateTime.toMillis()
                    })
                ]);
                done();
            });

            actions$.next(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION());
        });

        it('should emit UPDATE_KUBE_MANAGER_STATE_DOC_ACTION when pods are failed', (done) => {
            kubeManagerRequestService.getPods.mockReturnValue(throwError(() => new Error('fail')));

            effects.loadPods$.subscribe((action) => {
                expect(action).toEqual(
                    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.ERROR
                    })
                );
                done();
            });

            actions$.next(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION());
        });
    });

    describe('reloadPods$', () => {
        it('should emit LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION', (done) => {
            store.select.mockReturnValueOnce(of(LoadStatus.LOADED));

            effects.reloadPods$.subscribe((action) => {
                expect(action).toEqual(LOAD_KUBE_MANAGER_ENTITIES_TODO_ACTION());
                done();
            });

            actions$.next(SWITCH_CLUSTER_EVENT_ACTION());
        });
    });

    describe('loadPodsWithFilters$', () => {
        const filters: Partial<GetKubeManagerPodsRequest> = {
            pods: ['pod1', 'pod2'],
            nodes: ['node1'],
            namespaces: ['namespace1', 'namespace2'],
            containers: ['container1']
        };

        beforeEach(() => {
            kubeManagerRequestService.getPods.mockReturnValue(of(KUBE_MANAGER_PODS));
        });

        it('should call getPods with correct arguments', () => {
            const actions: Action[] = [];
            effects.loadPodsWithFilters$.subscribe((action) => {
                actions.push(action);
            });

            actions$.next(LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION({ filters }));

            expect(kubeManagerRequestService.getPods).toHaveBeenCalledWith(filters);
        });

        it('should emit SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION', (done) => {
            const actions: Action[] = [];
            effects.loadPodsWithFilters$.subscribe((action) => {
                actions.push(action);
                expect(actions).toEqual([
                    SET_ALL_KUBE_MANAGER_ENTITIES_DOC_ACTION({ list: KUBE_MANAGER_PODS }),
                    UPDATE_KUBE_MANAGER_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED
                    })
                ]);
                done();
            });

            actions$.next(LOAD_KUBE_MANAGER_ENTITIES_WITH_FILTERS_TODO_ACTION({ filters }));
        });
    });
});
