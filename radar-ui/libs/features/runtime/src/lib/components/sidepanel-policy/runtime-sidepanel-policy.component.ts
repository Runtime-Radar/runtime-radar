import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { KbqCodeBlockFile } from '@koobiq/components/code-block';
import { ChangeDetectionStrategy, Component, OnInit, inject } from '@angular/core';

import { RuntimeSidepanelPolicyProps } from '../../interfaces/runtime-sidepanel.interface';

@Component({
    templateUrl: './runtime-sidepanel-policy.component.html',
    styleUrl: './runtime-sidepanel-policy.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureSidepanelPolicyComponent implements OnInit {
    readonly props = inject<Partial<RuntimeSidepanelPolicyProps>>(KBQ_SIDEPANEL_DATA);

    files: KbqCodeBlockFile[] = [];

    ngOnInit() {
        if (this.props.yaml) {
            this.files.push({
                filename: 'source.yaml',
                content: this.props.yaml,
                language: 'yaml'
            });
        }
    }
}
