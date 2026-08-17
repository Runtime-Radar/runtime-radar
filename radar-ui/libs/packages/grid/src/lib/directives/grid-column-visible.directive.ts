import { Directive, HostBinding, Input } from '@angular/core';

@Directive({
    selector: '[columnVisible]',
    standalone: false
})
export class GridPackageColumnVisibleDirective {
    @HostBinding('style.display') display = '';

    @Input() set columnVisible(value: boolean | null | undefined) {
        this.display = value === false ? 'none' : '';
    }
}
