import { Action } from '@ngrx/store';
import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { Observable, forkJoin, of } from 'rxjs';
import { catchError, concatMap, map, switchMap, take } from 'rxjs/operators';

import { ROLE_LOAD_DONE_EVENT_ACTION } from '@cs/domains/role';
import { AVAILABLE_LOCALES, LoadStatus } from '@cs/core';

import { MitreRequestService } from '../services/mitre-request.service';
import { MitreTactic } from '../interfaces';
import { SET_ALL_MITRE_ENTITIES_DOC_ACTION, UPDATE_MITRE_STATE_DOC_ACTION } from './mitre-action.store';

@Injectable({
    providedIn: 'root'
})
export class MitreEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly availableLocales = inject<string[]>(AVAILABLE_LOCALES);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);

    private readonly mitreRequestService = inject(MitreRequestService);

    readonly loadMitreAttacks$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(ROLE_LOAD_DONE_EVENT_ACTION),
            concatMap(() =>
                forkJoin(
                    this.availableLocales.map((locale) =>
                        this.getMitreMatrixByLocale(locale).pipe(
                            map((mitre) => ({
                                locale,
                                mitre
                            }))
                        )
                    )
                )
            ),
            switchMap((list) => {
                const actions: Action[] = [
                    UPDATE_MITRE_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    })
                ];

                list.forEach((item) => {
                    if (item.mitre.length) {
                        actions.push(SET_ALL_MITRE_ENTITIES_DOC_ACTION({ list: item.mitre, locale: item.locale }));
                    }
                });

                return actions;
            })
        )
    );

    private getMitreMatrixByLocale(locale: string): Observable<MitreTactic[]> {
        return this.mitreRequestService.getMitreMatrix(locale).pipe(
            take(1),
            catchError(() => of([]))
        );
    }
}
