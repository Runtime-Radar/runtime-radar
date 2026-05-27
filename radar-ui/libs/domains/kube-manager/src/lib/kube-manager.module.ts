import { EffectsModule } from '@ngrx/effects';
import { NgModule } from '@angular/core';
import { StoreModule } from '@ngrx/store';

import { ApiModule } from '@cs/api';

import { KubeManagerEffectStore } from './stores/kube-manager-effect.store';
import { KubeManagerGroupNamespacesSortPipe } from './pipes/kube-manager-group-namespaces-sort.pipe';
import { KUBE_MANAGER_DOMAIN_KEY, kubeManagerDomainReducer } from './stores/kube-manager-selector.store';

@NgModule({
    imports: [
        ApiModule,
        StoreModule.forFeature(KUBE_MANAGER_DOMAIN_KEY, kubeManagerDomainReducer),
        EffectsModule.forFeature([KubeManagerEffectStore])
    ],
    declarations: [KubeManagerGroupNamespacesSortPipe],
    exports: [KubeManagerGroupNamespacesSortPipe]
})
export class KubeManagerDomainModule {}
