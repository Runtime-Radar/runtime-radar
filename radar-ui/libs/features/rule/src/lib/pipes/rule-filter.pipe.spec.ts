import { RULE, Rule, RuleType } from '@cs/domains/rule';

import { RuleFeatureFilterPipe } from './rule-filter.pipe';
import { RuleFilters } from '../interfaces/rule-filter.interface';

describe('RuleFeatureFilterPipe', () => {
    let pipe: RuleFeatureFilterPipe;

    beforeEach(() => {
        pipe = new RuleFeatureFilterPipe();
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should return empty rules', () => {
        expect(pipe.transform(null, undefined)).toEqual([]);
        expect(pipe.transform(undefined, undefined)).toEqual([]);
    });

    it('should return rules which are filtered by provided arguments', () => {
        const filters: RuleFilters = {
            name: 'name',
            type: [RuleType.TYPE_RUNTIME],
            notifySeverity: []
        };

        const rules: Rule[] = [
            {
                ...RULE,
                id: 'id1',
                name: 'name'
            },
            {
                ...RULE,
                id: 'id2',
                name: 'query'
            }
        ];

        expect(pipe.transform(rules, filters)).toEqual([RULE]);
    });
});
