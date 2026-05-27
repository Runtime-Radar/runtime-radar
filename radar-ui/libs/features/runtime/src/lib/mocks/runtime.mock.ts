import { DateTime } from 'luxon';

import { RuleSeverity } from '@cs/domains/rule';
import {
    GetRuntimeEventsResponse,
    RuntimeEvent,
    RuntimeEventCursorDirection,
    RuntimeEventProcess,
    RuntimeEventProcessorHistoryControl,
    RuntimeEventType
} from '@cs/domains/runtime';

import { RUNTIME_SETTINGS_TRACING_POLICIES_PROCESSES_KEY } from '../constants/runtime-config.constant';
import { RuntimeEventsPagination } from '../interfaces/runtime-events.interface';
import { RuntimeSettingForm } from '../interfaces/runtime-form.interface';
import { RuntimeEventContext, RuntimeEventFilters } from '../interfaces/runtime-filter.interface';

export const RUNTIME_EVENTS_CURSOR = '2025-09-18T10:00:00.000000000Z';

export const RUNTIME_SETTING_FORM: RuntimeSettingForm = {
    policies: {
        [RUNTIME_SETTINGS_TRACING_POLICIES_PROCESSES_KEY]: {
            isEnabled: false,
            name: 'name',
            description: '',
            yaml: ''
        },
        uuid1: {
            isEnabled: true,
            name: 'name',
            description: '',
            yaml: ''
        }
    },
    permissions: {
        uuid2: {
            isAllowedType: false,
            namespaces: ['namespace1'],
            pods: ['pod1'],
            labels: ['label1']
        },
        uuid3: {
            isAllowedType: true,
            namespaces: ['namespace2'],
            pods: ['pod2'],
            labels: ['label2']
        }
    },
    historyControl: RuntimeEventProcessorHistoryControl.ALL
};

export const RUNTIME_EVENTS_PAGINATION: RuntimeEventsPagination = {
    direction: RuntimeEventCursorDirection.RIGHT,
    cursor: RUNTIME_EVENTS_CURSOR
};

export const RUNTIME_EVENT_CONTEXT: RuntimeEventContext = {
    execId: '',
    parentExecId: '',
    context: undefined,
    activeContextId: undefined
};

export const RUNTIME_EVENT_FILTER: RuntimeEventFilters = {
    type: null,
    argument: '',
    binary: '',
    container: '',
    function: '',
    image: '',
    namespace: '',
    pod: '',
    period: '',
    hasThreats: false,
    hasIncident: false,
    detectors: [],
    rules: []
};

export const RUNTIME_EVENT_ID = 'eventId1';

export const RUNTIME_DATE_TIME = DateTime.fromISO('2025-09-18T12:00:00Z');

export const RUNTIME_EVENT_PROCESS: RuntimeEventProcess = {
    parent: null,
    process: {
        binary: '/bin/ls',
        arguments: '',
        exec_id: 'execId1',
        parent_exec_id: 'parentId1',
        start_time: '2025-09-18T10:00:00.000000000Z',
        pid: 1234,
        uid: 4321,
        cwd: '/',
        pod: null,
        cap: {
            permitted: [],
            effective: [],
            inheritable: []
        }
    }
};

export const RUNTIME_EVENT: RuntimeEvent = {
    id: 'id1',
    tetragon_version: 'v1',
    event: {
        node_name: 'name',
        time: '2025-09-18T10:00:00.000000000Z',
        [RuntimeEventType.EXEC]: RUNTIME_EVENT_PROCESS
    },
    threats: [],
    block_by: ['ruleId1'],
    notify_by: ['ruleId2'],
    incident_severity: RuleSeverity.MEDIUM,
    is_incident: false,
    detect_errors: []
};

export const RUNTIME_EVENT_RESPONSE: GetRuntimeEventsResponse = {
    runtime_events: [RUNTIME_EVENT],
    left_cursor: RUNTIME_EVENTS_CURSOR,
    right_cursor: RUNTIME_EVENTS_CURSOR
};

export const RUNTIME_EVENT_EMPTY_RESPONSE: GetRuntimeEventsResponse = {
    runtime_events: [],
    left_cursor: '',
    right_cursor: ''
};
