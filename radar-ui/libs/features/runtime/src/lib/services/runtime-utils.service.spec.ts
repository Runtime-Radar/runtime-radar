import { RuntimeMonitorConfig } from '@cs/domains/runtime';

import { RUNTIME_SETTING_FORM } from '../mocks/runtime.mock';
import { RuntimeFeatureConfigUtilsService } from './runtime-utils.service';

describe('RuntimeFeatureConfigUtilsService', () => {
    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('isEventFilterValueValid', () => {
        it('should return false for empty values', () => {
            expect(RuntimeFeatureConfigUtilsService.isEventFilterValueValid([])).toBe(false);
            expect(RuntimeFeatureConfigUtilsService.isEventFilterValueValid('')).toBe(false);
            expect(RuntimeFeatureConfigUtilsService.isEventFilterValueValid(null)).toBe(false);
        });

        it('should return true for non-empty values', () => {
            expect(RuntimeFeatureConfigUtilsService.isEventFilterValueValid(['key'])).toBe(true);
            expect(RuntimeFeatureConfigUtilsService.isEventFilterValueValid('key')).toBe(true);
        });
    });

    describe('convertSettingFormToMonitorConfig', () => {
        it('should return correct config', () => {
            const result: RuntimeMonitorConfig = {
                version: '1',
                tracing_policies: {
                    uuid1: {
                        name: 'name',
                        enabled: true,
                        description: undefined,
                        yaml: undefined
                    }
                },
                allow_list: [
                    {
                        namespace: ['namespace2'],
                        pod_regex: ['pod2'],
                        labels: ['label2']
                    }
                ],
                deny_list: [
                    {
                        namespace: ['namespace1'],
                        pod_regex: ['pod1'],
                        labels: ['label1']
                    }
                ]
            };

            expect(RuntimeFeatureConfigUtilsService.convertSettingFormToMonitorConfig(RUNTIME_SETTING_FORM)).toEqual(
                result
            );
        });
    });
});
