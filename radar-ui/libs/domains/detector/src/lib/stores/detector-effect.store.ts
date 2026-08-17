import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { HttpErrorResponse } from '@angular/common/http';
import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { Observable, combineLatest, map, of, switchMap, zip } from 'rxjs';
import { catchError, filter, mergeMap, take, tap } from 'rxjs/operators';

import { ApiErrorResponse } from '@cs/api';
import { I18nService } from '@cs/i18n';
import { LoadStatus } from '@cs/core';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';

import { DetectorRequestService } from '../services/detector-request.service';
import { getDetectorConfigs } from './detector-selector.store';
import {
    CREATE_RUNTIME_DETECTOR_ENTITIES_TODO_ACTION,
    DELETE_DETECTOR_ENTITY_DOC_ACTION,
    DELETE_RUNTIME_DETECTOR_ENTITY_TODO_ACTION,
    LOAD_DETECTOR_ENTITIES_TODO_ACTION,
    POLLING_LOAD_DETECTOR_ENTITIES_TODO_ACTION,
    SET_ALL_DETECTOR_ENTITIES_DOC_ACTION,
    SET_MANY_DETECTOR_CONFIG_ENTITIES_DOC_ACTION,
    SET_MANY_DETECTOR_ENTITIES_DOC_ACTION,
    UPSERT_DETECTOR_CONFIG_ENTITY_DOC_ACTION
} from './detector-action.store';
import { Detector, DetectorState, DetectorType, GetDetectorsRequest, GetDetectorsResponse } from '../interfaces';

const GET_DETECTORS_REQUEST: GetDetectorsRequest = {
    page_size: 1000
};

@Injectable({
    providedIn: 'root'
})
export class DetectorEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly toastService = inject(KbqToastService);

    private readonly detectorRequestService = inject(DetectorRequestService);
    private readonly i18nService = inject(I18nService);
    private readonly store = inject<Store<DetectorState>>(Store);

    readonly loadDetectors$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_DETECTOR_ENTITIES_TODO_ACTION),
            mergeMap(({ detectorType }) =>
                combineLatest([
                    of(detectorType),
                    this.getDetectors(detectorType).pipe(
                        take(1),
                        map((response) => response.detectors),
                        catchError(() => of(undefined))
                    )
                ])
            ),
            switchMap(([type, list]) => {
                if (list === undefined) {
                    return [
                        UPSERT_DETECTOR_CONFIG_ENTITY_DOC_ACTION({
                            config: {
                                id: type,
                                loadStatus: LoadStatus.ERROR,
                                lastUpdate: 0
                            }
                        })
                    ];
                }

                return [
                    SET_MANY_DETECTOR_ENTITIES_DOC_ACTION({
                        list: list.map((detector) => ({
                            ...detector,
                            id: `${detector.id}${detector.version}`,
                            key: detector.id,
                            type
                        }))
                    }),
                    UPSERT_DETECTOR_CONFIG_ENTITY_DOC_ACTION({
                        config: {
                            id: type,
                            loadStatus: LoadStatus.LOADED,
                            lastUpdate: this.dateAdapter.today().toMillis()
                        }
                    })
                ];
            })
        )
    );

    readonly reloadDetectors$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() =>
                this.store.select(getDetectorConfigs).pipe(
                    take(1),
                    map((list) => list.reduce<DetectorType[]>((acc, item) => (item ? [...acc, item.id] : acc), []))
                )
            ),
            switchMap((types) => {
                const requests = types.map((type) =>
                    this.getDetectors(type).pipe(
                        take(1),
                        map((response) => response.detectors),
                        catchError(() => of(undefined))
                    )
                );

                return combineLatest(requests).pipe(
                    map((responses) => ({
                        types,
                        responses
                    }))
                );
            }),
            mergeMap(({ types, responses }) => [
                SET_ALL_DETECTOR_ENTITIES_DOC_ACTION({
                    list: responses.flatMap((detectors, index) =>
                        detectors !== undefined
                            ? detectors.map((item) => ({
                                  ...item,
                                  id: `${item.id}${item.version}`,
                                  key: item.id,
                                  type: types[index]
                              }))
                            : []
                    )
                }),
                SET_MANY_DETECTOR_CONFIG_ENTITIES_DOC_ACTION({
                    config: types.map((type, index) => ({
                        id: type,
                        loadStatus: responses[index] !== undefined ? LoadStatus.LOADED : LoadStatus.ERROR,
                        lastUpdate: responses[index] !== undefined ? this.dateAdapter.today().toMillis() : 0
                    }))
                })
            ])
        )
    );

    readonly pollingDetectors$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(POLLING_LOAD_DETECTOR_ENTITIES_TODO_ACTION),
            mergeMap(({ detectorType }) =>
                combineLatest([
                    of(detectorType),
                    this.getDetectors(detectorType).pipe(
                        take(1),
                        map((response) => response.detectors)
                    )
                ])
            ),
            switchMap(([type, list]) => [
                SET_MANY_DETECTOR_ENTITIES_DOC_ACTION({
                    list: list.map((detector) => ({
                        ...detector,
                        id: `${detector.id}${detector.version}`,
                        key: detector.id,
                        type
                    }))
                }),
                UPSERT_DETECTOR_CONFIG_ENTITY_DOC_ACTION({
                    config: {
                        id: type,
                        loadStatus: LoadStatus.LOADED,
                        lastUpdate: this.dateAdapter.today().toMillis()
                    }
                })
            ])
        )
    );

    readonly createRuntimeDetectors$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_RUNTIME_DETECTOR_ENTITIES_TODO_ACTION),
            switchMap(({ base64list }) =>
                zip(
                    base64list.map((wasmBase64) =>
                        this.detectorRequestService.createRuntimeDetector({ wasmBase64 }).pipe(
                            catchError((error: HttpErrorResponse) => {
                                this.toastService.show({
                                    style: KbqToastStyle.Error,
                                    title: this.i18nService.translate(
                                        'Runtime.Pseudo.Notification.DetectorUploadFailed'
                                    ),
                                    caption: (error.error as ApiErrorResponse).message
                                });

                                return of({} as Detector);
                            })
                        )
                    )
                )
            ),
            map((detectors) => detectors.filter((item) => !!item.id)),
            filter((detectors) => !!detectors.length),
            map((detectors) =>
                detectors.map((item) => ({
                    ...item,
                    id: `${item.id}${item.version}`,
                    key: item.id,
                    type: DetectorType.RUNTIME
                }))
            ),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Success,
                    title: this.i18nService.translate('Runtime.Pseudo.Notification.DetectorCreated')
                });
            }),
            map((list) => SET_MANY_DETECTOR_ENTITIES_DOC_ACTION({ list }))
        )
    );

    readonly deleteRuntimeDetector$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DELETE_RUNTIME_DETECTOR_ENTITY_TODO_ACTION),
            switchMap((item) =>
                this.detectorRequestService.deleteRuntimeDetector({ id: item.key, version: item.version }).pipe(take(1))
            ),
            filter((id) => !!id),
            tap(() => {
                this.toastService.show({
                    style: KbqToastStyle.Contrast,
                    title: this.i18nService.translate('Runtime.Pseudo.Notification.DetectorDeleted')
                });
            }),
            map((id) => DELETE_DETECTOR_ENTITY_DOC_ACTION({ id }))
        )
    );

    private getDetectors(type: DetectorType): Observable<GetDetectorsResponse> {
        switch (type) {
            case DetectorType.RUNTIME:
                return this.detectorRequestService.getRuntimeDetectors(GET_DETECTORS_REQUEST);
        }
    }
}
