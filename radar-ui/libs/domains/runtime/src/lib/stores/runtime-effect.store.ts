import { Action, Store } from '@ngrx/store';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { Injectable, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';
import { NEVER, Observable, combineLatest, forkJoin, of } from 'rxjs';
import { catchError, concatMap, debounceTime, filter, map, mergeMap, switchMap, take, tap } from 'rxjs/operators';

import { I18nService } from '@cs/i18n';
import { SIGN_OUT_EVENT_ACTION } from '@cs/domains/auth';
import { SWITCH_CLUSTER_EVENT_ACTION } from '@cs/domains/cluster';
import { CoreNavigationStoreService, CoreWindowService, LoadStatus, CoreUtilsService as utils } from '@cs/core';

import { RuntimeRequestService } from '../services/runtime-request.service';
import { RuntimeHelperService as runtimeHelper } from '../services/runtime-helper.service';
import {
    CHECK_RUNTIME_CHANGES_TODO_ACTION,
    CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION,
    CREATE_RUNTIME_CONFIG_TODO_ACTION,
    DEACTIVATE_RUNTIME_CONFIG_TODO_ACTION,
    GET_RUNTIME_CONFIG_STATUS_TODO_ACTION,
    HIDE_RUNTIME_OVERLAY_TODO_ACTION,
    LOAD_RUNTIME_CONFIG_TODO_ACTION,
    RESET_RUNTIME_CONFIG_TODO_ACTION,
    SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION,
    UPDATE_RUNTIME_STATE_DOC_ACTION
} from './runtime-action.store';
import {
    RuntimeConfigStatus,
    RuntimeEventProcessorHistoryControl,
    RuntimeMonitorConfig,
    RuntimeState
} from '../interfaces';
import {
    getRuntimeEventProcessorHistoryControl,
    getRuntimeIsExpertMode,
    getRuntimeLoadStatus,
    getRuntimeMonitorConfig
} from './runtime-selector.store';

const RUNTIME_EXPERT_MODE_LOCAL_STORAGE_KEY = 'xprtmd';

@Injectable({
    providedIn: 'root'
})
export class RuntimeEffectStore {
    private readonly actions$ = inject(Actions);
    private readonly toastService = inject(KbqToastService);

    private readonly coreNavigationStoreService = inject(CoreNavigationStoreService);
    private readonly coreWindowService = inject(CoreWindowService);
    private readonly i18nService = inject(I18nService);
    private readonly runtimeRequestService = inject(RuntimeRequestService);
    private readonly store = inject<Store<RuntimeState>>(Store);

    readonly checkExpertMode$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_RUNTIME_CONFIG_TODO_ACTION),
            map(() => this.coreWindowService.localStorage.getItem(RUNTIME_EXPERT_MODE_LOCAL_STORAGE_KEY)),
            map((value) =>
                UPDATE_RUNTIME_STATE_DOC_ACTION({
                    isExpertMode: value ? value === 'true' : false
                })
            )
        )
    );

    readonly getConfigStatus$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_RUNTIME_CONFIG_TODO_ACTION, GET_RUNTIME_CONFIG_STATUS_TODO_ACTION),
            switchMap(() => this.runtimeRequestService.getRuntimeMonitorStatus()),
            map((status) => UPDATE_RUNTIME_STATE_DOC_ACTION({ status }))
        )
    );

    readonly loadGrafanaUrl$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_RUNTIME_CONFIG_TODO_ACTION),
            switchMap(() => this.runtimeRequestService.getGrafanaUrl()),
            map((grafanaUrl) => UPDATE_RUNTIME_STATE_DOC_ACTION({ grafanaUrl }))
        )
    );

    readonly loadConfig$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(LOAD_RUNTIME_CONFIG_TODO_ACTION),
            switchMap(() =>
                forkJoin({
                    config: this.runtimeRequestService.getRuntimeMonitor().pipe(
                        map((response) => response.config),
                        catchError(() => of(undefined))
                    ),
                    historyControl: this.runtimeRequestService.getEventProcessor().pipe(
                        map((response) => response.config.history_control),
                        catchError(() => of(undefined))
                    )
                })
            ),
            map(({ config, historyControl }) => {
                if (config === undefined) {
                    return UPDATE_RUNTIME_STATE_DOC_ACTION({
                        loadStatus: LoadStatus.ERROR
                    });
                }

                return UPDATE_RUNTIME_STATE_DOC_ACTION({
                    loadStatus: LoadStatus.LOADED,
                    historyControl,
                    config
                });
            })
        )
    );

    readonly reloadConfig$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_CLUSTER_EVENT_ACTION),
            switchMap(() => this.store.select(getRuntimeLoadStatus).pipe(take(1))),
            filter((status) => status !== LoadStatus.INIT),
            map(() => LOAD_RUNTIME_CONFIG_TODO_ACTION())
        )
    );

    readonly resetConfig$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(RESET_RUNTIME_CONFIG_TODO_ACTION),
            switchMap(() => this.runtimeRequestService.resetConfigToDefault().pipe(take(1))),
            mergeMap((isConfigReseted) => {
                if (!isConfigReseted) {
                    return of(undefined);
                }

                return this.runtimeRequestService.getRuntimeMonitor().pipe(
                    map((response) => response.config),
                    catchError(() => of(undefined))
                );
            }),
            tap((config) => {
                this.toastService.show({
                    style: config === undefined ? KbqToastStyle.Warning : KbqToastStyle.Success,
                    title: this.i18nService.translate(
                        config === undefined
                            ? 'Runtime.Pseudo.Notification.ResetFailed'
                            : 'Runtime.Pseudo.Notification.Reseted'
                    )
                });
            }),
            filter((config) => config !== undefined),
            switchMap((config) => [
                UPDATE_RUNTIME_STATE_DOC_ACTION({
                    config,
                    hasChanges: false,
                    hasPoliciesChanges: false,
                    configStatus: RuntimeConfigStatus.STAY
                }),
                GET_RUNTIME_CONFIG_STATUS_TODO_ACTION()
            ])
        )
    );

    readonly deactivateState$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(DEACTIVATE_RUNTIME_CONFIG_TODO_ACTION),
            map(() =>
                UPDATE_RUNTIME_STATE_DOC_ACTION({
                    loadStatus: LoadStatus.INIT // @todo: set separate status e.g. PARTIAL
                })
            )
        )
    );

    readonly createRuntimeMonitor$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_RUNTIME_CONFIG_TODO_ACTION),
            switchMap((action) =>
                this.store.select(getRuntimeMonitorConfig).pipe(
                    take(1),
                    map((config: RuntimeMonitorConfig) => {
                        const previous = runtimeHelper.convertConfigToDiffValues(config);
                        const current = { ...action.config, historyControl: RuntimeEventProcessorHistoryControl.NONE };

                        return !utils.isEqual(previous, current) ? action.config : undefined;
                    })
                )
            ),
            filter((config) => !!config),
            switchMap((config) =>
                this.runtimeRequestService.createRuntimeMonitor(config).pipe(
                    take(1),
                    map((response) => response.config)
                )
            ),
            filter((config) => config && !!Object.keys(config).length),
            switchMap((config) => [
                GET_RUNTIME_CONFIG_STATUS_TODO_ACTION(),
                CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION(),
                UPDATE_RUNTIME_STATE_DOC_ACTION({
                    config,
                    configStatus: RuntimeConfigStatus.STAY
                })
            ])
        )
    );

    readonly createEventProcessor$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CREATE_RUNTIME_CONFIG_TODO_ACTION),
            switchMap((action) =>
                this.store.select(getRuntimeEventProcessorHistoryControl).pipe(
                    take(1),
                    map((historyControl) =>
                        historyControl !== action.historyControl ? action.historyControl : undefined
                    )
                )
            ),
            filter((historyControl) => !!historyControl),
            switchMap((historyControl) =>
                this.runtimeRequestService.createEventProcessor(historyControl).pipe(
                    take(1),
                    map((response) => response.config.history_control)
                )
            ),
            filter((historyControl) => !!historyControl),
            switchMap((historyControl) => [
                CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION(),
                UPDATE_RUNTIME_STATE_DOC_ACTION({ historyControl })
            ])
        )
    );

    readonly showCreateNotification$: Observable<Action> = createEffect(
        () =>
            this.actions$.pipe(
                ofType(CREATE_RUNTIME_CONFIG_NOTIFICATION_TODO_ACTION),
                debounceTime(1000),
                tap(() => {
                    this.toastService.show({
                        style: KbqToastStyle.Success,
                        title: this.i18nService.translate('Runtime.Pseudo.Notification.Created')
                    });
                }),
                concatMap(() => NEVER)
            ),
        { dispatch: false }
    );

    readonly checkChanges$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(CHECK_RUNTIME_CHANGES_TODO_ACTION),
            switchMap((action) =>
                combineLatest([
                    this.store.select(getRuntimeEventProcessorHistoryControl).pipe(take(1)),
                    this.store.select(getRuntimeMonitorConfig).pipe(take(1))
                ]).pipe(
                    map(([historyControl, config]) => ({
                        previous: runtimeHelper.convertConfigToDiffValues(config, historyControl),
                        current: action.config
                    }))
                )
            ),
            map(({ previous, current }) => {
                const hasChanges = !utils.isEqual(previous, current);

                return UPDATE_RUNTIME_STATE_DOC_ACTION({
                    hasChanges,
                    hasPoliciesChanges: !utils.isEqual(
                        Object.entries(previous.tracing_policies).map(([key, { enabled, ...rest }]) => [key, rest]),
                        Object.entries(current.tracing_policies).map(([key, { enabled, ...rest }]) => [key, rest])
                    ),
                    configStatus: hasChanges ? RuntimeConfigStatus.MODIFY : RuntimeConfigStatus.STAY
                });
            })
        )
    );

    readonly switchExpertMode$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION),
            switchMap(() => this.store.select(getRuntimeIsExpertMode).pipe(take(1))),
            tap((isExpertMode) => {
                this.coreWindowService.localStorage.setItem(
                    RUNTIME_EXPERT_MODE_LOCAL_STORAGE_KEY,
                    (!isExpertMode).toString()
                );
            }),
            map((isExpertMode) => UPDATE_RUNTIME_STATE_DOC_ACTION({ isExpertMode: !isExpertMode }))
        )
    );

    readonly deactivateExpertMode$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(SIGN_OUT_EVENT_ACTION),
            map(() => UPDATE_RUNTIME_STATE_DOC_ACTION({ isExpertMode: false })),
            tap(() => {
                this.coreWindowService.localStorage.removeItem(RUNTIME_EXPERT_MODE_LOCAL_STORAGE_KEY);
            })
        )
    );

    readonly hideOverlay$: Observable<Action> = createEffect(() =>
        this.actions$.pipe(
            ofType(HIDE_RUNTIME_OVERLAY_TODO_ACTION),
            tap(() => {
                this.coreNavigationStoreService.setLoadStatus(LoadStatus.LOADED);
            }),
            map(() =>
                UPDATE_RUNTIME_STATE_DOC_ACTION({
                    isOverlayed: false
                })
            )
        )
    );
}
