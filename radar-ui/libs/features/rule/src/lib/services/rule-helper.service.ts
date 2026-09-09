import { Injectable } from '@angular/core';

import { RuleForm } from '@cs/packages/rule';
import { RuleBlockEntity, RuleNotifyEntity, RuleSeverity, RuleWhiteList } from '@cs/domains/rule';

@Injectable({
    providedIn: 'root'
})
export class RuleFeatureHelperService {
    static convertFormValuesToBlockEntity(form: RuleForm): RuleBlockEntity | null {
        if (form.blockSeverity === RuleSeverity.NONE) {
            return null;
        }

        return {
            severity: form.blockSeverity,
            verdict: null
        };
    }

    static convertFormValuesToNotifyEntity(form: RuleForm): RuleNotifyEntity | null {
        if (form.notifySeverity === RuleSeverity.NONE) {
            return null;
        }

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
