import { KubeManagerNamespaceGroup, KubeManagerPod, KubeManagerPodPhase } from '../interfaces';

export const KUBE_MANAGER_PODS: KubeManagerPod[] = [
    {
        uid: 'uid1',
        name: 'name1',
        namespace: 'namespace1',
        node_name: 'node1',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: [
            {
                name: 'cntr11',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr12',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr13',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
    },
    {
        uid: 'uid2',
        name: 'name2',
        namespace: 'namespace1',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: [
            {
                name: 'cntr21',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr22',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
    },
    {
        uid: 'uid3',
        name: 'name3',
        namespace: 'namespace2',
        node_name: 'node1',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: [
            {
                name: 'cntr31',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
    },
    {
        uid: 'uid4',
        name: 'name4',
        namespace: 'namespace2',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: [
            {
                name: 'cntr41',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr42',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr43',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr44',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr45',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
    },
    {
        uid: 'uid5',
        name: 'name5',
        namespace: 'namespace2',
        node_name: 'node1',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: []
    },
    {
        uid: 'uid6',
        name: 'name6',
        namespace: 'namespace3',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: [
            {
                name: 'cntr61',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr62',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr63',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
    },
    {
        uid: 'uid7',
        name: 'name7',
        namespace: 'namespace4',
        node_name: 'node2',
        phase: KubeManagerPodPhase.RUNNING,
        owners: [],
        containers: [
            {
                name: 'cntr71',
                image_url: 'url',
                image_digest: 'digest'
            },
            {
                name: 'cntr71',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
    },
    {
        uid: 'uid8',
        name: 'name8',
        namespace: 'namespace1',
        node_name: '',
        phase: KubeManagerPodPhase.PENDING,
        owners: [],
        containers: [
            {
                name: 'cntr81',
                image_url: 'url',
                image_digest: 'digest'
            }
        ]
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
