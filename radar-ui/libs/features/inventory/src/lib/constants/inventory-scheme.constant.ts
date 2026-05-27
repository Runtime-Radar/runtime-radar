import { InventorySidepanelContextType } from '../interfaces/inventory-sidepanel.interface';

export const INVENTORY_NODE_BG_COLORS: string[] = ['#f3f4f6', '#e0edff', '#dbf5d1', '#ecdbf6', '#ffeccc'];

export const INVENTORY_NODE_TEXT_COLORS: string[] = ['#21222c', '#0066ff', '#367d1c', '#861fc6', '#a86b00'];

export const INVENTORY_CONTEXT_WEIGHT = new Map<InventorySidepanelContextType, number>([
    [InventorySidepanelContextType.NONE, 0],
    [InventorySidepanelContextType.NODE, 1],
    [InventorySidepanelContextType.NAMESPACE, 2],
    [InventorySidepanelContextType.POD, 3],
    [InventorySidepanelContextType.CONTAINER, 4]
]);
