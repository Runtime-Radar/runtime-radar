import { RuleSeverity } from '@cs/domains/rule';

export interface RuleForm {
    name: string;
    namespaces: string[];
    notifySeverity: RuleSeverity;
    mailIds: string[];
    detectors: string[];
    pods: string[];
    containers: string[];
    nodes: string[];
    binaries: string[];
    imageNames: string[];
    registries: string[];
}
