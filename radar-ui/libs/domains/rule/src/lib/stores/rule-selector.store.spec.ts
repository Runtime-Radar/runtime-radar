import { RULE } from '../mocks/rule.mock';
import { RuleEntityState } from '../interfaces/state/rule-state.interface';
import { adapter } from './rule-reducer.store';
import { getRulesByNotificationId } from './rule-selector.store';
import { Rule, RuleEntity, RuleSeverity } from '../interfaces';

const RULE_ENTITY_ONE: RuleEntity = {
    version: '1',
    notify: {
        targets: ['notificationId1'],
        severity: RuleSeverity.LOW,
        verdict: null
    },
    whitelist: {
        threats: [],
        binaries: []
    }
};
const RULE_ENTITY_TWO: RuleEntity = {
    version: '1',
    notify: {
        targets: ['notificationId2'],
        severity: RuleSeverity.MEDIUM,
        verdict: null
    },
    whitelist: {
        threats: [],
        binaries: []
    }
};

describe('RuleDomainReducer', () => {
    let entityState: RuleEntityState = adapter.getInitialState();
    const rulesWithNotify: Rule[] = [
        {
            ...RULE,
            id: 'id2',
            rule: RULE_ENTITY_ONE
        },
        {
            ...RULE,
            id: 'id3',
            rule: RULE_ENTITY_TWO
        }
    ];

    beforeEach(() => {
        entityState = adapter.setAll([...rulesWithNotify, RULE], adapter.getInitialState());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('getRulesByNotificationId', () => {
        it('should return rules filtered by notificationId', () => {
            const result = getRulesByNotificationId('notificationId1').projector(entityState);

            expect(result).toEqual([rulesWithNotify.at(0)]);
        });
    });
});
