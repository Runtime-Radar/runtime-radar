import { InventorySidepanelContext, InventorySidepanelContextType } from '../interfaces/inventory-sidepanel.interface';

export const INVENTORY_SIDEPANEL_CONTEXT_NODE: InventorySidepanelContext[] = [
    {
        id: 'nodeId1',
        sidepanelId: 'node:nodeId1',
        path: 'nodeId1',
        type: InventorySidepanelContextType.NODE
    },
    {
        id: 'namespaceId1',
        sidepanelId: 'namespace:namespaceId1',
        path: 'nodeId1:namespaceId1',
        type: InventorySidepanelContextType.NAMESPACE
    },
    {
        id: 'podId1',
        sidepanelId: 'pod:podId1',
        path: 'nodeId1:namespaceId1:podId1',
        type: InventorySidepanelContextType.POD
    },
    {
        id: 'containerId1',
        sidepanelId: 'container:containerId1',
        path: 'nodeId1:namespaceId1:podId1:containerId1',
        type: InventorySidepanelContextType.CONTAINER
    },
    {
        id: 'containerId2',
        sidepanelId: 'container:containerId2',
        path: 'nodeId1:namespaceId1:podId1:containerId2',
        type: InventorySidepanelContextType.CONTAINER
    },
    {
        id: 'podId2',
        sidepanelId: 'pod:podId2',
        path: 'nodeId1:namespaceId1:podId2',
        type: InventorySidepanelContextType.POD
    },
    {
        id: 'containerId3',
        sidepanelId: 'container:containerId3',
        path: 'nodeId1:namespaceId1:podId2:containerId3',
        type: InventorySidepanelContextType.CONTAINER
    },
    {
        id: 'namespaceId2',
        sidepanelId: 'namespace:namespaceId2',
        path: 'nodeId1:namespaceId2',
        type: InventorySidepanelContextType.NAMESPACE
    },
    {
        id: 'podId3',
        sidepanelId: 'pod:podId3',
        path: 'nodeId1:namespaceId2:podId3',
        type: InventorySidepanelContextType.POD
    },
    {
        id: 'containerId4',
        sidepanelId: 'container:containerId4',
        path: 'nodeId1:namespaceId2:podId3:containerId4',
        type: InventorySidepanelContextType.CONTAINER
    },
    {
        id: 'nodeId2',
        sidepanelId: 'node:nodeId2',
        path: 'nodeId2',
        type: InventorySidepanelContextType.NODE
    },
    {
        id: 'namespaceId3',
        sidepanelId: 'namespace:namespaceId3',
        path: 'nodeId2:namespaceId3',
        type: InventorySidepanelContextType.NAMESPACE
    },
    {
        id: 'podId4',
        sidepanelId: 'pod:podId4',
        path: 'nodeId2:namespaceId3:podId4',
        type: InventorySidepanelContextType.POD
    },
    {
        id: 'containerId5',
        sidepanelId: 'container:containerId5',
        path: 'nodeId2:namespaceId3:podId4:containerId5',
        type: InventorySidepanelContextType.CONTAINER
    }
];
