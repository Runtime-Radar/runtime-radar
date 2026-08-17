import {
    ChangeDetectionStrategy,
    Component,
    ElementRef,
    EventEmitter,
    HostListener,
    Input,
    OnDestroy,
    Output,
    ViewChild,
    booleanAttribute,
    inject,
    signal
} from '@angular/core';

import { RuntimeFeatureFilterContextDropdownService } from './runtime-filter-context-dropdown.service';
import {
    RuntimeEventFilterContextDropdown,
    RuntimeEventFilterContextDropdownType,
    RuntimeEventFilterKey
} from '../../interfaces/runtime-filter.interface';

@Component({
    selector: 'cs-runtime-feature-filter-context-dropdown-component',
    templateUrl: './runtime-filter-context-dropdown.component.html',
    styleUrl: './runtime-filter-context-dropdown.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RuntimeFeatureFilterContextDropdownComponent implements OnDestroy {
    private readonly filterContextDropdownService = inject(RuntimeFeatureFilterContextDropdownService);

    @ViewChild('dropdown', { static: false }) dropdown!: ElementRef<HTMLElement>;

    @Input({ required: true }) key!: RuntimeEventFilterKey;

    @Input() value?: string;

    @Input() constraint?: string | null;

    @Input({ transform: booleanAttribute }) isDropdownAvailable = false;

    @Output() contextChange = new EventEmitter<RuntimeEventFilterContextDropdown>();

    readonly isVisible = signal(false);
    readonly position = signal({
        top: 0,
        left: 0
    });

    readonly dropdownType = RuntimeEventFilterContextDropdownType;

    private readonly close = () => {
        this.isVisible.set(false);
    };

    ngOnDestroy() {
        this.filterContextDropdownService.unregister();
    }

    @HostListener('document:click', ['$event'])
    onDocumentClick(event: MouseEvent) {
        if (!this.isVisible()) {
            return;
        }

        if (this.dropdown && !this.dropdown.nativeElement.contains(event.target as Node)) {
            this.filterContextDropdownService.unregister();
        }
    }

    open(event: MouseEvent) {
        this.filterContextDropdownService.register(this.close);
        this.isVisible.set(true);
        this.position.set({
            top: event.clientY,
            left: event.clientX
        });

        event.preventDefault();
    }

    change(type: RuntimeEventFilterContextDropdownType) {
        this.filterContextDropdownService.unregister();
        this.contextChange.emit({
            type,
            key: this.key,
            value: this.value || ''
        });
    }
}
