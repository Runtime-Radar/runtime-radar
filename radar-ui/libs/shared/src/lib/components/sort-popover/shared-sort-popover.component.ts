import { BehaviorSubject } from 'rxjs';
import { KbqBadgeColors } from '@koobiq/components/badge';
import { KbqPopoverTrigger } from '@koobiq/components/popover';
import { PopUpPlacements } from '@koobiq/components/core';
import {
    ChangeDetectionStrategy,
    Component,
    EventEmitter,
    Input,
    OnInit,
    Output,
    ViewChild,
    inject
} from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';

import { AbstractSorts, FormScheme, SortField, SortKey, CoreUtilsService as utils } from '@cs/core';

import { SortKeyOption } from './shared-sort-popover.interface';

const SORT_KEY: SortKeyOption[] = [
    {
        id: SortKey.NONE,
        localizationKey: 'Common.Pseudo.SortKey.None'
    },
    {
        id: SortKey.ASC,
        localizationKey: 'Common.Pseudo.SortKey.Asc'
    },
    {
        id: SortKey.DESC,
        localizationKey: 'Common.Pseudo.SortKey.Desc'
    }
];

@Component({
    selector: 'cs-sort-popover-component',
    templateUrl: './shared-sort-popover.component.html',
    styleUrl: './shared-sort-popover.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedSortPopoverComponent<T extends AbstractSorts> implements OnInit {
    private readonly formBuilder = inject(FormBuilder);

    @ViewChild('kbqPopover', { static: false }) popover!: KbqPopoverTrigger;

    @Input() id?: string;

    @Input({ required: true }) config!: T;

    @Output() sortChange = new EventEmitter<T>();

    testLocator = this.id ? `${this.id}-` : '';

    sortForm: FormGroup<FormScheme<AbstractSorts>> = this.formBuilder.group({});

    readonly filtersCounter$ = new BehaviorSubject(0);

    readonly isResetIconVisible$ = new BehaviorSubject(false);

    readonly tooltipPlacements = PopUpPlacements;

    readonly badgeColors = KbqBadgeColors;

    readonly sortField = SortField;

    readonly sortKeyOptions = SORT_KEY;

    private originConfig: AbstractSorts = {};

    ngOnInit() {
        this.originConfig = { ...this.config };
        Object.entries(this.config).forEach(([key, value]) => {
            this.sortForm.addControl(key, this.formBuilder.control(value, { nonNullable: true }));
        });

        const values = utils.getFormValues<T>(this.sortForm.controls);
        this.calculateCounterStates(values);
    }

    confirm() {
        const values = utils.getFormValues<T>(this.sortForm.controls);
        this.calculateCounterStates(values);
        this.sortChange.emit(values);
        this.popover.hide();
    }

    reset() {
        this.sortForm.setValue(this.originConfig, { onlySelf: true });
        this.confirm();
    }

    private calculateCounterStates(values: T) {
        this.filtersCounter$.next(
            Object.entries(values).filter(([key, value]) => value !== (this.originConfig[key as SortField] as SortKey))
                .length
        );
        this.isResetIconVisible$.next(!utils.isEqual(this.originConfig, values));
    }
}
