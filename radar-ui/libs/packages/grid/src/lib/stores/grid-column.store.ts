import { ComponentStore } from '@ngrx/component-store';
import { Observable } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import { CoreWindowService, CoreUtilsService as utils } from '@cs/core';

import { GridColumnKey, GridColumnState, GridColumns } from '../interfaces/grid.interface';

@Injectable()
export class GridPackageColumnComponentStore extends ComponentStore<GridColumnState> {
    private readonly coreWindowService = inject(CoreWindowService);

    private initialColumns: Partial<GridColumns> | null = null;

    readonly columns$: Observable<Partial<GridColumns>> = this.select((entities) => entities.columns);

    readonly setInitialState = this.updater((_, data: GridColumnState) => {
        this.initialColumns = { ...data.columns };

        return {
            id: data.id,
            columns: this.getColumnFromStorage(data.id, data.columns) || data.columns
        };
    });

    readonly update = this.updater((state: GridColumnState, key: GridColumnKey) => {
        const columns = {
            ...state.columns,
            [key]: !state.columns[key]
        };

        this.coreWindowService.localStorage.setItem(state.id, JSON.stringify(columns));

        return {
            id: state.id,
            columns
        };
    });

    readonly reset = this.updater((state: GridColumnState) => {
        const columns = this.initialColumns || state.columns;
        this.coreWindowService.localStorage.setItem(state.id, JSON.stringify(columns));

        return {
            id: state.id,
            columns
        };
    });

    constructor() {
        super({} as GridColumnState);
    }

    private getColumnFromStorage(key: string, columns: Partial<GridColumns>): Partial<GridColumns> | null {
        const storage = this.coreWindowService.localStorage.getItem(key);
        if (storage) {
            const data = JSON.parse(storage);
            return utils.isEqual(Object.keys(columns), Object.keys(data)) ? data : null;
        }

        return null;
    }
}
