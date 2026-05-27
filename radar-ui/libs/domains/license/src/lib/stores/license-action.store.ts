import { createAction, props } from '@ngrx/store';

import { LicenseState } from '../interfaces';

export const UPDATE_LICENSE_STATE_DOC_ACTION = createAction(
    '[License] (Doc) Update State',
    props<Partial<LicenseState>>()
);
