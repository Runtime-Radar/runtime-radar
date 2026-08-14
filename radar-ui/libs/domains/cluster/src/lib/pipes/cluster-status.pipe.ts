import { Pipe, PipeTransform } from '@angular/core';

import { CLUSTER_STATUS } from '../constants/cluster.constant';
import { ClusterStatus } from '../interfaces/contract/cluster-contract.interface';

@Pipe({
    name: 'clusterStatusLocalization',
    pure: false,
    standalone: false
})
export class ClusterStatusLocalizationPipe implements PipeTransform {
    transform(type?: ClusterStatus): string {
        const value = CLUSTER_STATUS.find((item) => item.id === type);

        return value ? value.localizationKey : '';
    }
}
