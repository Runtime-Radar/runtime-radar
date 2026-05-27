import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { InventoryFeatureMapContainer } from './containers/map/inventory-map.container';

const routes: Routes = [
    {
        path: '',
        component: InventoryFeatureMapContainer,
        data: {
            localizationTitleKey: 'Inventory.MapPage.Header.Title'
        }
    }
];

@NgModule({
    imports: [RouterModule.forChild(routes)],
    exports: [RouterModule]
})
export class InventoryFeatureRoutingModule {}
