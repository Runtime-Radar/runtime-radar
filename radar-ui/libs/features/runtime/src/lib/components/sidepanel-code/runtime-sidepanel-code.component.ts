import { KBQ_SIDEPANEL_DATA } from '@koobiq/components/sidepanel';
import { KbqCodeBlockFile } from '@koobiq/components/code-block';
import { ChangeDetectionStrategy, Component, OnInit, inject } from '@angular/core';

import { RuntimeSidepanelCodeProps } from '../../interfaces/runtime-sidepanel.interface';

const RUNTIME_CODE_SPACE_INDENT = 4;

@Component({
    templateUrl: './runtime-sidepanel-code.component.html',
    styleUrl: './runtime-sidepanel-code.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureSidepanelCodeComponent implements OnInit {
    readonly props = inject<RuntimeSidepanelCodeProps>(KBQ_SIDEPANEL_DATA);

    files: KbqCodeBlockFile[] = [];

    ngOnInit() {
        this.files.push({
            filename: 'event.json',
            content: this.format(this.props.content),
            language: 'json'
        });
    }

    private format(content: string) {
        return JSON.stringify(JSON.parse(content), null, RUNTIME_CODE_SPACE_INDENT);
    }
}
