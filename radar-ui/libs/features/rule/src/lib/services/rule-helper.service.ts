import { Injectable } from '@angular/core';

import { RuleForm } from '@cs/packages/rule';
import { RuleNotifyEntity, RuleWhiteList } from '@cs/domains/rule';

@Injectable({
    providedIn: 'root'
})
export class RuleFeatureHelperService {
    static convertFormValuesToNotifyEntity(form: RuleForm): RuleNotifyEntity | null {
        return {
            severity: form.notifySeverity,
            verdict: null,
            targets: form.mailIds
        };
    }

    static convertWhiteListToRequestNode(form: RuleForm): RuleWhiteList {
        const node: RuleWhiteList = {
            threats: [],
            binaries: []
        };

        node.threats.push(...form.detectors);
        node.binaries.push(...form.binaries);

        return node;
    }
}
