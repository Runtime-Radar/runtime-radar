import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { ClusterDomainModule } from '@cs/domains/cluster';
import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';

import { ClusterFeatureAccessFormComponent } from './components/access-form/cluster-access-form.component';
import { ClusterFeatureCreateContainer } from './containers/create/cluster-create.container';
import { ClusterFeatureDataBaseFormComponent } from './components/database-form/cluster-database-form.component';
import { ClusterFeatureDeleteUnregisteredModalContainer } from './containers/delete-unregistered-modal/cluster-delete-unregistered-modal.container';
import { ClusterFeatureDetailsContainer } from './containers/details/cluster-details.container';
import { ClusterFeatureEditPopoverComponent } from './components/edit-popover/cluster-edit-popover.component';
import { ClusterFeatureHoursFormatterPipe } from './pipes/cluster-hours-formatter.pipe';
import { ClusterFeatureIngressFormComponent } from './components/ingress-form/cluster-ingress-form.component';
import { ClusterFeatureListContainer } from './containers/list/cluster-list.container';
import { ClusterFeatureMetricFormComponent } from './components/metric-form/cluster-metric-form.component';
import { ClusterFeatureRabbitFormComponent } from './components/rabbit-form/cluster-rabbit-form.component';
import { ClusterFeatureRegistryFormComponent } from './components/registry-form/cluster-registry-form.component';
import { ClusterFeatureRoutingModule } from './cluster-routing.module';
import { ClusterFeatureStepperComponent } from './components/stepper/cluster-stepper.component';

@NgModule({
    imports: [
        ClusterDomainModule,
        ClusterFeatureRoutingModule,
        CommonModule,
        FormsModule,
        I18nModule,
        ReactiveFormsModule,
        SharedModule
    ],
    declarations: [
        ClusterFeatureAccessFormComponent,
        ClusterFeatureCreateContainer,
        ClusterFeatureDataBaseFormComponent,
        ClusterFeatureDeleteUnregisteredModalContainer,
        ClusterFeatureDetailsContainer,
        ClusterFeatureEditPopoverComponent,
        ClusterFeatureHoursFormatterPipe,
        ClusterFeatureIngressFormComponent,
        ClusterFeatureListContainer,
        ClusterFeatureMetricFormComponent,
        ClusterFeatureRabbitFormComponent,
        ClusterFeatureRegistryFormComponent,
        ClusterFeatureStepperComponent
    ]
})
export class ClusterFeatureModule {}
