import { DateTime } from 'luxon';
import { KbqAlertColors } from '@koobiq/components/alert';
import { ActivatedRoute, Router } from '@angular/router';
import {
    BehaviorSubject,
    Observable,
    catchError,
    combineLatest,
    filter,
    map,
    of,
    shareReplay,
    switchMap,
    take,
    tap
} from 'rxjs';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { DateAdapter, PopUpPlacements } from '@koobiq/components/core';
import { KbqSidepanelConfig, KbqSidepanelPosition, KbqSidepanelService } from '@koobiq/components/sidepanel';
import { KbqToastService, KbqToastStyle } from '@koobiq/components/toast';

import { I18nService } from '@cs/i18n';
import {
    GetRuntimeEventsResponse,
    RUNTIME_CONTEXT,
    RuntimeContext,
    RuntimeDetectError,
    RuntimeEvent,
    RuntimeEventCursorDirection,
    RuntimeEventEntity,
    RuntimeEventProcess,
    RuntimeEventThreat,
    RuntimeEventType,
    RuntimeRequestService
} from '@cs/domains/runtime';
import { LoadStatus, RouterName } from '@cs/core';
import { PermissionName, RolePermissionMap } from '@cs/domains/role';
import { RULE_SEVERITIES, RuleSeverity } from '@cs/domains/rule';

import { RUNTIME_DETAILS_LIST_ITEMS_LIMIT } from '../../constants/runtime-config.constant';
import { RUNTIME_FILTER_INITIAL_STATE } from '../../constants/runtime-filter.constant';
import { RuntimeEventContext } from '../../interfaces/runtime-filter.interface';
import { RuntimeFeatureRequestAdapterService } from '../../services/runtime-request-adapter.service';
import { RuntimeFeatureSidepanelCodeComponent } from '../../components/sidepanel-code/runtime-sidepanel-code.component';
import { RuntimeFeatureSidepanelThreatsComponent } from '../../components/sidepanel-threats/runtime-sidepanel-threats.component';
import { RuntimeRouterName } from '../../interfaces/runtime-navigation.interface';
import { GetRuntimeEventsResponseExtended, RuntimeEventsPagination } from '../../interfaces/runtime-events.interface';
import { RuntimeSidepanelCodeProps, RuntimeSidepanelThreatsProps } from '../../interfaces/runtime-sidepanel.interface';

const RUNTIME_DETAILS_PAGINATOR_PAGE_SIZE = 5;

@Component({
    templateUrl: './runtime-details.container.html',
    styleUrl: './runtime-details.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureDetailsContainer {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly route = inject(ActivatedRoute);
    private readonly router = inject(Router);
    private readonly sidepanelService = inject(KbqSidepanelService);
    private readonly toastService = inject(KbqToastService);

    private readonly i18nService = inject(I18nService);
    private readonly runtimeFeatureRequestAdapterService = inject(RuntimeFeatureRequestAdapterService);
    private readonly runtimeRequestService = inject(RuntimeRequestService);

    readonly loadStatus$ = new BehaviorSubject<LoadStatus>(LoadStatus.INIT);

    readonly cardLoadStatus$ = new BehaviorSubject<LoadStatus>(LoadStatus.INIT);

    readonly activeCursor$ = new BehaviorSubject<RuntimeEventsPagination>({
        direction: RuntimeEventCursorDirection.RIGHT,
        cursor: this.dateAdapter.today().toJSDate().toISOString() // RFC3339
    });

    readonly event$: Observable<RuntimeEvent> = this.route.params.pipe(
        map((params) => params['eventId']),
        tap(() => this.cardLoadStatus$.next(LoadStatus.IN_PROGRESS)),
        switchMap((eventId) =>
            this.runtimeRequestService.getEvent(eventId).pipe(
                take(1),
                tap((event: RuntimeEvent) => {
                    this.ruleIds = [...event.block_by, ...event.notify_by];
                    this.eventType = Object.values(RuntimeEventType).find(
                        (item) => event.event[item] !== undefined
                    ) as RuntimeEventType;
                    this.eventEntity = this.eventType ? event.event[this.eventType] : null;
                }),
                catchError(() => {
                    this.cardLoadStatus$.next(LoadStatus.ERROR);
                    this.router.navigate([RouterName.ERROR]);

                    return of({} as RuntimeEvent);
                })
            )
        ),
        filter((runtimeEvent) => runtimeEvent !== null),
        tap(() => this.cardLoadStatus$.next(LoadStatus.LOADED)),
        shareReplay({
            bufferSize: 1,
            refCount: true
        })
    );

    private eventsResponseBuffer?: GetRuntimeEventsResponse;
    readonly eventsResponse$: Observable<GetRuntimeEventsResponseExtended> = combineLatest([
        this.event$,
        this.activeCursor$
    ]).pipe(
        tap(() => this.loadStatus$.next(LoadStatus.IN_PROGRESS)),
        switchMap(([_, pagination]) =>
            this.runtimeFeatureRequestAdapterService
                .getEvents(
                    pagination,
                    RUNTIME_FILTER_INITIAL_STATE,
                    this.getEventContext(),
                    RUNTIME_DETAILS_PAGINATOR_PAGE_SIZE
                )
                .pipe(
                    map((response) => {
                        if (this.eventsResponseBuffer !== undefined && !response.runtime_events.length) {
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
        )
    );

    ruleIds: string[] = [];

    eventType: RuntimeEventType | null = null;

    eventEntity?: RuntimeEventProcess | null;

    /* eslint @typescript-eslint/dot-notation: "off" */
    readonly permissions: RolePermissionMap = this.route.snapshot.data['permissions'];

    readonly permissionName = PermissionName;

    readonly ruleSeverityOptions = RULE_SEVERITIES;

    readonly ruleSeverity = RuleSeverity;

    readonly tooltipPlacements = PopUpPlacements;

    readonly runtimeEventCursorDirection = RuntimeEventCursorDirection;

    readonly loadStatus = LoadStatus;

    readonly alertColors = KbqAlertColors;

    readonly runtimeContextOptions = RUNTIME_CONTEXT;

    expandThreatsLimit: number | undefined = RUNTIME_DETAILS_LIST_ITEMS_LIMIT;

    expandErrorsLimit: number | undefined = RUNTIME_DETAILS_LIST_ITEMS_LIMIT;

    goToStartPage() {
        this.activeCursor$.next({
            direction: RuntimeEventCursorDirection.RIGHT,
            cursor: this.dateAdapter.today().toJSDate().toISOString() // RFC3339
        });
    }

    goToEndPage() {
        this.activeCursor$.next({
            direction: RuntimeEventCursorDirection.LEFT,
            cursor: this.dateAdapter.today().minus({ years: 10 }).toJSDate().toISOString() // RFC3339
        });
    }

    goToListPage() {
        this.router.navigate([RouterName.RUNTIME, RuntimeRouterName.EVENTS], {
            queryParams: this.getEventContext()
        });
    }

    goToListPageWithContext(context: RuntimeContext) {
        const { execId, parentExecId } = this.getEventContext();
        this.router.navigate([RouterName.RUNTIME, RuntimeRouterName.EVENTS], {
            queryParams: {
                context,
                execId,
                parentExecId
            }
        });
    }

    changePage(direction: RuntimeEventCursorDirection, cursor: string) {
        this.activeCursor$.next({
            direction,
            cursor
        });
    }

    openViewCodeSidepanel(entity: RuntimeEventEntity) {
        const config: KbqSidepanelConfig<RuntimeSidepanelCodeProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                content: JSON.stringify(entity)
            }
        };

        this.sidepanelService
            .open(RuntimeFeatureSidepanelCodeComponent, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe();
    }

    openThreatsSidepanel(threats: RuntimeEventThreat[], errors: RuntimeDetectError[]) {
        const config: KbqSidepanelConfig<RuntimeSidepanelThreatsProps> = {
            position: KbqSidepanelPosition.Right,
            hasBackdrop: true,
            data: {
                threats,
                effectives: [],
                errors
            }
        };

        this.sidepanelService
            .open(RuntimeFeatureSidepanelThreatsComponent, config)
            .afterClosed()
            .pipe(take(1))
            .subscribe();
    }

    expandThreats() {
        this.expandThreatsLimit = undefined;
    }

    expandErrors() {
        this.expandErrorsLimit = undefined;
    }

    private getEventContext(): RuntimeEventContext {
        return {
            context: RuntimeContext.SIBLING,
            execId: this.eventEntity ? this.eventEntity.process.exec_id : '',
            parentExecId: this.eventEntity ? this.eventEntity.process.parent_exec_id : ''
        };
    }
}
