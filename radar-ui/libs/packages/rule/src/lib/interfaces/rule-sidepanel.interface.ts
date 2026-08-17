import { Observable } from 'rxjs';

import { DetectorExtended } from '@cs/domains/detector';
import { Notification } from '@cs/domains/notification';
import { PermissionType } from '@cs/domains/role';
import { Rule } from '@cs/domains/rule';

export interface RuleSidepanelFormProps {
    rule: Partial<Rule>;
    isEdit: boolean;
}

export interface RuleSidepanelInfoProps {
    rule$: Observable<Rule | undefined>;
    detectors$: Observable<DetectorExtended[]>;
    notifications$: Observable<Notification[]>;
    permissions?: Map<PermissionType, boolean>;
    isDeleted: boolean;
    updateHandler?: (rule: Rule) => void;
    deleteHandler?: (id: string) => void;
}
