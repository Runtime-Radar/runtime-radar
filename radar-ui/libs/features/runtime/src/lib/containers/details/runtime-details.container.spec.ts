import { DateAdapter } from '@koobiq/components/core';
import { DateFormatter } from '@koobiq/date-formatter';
import { KbqSidepanelService } from '@koobiq/components/sidepanel';
import { KbqToastService } from '@koobiq/components/toast';
import { provideAutoSpy } from 'jest-auto-spies';
import { render } from '@testing-library/angular';
import { subscribeSpyTo } from '@hirez_io/observer-spy';
import { ActivatedRoute, Router } from '@angular/router';
import {
    DefaultTranspiler,
    TRANSLOCO_FALLBACK_STRATEGY,
    TRANSLOCO_INTERCEPTOR,
    TRANSLOCO_MISSING_HANDLER,
    TRANSLOCO_TRANSPILER,
    TranslocoModule
} from '@jsverse/transloco';
import { of, throwError } from 'rxjs';

import { I18nLocale, I18nService } from '@cs/i18n';
import { LoadStatus, RouterName } from '@cs/core';
import { RuntimeContext, RuntimeEventType, RuntimeRequestService } from '@cs/domains/runtime';

import { RuntimeFeatureDetailsContainer } from './runtime-details.container';
import { RuntimeFeatureRequestAdapterService } from '../../services/runtime-request-adapter.service';
import { RuntimeRouterName } from '../../interfaces/runtime-navigation.interface';
import {
    RUNTIME_DATE_TIME,
    RUNTIME_EVENT,
    RUNTIME_EVENT_EMPTY_RESPONSE,
    RUNTIME_EVENT_ID,
    RUNTIME_EVENT_PROCESS,
    RUNTIME_EVENT_RESPONSE
} from '../../mocks/runtime.mock';

// @todo: move this provider into a common mock.
const translocoProviders = [
    { provide: TRANSLOCO_MISSING_HANDLER, useValue: { handle: () => '' } },
    { provide: TRANSLOCO_FALLBACK_STRATEGY, useValue: { getNextLang: () => I18nLocale.EN } },
    { provide: TRANSLOCO_TRANSPILER, useClass: DefaultTranspiler },
    {
        provide: TRANSLOCO_INTERCEPTOR,
        useValue: { preProcess: (v) => v, postProcess: (v) => v, preSaveTranslation: (v) => v }
    }
];

const dateProviders = [{ provide: DateAdapter, useValue: { today: () => RUNTIME_DATE_TIME } }];

const activatedRouteProviders = [
    {
        provide: ActivatedRoute,
        useValue: { params: of({ eventId: RUNTIME_EVENT_ID }), snapshot: { data: { permissions: {} } } }
    }
];

describe('RuntimeFeatureDetailsContainer', () => {
    let instance: RuntimeFeatureDetailsContainer;
    let runtimeFeatureRequestAdapterService: jest.Mocked<RuntimeFeatureRequestAdapterService>;
    let runtimeRequestService: jest.Mocked<RuntimeRequestService>;
    let router: jest.Mocked<Router>;
    let toastService: jest.Mocked<KbqToastService>;

    beforeEach(async () => {
        const { fixture } = await render(RuntimeFeatureDetailsContainer, {
            imports: [TranslocoModule],
            providers: [
                ...translocoProviders,
                ...dateProviders,
                ...activatedRouteProviders,
                provideAutoSpy(RuntimeRequestService, { methodsToSpyOn: ['getEvent'] }),
                provideAutoSpy(RuntimeFeatureRequestAdapterService, { methodsToSpyOn: ['getEvents'] }),
                provideAutoSpy(I18nService),
                provideAutoSpy(KbqToastService),
                provideAutoSpy(KbqSidepanelService),
                provideAutoSpy(Router),
                provideAutoSpy(DateFormatter)
            ]
        });

        instance = fixture.componentInstance;
        runtimeRequestService = fixture.debugElement.injector.get(
            RuntimeRequestService
        ) as jest.Mocked<RuntimeRequestService>;
        runtimeRequestService.getEvent.mockReturnValue(of(RUNTIME_EVENT));
        runtimeFeatureRequestAdapterService = fixture.debugElement.injector.get(
            RuntimeFeatureRequestAdapterService
        ) as jest.Mocked<RuntimeFeatureRequestAdapterService>;
        runtimeFeatureRequestAdapterService.getEvents.mockReturnValue(of(RUNTIME_EVENT_RESPONSE));
        router = fixture.debugElement.injector.get(Router) as jest.Mocked<Router>;
        toastService = fixture.debugElement.injector.get(KbqToastService) as jest.Mocked<KbqToastService>;
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('event$', () => {
        it('should set ruleIds', () => {
            runtimeRequestService.getEvent(RUNTIME_EVENT_ID);

            subscribeSpyTo(instance.event$).getLastValue();

            expect(instance.ruleIds).toEqual(['ruleId1', 'ruleId2']);
        });

        it('should set eventType', () => {
            runtimeRequestService.getEvent(RUNTIME_EVENT_ID);

            subscribeSpyTo(instance.event$).getLastValue();

            expect(instance.eventType).toEqual(RuntimeEventType.EXEC);
        });

        it('should set eventEntity', () => {
            runtimeRequestService.getEvent(RUNTIME_EVENT_ID);

            subscribeSpyTo(instance.event$).getLastValue();

            expect(instance.eventEntity).toEqual(RUNTIME_EVENT_PROCESS);
        });

        it('should set load status', () => {
            subscribeSpyTo(instance.event$);

            expect(instance.cardLoadStatus$.value).toEqual(LoadStatus.LOADED);
        });
    });

    describe('eventsResponse$', () => {
        it('should return events', () => {
            const result = subscribeSpyTo(instance.eventsResponse$).getLastValue();

            expect(result).toEqual({
                ...RUNTIME_EVENT_RESPONSE,
                isPrevResponse: false
            });
        });

        it('should show contrast toast', () => {
            instance['eventsResponseBuffer'] = RUNTIME_EVENT_RESPONSE;
            runtimeFeatureRequestAdapterService.getEvents.mockReturnValueOnce(of(RUNTIME_EVENT_EMPTY_RESPONSE));

            subscribeSpyTo(instance.eventsResponse$).getLastValue();

            expect(toastService.show).toHaveBeenCalled();
        });

        it('should return empty response', () => {
            runtimeFeatureRequestAdapterService.getEvents.mockReturnValueOnce(throwError(() => new Error('fail')));

            const result = subscribeSpyTo(instance.eventsResponse$).getLastValue();

            expect(result).toEqual({
                ...RUNTIME_EVENT_EMPTY_RESPONSE,
                isPrevResponse: true
            });
        });
    });

    describe('goToListPageWithContext', () => {
        it('should navigate to runtime event page', () => {
            instance['eventEntity'] = RUNTIME_EVENT_PROCESS;

            instance.goToListPageWithContext(RuntimeContext.SIBLING);

            expect(router.navigate).toHaveBeenCalledWith([RouterName.RUNTIME, RuntimeRouterName.EVENTS], {
                queryParams: {
                    context: RuntimeContext.SIBLING,
                    execId: RUNTIME_EVENT_PROCESS.process.exec_id,
                    parentExecId: RUNTIME_EVENT_PROCESS.process.parent_exec_id
                }
            });
        });
    });

    describe('goToListPage', () => {
        it('should navigate to runtime list page', () => {
            instance.goToListPage();

            expect(router.navigate).toHaveBeenCalledWith([RouterName.RUNTIME, RuntimeRouterName.EVENTS], {
                queryParams: {
                    context: RuntimeContext.SIBLING,
                    execId: '',
                    parentExecId: ''
                }
            });
        });
    });
});
