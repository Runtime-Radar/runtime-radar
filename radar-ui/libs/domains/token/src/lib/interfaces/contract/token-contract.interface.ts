import { RolePermission } from '@cs/domains/role';

export enum TokenPermissionName {
    CLUSTERS = 'clusters',
    RULES = 'rules',
    EVENTS = 'events',
    SYSTEM = 'system_settings'
}

export type TokenPermissions = {
    [key in TokenPermissionName]: RolePermission;
};

export interface Token {
    id: string;
    name: string;
    permissions: TokenPermissions;
    expires_at: string | null; // RFC3339
    invalidated_at?: string; // RFC3339
    access_token?: string;
}
