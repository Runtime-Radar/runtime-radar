import { DetectorExtended } from '@cs/domains/detector';
import { PermissionType } from '@cs/domains/role';
import { RuleSeverity } from '@cs/domains/rule';
import {
    RuntimeCapabilityType,
    RuntimeDetectError,
    RuntimeEventProcess,
    RuntimeEventThreat,
    RuntimeEventType
} from '@cs/domains/runtime';

export interface RuntimeSidepanelPolicyFormProps {
    isEdit: boolean;
    name: string;
    description: string;
    yaml: string;
}

export interface RuntimeSidepanelPermissionFormProps {
    isEdit: boolean;
    isAllowedType: boolean;
    namespaces: string[];
    pods: string[];
    labels: string[];
}

export interface RuntimeSidepanelCodeProps {
    content: string;
    nextHandler?: (isNext: boolean) => void;
    prevHandler?: (isNext: boolean) => void;
}

export interface RuntimeSidepanelPolicyProps {
    name: string;
    description: string;
    yaml: string;
}

export interface RuntimeSidepanelThreatsProps {
    threats: RuntimeEventThreat[];
    effectives: RuntimeCapabilityType[];
    errors: RuntimeDetectError[];
}

export interface RuntimeSidepanelIncidentProps {
    type: RuntimeEventType;
    time: string;
    severity: RuleSeverity;
    process: RuntimeEventProcess;
    threats: RuntimeEventThreat[];
    ruleIds: string[];
    permissions: Map<PermissionType, boolean>;
}

export interface RuntimeSidepanelDetectorProps {
    detector: DetectorExtended;
    permissions: Map<PermissionType, boolean>;
    deleteHandler?: (key: string, version: number) => void;
}
