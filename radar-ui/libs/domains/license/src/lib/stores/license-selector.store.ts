import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { LicenseState } from '../interfaces';
import { licenseReducer } from './license-reducer.store';

export const LICENSE_DOMAIN_KEY = 'license';

export interface LicenseDomainState {
    readonly domain: LicenseState;
}

const selectLicenseDomainState = createFeatureSelector<LicenseDomainState>(LICENSE_DOMAIN_KEY);
const selectLicenseState = createSelector(selectLicenseDomainState, (state: LicenseDomainState) => state.domain);

export const getAppVersion = createSelector(selectLicenseState, (state: LicenseState) => state.appVersion);

export const getHostAppVersion = createSelector(selectLicenseState, (state: LicenseState) => state.hostAppVersion);

export const getCentralUrl = createSelector(selectLicenseState, (state: LicenseState) => state.centralUrl);

export const licenseDomainReducer: ActionReducerMap<LicenseDomainState> = {
    domain: licenseReducer
};
