export enum SortField {
    SEVERITY = 'severity',
    VERDICT = 'verdict',
    TIME = 'time',
    REGISTERED_AT = 'registered_at',
    CREATED_AT = 'created_at'
}

export enum SortKey {
    NONE = 'none',
    ASC = 'asc nulls first',
    DESC = 'desc nulls last'
}

export type AbstractSorts = Partial<{
    [key in SortField]: SortKey;
}>;

export interface AbstractFilters {
    [key: string]: unknown;
}
