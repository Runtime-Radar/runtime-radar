import { KubeManagerNamespaceGroup, KubeManagerPod, KubeManagerPodPhase } from '../interfaces';

export const KUBE_MANAGER_PODS: KubeManagerPod[] = [
    {
        uid: 'uid1',
        name: 'name1',
        namespace: 'namespace1',
        node_name: 'node1',
        phase: KubeManagerPodPhase.RUNNING,
        containers: ['cntr11', 'cntr12', 'cntr13']
    },
    {
        uid: 'uid2',
        name: 'name2',
        namespace: 'namespace1',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        containers: ['cntr21', 'cntr22']
    },
    {
        uid: 'uid3',
        name: 'name3',
        namespace: 'namespace2',
        node_name: 'node1',
        phase: KubeManagerPodPhase.RUNNING,
        containers: ['cntr31']
    },
    {
        uid: 'uid4',
        name: 'name4',
        namespace: 'namespace2',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        containers: ['cntr41', 'cntr42', 'cntr43', 'cntr44', 'cntr45']
    },
    {
        uid: 'uid5',
        name: 'name5',
        namespace: 'namespace2',
        node_name: 'node1',
        phase: KubeManagerPodPhase.RUNNING,
        containers: []
    },
    {
        uid: 'uid6',
        name: 'name6',
        namespace: 'namespace3',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        containers: ['cntr61', 'cntr62', 'cntr63']
    },
    {
        uid: 'uid7',
        name: 'name7',
        namespace: 'namespace4',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        containers: ['cntr71', 'cntr72']
    },
    {
        uid: 'uid8',
        name: 'name8',
        namespace: 'namespace1',
        node_name: '',
        phase: KubeManagerPodPhase.PENDING,
        containers: ['cntr81']
    }
];

export const KUBE_MANAGER_NAMESPACE_GROUP_1: KubeManagerNamespaceGroup = {
    id: 'groupId0',
    namespaces: [
        {
            id: 'namespaceId0',
            namespace: 'namespace1',
            nodes: ['node1', 'node2'],
            podCount: 2,
            isVisible: null
        }
    ]
};

export const KUBE_MANAGER_NAMESPACE_GROUP_2: KubeManagerNamespaceGroup = {
    id: 'groupId1',
    namespaces: [
        {
            id: 'namespaceId1',
            namespace: 'namespace2',
            nodes: ['node1', 'node2'],
            podCount: 3,
            isVisible: null
        }
    ]
};

export const KUBE_MANAGER_NAMESPACE_GROUP_3: KubeManagerNamespaceGroup = {
    id: 'groupId2',
    namespaces: [
        {
            id: 'namespaceId2',
            namespace: 'namespace3',
            nodes: ['node2'],
            podCount: 1,
            isVisible: null
        },
        {
            id: 'namespaceId3',
            namespace: 'namespace4',
            nodes: ['node2'],
            podCount: 1,
            isVisible: null
        }
    ]
};
