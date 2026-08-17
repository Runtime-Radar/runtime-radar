import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';

import { GridPackageColumnViewIconComponent } from './components/column-view-icon/grid-column-view-icon.component';
import { GridPackageColumnViewPopoverComponent } from './components/column-view-popover/grid-column-view-popover.component';
import { GridPackageColumnVisibleDirective } from './directives/grid-column-visible.directive';

@NgModule({
    imports: [CommonModule, FormsModule, I18nModule, ReactiveFormsModule, SharedModule],
    declarations: [
        GridPackageColumnViewIconComponent,
        GridPackageColumnViewPopoverComponent,
        GridPackageColumnVisibleDirective
    ],
    exports: [
        GridPackageColumnViewIconComponent,
        GridPackageColumnViewPopoverComponent,
        GridPackageColumnVisibleDirective
    ]
})
export class GridPackageModule {}
