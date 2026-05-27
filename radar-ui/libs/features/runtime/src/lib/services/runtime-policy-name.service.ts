import { Injectable } from '@angular/core';

@Injectable({
    providedIn: 'root'
})
export class RuntimeFeaturePolicyNameService {
    private readonly names: string[] = [];

    get(): string[] {
        return this.names;
    }

    set(key: string) {
        this.names.push(key);
    }

    setAll(keys: string[]) {
        this.clear();
        this.names.push(...keys);
    }

    clear() {
        this.names.splice(0, this.names.length);
    }

    replace(key: string, oldKey: string) {
        const index = this.names.indexOf(oldKey);
        if (index !== -1) {
            this.names.splice(index, 1, key);
        }
    }

    remove(key: string) {
        const index = this.names.indexOf(key);
        if (index !== -1) {
            this.names.splice(index, 1);
        }
    }
}
