import { RuleSeverity, RuleSeverityOrder, RuleVerdict, RuleVerdictOrder } from '../interfaces';

export const RULE_SEVERITY_ORDER: RuleSeverityOrder = {
    [RuleSeverity.NONE]: 4,
    [RuleSeverity.LOW]: 3,
    [RuleSeverity.MEDIUM]: 2,
    [RuleSeverity.HIGH]: 1,
    [RuleSeverity.CRITICAL]: 0
};

export const RULE_VERDICT_ORDER: RuleVerdictOrder = {
    [RuleVerdict.NONE]: 2,
    [RuleVerdict.UNWANTED]: 1,
    [RuleVerdict.DANGEROUS]: 0
};
