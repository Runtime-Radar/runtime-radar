import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { I18nModule } from '@cs/i18n';
import { SharedModule } from '@cs/shared';

import { RulePackageSeverityBgColorDirective } from './directives/rule-severity-bgcolor.directive';
import { RulePackageSeverityColorDirective } from './directives/rule-severity-color.directive';
import { RulePackageSeverityComponent } from './components/severity/rule-severity.component';
import { RulePackageSeverityGradeComponent } from './components/severity-grade/rule-severity-grade.component';
import { RulePackageSeverityLabelComponent } from './components/severity-label/rule-severity-label.component';
import { RulePackageSeverityLocalizationPipe } from './pipes/rule-severity.pipe';
import { RulePackageSeverityRadioComponent } from './components/severity-radio/rule-severity-radio.component';
import { RulePackageSidepanelFormComponent } from './components/sidepanel-form/rule-sidepanel-form.component';
import { RulePackageSidepanelInfoComponent } from './components/sidepanel-info/rule-sidepanel-info.component';
import { RulePackageVerdictLocalizationPipe } from './pipes/rule-verdict.pipe';
import { RulePackageVerdictRadioComponent } from './components/verdict-radio/rule-verdict-radio.component';

@NgModule({
    imports: [CommonModule, FormsModule, I18nModule, ReactiveFormsModule, SharedModule],
    declarations: [
        RulePackageSeverityBgColorDirective,
        RulePackageSeverityColorDirective,
        RulePackageSeverityComponent,
        RulePackageSeverityGradeComponent,
        RulePackageSeverityLabelComponent,
        RulePackageSeverityRadioComponent,
        RulePackageSeverityLocalizationPipe,
        RulePackageSidepanelFormComponent,
        RulePackageSidepanelInfoComponent,
        RulePackageVerdictLocalizationPipe,
        RulePackageVerdictRadioComponent
    ],
    exports: [
        RulePackageSeverityBgColorDirective,
        RulePackageSeverityColorDirective,
        RulePackageSeverityComponent,
        RulePackageSeverityGradeComponent,
        RulePackageSeverityLabelComponent,
        RulePackageSeverityRadioComponent,
        RulePackageSeverityLocalizationPipe,
        RulePackageSidepanelFormComponent,
        RulePackageSidepanelInfoComponent,
        RulePackageVerdictLocalizationPipe,
        RulePackageVerdictRadioComponent
    ]
})
export class RulePackageModule {}
