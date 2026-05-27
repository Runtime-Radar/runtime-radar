import { LoadStatus } from '@cs/core';

export interface AuthTokens {
    accessToken: string;
    refreshToken: string;
}

export interface AuthCredentials {
    userId: string;
    username: string;
    email: string;
    roleId: string;
    authType: string;
    passwordChangeTimestamp: number;
}

export interface AuthState extends AuthCredentials {
    loadStatus: LoadStatus;
    originPath: string;
}
