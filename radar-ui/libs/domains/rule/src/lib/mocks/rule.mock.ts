import { DateTime } from 'luxon';

import { Rule, RuleEntity, RuleScope, RuleSeverity, RuleType } from '../interfaces';

export const RULE_DATE_TIME = DateTime.fromISO('2025-09-18T12:00:00Z');

const RULE_ENTITY: RuleEntity = {
    version: '1',
    block: {
        severity: RuleSeverity.LOW,
        verdict: null
    },
    notify: {
        severity: RuleSeverity.MEDIUM,
        verdict: null,
        targets: []
    },
    whitelist: {
        threats: [],
        binaries: []
    }
};

const RULE_SCOPE: RuleScope = {
    version: '1',
    image_names: [],
    registries: [],
    clusters: [],
    namespaces: [],
    pods: [],
    containers: [],
    nodes: []
};

export const RULE: Rule = {
    id: 'id1',
    name: 'name',
    type: RuleType.TYPE_RUNTIME,
    rule: RULE_ENTITY,
    scope: RULE_SCOPE
};
