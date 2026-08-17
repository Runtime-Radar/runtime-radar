import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { PopUpPlacements } from '@koobiq/components/core';
import { ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output, ViewChild } from '@angular/core';

import { Cluster } from '@cs/domains/cluster';

import { ClusterEditPopoverOutputs } from '../../interfaces/cluster-popover.interface';

@Component({
    selector: 'cs-cluster-feature-edit-popover-component',
    templateUrl: './cluster-edit-popover.component.html',
    styleUrl: './cluster-edit-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ClusterFeatureEditPopoverComponent implements OnInit {
    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input({ required: true }) cluster!: Cluster;

    @Input() disabled?: boolean;

    @Output() nameChange = new EventEmitter<ClusterEditPopoverOutputs>();

    name = '';

    readonly tooltipPlacements = PopUpPlacements;

    ngOnInit() {
        this.name = this.cluster.name;
    }

    confirm() {
        this.popover.hide();
        this.nameChange.emit({
            id: this.cluster.id,
            name: this.name,
            config: this.cluster.config
        });
    }

    cancel() {
        this.popover.hide();
    }
}
