import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { I18nModule } from '@cs/i18n';
import { KubeManagerDomainModule } from '@cs/domains/kube-manager';
import { RulePackageModule } from '@cs/packages/rule';
import { SharedModule } from '@cs/shared';

import { InventoryFeatureDetectorTableComponent } from './components/detector-table/inventory-detector-table.component';
import { InventoryFeatureDragAreaDirective } from './directives/inventory-drag-area.directive';
import { InventoryFeatureLegendComponent } from './components/legend/inventory-legend.component';
import { InventoryFeatureMapContainer } from './containers/map/inventory-map.container';
import { InventoryFeatureNamespaceComponent } from './components/namespace/inventory-namespace.component';
import { InventoryFeaturePodComponent } from './components/pod/inventory-pod.component';
import { InventoryFeatureRoutingModule } from './inventory-routing.module';
import { InventoryFeatureSidepanelContainerComponent } from './components/sidepanel-container/inventory-sidepanel-container.component';
import { InventoryFeatureSidepanelNamespaceComponent } from './components/sidepanel-namespace/inventory-sidepanel-namespace.component';
import { InventoryFeatureSidepanelNodeContainer } from './containers/sidepanel-node/inventory-sidepanel-node.container';
import { InventoryFeatureSidepanelPodComponent } from './components/sidepanel-pod/inventory-sidepanel-pod.component';

@NgModule({
    imports: [
        CommonModule,
        FormsModule,
        I18nModule,
        KubeManagerDomainModule,
        ReactiveFormsModule,
        RulePackageModule,
        SharedModule,
        InventoryFeatureRoutingModule
    ],
    declarations: [
        InventoryFeatureDetectorTableComponent,
        InventoryFeatureDragAreaDirective,
        InventoryFeatureLegendComponent,
        InventoryFeatureMapContainer,
        InventoryFeatureNamespaceComponent,
        InventoryFeaturePodComponent,
        InventoryFeatureSidepanelContainerComponent,
        InventoryFeatureSidepanelNamespaceComponent,
        InventoryFeatureSidepanelNodeContainer,
        InventoryFeatureSidepanelPodComponent
    ]
})
export class InventoryFeatureModule {}
