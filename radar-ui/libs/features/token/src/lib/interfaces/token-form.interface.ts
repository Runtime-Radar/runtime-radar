import { DateTime } from 'luxon';

import { TokenPermissionName } from '@cs/domains/token';

export enum TokenExpiryDatePreset {
    WEEK = 'WEEK',
    MONTH = 'MONTH',
    QUARTER = 'QUARTER',
    INDEFINITELY = 'INDEFINITELY',
    CUSTOM = 'CUSTOM'
}

export interface TokenExpiryDatePresetOption {
    id: TokenExpiryDatePreset;
    localizationKey: string;
}

export enum TokenPermissionType {
    CREATE = 'canCreate',
    READ = 'canRead',
    UPDATE = 'canUpdate',
    DELETE = 'canDelete',
    EXECUTE = 'canExecute'
}

export type TokenPermissionForm = {
    [key in TokenPermissionType]: boolean;
};

export type TokenPermissionRecord = {
    [key in TokenPermissionName]: TokenPermissionForm;
};

export interface TokenForm {
    name: string;
    date: DateTime;
    preset: TokenExpiryDatePreset;
    permissions: TokenPermissionRecord;
}
