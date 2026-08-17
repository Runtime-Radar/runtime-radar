import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { DetectorDomainModule } from '@cs/domains/detector';
import { GridPackageModule } from '@cs/packages/grid';
import { I18nModule } from '@cs/i18n';
import { NotificationDomainModule } from '@cs/domains/notification';
import { RuleDomainModule } from '@cs/domains/rule';
import { RulePackageModule } from '@cs/packages/rule';
import { RuntimeDomainModule } from '@cs/domains/runtime';
import { SharedModule } from '@cs/shared';

import { RuntimeFeatureByteFormatterPipe } from './pipes/runtime-byte-formatter.pipe';
import { RuntimeFeatureContextPopoverComponent } from './components/context-popover/runtime-context-popover.component';
import { RuntimeFeatureDateTimePeriodPickerComponent } from './components/datetime-period-picker/runtime-datetime-period-picker.component';
import { RuntimeFeatureDetailsContainer } from './containers/details/runtime-details.container';
import { RuntimeFeatureDetectorsContainer } from './containers/detectors/runtime-detectors.container';
import { RuntimeFeatureEventCounterComponent } from './components/event-counter/runtime-event-counter.component';
import { RuntimeFeatureEventTypeIconDirective } from './directives/runtime-event-type-icon.directive';
import { RuntimeFeatureEventsContainer } from './containers/events/runtime-events.container';
import { RuntimeFeatureEventsGridContainer } from './containers/events-grid/runtime-events-grid.container';
import { RuntimeFeatureFilterContextDropdownComponent } from './components/filter-context-dropdown/runtime-filter-context-dropdown.component';
import { RuntimeFeatureFilterPopoverComponent } from './components/filter-popover/runtime-filter-popover.component';
import { RuntimeFeatureHighlightReasonFragmentDirective } from './directives/runtime-highlight-reason-fragment.directive';
import { RuntimeFeatureHistoryDropdownComponent } from './components/history-dropdown/runtime-history-dropdown.component';
import { RuntimeFeatureHistoryLabelPipe } from './pipes/runtime-history-label.pipe';
import { RuntimeFeatureMitreTacticComponent } from './components/mitre-tactic/runtime-mitre-tactic.component';
import { RuntimeFeatureNanosecondsFormatterPipe } from './pipes/runtime-nanoseconds-formatter.pipe';
import { RuntimeFeatureOverlayDirective } from './directives/runtime-overlay.directive';
import { RuntimeFeaturePermissionsFilterPipe } from './pipes/runtime-permissions-filter.pipe';
import { RuntimeFeaturePresetDropdownComponent } from './components/preset-dropdown/runtime-preset-dropdown.component';
import { RuntimeFeatureRoutingModule } from './runtime-routing.module';
import { RuntimeFeatureRulesContainer } from './containers/rules/runtime-rules.container';
import { RuntimeFeatureSettingsContainer } from './containers/settings/runtime-settings.container';
import { RuntimeFeatureSeverityThreatsCounterPipe } from './pipes/runtime-severity-threats-counter.pipe';
import { RuntimeFeatureSidepanelCodeComponent } from './components/sidepanel-code/runtime-sidepanel-code.component';
import { RuntimeFeatureSidepanelDetectorComponent } from './components/sidepanel-detector/runtime-sidepanel-detector.component';
import { RuntimeFeatureSidepanelIncidentComponent } from './components/sidepanel-incident/runtime-sidepanel-incident.component';
import { RuntimeFeatureSidepanelPermissionFormComponent } from './components/sidepanel-permission-form/runtime-sidepanel-permission-form.component';
import { RuntimeFeatureSidepanelPolicyComponent } from './components/sidepanel-policy/runtime-sidepanel-policy.component';
import { RuntimeFeatureSidepanelPolicyFormComponent } from './components/sidepanel-policy-form/runtime-sidepanel-policy-form.component';
import { RuntimeFeatureSidepanelThreatsComponent } from './components/sidepanel-threats/runtime-sidepanel-threats.component';
import { RuntimeFeatureUploadDetectorModalComponent } from './components/upload-detector-modal/runtime-upload-detector-modal.component';

@NgModule({
    imports: [
        CommonModule,
        DetectorDomainModule,
        FormsModule,
        GridPackageModule,
        I18nModule,
        NotificationDomainModule,
        ReactiveFormsModule,
        RuleDomainModule,
        RulePackageModule,
        RuntimeDomainModule,
        RuntimeFeatureRoutingModule,
        SharedModule
    ],
    declarations: [
        RuntimeFeatureByteFormatterPipe,
        RuntimeFeatureContextPopoverComponent,
        RuntimeFeatureDateTimePeriodPickerComponent,
        RuntimeFeatureDetailsContainer,
        RuntimeFeatureDetectorsContainer,
        RuntimeFeatureEventCounterComponent,
        RuntimeFeatureEventTypeIconDirective,
        RuntimeFeatureEventsContainer,
        RuntimeFeatureEventsGridContainer,
        RuntimeFeatureFilterContextDropdownComponent,
        RuntimeFeatureFilterPopoverComponent,
        RuntimeFeatureHistoryDropdownComponent,
        RuntimeFeatureHistoryLabelPipe,
        RuntimeFeatureHighlightReasonFragmentDirective,
        RuntimeFeatureMitreTacticComponent,
        RuntimeFeatureNanosecondsFormatterPipe,
        RuntimeFeatureOverlayDirective,
        RuntimeFeaturePermissionsFilterPipe,
        RuntimeFeaturePresetDropdownComponent,
        RuntimeFeatureRulesContainer,
        RuntimeFeatureSettingsContainer,
        RuntimeFeatureSeverityThreatsCounterPipe,
        RuntimeFeatureSidepanelCodeComponent,
        RuntimeFeatureSidepanelDetectorComponent,
        RuntimeFeatureSidepanelIncidentComponent,
        RuntimeFeatureSidepanelPermissionFormComponent,
        RuntimeFeatureSidepanelPolicyComponent,
        RuntimeFeatureSidepanelPolicyFormComponent,
        RuntimeFeatureSidepanelThreatsComponent,
        RuntimeFeatureUploadDetectorModalComponent
    ]
})
export class RuntimeFeatureModule {}
