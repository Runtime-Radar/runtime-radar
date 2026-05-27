import { BehaviorSubject } from 'rxjs';
import { Injectable } from '@angular/core';

import { KbqSidepanelService } from '@koobiq/components/sidepanel';

import { INVENTORY_CONTEXT_WEIGHT } from '../constants/inventory-scheme.constant';
import { InventorySidepanelContext, InventorySidepanelContextType } from '../interfaces/inventory-sidepanel.interface';

export const DEFAULT_CONTEXT: InventorySidepanelContext = {
    id: '',
    sidepanelId: '',
    path: '',
    type: InventorySidepanelContextType.NONE
};

@Injectable({
    providedIn: 'root'
})
export class InventoryFeatureSidepanelContextService {
    private context: InventorySidepanelContext[] = [];

    readonly context$ = new BehaviorSubject<InventorySidepanelContext>(DEFAULT_CONTEXT);

    constructor(private readonly sidepanelService: KbqSidepanelService) {}

    get(): InventorySidepanelContext[] {
        return this.context;
    }

    set(item: InventorySidepanelContext) {
        this.context.push(item);
        this.context$.next(item);
    }

    remove(id: string) {
        const ctx = this.context.find((item) => item.id === id);
        if (ctx) {
            this.context.pop();
            this.context$.next(this.context.at(-1) || DEFAULT_CONTEXT);
        }
    }

    slice(item: InventorySidepanelContext) {
        const ids: string[] = [];
        for (let i = this.context.length - 1; i >= 0; i--) {
            const context = this.context[i];
            const curr = INVENTORY_CONTEXT_WEIGHT.get(item.type) || -1;
            const prev = INVENTORY_CONTEXT_WEIGHT.get(context.type) || -1;

            if (
                (curr === prev && item.sidepanelId !== context.sidepanelId) ||
                ((curr < prev || curr > prev) && item.path.indexOf(context.path) < 0)
            ) {
                ids.push(context.id);
                this.closeSidepanel(context.sidepanelId);
            }
        }

        this.context$.next(item);
        const entities = this.context.some((ctx) => ctx.id === item.id) ? [...this.context] : [...this.context, item];
        this.context = entities.filter((ctx) => !ids.includes(ctx.id));
    }

    private closeSidepanel(id: string) {
        const panel = this.sidepanelService.getSidepanelById(id);
        if (panel) {
            panel.close();
        } else {
            console.warn('sidepanel must be provided');
        }
    }
}
