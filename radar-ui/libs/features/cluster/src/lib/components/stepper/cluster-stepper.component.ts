import { KbqTabChangeEvent } from '@koobiq/components/tabs';
import {
    ChangeDetectionStrategy,
    Component,
    ContentChildren,
    EventEmitter,
    Input,
    Output,
    QueryList,
    TemplateRef,
    computed,
    signal
} from '@angular/core';

import { CLUSTER_STEPPER_TABS } from '../../constants/cluster-stepper.constant';
import { ClusterStepName } from '../../interfaces/cluster-stepper.interface';

@Component({
    selector: 'cs-cluster-feature-stepper-component',
    templateUrl: './cluster-stepper.component.html',
    styleUrl: './cluster-stepper.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureStepperComponent {
    @ContentChildren('step', { read: TemplateRef }) steps!: QueryList<TemplateRef<never>>;

    @Input() isStepValid: boolean | null = true;

    @Input() set step(step: ClusterStepName | null) {
        if (step) {
            this.activeTabIndex$.set(this.clusterStepperTabs.findIndex((tab) => tab.id === step));
        }
    }

    @Output() stepChange = new EventEmitter<ClusterStepName>();

    @Output() completeChange = new EventEmitter<void>();

    readonly clusterStepperTabs = CLUSTER_STEPPER_TABS;

    readonly activeTabIndex$ = signal(0);

    readonly isFirstStep$ = computed(() => this.activeTabIndex$() === 0);

    readonly isLastStep$ = computed(() => this.activeTabIndex$() === this.clusterStepperTabs.length - 1);

    onSelectedTabChange(tab: KbqTabChangeEvent) {
        const id = this.clusterStepperTabs.at(tab.index)?.id;
        if (id) {
            this.activeTabIndex$.set(tab.index);
            this.stepChange.emit(id);
        }
    }

    prev() {
        if (!this.isFirstStep$()) {
            this.activeTabIndex$.update((current) => current - 1);
        }
    }

    next() {
        if (!this.isLastStep$()) {
            this.activeTabIndex$.update((current) => current + 1);
        }
    }

    complete() {
        this.completeChange.emit();
    }
}
