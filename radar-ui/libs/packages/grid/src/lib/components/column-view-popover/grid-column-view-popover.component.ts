import { BehaviorSubject } from 'rxjs';
import { KbqBadgeColors } from '@koobiq/components/badge';
import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { PopUpPlacements } from '@koobiq/components/core';
import {
    ChangeDetectionStrategy,
    Component,
    EventEmitter,
    Input,
    OnChanges,
    Output,
    SimpleChanges,
    ViewChild
} from '@angular/core';

import { GridColumnKey, GridColumnOption, GridColumns } from '../../interfaces/grid.interface';

@Component({
    selector: 'cs-grid-package-column-view-popover-component',
    templateUrl: './grid-column-view-popover.component.html',
    styleUrl: './grid-column-view-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class GridPackageColumnViewPopoverComponent implements OnChanges {
    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input({ required: true }) options!: GridColumnOption[];

    @Input({ required: true }) columns!: Partial<GridColumns>;

    @Input({ required: true }) localizationFn!: any;

    @Output() checkboxChange = new EventEmitter<GridColumnKey>();

    @Output() checkboxesReset = new EventEmitter();

    readonly hiddenColumnsCounter$ = new BehaviorSubject(0);

    readonly tooltipPlacements = PopUpPlacements;

    readonly badgeColors = KbqBadgeColors;

    ngOnChanges(changes: SimpleChanges) {
        const state: Partial<GridColumns> = changes['columns'].currentValue;
        this.hiddenColumnsCounter$.next(Object.values(state).filter((item) => !item).length);
    }
}
