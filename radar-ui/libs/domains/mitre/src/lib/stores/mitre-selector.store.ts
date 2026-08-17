import { InjectionToken } from '@angular/core';
import { ActionReducerMap, createFeatureSelector, createSelector } from '@ngrx/store';

import { MitreEntityState, MitreState, MitreTactic, MitreTechnique } from '../interfaces';
import { mitreEntitySelector, mitreReducerFactory } from './mitre-reducer.store';

export const MITRE_DOMAIN_KEY = 'mitre';

export interface MitreDomainState {
    readonly domain: MitreState;
}

export const MITRE_DOMAIN_REDUCER = new InjectionToken<ActionReducerMap<MitreDomainState>>('mitreDomainReducer');

const selectMitreDomainState = createFeatureSelector<MitreDomainState>(MITRE_DOMAIN_KEY);
const selectMitreState = createSelector(selectMitreDomainState, (state: MitreDomainState) => state.domain);
const selectMitreEntityState = (locale: string) =>
    createSelector(selectMitreState, (state: MitreState) => state.list[locale]);

export const getMitreTactic = (id: string, locale: string) =>
    createSelector(selectMitreEntityState(locale), (state: MitreEntityState): MitreTactic | undefined =>
        mitreEntitySelector.selectAll(state).find((item) => item.id === id)
    );

export const getMitreTechnique = (id: string, tacticId: string, locale: string) =>
    createSelector(selectMitreEntityState(locale), (state: MitreEntityState): MitreTechnique | undefined => {
        const tactic = mitreEntitySelector.selectAll(state).find((item) => item.id === tacticId);
        return tactic ? tactic.techniques.find((item) => item.id === id) : undefined;
    });

export function mitreDomainReducerFactory(availableLocales: string[]): ActionReducerMap<MitreDomainState> {
    return {
        domain: mitreReducerFactory(availableLocales)
    };
}
