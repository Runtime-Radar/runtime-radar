export enum GridColumnKey {
    FUNCTION = 'function',
    BINARY = 'binary',
    ARGUMENT = 'argument',
    SOURCE = 'source',
    SEVERITY = 'severity',
    DATE_FROM = 'dateFrom',
    DATE_TO = 'dateTo'
}

export interface GridColumnOption {
    id: GridColumnKey;
    localizationKey: string;
}

export type GridColumns = {
    [key in GridColumnKey]: boolean;
};

export interface GridColumnState {
    id: string;
    columns: Partial<GridColumns>;
}
