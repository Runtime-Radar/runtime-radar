import { Store } from '@ngrx/store';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { MitreState } from '../interfaces';
import { getMitreTactic, getMitreTechnique } from '../stores/mitre-selector.store';

@Injectable({
    providedIn: 'root'
})
export class MitreStoreService {
    private readonly store = inject<Store<MitreState>>(Store);

    readonly tactic$ = (id: string, locale: string): Observable<string> =>
        this.store.select(getMitreTactic(id, locale)).pipe(map((tactic) => (tactic ? tactic.name : id)));

    readonly technique$ = (id: string, tacticId: string, locale: string): Observable<string> =>
        this.store
            .select(getMitreTechnique(id, tacticId, locale))
            .pipe(map((technique) => (technique ? technique.name : id)));
}
