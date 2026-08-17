import { Pipe, PipeTransform } from '@angular/core';

import { KubeManagerNamespace, KubeManagerNamespaceGroup } from '@cs/domains/kube-manager';

@Pipe({
    name: 'kubeManagerGroupNamespacesSort',
    standalone: false
})
export class KubeManagerGroupNamespacesSortPipe implements PipeTransform {
    transform(values?: KubeManagerNamespaceGroup[] | null): KubeManagerNamespaceGroup[] {
        if (!values) {
            return [];
        }

        return [...values].sort(
            (a, b) => new Set(this.getNodes(a.namespaces)).size - new Set(this.getNodes(b.namespaces)).size
        );
    }

    private getNodes = (namespaces: KubeManagerNamespace[]): string[] => {
        return namespaces.reduce<string[]>((acc, item) => [...acc, ...item.nodes], []);
    };
}
