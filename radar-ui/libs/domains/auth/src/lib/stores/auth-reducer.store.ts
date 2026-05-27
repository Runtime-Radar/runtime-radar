import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import { AUTH_DEFAULT_ORIGIN_PATH } from '../constants/auth.constant';
import { AuthCredentials, AuthState } from '../interfaces/state/auth-state.interface';
import { RESET_AUTH_CREDENTIALS_DOC_ACTION, UPDATE_AUTH_STATE_DOC_ACTION } from './auth-action.store';

const INITIAL_STATE_CREDENTIALS: AuthCredentials = {
    userId: '',
    username: '',
    email: '',
    roleId: '',
    authType: 'internal',
    passwordChangeTimestamp: 0
};

const INITIAL_STATE_CONFIG: Omit<AuthState, keyof AuthCredentials> = {
    loadStatus: LoadStatus.INIT,
    originPath: AUTH_DEFAULT_ORIGIN_PATH
};

const reducer: ActionReducer<AuthState, Action> = createReducer(
    { ...INITIAL_STATE_CONFIG, ...INITIAL_STATE_CREDENTIALS },
    on(UPDATE_AUTH_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values })),
    on(RESET_AUTH_CREDENTIALS_DOC_ACTION, (state) => ({
        ...state,
        ...INITIAL_STATE_CREDENTIALS,
        loadStatus: LoadStatus.ERROR
    }))
);

export function authReducer(state: AuthState | undefined, action: Action): AuthState {
    return reducer(state, action);
}
