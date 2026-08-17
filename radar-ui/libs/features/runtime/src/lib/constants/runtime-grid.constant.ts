import { GridColumnKey, GridColumnOption, GridColumns } from '@cs/packages/grid';

export const RUNTIME_EVENT_GRID_COLUMNS: GridColumnOption[] = [
    {
        id: GridColumnKey.FUNCTION,
        localizationKey: 'Runtime.EventsPage.Table.Function'
    },
    {
        id: GridColumnKey.BINARY,
        localizationKey: 'Runtime.EventsPage.Table.Binary'
    },
    {
        id: GridColumnKey.ARGUMENT,
        localizationKey: 'Runtime.EventsPage.Table.Argument'
    },
    {
        id: GridColumnKey.SEVERITY,
        localizationKey: 'Runtime.EventsPage.Table.Severity'
    },
    {
        id: GridColumnKey.DATE_FROM,
        localizationKey: 'Runtime.EventsPage.Table.Date'
    }
];

export const RUNTIME_EVENT_GRID_COLUMNS_INITIAL_VALUES: Partial<GridColumns> = {
    [GridColumnKey.FUNCTION]: true,
    [GridColumnKey.BINARY]: true,
    [GridColumnKey.ARGUMENT]: true,
    [GridColumnKey.SEVERITY]: true,
    [GridColumnKey.DATE_FROM]: true
};
