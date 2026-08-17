import { AbstractFilters } from '@cs/core';
import { Rule } from '@cs/domains/rule';
import { RuntimeContext, RuntimeEventType } from '@cs/domains/runtime';

export enum RuntimeEventFilterKey {
    NAMESPACE = 'namespace',
    POD = 'pod',
    BINARY = 'binary',
    ARGUMENT = 'argument',
    TYPE = 'type',
    FUNCTION = 'function',
    CONTAINER = 'container',
    IMAGE = 'image',
    PERIOD = 'period'
}

export enum RuntimeEventDateTimePeriod {
    ONE_MINUTE = 'ONE_MINUTE',
    TEN_MINUTES = 'TEN_MINUTES',
    ONE_HOUR = 'ONE_HOUR',
    ONE_DAY = 'ONE_DAY',
    CUSTOM = 'CUSTOM'
}

export interface RuntimeEventFilters extends AbstractFilters {
    type: RuntimeEventType | null;
    argument: string;
    binary: string;
    container: string;
    function: string;
    image: string;
    namespace: string;
    pod: string;
    period: string; // RFC3339
    hasThreats: boolean;
    hasIncident: boolean;
    detectors: string[];
    rules: string[];
}

export interface RuntimeEventContext extends AbstractFilters {
    activeContextId?: string;
    context?: RuntimeContext;
    execId: string;
    parentExecId: string;
}

export enum RuntimeEventFilterContextDropdownType {
    ADD,
    REMOVE
}

export interface RuntimeEventFilterContextDropdown {
    type: RuntimeEventFilterContextDropdownType;
    key: RuntimeEventFilterKey;
    value: string;
}

export interface RuntimeEventFilterRuleNode extends Pick<Rule, 'id' | 'name'> {
    isExtra: boolean;
}

export interface RuntimeEventDateTimePeriodOption {
    id: RuntimeEventDateTimePeriod;
    localizationKey: string;
}
