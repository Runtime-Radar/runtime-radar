import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { authReducer } from './auth-reducer.store';
import { AuthCredentials, AuthState } from '../interfaces/state/auth-state.interface';

export const AUTH_DOMAIN_KEY = 'auth';

export interface AuthDomainState {
    readonly domain: AuthState;
}

const selectAuthDomainState = createFeatureSelector<AuthDomainState>(AUTH_DOMAIN_KEY);
const selectAuthState = createSelector(selectAuthDomainState, (state: AuthDomainState) => state.domain);

export const getAuthCredentials = createSelector(
    selectAuthState,
    (state: AuthState): AuthCredentials => ({
        userId: state.userId,
        username: state.username,
        email: state.email,
        roleId: state.roleId,
        authType: state.authType,
        passwordChangeTimestamp: state.passwordChangeTimestamp
    })
);

export const getAuthLoadStatus = createSelector(selectAuthState, (state: AuthState) => state.loadStatus);

export const getAuthOriginPath = createSelector(selectAuthState, (state: AuthState) => state.originPath);

export const authDomainReducer: ActionReducerMap<AuthDomainState> = {
    domain: authReducer
};
