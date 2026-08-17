import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';

import { SignInFeaturePageContainer } from './containers/page/sign-in-page.container';
import { SignInFeatureRoutingModule } from './sign-in-routing.module';

@NgModule({
    imports: [CommonModule, FormsModule, I18nModule, ReactiveFormsModule, SharedModule, SignInFeatureRoutingModule],
    declarations: [SignInFeaturePageContainer]
})
export class SignInFeatureModule {}
