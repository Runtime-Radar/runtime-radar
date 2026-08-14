import { KUBE_MANAGER_PODS } from '../mocks/kube-manager.mock';
import { adapter } from './kube-manager-reducer.store';
import { KubeManagerEntityState, KubeManagerNamespace, KubeManagerPodPhase } from '../interfaces';
import {
    getKubeManagerGroupNamespaces,
    getKubeManagerNamespaces,
    getKubeManagerNodes,
    getKubeManagerPodsByNamespace
} from './kube-manager-selector.store';

describe('KubeManagerDomainReducer', () => {
    let entityState: KubeManagerEntityState = adapter.getInitialState();

    const uniqueNamespaces: KubeManagerNamespace[] = [
        {
            id: 'namespaceId0',
            namespace: 'namespace1',
            nodes: ['node1', 'node2'],
            podCount: 3,
            isVisible: null
        },
        {
            id: 'namespaceId1',
            namespace: 'namespace2',
            nodes: ['node1', 'node2'],
            podCount: 3,
            isVisible: null
        },
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
    ];

    beforeEach(() => {
        entityState = adapter.setAll(KUBE_MANAGER_PODS, adapter.getInitialState());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('getKubeManagerNodes', () => {
        it('should return unique nodes', () => {
            const result = getKubeManagerNodes.projector(entityState);

            expect(result).toEqual([
                {
                    id: 'nodeId0',
                    node_name: 'node1'
                },
                {
                    id: 'nodeId1',
                    node_name: 'node2'
                }
            ]);
        });
    });

    describe('getKubeManagerNamespaces', () => {
        it('should return unique namespaces', () => {
            const result = getKubeManagerNamespaces().projector(entityState);

            expect(result).toEqual(uniqueNamespaces);
        });

        it('should return unique namespaces filtered by node', () => {
            const result = getKubeManagerNamespaces('node1').projector(entityState);

            expect(result).toEqual([
                {
                    id: 'namespaceId0',
                    namespace: 'namespace1',
                    nodes: ['node1', 'node2'],
                    podCount: 3,
                    isVisible: true
                },
                {
                    id: 'namespaceId1',
                    namespace: 'namespace2',
                    nodes: ['node1', 'node2'],
                    podCount: 3,
                    isVisible: true
                },
                {
                    id: 'namespaceId2',
                    namespace: 'namespace3',
                    nodes: ['node2'],
                    podCount: 1,
                    isVisible: false
                },
                {
                    id: 'namespaceId3',
                    namespace: 'namespace4',
                    nodes: ['node2'],
                    podCount: 1,
                    isVisible: false
                }
            ]);
        });
    });

    describe('getKubeManagerGroupNamespaces', () => {
        it('should return grouped namespaces', () => {
            const result = getKubeManagerGroupNamespaces().projector(uniqueNamespaces);

            expect(result).toEqual([
                {
                    id: 'groupId0',
                    namespaces: [
                        {
                            id: 'namespaceId0',
                            namespace: 'namespace1',
                            nodes: ['node1', 'node2'],
                            podCount: 3,
                            isVisible: null
                        }
                    ]
                },
                {
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
                },
                {
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
                }
            ]);
        });
    });

    describe('getKubeManagerPodsByNamespace', () => {
        it('should return pods filtered by namespace', () => {
            const result = getKubeManagerPodsByNamespace('namespace1').projector(entityState);

            expect(result).toEqual([
                {
                    uid: 'uid1',
                    name: 'name1',
                    namespace: 'namespace1',
                    node_name: 'node1',
                    phase: KubeManagerPodPhase.RUNNING,
                    owners: [],
                    isVisible: null,
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
                    isVisible: null,
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
                    uid: 'uid8',
                    name: 'name8',
                    namespace: 'namespace1',
                    node_name: '',
                    phase: KubeManagerPodPhase.PENDING,
                    owners: [],
                    isVisible: null,
                    containers: [
                        {
                            name: 'cntr81',
                            image_url: 'url',
                            image_digest: 'digest'
                        }
                    ]
                }
            ]);
        });

        it('should return pods filtered by namespace and node', () => {
            const result = getKubeManagerPodsByNamespace('namespace1', 'node1').projector(entityState);

            expect(result).toEqual([
                {
                    uid: 'uid1',
                    name: 'name1',
                    namespace: 'namespace1',
                    node_name: 'node1',
                    phase: KubeManagerPodPhase.RUNNING,
                    owners: [],
                    isVisible: true,
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
                    isVisible: false,
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
                    uid: 'uid8',
                    name: 'name8',
                    namespace: 'namespace1',
                    node_name: '',
                    phase: KubeManagerPodPhase.PENDING,
                    owners: [],
                    isVisible: false,
                    containers: [
                        {
                            name: 'cntr81',
                            image_url: 'url',
                            image_digest: 'digest'
                        }
                    ]
                }
            ]);
        });
    });
});
