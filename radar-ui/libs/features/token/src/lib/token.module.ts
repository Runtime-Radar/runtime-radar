import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';
import { TokenDomainModule } from '@cs/domains/token';

import { TokenFeatureExpirationColorDirective } from './directives/token-expiration-color.directive';
import { TokenFeatureExpirationLabelPipe } from './pipes/token-expiration-label.pipe';
import { TokenFeatureListContainer } from './containers/list/token-list.container';
import { TokenFeaturePermissionTypePipe } from './pipes/token-permission-type.pipe';
import { TokenFeatureRoutingModule } from './token-routing.module';
import { TokenFeatureSidepanelFormComponent } from './components/sidepanel-form/token-sidepanel-form.component';

@NgModule({
    imports: [
        CommonModule,
        FormsModule,
        I18nModule,
        ReactiveFormsModule,
        SharedModule,
        TokenDomainModule,
        TokenFeatureRoutingModule
    ],
    declarations: [
        TokenFeatureExpirationColorDirective,
        TokenFeatureExpirationLabelPipe,
        TokenFeatureListContainer,
        TokenFeaturePermissionTypePipe,
        TokenFeatureSidepanelFormComponent
    ]
})
export class TokenFeatureModule {}
