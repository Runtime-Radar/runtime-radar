import { Injectable } from '@angular/core';

@Injectable({
    providedIn: 'root'
})
export class RuntimeFeatureFilterContextDropdownService {
    private dropdown: (() => void) | null = null;

    register(closeFn: () => void) {
        if (this.dropdown !== null) {
            this.dropdown();
        }

        this.dropdown = closeFn;
    }

    unregister() {
        if (this.dropdown !== null) {
            this.dropdown();
        }

        this.dropdown = null;
    }
}
