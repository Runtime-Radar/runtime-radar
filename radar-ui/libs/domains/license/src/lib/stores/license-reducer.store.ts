import { Action, ActionReducer, createReducer, on } from '@ngrx/store';

import { LicenseState } from '../interfaces';
import { UPDATE_LICENSE_STATE_DOC_ACTION } from './license-action.store';

const INITIAL_STATE: LicenseState = {
    appVersion: '',
    hostAppVersion: '',
    centralUrl: ''
};

const reducer: ActionReducer<LicenseState, Action> = createReducer(
    INITIAL_STATE,
    on(UPDATE_LICENSE_STATE_DOC_ACTION, (state, values) => ({ ...state, ...values }))
);

export function licenseReducer(state: LicenseState | undefined, action: Action): LicenseState {
    return reducer(state, action);
}
