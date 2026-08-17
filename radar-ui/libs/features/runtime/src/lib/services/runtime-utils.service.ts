import { Injectable } from '@angular/core';

import { RuntimeMonitorConfig, RuntimeMonitorTracingPolicies } from '@cs/domains/runtime';

import { RUNTIME_SETTINGS_TRACING_POLICIES_PROCESSES_KEY } from '../constants/runtime-config.constant';
import { RuntimeEventFilterEntity } from '../interfaces/runtime-state.interface';
import { RuntimeSettingForm } from '../interfaces/runtime-form.interface';

@Injectable({
    providedIn: 'root'
})
export class RuntimeFeatureConfigUtilsService {
    static isEventFilterValueValid(values: RuntimeEventFilterEntity[keyof RuntimeEventFilterEntity]): boolean {
        return (Array.isArray(values) && !!values.length) || (!Array.isArray(values) && !!values);
    }

    static convertSettingFormToMonitorConfig(formValues: RuntimeSettingForm): RuntimeMonitorConfig {
        const permissionsValues = Object.values(formValues.permissions);
        const tracingPoliciesValues = Object.entries(formValues.policies)
            .filter(([key, _]) => key !== RUNTIME_SETTINGS_TRACING_POLICIES_PROCESSES_KEY)
            .reduce<RuntimeMonitorTracingPolicies>((acc, [key, value]) => {
                acc[key] = {
                    name: value.name,
                    enabled: value.isEnabled,
                    description: value.description || undefined,
                    yaml: value.yaml || undefined
                };

                return acc;
            }, {});

        return {
            version: '1', // @todo: create environment constant
            tracing_policies: tracingPoliciesValues,
            allow_list: permissionsValues
                .filter((item) => item.isAllowedType)
                .map((item) => ({
                    namespace: item.namespaces,
                    pod_regex: item.pods,
                    labels: item.labels
                })),
            deny_list: permissionsValues
                .filter((item) => !item.isAllowedType)
                .map((item) => ({
                    namespace: item.namespaces,
                    pod_regex: item.pods,
                    labels: item.labels
                }))
        };
    }
}
