import { Observable } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import { DEFAULT_PAGINATOR_PAGE_SIZE } from '@cs/shared';
import {
    GetRuntimeEventsRequest,
    GetRuntimeEventsResponse,
    RuntimeContext,
    RuntimeFilterRequest,
    RuntimeRequestService
} from '@cs/domains/runtime';

import { RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR } from '../constants/runtime-filter.constant';
import { RuntimeEventsPagination } from '../interfaces/runtime-events.interface';
import { RuntimeFeatureConfigUtilsService as runtimeConfigUtils } from './runtime-utils.service';
import { RuntimeEventContext, RuntimeEventFilters } from '../interfaces/runtime-filter.interface';

@Injectable({
    providedIn: 'root'
})
export class RuntimeFeatureRequestAdapterService {
    private readonly runtimeRequestService = inject(RuntimeRequestService);

    getEvents(
        pagination: RuntimeEventsPagination,
        filters: RuntimeEventFilters,
        context: RuntimeEventContext,
        pageSize = DEFAULT_PAGINATOR_PAGE_SIZE
    ): Observable<GetRuntimeEventsResponse> {
        const hasValues = !!Object.values({ ...filters, ...context }).filter((value) =>
            runtimeConfigUtils.isEventFilterValueValid(value)
        ).length;
        const request: GetRuntimeEventsRequest = {
            cursor: pagination.cursor,
            slice_size: pageSize
        };

        if (!hasValues) {
            return this.runtimeRequestService.getEvents(pagination.direction, request);
        }

        const [from, to] = filters.period ? filters.period.split(RUNTIME_FILTER_DATETIME_PERIOD_SEPARATOR) : [];
        const contextRequest = this.getContextRequest(context);

        const filterRequest: RuntimeFilterRequest = {
            event_type: filters.type ? [filters.type.toString().toLocaleUpperCase()] : [],
            kprobe_function_name: filters.function ? [filters.function] : [],
            process_pod_namespace: filters.namespace ? [filters.namespace] : [],
            process_pod_name: filters.pod ? [filters.pod] : [],
            process_pod_container_name: filters.container ? [filters.container] : [],
            process_pod_container_image_name: filters.image ? [filters.image] : [],
            process_binary: filters.binary ? [filters.binary] : [],
            process_arguments: filters.argument ? [filters.argument] : [],
            process_exec_id: contextRequest.execId,
            process_parent_exec_id: contextRequest.parentExecId,
            has_threats: filters.hasThreats || undefined,
            has_incident: filters.hasIncident || undefined,
            threats_detectors: filters.hasThreats ? filters.detectors : [],
            rules: filters.hasIncident ? filters.rules : [],
            node_name: [],
            period: {
                from: from || null,
                to: to || null
            }
        };

        return this.runtimeRequestService.getEventsByFilter(pagination.direction, {
            ...request,
            filter: filterRequest
        });
    }

    private getContextRequest(values: RuntimeEventContext): RuntimeEventContext {
        switch (values.context) {
            case RuntimeContext.PARENT:
                return {
                    context: values.context,
                    execId: values.parentExecId,
                    parentExecId: ''
                };
                break;
            case RuntimeContext.SIBLING:
                return {
                    context: values.context,
                    execId: '',
                    parentExecId: values.parentExecId
                };
                break;
            case RuntimeContext.CHILDREN:
                return {
                    context: values.context,
                    execId: '',
                    parentExecId: values.execId
                };
                break;
            case RuntimeContext.CURRENT:
                return {
                    context: values.context,
                    execId: values.execId,
                    parentExecId: ''
                };
                break;
            default:
                return {
                    context: undefined,
                    execId: '',
                    parentExecId: ''
                };
                break;
        }
    }
}
