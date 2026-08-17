import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ActivatedRoute, Params, Router } from '@angular/router';
import {
    BehaviorSubject,
    Observable,
    Subject,
    catchError,
    combineLatest,
    concatMap,
    filter,
    forkJoin,
    map,
    of,
    shareReplay,
    switchMap,
    take,
    tap
} from 'rxjs';
import { ChangeDetectionStrategy, Component, DestroyRef, OnInit, inject } from '@angular/core';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';

import { ApiPathService } from '@cs/api';
import { GridPackageColumnComponentStore } from '@cs/packages/grid';
import { I18nService } from '@cs/i18n';
import { ClusterStoreService, RegisteredCluster } from '@cs/domains/cluster';
import { DetectorExtended, DetectorStoreService, DetectorType } from '@cs/domains/detector';
import { GetRuleResponse, Rule, RuleRequestService, RuleStoreService, RuleType } from '@cs/domains/rule';
import { GetRuntimeEventsResponse, RuntimeEventCursorDirection } from '@cs/domains/runtime';
import { LoadStatus, RouterName } from '@cs/core';
import { PermissionName, RolePermissionMap } from '@cs/domains/role';

import { RUNTIME_NAVIGATION_TABS } from '../../constants/runtime-navigation.constant';
import { RuntimeFeatureEventFilterComponentStore } from '../../stores/runtime-event-filter.store';
import { RuntimeFeatureRequestAdapterService } from '../../services/runtime-request-adapter.service';
import { RuntimeRouterName } from '../../interfaces/runtime-navigation.interface';
import { RuntimeFeatureConfigUtilsService as runtimeConfigUtils } from '../../services/runtime-utils.service';
import {
    GetRuntimeEventsResponseExtended,
    RuntimeEventsGridContextId,
    RuntimeEventsPagination
} from '../../interfaces/runtime-events.interface';
import {
    RUNTIME_EVENT_GRID_COLUMNS,
    RUNTIME_EVENT_GRID_COLUMNS_INITIAL_VALUES
} from '../../constants/runtime-grid.constant';
import {
    RUNTIME_FILTER_INITIAL_CONTEXT_STATE,
    RUNTIME_FILTER_INITIAL_STATE
} from '../../constants/runtime-filter.constant';
import {
    RuntimeEventContext,
    RuntimeEventFilterContextDropdown,
    RuntimeEventFilterContextDropdownType,
    RuntimeEventFilterKey,
    RuntimeEventFilterRuleNode,
    RuntimeEventFilters
} from '../../interfaces/runtime-filter.interface';

@Component({
    templateUrl: './runtime-events.container.html',
    styleUrl: './runtime-events.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false,
    providers: [RuntimeFeatureEventFilterComponentStore, GridPackageColumnComponentStore]
})
export class RuntimeFeatureEventsContainer implements OnInit {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly destroyRef = inject(DestroyRef);
    private readonly route = inject(ActivatedRoute);
    private readonly router = inject(Router);
    private readonly toastService = inject(KbqToastService);

    private readonly apiPathService = inject(ApiPathService);
    private readonly clusterStoreService = inject(ClusterStoreService);
    private readonly detectorStoreService = inject(DetectorStoreService);
    private readonly gridPackageColumnComponentStore = inject(GridPackageColumnComponentStore);
    private readonly i18nService = inject(I18nService);
    private readonly ruleRequestService = inject(RuleRequestService);
    private readonly ruleStoreService = inject(RuleStoreService);
    private readonly runtimeFeatureEventFilterComponentStore = inject(RuntimeFeatureEventFilterComponentStore);
    private readonly runtimeFeatureRequestAdapterService = inject(RuntimeFeatureRequestAdapterService);

    readonly updateCounter$ = new Subject<string>();

    readonly activeClusterHost$ = this.apiPathService.host$;

    /* eslint @typescript-eslint/dot-notation: "off" */
    private readonly direction = this.route.snapshot.queryParams['direction'] as
        | RuntimeEventCursorDirection
        | undefined;
    private readonly cursor = this.route.snapshot.queryParams['cursor'] as string | undefined;
    readonly activeCursor$ = new BehaviorSubject<RuntimeEventsPagination>({
        direction: this.direction || RuntimeEventCursorDirection.RIGHT,
        cursor: this.cursor || this.dateAdapter.today().toJSDate().toISOString() // RFC3339
    });

    // this subject stores value shows whether click on cursor navigation happened
    readonly isClickWithCursorCheck$ = new BehaviorSubject(false);

    readonly loadStatus$ = new BehaviorSubject<LoadStatus>(LoadStatus.INIT);

    readonly activeContextId$: Observable<string | undefined> =
        this.runtimeFeatureEventFilterComponentStore.context$.pipe(map((context) => context.activeContextId));

    readonly isFilterExist$: Observable<boolean> = combineLatest([
        this.runtimeFeatureEventFilterComponentStore.filters$,
        this.runtimeFeatureEventFilterComponentStore.context$
    ]).pipe(
        map(
            ([filters, context]) =>
                !!Object.values({ ...filters, ...context }).filter((value) =>
                    runtimeConfigUtils.isEventFilterValueValid(value)
                ).length
        )
    );

    private eventsResponseBuffer?: GetRuntimeEventsResponse;
    readonly eventsResponse$: Observable<GetRuntimeEventsResponseExtended> = combineLatest([
        this.activeCursor$,
        this.activeClusterHost$
    ]).pipe(
        tap(() => this.loadStatus$.next(LoadStatus.IN_PROGRESS)),
        switchMap(([pagination]) =>
            combineLatest([
                this.runtimeFeatureEventFilterComponentStore.filters$,
                this.runtimeFeatureEventFilterComponentStore.context$
            ]).pipe(
                take(1),
                map(([filters, context]) => ({ pagination, filters, context }))
            )
        ),
        switchMap(({ pagination, filters, context }) =>
            this.runtimeFeatureRequestAdapterService.getEvents(pagination, filters, context).pipe(
                switchMap((response) =>
                    this.isClickWithCursorCheck$.pipe(
                        take(1),
                        map((isClickWithCursorCheck) => ({ response, isClickWithCursorCheck }))
                    )
                ),
                map(({ response, isClickWithCursorCheck }) => {
                    // cursor check is needed to detect whether backend returns empty result or empty cursor
                    if (
                        this.eventsResponseBuffer !== undefined &&
                        !response.runtime_events.length &&
                        isClickWithCursorCheck
                    ) {
                        this.loadStatus$.next(LoadStatus.LOADED);
                        this.toastService.show({
                            style: KbqToastStyle.Contrast,
                            title: this.i18nService.translate('Runtime.Pseudo.Notification.CursorLimited')
                        });

                        return {
                            ...this.eventsResponseBuffer,
                            isPrevResponse: true
                        };
                    }

                    this.loadStatus$.next(LoadStatus.LOADED);
                    this.eventsResponseBuffer = { ...response };

                    return {
                        ...this.eventsResponseBuffer,
                        isPrevResponse: false
                    };
                }),
                catchError(() => {
                    this.loadStatus$.next(LoadStatus.ERROR);

                    return of({
                        runtime_events: [],
                        left_cursor: '',
                        right_cursor: '',
                        isPrevResponse: true
                    });
                })
            )
        ),
        shareReplay({
            bufferSize: 1,
            refCount: true
        })
    );

    private readonly routeWithFilterAndCursor$: Observable<Params> = this.eventsResponse$.pipe(
        filter((response) => !response.isPrevResponse),
        switchMap(() =>
            combineLatest([this.activeCursor$, this.runtimeFeatureEventFilterComponentStore.queryParams$]).pipe(
                take(1),
                map(([pagination, queryParams]) => ({ ...queryParams, ...pagination }) as Params)
            )
        ),
        tap((queryParams) => {
            this.router.navigate([RouterName.RUNTIME, RuntimeRouterName.EVENTS], {
                queryParams,
                replaceUrl: true
            });
        })
    );

    readonly detectors$: Observable<DetectorExtended[]> = this.detectorStoreService.detectors$([DetectorType.RUNTIME]);

    readonly ruleNodes$: Observable<RuntimeEventFilterRuleNode[]> = this.ruleStoreService.rules$.pipe(
        map((rules) => rules.filter((item) => item.type === RuleType.TYPE_RUNTIME)),
        switchMap((rules) =>
            this.filterComponentStore.filters$.pipe(
                map((filters) => {
                    const ids = rules.map((item) => item.id);

                    return (filters?.rules || []).filter((id) => !ids.includes(id));
                }),
                // @todo: refactor getRule to get an ability to provide ids
                concatMap((ruleIds) =>
                    ruleIds.length
                        ? forkJoin(ruleIds.map((id) => this.ruleRequestService.getRule(id)))
                        : of([] as GetRuleResponse[])
                ),
                map((response: GetRuleResponse[]) => ({
                    deletedRules: response.map((item) => item.rule),
                    rules
                }))
            )
        ),
        map(({ rules, deletedRules }) => {
            const convertToNode = (entities: Rule[], isExtra: boolean) =>
                entities.map<RuntimeEventFilterRuleNode>((item) => ({
                    id: item.id,
                    name: item.name,
                    isExtra
                }));

            return convertToNode(rules, false).concat(convertToNode(deletedRules, true));
        })
    );

    readonly clusters$: Observable<RegisteredCluster[]> = this.clusterStoreService.registeredClusters$;

    readonly filterComponentStore = this.runtimeFeatureEventFilterComponentStore;

    readonly gridColumnComponentStore = this.gridPackageColumnComponentStore;

    readonly loadStatus = LoadStatus;

    readonly runtimeEventCursorDirection = RuntimeEventCursorDirection;

    readonly routerName = RouterName;

    readonly runtimeRouterName = RuntimeRouterName;

    readonly runtimeNavigationTabs = RUNTIME_NAVIGATION_TABS;

    readonly runtimeEventGridColumns = RUNTIME_EVENT_GRID_COLUMNS;

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionName = PermissionName;

    ngOnInit() {
        this.routeWithFilterAndCursor$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
        this.gridPackageColumnComponentStore.setInitialState({
            id: 'rntmgrd',
            columns: RUNTIME_EVENT_GRID_COLUMNS_INITIAL_VALUES
        });
    }

    tabChange(path?: string) {
        this.router.navigate([RouterName.RUNTIME, path]);
    }

    goToStartPage() {
        const cursor = this.dateAdapter.today().toJSDate().toISOString(); // RFC3339

        this.isClickWithCursorCheck$.next(false);
        this.updateCounter$.next(cursor);
        this.activeCursor$.next({
            direction: RuntimeEventCursorDirection.RIGHT,
            cursor
        });
    }

    goToEndPage() {
        this.isClickWithCursorCheck$.next(false);
        this.activeCursor$.next({
            direction: RuntimeEventCursorDirection.LEFT,
            cursor: this.dateAdapter.today().minus({ years: 10 }).toJSDate().toISOString() // RFC3339
        });
    }

    resetFilterAndContext() {
        this.runtimeFeatureEventFilterComponentStore.update({
            ...RUNTIME_FILTER_INITIAL_STATE,
            ...RUNTIME_FILTER_INITIAL_CONTEXT_STATE
        });
        this.goToStartPage();
    }

    setContextId(values: RuntimeEventsGridContextId) {
        this.runtimeFeatureEventFilterComponentStore.setContextIds(
            values.execId,
            values.parentExecId,
            values.context,
            values.id
        );
        this.goToStartPage();
    }

    changeFilterEntity(filter: RuntimeEventFilterContextDropdown) {
        if (filter.key === RuntimeEventFilterKey.NAMESPACE) {
            const [namespace, pod] = filter.value.split('/');
            this.runtimeFeatureEventFilterComponentStore.update({
                [RuntimeEventFilterKey.NAMESPACE]:
                    filter.type === RuntimeEventFilterContextDropdownType.ADD ? namespace : '',
                [RuntimeEventFilterKey.POD]: filter.type === RuntimeEventFilterContextDropdownType.ADD ? pod : ''
            });
        } else {
            this.runtimeFeatureEventFilterComponentStore.update({
                [filter.key]: filter.type === RuntimeEventFilterContextDropdownType.ADD ? filter.value : ''
            });
        }

        this.goToStartPage();
    }

    changeFilter(filters: RuntimeEventFilters) {
        this.runtimeFeatureEventFilterComponentStore.update(filters);
        this.goToStartPage();
    }

    changeContext(context: RuntimeEventContext) {
        this.runtimeFeatureEventFilterComponentStore.setContextIds(
            context.execId,
            context.parentExecId,
            context.context,
            context.activeContextId
        );
        this.goToStartPage();
    }

    changePage(direction: RuntimeEventCursorDirection, cursor: string) {
        this.isClickWithCursorCheck$.next(true);
        this.activeCursor$.next({
            direction,
            cursor
        });
    }

    switchCluster(id: string) {
        this.clusterStoreService.switchCluster(id);
    }
}
