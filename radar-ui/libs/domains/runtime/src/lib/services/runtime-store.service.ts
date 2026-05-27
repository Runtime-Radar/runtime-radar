import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Store } from '@ngrx/store';

import {
    CHECK_RUNTIME_CHANGES_TODO_ACTION,
    CREATE_RUNTIME_CONFIG_TODO_ACTION,
    HIDE_RUNTIME_OVERLAY_TODO_ACTION,
    RESET_RUNTIME_CONFIG_TODO_ACTION,
    SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION
} from '../stores/runtime-action.store';
import {
    RuntimeEventProcessorHistoryControl,
    RuntimeMonitorConfig,
    RuntimeMonitorConfigExtended,
    RuntimeMonitorConfigStatus,
    RuntimeState
} from '../interfaces';
import {
    getRuntimeEventProcessorHistoryControl,
    getRuntimeGrafanaUrl,
    getRuntimeHasChanges,
    getRuntimeHasPoliciesChanges,
    getRuntimeIsExpertMode,
    getRuntimeIsOverlayed,
    getRuntimeMonitorConfig,
    getRuntimeMonitorConfigStatus
} from '../stores/runtime-selector.store';

@Injectable({
    providedIn: 'root'
})
export class RuntimeStoreService {
    readonly runtimeGrafanaUrl$: Observable<string> = this.store.select(getRuntimeGrafanaUrl);

    readonly runtimeMonitorConfig$: Observable<RuntimeMonitorConfig> = this.store.select(getRuntimeMonitorConfig);

    readonly runtimeMonitorConfigStatus$: Observable<RuntimeMonitorConfigStatus | undefined> =
        this.store.select(getRuntimeMonitorConfigStatus);

    readonly eventProcessorHistoryControl$: Observable<RuntimeEventProcessorHistoryControl | undefined> =
        this.store.select(getRuntimeEventProcessorHistoryControl);

    readonly runtimeHasChanges$: Observable<boolean> = this.store.select(getRuntimeHasChanges);

    readonly runtimeHasPoliciesChanges$: Observable<boolean> = this.store.select(getRuntimeHasPoliciesChanges);

    readonly runtimeIsExpertMode$: Observable<boolean> = this.store.select(getRuntimeIsExpertMode);

    readonly runtimeIsOverlayed$: Observable<boolean> = this.store.select(getRuntimeIsOverlayed);

    constructor(private readonly store: Store<RuntimeState>) {}

    createConfig(config: RuntimeMonitorConfig, historyControl: RuntimeEventProcessorHistoryControl) {
        this.store.dispatch(
            CREATE_RUNTIME_CONFIG_TODO_ACTION({
                config,
                historyControl
            })
        );
    }

    resetConfig() {
        this.store.dispatch(RESET_RUNTIME_CONFIG_TODO_ACTION());
    }

    checkChanges(config: RuntimeMonitorConfigExtended) {
        this.store.dispatch(CHECK_RUNTIME_CHANGES_TODO_ACTION({ config }));
    }

    switchExpertMode() {
        this.store.dispatch(SWITCH_RUNTIME_EXPERT_MODE_TODO_ACTION());
    }

    hideOverlay() {
        this.store.dispatch(HIDE_RUNTIME_OVERLAY_TODO_ACTION());
    }
}
