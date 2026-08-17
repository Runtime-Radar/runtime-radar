import { AbstractFilters } from '@cs/core';

export interface InventoryFilters extends AbstractFilters {
    nodes: string[];
    namespaces: string[];
    pods: string[];
    containers: string[];
}
