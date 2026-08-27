import {
    RuleCombinedSeverityVerdict,
    RuleCombinedSeverityVerdictOption,
    RuleSeverity,
    RuleSeverityOption,
    RuleType,
    RuleTypeOption,
    RuleVerdict,
    RuleVerdictOption
} from '../interfaces';

export const RULE_SEVERITIES: RuleSeverityOption[] = [
    {
        id: RuleSeverity.LOW,
        localizationKey: 'Common.Pseudo.Severity.Low',
        testId: 'low-radio'
    },
    {
        id: RuleSeverity.MEDIUM,
        localizationKey: 'Common.Pseudo.Severity.Medium',
        testId: 'medium-radio'
    },
    {
        id: RuleSeverity.HIGH,
        localizationKey: 'Common.Pseudo.Severity.High',
        testId: 'high-radio'
    },
    {
        id: RuleSeverity.CRITICAL,
        localizationKey: 'Common.Pseudo.Severity.Critical',
        testId: 'critical-radio'
    },
    {
        id: RuleSeverity.NONE,
        localizationKey: 'Common.Pseudo.Severity.None',
        testId: 'none-radio'
    }
];

export const RULE_VERDICTS: RuleVerdictOption[] = [
    {
        id: RuleVerdict.UNWANTED,
        localizationKey: 'Common.Pseudo.Verdict.Unwanted',
        testId: 'unwanted-radio'
    },
    {
        id: RuleVerdict.DANGEROUS,
        localizationKey: 'Common.Pseudo.Verdict.Dangerous',
        testId: 'dangerous-radio'
    },
    {
        id: RuleVerdict.NONE,
        localizationKey: 'Common.Pseudo.Verdict.None',
        testId: 'none-radio'
    }
];

export const RULE_COMBINED_SEVERITIES_VERDICTS: RuleCombinedSeverityVerdictOption[] = [
    {
        id: RuleCombinedSeverityVerdict.LOW,
        localizationKey: 'Common.Pseudo.Severity.Low',
        testId: 'low-radio'
    },
    {
        id: RuleCombinedSeverityVerdict.MEDIUM,
        localizationKey: 'Common.Pseudo.Severity.Medium',
        testId: 'medium-radio'
    },
    {
        id: RuleCombinedSeverityVerdict.HIGH,
        localizationKey: 'Common.Pseudo.Severity.High',
        testId: 'high-radio'
    },
    {
        id: RuleCombinedSeverityVerdict.CRITICAL,
        localizationKey: 'Common.Pseudo.Severity.Critical',
        testId: 'critical-radio'
    },
    {
        id: RuleCombinedSeverityVerdict.UNWANTED,
        localizationKey: 'Common.Pseudo.Verdict.Unwanted',
        testId: 'unwanted-radio'
    },
    {
        id: RuleCombinedSeverityVerdict.DANGEROUS,
        localizationKey: 'Common.Pseudo.Verdict.Dangerous',
        testId: 'dangerous-radio'
    },
    {
        id: RuleCombinedSeverityVerdict.NONE,
        localizationKey: 'Common.Pseudo.Severity.None',
        testId: 'none-radio'
    }
];

export const RULE_TYPE: RuleTypeOption[] = [
    {
        id: RuleType.TYPE_RUNTIME,
        localizationKey: 'Common.Pseudo.ScanType.Runtime'
    }
];
