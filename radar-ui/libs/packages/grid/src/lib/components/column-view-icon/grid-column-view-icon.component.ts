import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output } from '@angular/core';

import { GridColumnKey } from '../../interfaces/grid.interface';

@Component({
    selector: 'cs-grid-package-column-view-icon-component',
    templateUrl: './grid-column-view-icon.component.html',
    styleUrl: './grid-column-view-icon.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class GridPackageColumnViewIconComponent {
    @Input({ required: true }) id!: GridColumnKey;

    @Output() viewChange = new EventEmitter<GridColumnKey>();
}
