import { KubeManagerGroupNamespacesSortPipe } from './kube-manager-group-namespaces-sort.pipe';
import {
    KUBE_MANAGER_NAMESPACE_GROUP_1,
    KUBE_MANAGER_NAMESPACE_GROUP_2,
    KUBE_MANAGER_NAMESPACE_GROUP_3
} from '../mocks/kube-manager.mock';

describe('KubeManagerGroupNamespacesSortPipe', () => {
    let pipe: KubeManagerGroupNamespacesSortPipe;

    beforeEach(() => {
        pipe = new KubeManagerGroupNamespacesSortPipe();
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should return empty groups', () => {
        expect(pipe.transform()).toEqual([]);
    });

    it('should return sorted groups', () => {
        const values = [KUBE_MANAGER_NAMESPACE_GROUP_1, KUBE_MANAGER_NAMESPACE_GROUP_2, KUBE_MANAGER_NAMESPACE_GROUP_3];

        expect(pipe.transform(values)).toEqual([
            KUBE_MANAGER_NAMESPACE_GROUP_3,
            KUBE_MANAGER_NAMESPACE_GROUP_1,
            KUBE_MANAGER_NAMESPACE_GROUP_2
        ]);
    });
});
