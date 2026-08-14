import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, OnChanges, Output, ViewChild } from '@angular/core';

import { RUNTIME_CONTEXT, RuntimeContext } from '@cs/domains/runtime';

import { RUNTIME_FILTER_INITIAL_CONTEXT_STATE } from '../../constants/runtime-filter.constant';
import { RuntimeEventContext } from '../../interfaces/runtime-filter.interface';

@Component({
    selector: 'cs-runtime-feature-context-popover-component',
    templateUrl: './runtime-context-popover.component.html',
    styleUrl: './runtime-context-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureContextPopoverComponent implements OnChanges {
    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input() eventId?: string | null;

    @Input() context?: RuntimeEventContext | null;

    @Output() contextChange = new EventEmitter<RuntimeEventContext>();

    contextValue?: RuntimeContext;

    readonly tooltipPlacements = PopUpPlacements;

    readonly runtimeContextOptions = RUNTIME_CONTEXT;

    ngOnChanges() {
        if (this.context) {
            this.contextValue = this.context.context;
        }
    }

    confirm() {
        this.contextChange.emit({
            context: this.contextValue,
            execId: this.context ? this.context.execId : '',
            parentExecId: this.context ? this.context.parentExecId : '',
            activeContextId: this.contextValue && this.eventId ? this.eventId : ''
        });
        this.popover.hide();
    }

    reset() {
        this.contextChange.emit(RUNTIME_FILTER_INITIAL_CONTEXT_STATE);
        this.popover.hide();
    }

    cancel() {
        this.popover.hide();
    }
}
