import { DateTime } from 'luxon';
import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output } from '@angular/core';

import { CoreUtilsService as utils } from '@cs/core';

import { RUNTIME_FILTER_LABEL_LIMIT } from '../../constants/runtime-filter.constant';
import { RuntimeEventFilterEntity } from '../../interfaces/runtime-state.interface';
import { RuntimeEventContext, RuntimeEventFilters } from '../../interfaces/runtime-filter.interface';

@Component({
    selector: 'cs-runtime-feature-history-dropdown-component',
    templateUrl: './runtime-history-dropdown.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureHistoryDropdownComponent {
    @Input() history?: RuntimeEventFilterEntity[] | null;

    @Output() historyChange = new EventEmitter<RuntimeEventFilters>();

    readonly selectedFilterLabelLimit = RUNTIME_FILTER_LABEL_LIMIT;

    readonly dateTimeShortFormat = DateTime.DATETIME_SHORT;

    readonly tooltipPlacements = PopUpPlacements;

    select(item: RuntimeEventFilterEntity) {
        const filters = utils.omit<
            RuntimeEventFilterEntity,
            keyof (RuntimeEventContext & Pick<RuntimeEventFilterEntity, 'id'>)
        >(item, ['id', 'context', 'execId', 'parentExecId']) as RuntimeEventFilters;
        this.historyChange.emit(filters);
    }
}
