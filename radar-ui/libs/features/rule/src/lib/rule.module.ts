import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { DetectorDomainModule } from '@cs/domains/detector';
import { I18nModule } from '@cs/i18n';
import { NotificationDomainModule } from '@cs/domains/notification';
import { RuleDomainModule } from '@cs/domains/rule';
import { RulePackageModule } from '@cs/packages/rule';
import { SharedModule } from '@cs/shared';

import { RuleFeatureFilterPipe } from './pipes/rule-filter.pipe';
import { RuleFeatureListContainer } from './containers/list/rule-list.container';
import { RuleFeaturePanelFilterComponent } from './components/panel-filter/rule-panel-filter.component';
import { RuleFeatureRoutingModule } from './rule-routing.module';

@NgModule({
    imports: [
        CommonModule,
        DetectorDomainModule,
        FormsModule,
        I18nModule,
        NotificationDomainModule,
        ReactiveFormsModule,
        RuleDomainModule,
        RuleFeatureRoutingModule,
        RulePackageModule,
        SharedModule
    ],
    declarations: [RuleFeatureFilterPipe, RuleFeatureListContainer, RuleFeaturePanelFilterComponent]
})
export class RuleFeatureModule {}
