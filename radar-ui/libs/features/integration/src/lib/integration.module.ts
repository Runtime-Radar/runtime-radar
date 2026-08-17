import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { DetectorDomainModule } from '@cs/domains/detector';
import { I18nModule } from '@cs/i18n';
import { IntegrationDomainModule } from '@cs/domains/integration';
import { NotificationDomainModule } from '@cs/domains/notification';
import { RuleDomainModule } from '@cs/domains/rule';
import { RulePackageModule } from '@cs/packages/rule';
import { SharedModule } from '@cs/shared';

import { IntegrationFeatureCollapseCardContainer } from './containers/collapse-card/integration-collapse-card.container';
import { IntegrationFeatureEmailFormComponent } from './components/email-form/integration-email-form.component';
import { IntegrationFeatureListContainer } from './containers/list/integration-list.container';
import { IntegrationFeatureProtocolCheckboxGroupComponent } from './components/protocol-checkbox-group/integration-protocol-checkbox-group.component';
import { IntegrationFeatureRoutingModule } from './integration-routing.module';
import { IntegrationFeatureSidepanelFormComponent } from './components/sidepanel-form/integration-sidepanel-form.component';
import { IntegrationFeatureSidepanelRecipientFormComponent } from './components/sidepanel-recipient-form/integration-sidepanel-recipient-form.component';
import { IntegrationFeatureSyslogFormComponent } from './components/syslog-form/integration-syslog-form.component';
import { IntegrationFeatureWebhookFormComponent } from './components/webhook-form/integration-webhook-form.component';

@NgModule({
    imports: [
        CommonModule,
        DetectorDomainModule,
        FormsModule,
        I18nModule,
        IntegrationDomainModule,
        IntegrationFeatureRoutingModule,
        NotificationDomainModule,
        ReactiveFormsModule,
        RuleDomainModule,
        RulePackageModule,
        SharedModule
    ],
    declarations: [
        IntegrationFeatureCollapseCardContainer,
        IntegrationFeatureEmailFormComponent,
        IntegrationFeatureListContainer,
        IntegrationFeatureProtocolCheckboxGroupComponent,
        IntegrationFeatureSidepanelFormComponent,
        IntegrationFeatureSidepanelRecipientFormComponent,
        IntegrationFeatureSyslogFormComponent,
        IntegrationFeatureWebhookFormComponent
    ]
})
export class IntegrationFeatureModule {}
