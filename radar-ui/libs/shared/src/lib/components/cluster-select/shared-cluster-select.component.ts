import { ChangeDetectionStrategy, Component, EventEmitter, Input, Output } from '@angular/core';

import { RegisteredCluster } from '@cs/domains/cluster';

@Component({
    selector: 'cs-cluster-select-component',
    templateUrl: './shared-cluster-select.component.html',
    styleUrl: './shared-cluster-select.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedClusterSelectComponent {
    @Input({ required: true }) clusters!: RegisteredCluster[] | null;

    @Input({ required: true }) activeClusterHost!: string | null;

    @Input() testLocator?: string;

    @Output() selectClusterId = new EventEmitter<string>();

    change(id: string) {
        this.selectClusterId.emit(id);
    }
}
