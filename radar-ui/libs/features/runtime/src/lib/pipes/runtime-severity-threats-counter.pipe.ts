import { Pipe, PipeTransform } from '@angular/core';

import { RuleSeverity } from '@cs/domains/rule';
import { RuntimeEventThreat } from '@cs/domains/runtime';

@Pipe({
    name: 'runtimeSeverityThreatsCounter',
    standalone: false
})
export class RuntimeFeatureSeverityThreatsCounterPipe implements PipeTransform {
    transform(threats?: RuntimeEventThreat[], type?: RuleSeverity): number {
        if (!threats || !type) {
            return 0;
        }

        return threats.filter((item) => item.severity === type).length;
    }
}
