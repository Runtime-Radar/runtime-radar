import { of } from 'rxjs';
import { Spy, createSpyFromClass } from 'jest-auto-spies';
import { createEnvironmentInjector, runInInjectionContext } from '@angular/core';

import { DEFAULT_PAGINATOR_PAGE_SIZE } from '@cs/shared';
import {
    GetRuntimeEventsResponse,
    RuntimeContext,
    RuntimeEventCursorDirection,
    RuntimeEventType,
    RuntimeFilterRequest,
    RuntimeRequestService
} from '@cs/domains/runtime';

import { RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR } from '../constants/runtime-filter.constant';
import { RuntimeFeatureRequestAdapterService } from './runtime-request-adapter.service';
import {
    RUNTIME_EVENTS_CURSOR,
    RUNTIME_EVENTS_PAGINATION,
    RUNTIME_EVENT_CONTEXT,
    RUNTIME_EVENT_FILTER
} from '../mocks/runtime.mock';
import { RuntimeEventContext, RuntimeEventFilters } from '../interfaces/runtime-filter.interface';

describe('RuntimeFeatureRequestAdapterService', () => {
    let runtimeRequestService: Spy<RuntimeRequestService>;
    let service: RuntimeFeatureRequestAdapterService;

    const getEventsResponse: GetRuntimeEventsResponse = {
        runtime_events: [],
        left_cursor: RUNTIME_EVENTS_CURSOR,
        right_cursor: RUNTIME_EVENTS_CURSOR
    };

    beforeEach(() => {
        runtimeRequestService = createSpyFromClass(RuntimeRequestService, {
            methodsToSpyOn: ['getEvents', 'getEventsByFilter']
        });

        const injector = createEnvironmentInjector(
            [
                {
                    provide: RuntimeRequestService,
                    useValue: runtimeRequestService
                }
            ],
            null as any
        );

        service = runInInjectionContext(injector, () => new RuntimeFeatureRequestAdapterService());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should call getEvents when filters are empty', () => {
        runtimeRequestService.getEvents.mockReturnValue(of(getEventsResponse));

        service.getEvents(
            RUNTIME_EVENTS_PAGINATION,
            RUNTIME_EVENT_FILTER,
            RUNTIME_EVENT_CONTEXT,
            DEFAULT_PAGINATOR_PAGE_SIZE
        );

        expect(runtimeRequestService.getEvents).toHaveBeenCalledWith(RuntimeEventCursorDirection.RIGHT, {
            cursor: RUNTIME_EVENTS_CURSOR,
            slice_size: DEFAULT_PAGINATOR_PAGE_SIZE
        });
    });

    it('should call getEventsByFilter with transformed request when filters are provided', () => {
        const context: RuntimeEventContext = {
            ...RUNTIME_EVENT_CONTEXT,
            execId: 'execId',
            parentExecId: 'parentExecId',
            context: RuntimeContext.CURRENT
        };
        const filters: RuntimeEventFilters = {
            ...RUNTIME_EVENT_FILTER,
            type: RuntimeEventType.EXEC,
            argument: 'argument',
            binary: 'binary',
            container: 'container',
            function: 'function',
            image: 'nginx:latest',
            namespace: 'namespace',
            pod: 'pod',
            hasThreats: true,
            hasIncident: true,
            detectors: ['detectorId'],
            rules: ['ruleId'],
            period: `2025-09-18${RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR}2025-09-19`
        };
        const filterResponse: RuntimeFilterRequest = {
            event_type: [RuntimeEventType.EXEC.toString().toLocaleUpperCase()],
            kprobe_function_name: ['function'],
            process_pod_namespace: ['namespace'],
            process_pod_name: ['pod'],
            process_pod_container_name: ['container'],
            process_pod_container_image_name: ['nginx:latest'],
            process_binary: ['binary'],
            process_arguments: ['argument'],
            process_exec_id: 'execId',
            process_parent_exec_id: '',
            node_name: [],
            has_threats: true,
            has_incident: true,
            threats_detectors: ['detectorId'],
            rules: ['ruleId'],
            period: {
                from: '2025-09-18',
                to: '2025-09-19'
            }
        };
        runtimeRequestService.getEventsByFilter.mockReturnValue(of(getEventsResponse));

        service.getEvents(RUNTIME_EVENTS_PAGINATION, filters, context, DEFAULT_PAGINATOR_PAGE_SIZE);

        expect(runtimeRequestService.getEventsByFilter).toHaveBeenCalledWith(RuntimeEventCursorDirection.RIGHT, {
            cursor: RUNTIME_EVENTS_CURSOR,
            slice_size: DEFAULT_PAGINATOR_PAGE_SIZE,
            filter: filterResponse
        });
    });
});
