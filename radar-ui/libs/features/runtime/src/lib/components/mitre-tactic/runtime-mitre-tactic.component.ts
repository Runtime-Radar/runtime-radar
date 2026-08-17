import { KbqBadgeColors } from '@koobiq/components/badge';
import { ChangeDetectionStrategy, Component, Input, booleanAttribute, inject } from '@angular/core';
import { Observable, switchMap } from 'rxjs';

import { I18nService } from '@cs/i18n';
import { MitreStoreService } from '@cs/domains/mitre';

@Component({
    selector: 'cs-runtime-feature-mitre-tactic-component',
    templateUrl: './runtime-mitre-tactic.component.html',
    styleUrl: './runtime-mitre-tactic.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureMitreTacticComponent {
    private readonly i18nService = inject(I18nService);
    private readonly mitreStoreService = inject(MitreStoreService);

    @Input({ required: true }) id!: string;

    @Input() tacticId?: string;

    @Input() techniques?: string[];

    @Input({ transform: booleanAttribute }) isOutline = true;

    @Input({ transform: booleanAttribute }) isCompact = false;

    readonly tactic$ = (id: string): Observable<string> =>
        this.i18nService.locale$.pipe(switchMap((locale) => this.mitreStoreService.tactic$(id, locale)));

    readonly technique$ = (id: string, tacticId: string): Observable<string> =>
        this.i18nService.locale$.pipe(switchMap((locale) => this.mitreStoreService.technique$(id, tacticId, locale)));

    readonly badgeColors = KbqBadgeColors;
}
