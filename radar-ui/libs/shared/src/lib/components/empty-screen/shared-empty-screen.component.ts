import { ChangeDetectionStrategy, Component, Input, booleanAttribute } from '@angular/core';

@Component({
    selector: 'cs-empty-screen-component',
    templateUrl: './shared-empty-screen.component.html',
    styleUrl: './shared-empty-screen.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedEmptyScreenComponent {
    @Input({ required: true }) title!: string;

    @Input({ required: true }) description!: string;

    @Input({ transform: booleanAttribute }) isCustomDescription = false;

    @Input() size?: string; // small

    @Input() imageUrl?: string;
}
