import { RuleForm } from '@cs/packages/rule';
import { RuleSeverity } from '@cs/domains/rule';

export const RULE_FORM: RuleForm = {
    name: 'name',
    namespaces: [],
    notifySeverity: RuleSeverity.MEDIUM,
    blockSeverity: RuleSeverity.LOW,
    mailIds: ['mail@example.com'],
    detectors: ['detector1'],
    pods: [],
    containers: [],
    nodes: [],
    binaries: ['binary1'],
    imageNames: ['image1'],
    registries: ['reg1']
};
