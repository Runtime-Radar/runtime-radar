import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';
import { UserDomainModule } from '@cs/domains/user';

import { UserFeatureListContainer } from './containers/list/user-list.container';
import { UserFeatureRoutingModule } from './user-routing.module';
import { UserFeatureSidepanelUserFormComponent } from './components/sidepanel-user-form/user-sidepanel-user-form.component';

@NgModule({
    imports: [
        CommonModule,
        FormsModule,
        I18nModule,
        ReactiveFormsModule,
        SharedModule,
        UserDomainModule,
        UserFeatureRoutingModule
    ],
    declarations: [UserFeatureListContainer, UserFeatureSidepanelUserFormComponent]
})
export class UserFeatureModule {}
