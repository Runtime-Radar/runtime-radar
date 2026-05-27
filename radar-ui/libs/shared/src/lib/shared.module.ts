import { LetDirective } from '@ngrx/component';
import { NgModule } from '@angular/core';
import { CommonModule, NgOptimizedImage } from '@angular/common';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { KbqAccordionModule } from '@koobiq/components/accordion';
import { KbqAlertModule } from '@koobiq/components/alert';
import { KbqBadgeModule } from '@koobiq/components/badge';
import { KbqButtonModule } from '@koobiq/components/button';
import { KbqButtonToggleModule } from '@koobiq/components/button-toggle';
import { KbqCheckboxModule } from '@koobiq/components/checkbox';
import { KbqCodeBlockModule } from '@koobiq/components/code-block';
import { KbqDatepickerModule } from '@koobiq/components/datepicker';
import { KbqDividerModule } from '@koobiq/components/divider';
import { KbqDlModule } from '@koobiq/components/dl';
import { KbqDropdownModule } from '@koobiq/components/dropdown';
import { KbqEmptyStateModule } from '@koobiq/components/empty-state';
import { KbqFormFieldModule } from '@koobiq/components/form-field';
import { KbqIconModule } from '@koobiq/components/icon';
import { KbqInputModule } from '@koobiq/components/input';
import { KbqLinkModule } from '@koobiq/components/link';
import { KbqLoaderOverlayModule } from '@koobiq/components/loader-overlay';
import { KbqPopoverModule } from '@koobiq/components/popover';
import { KbqProgressBarModule } from '@koobiq/components/progress-bar';
import { KbqProgressSpinnerModule } from '@koobiq/components/progress-spinner';
import { KbqRadioModule } from '@koobiq/components/radio';
import { KbqSelectModule } from '@koobiq/components/select';
import { KbqTableModule } from '@koobiq/components/table';
import { KbqTabsModule } from '@koobiq/components/tabs';
import { KbqTagsModule } from '@koobiq/components/tags';
import { KbqTextareaModule } from '@koobiq/components/textarea';
import { KbqTimepickerModule } from '@koobiq/components/timepicker';
import { KbqToggleModule } from '@koobiq/components/toggle';
import { KbqToolTipModule } from '@koobiq/components/tooltip';
import { KbqTreeModule } from '@koobiq/components/tree';
import { KbqTreeSelectModule } from '@koobiq/components/tree-select';
import {
    KBQ_TOAST_CONFIG,
    KbqToastConfig,
    KbqToastModule,
    KbqToastPosition,
    KbqToastService
} from '@koobiq/components/toast';
import { KbqHighlightModule, KbqPseudoCheckboxModule } from '@koobiq/components/core';
import { KbqModalModule, KbqModalService } from '@koobiq/components/modal';
import { KbqSidepanelModule, KbqSidepanelService } from '@koobiq/components/sidepanel';

import { I18nModule } from '@cs/i18n';

import { SharedClipboardComponent } from './components/clipboard/shared-clipboard.component';
import { SharedClusterSelectComponent } from './components/cluster-select/shared-cluster-select.component';
import { SharedDateFormatterPipe } from './pipes/date-formatter/shared-date-formatter.pipe';
import { SharedDatetimePickerComponent } from './components/datetime-picker/shared-datetime-picker.component';
import { SharedDetectorTreeSelectComponent } from './components/detector-tree-select/shared-detector-tree-select.component';
import { SharedEmptyScreenComponent } from './components/empty-screen/shared-empty-screen.component';
import { SharedEventActionComponent } from './components/event-action/shared-event-action.component';
import { SharedHoursFormatterPipe } from './pipes/hours-formatter/shared-hours-formatter.pipe';
import { SharedMultipleSelectComponent } from './components/multiple-select/shared-multiple-select.component';
import { SharedNumberFormatterPipe } from './pipes/number-formatter/shared-number-formatter.pipe';
import { SharedOverlayDirective } from './directives/overlay/shared-overlay.directive';
import { SharedPaginatorComponent } from './components/paginator/shared-paginator.component';
import { SharedPasswordModalComponent } from './components/password-modal/password-modal.component';
import { SharedRuleSidepanelComponent } from './components/rule-sidepanel/shared-rule-sidepanel.component';
import { SharedRuleSidepanelFormComponent } from './components/rule-sidepanel-form/shared-rule-sidepanel-form.component';
import { SharedSeverityBgColorDirective } from './directives/severity-bgcolor/shared-severity-bgcolor.directive';
import { SharedSeverityColorDirective } from './directives/severity-color/shared-severity-color.directive';
import { SharedSeverityComponent } from './components/severity/shared-severity.component';
import { SharedSeverityLabelComponent } from './components/severity-label/shared-severity-label.component';
import { SharedSeverityLocalizationPipe } from './pipes/severity/shared-severity.pipe';
import { SharedSeverityRadioComponent } from './components/severity-radio/shared-severity-radio.component';
import { SharedTabsComponent } from './components/tabs/shared-tabs.component';
import { SharedTagInputComponent } from './components/tag-input/shared-tag-input.component';
import { SharedVerdictLocalizationPipe } from './pipes/verdict/shared-verdict.pipe';
import { sharedCodeBlockLocalizationProvider } from './providers/shared-code-block-localization.provider';

const KOOBIQ_TOAST_CONFIG: KbqToastConfig = {
    position: KbqToastPosition.TOP_RIGHT,
    duration: 3000,
    delay: 2000,
    onTop: true,
    indent: {
        vertical: 16,
        horizontal: 16
    }
};

const koobiqModules = [
    KbqAccordionModule,
    KbqAlertModule,
    KbqBadgeModule,
    KbqButtonModule,
    KbqButtonToggleModule,
    KbqCheckboxModule,
    KbqCodeBlockModule,
    KbqDatepickerModule,
    KbqDividerModule,
    KbqDlModule,
    KbqDropdownModule,
    KbqHighlightModule,
    KbqFormFieldModule,
    KbqIconModule,
    KbqInputModule,
    KbqLinkModule,
    KbqLoaderOverlayModule,
    KbqModalModule,
    KbqPopoverModule,
    KbqProgressBarModule,
    KbqProgressSpinnerModule,
    KbqPseudoCheckboxModule,
    KbqRadioModule,
    KbqSidepanelModule,
    KbqSelectModule,
    KbqTableModule,
    KbqTabsModule,
    KbqTagsModule,
    KbqTextareaModule,
    KbqTimepickerModule,
    KbqToggleModule,
    KbqToolTipModule,
    KbqToastModule,
    KbqTreeModule,
    KbqTreeSelectModule
];

const koobiqServices = [KbqModalService, KbqSidepanelService, KbqToastService];

const koobiqImports = [
    KbqAlertModule,
    KbqBadgeModule,
    KbqButtonModule,
    KbqCheckboxModule,
    KbqDatepickerModule,
    KbqDividerModule,
    KbqEmptyStateModule,
    KbqHighlightModule,
    KbqFormFieldModule,
    KbqIconModule,
    KbqInputModule,
    KbqModalModule,
    KbqProgressBarModule,
    KbqSidepanelModule,
    KbqSelectModule,
    KbqTabsModule,
    KbqTagsModule,
    KbqTimepickerModule,
    KbqToolTipModule,
    KbqTreeModule,
    KbqTreeSelectModule
];

const appDeclarations = [
    SharedClusterSelectComponent,
    SharedClipboardComponent,
    SharedDateFormatterPipe,
    SharedDatetimePickerComponent,
    SharedDetectorTreeSelectComponent,
    SharedEmptyScreenComponent,
    SharedEventActionComponent,
    SharedHoursFormatterPipe,
    SharedMultipleSelectComponent,
    SharedNumberFormatterPipe,
    SharedOverlayDirective,
    SharedPaginatorComponent,
    SharedPasswordModalComponent,
    SharedRuleSidepanelComponent,
    SharedRuleSidepanelFormComponent,
    SharedSeverityComponent,
    SharedSeverityBgColorDirective,
    SharedSeverityColorDirective,
    SharedSeverityRadioComponent,
    SharedSeverityLocalizationPipe,
    SharedSeverityLabelComponent,
    SharedVerdictLocalizationPipe,
    SharedTabsComponent,
    SharedTagInputComponent
];

@NgModule({
    imports: [
        ...koobiqImports,
        CommonModule,
        FormsModule,
        ReactiveFormsModule,
        NgOptimizedImage,
        I18nModule,
        LetDirective
    ],
    declarations: [...appDeclarations],
    providers: [
        ...koobiqServices,
        ...sharedCodeBlockLocalizationProvider(),
        {
            provide: KBQ_TOAST_CONFIG,
            useValue: KOOBIQ_TOAST_CONFIG
        }
    ],
    exports: [...koobiqModules, ...appDeclarations, NgOptimizedImage, LetDirective]
})
export class SharedModule {}
