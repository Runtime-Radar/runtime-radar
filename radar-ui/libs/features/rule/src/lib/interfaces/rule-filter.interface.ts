import { AbstractFilters } from '@cs/core';
import { RuleSeverity, RuleType } from '@cs/domains/rule';

export interface RuleFiltersPopover extends AbstractFilters {
    type: RuleType[];
    notifySeverity: RuleSeverity[];
}

export interface RuleFilters extends RuleFiltersPopover {
    name: string;
}
