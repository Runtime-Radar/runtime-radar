import { createAction, props } from '@ngrx/store';

import { MitreState, MitreTactic } from '../interfaces';

export const UPDATE_MITRE_STATE_DOC_ACTION = createAction(
    '[Mitre] (Doc) Update State',
    props<Partial<Omit<MitreState, 'list'>>>()
);

export const SET_ALL_MITRE_ENTITIES_DOC_ACTION = createAction(
    '[Mitre] (Doc) Set All',
    props<{ list: MitreTactic[]; locale: string }>()
);
