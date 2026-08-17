import { KbqAlertColors } from '@koobiq/components/alert';
import { KbqBadgeColors } from '@koobiq/components/badge';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, map } from 'rxjs';

import { PermissionType } from '@cs/domains/role';
import { DetectorExtended, DetectorType } from '@cs/domains/detector';
import { Rule, RuleWhiteList } from '@cs/domains/rule';

import { RuleSidepanelInfoProps } from '../../interfaces/rule-sidepanel.interface';

@Component({
    templateUrl: './rule-sidepanel-info.component.html',
    styleUrl: './rule-sidepanel-info.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RulePackageSidepanelInfoComponent {
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    readonly props = inject<RuleSidepanelInfoProps>(KBQ_SIDEPANEL_DATA);

    readonly alertColors = KbqAlertColors;

    readonly badgeColors = KbqBadgeColors;

    readonly permissionType = PermissionType;

    readonly detectors$ = (whiteList: RuleWhiteList): Observable<DetectorExtended[]> =>
        this.props.detectors$.pipe(
            map((detectors) => {
                const { binaries, ...rest } = whiteList;
                const keys = Object.values(rest).flat();

                return keys.map((key) => {
                    const detector = detectors.find((item) => item.key === key);
                    const defaultVersion = 1;
                    const emptyDetector: DetectorExtended = {
                        id: `${key}${defaultVersion}`,
                        key,
                        name: '',
                        description: '',
                        type: DetectorType.RUNTIME,
                        version: defaultVersion,
                        tactics_covered: []
                    };

                    return detector || emptyDetector;
                });
            })
        );

    updateRule(rule: Rule) {
        if (this.props.updateHandler) {
            this.props.updateHandler(rule);
        }
    }

    deleteRule(id: string) {
        if (this.props.deleteHandler) {
            this.props.deleteHandler(id);
            this.sidepanelRef.close(undefined);
        }
    }
}
