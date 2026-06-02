import { map } from 'rxjs';
import { ChangeDetectionStrategy, Component } from '@angular/core';

import { ClusterStoreService } from '@cs/domains/cluster';
import { LicenseStoreService } from '@cs/domains/license';

import { ClusterFeatureFormComponentStore } from '../../stores/cluster-form.store';
import { ClusterFormType } from '../../interfaces/cluster-form-state.interface';
import { ClusterStepName } from '../../interfaces/cluster-stepper.interface';
import { ClusterFeatureRequestUtilsService as clusterRequestUtils } from '../../services/cluster-request-utils.service';
import {
    ClusterAccessForm,
    ClusterCreateFormOutputs,
    ClusterDataBaseForm,
    ClusterIngressForm,
    ClusterMetricForm,
    ClusterRabbitForm,
    ClusterRegistryForm
} from '../../interfaces/cluster-form.interface';

@Component({
    templateUrl: './cluster-create.container.html',
    styleUrl: './cluster-create.container.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    providers: [ClusterFeatureFormComponentStore]
})
export class ClusterFeatureCreateContainer {
    readonly registryForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterRegistryForm>('registry');

    readonly clickhouseForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterDataBaseForm>('clickhouse');

    readonly postgresForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterDataBaseForm>('postgres');

    readonly redisForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterDataBaseForm>('redis');

    readonly rabbitForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterRabbitForm>('rabbit');

    readonly metricForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterMetricForm>('metric');

    readonly ingressForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterIngressForm>('ingress');

    readonly accessForm$ = this.clusterFeatureFormComponentStore.partialForm$<ClusterAccessForm>('access');

    readonly step$ = this.clusterFeatureFormComponentStore.form$.pipe(map((state) => state.step));

    readonly centralUrl$ = this.licenseStoreService.centralUrl$;

    readonly isStepValid$ = this.step$.pipe(
        map((step) => {
            switch (step) {
                case ClusterStepName.REGISTRY:
                    return this.formValidations.get('registry') || false;
                case ClusterStepName.DATABASE:
                    return (
                        (this.formValidations.get('clickhouse') &&
                            this.formValidations.get('rabbit') &&
                            this.formValidations.get('postgres') &&
                            this.formValidations.get('redis')) ||
                        false
                    );
                case ClusterStepName.METRIC:
                    return this.formValidations.get('metric') || false;
                case ClusterStepName.INGRESS:
                    return this.formValidations.get('ingress') || false;
                case ClusterStepName.ACCESS:
                    return this.formValidations.get('access') || false;
                default:
                    return false;
            }
        })
    );

    private readonly formValidations = new Map<ClusterFormType, boolean>([]);

    constructor(
        private readonly clusterStoreService: ClusterStoreService,
        private readonly licenseStoreService: LicenseStoreService,
        private readonly clusterFeatureFormComponentStore: ClusterFeatureFormComponentStore
    ) {}

    changeRegistryForm(registry: ClusterCreateFormOutputs<ClusterRegistryForm>) {
        this.formValidations.set('registry', registry.isValid);
        this.clusterFeatureFormComponentStore.update({ registry: registry.form });
    }

    changeClickhouseForm(clickhouse: ClusterCreateFormOutputs<ClusterDataBaseForm>) {
        this.formValidations.set('clickhouse', clickhouse.isValid);
        this.clusterFeatureFormComponentStore.update({ clickhouse: clickhouse.form });
    }

    changePostgresForm(postgres: ClusterCreateFormOutputs<ClusterDataBaseForm>) {
        this.formValidations.set('postgres', postgres.isValid);
        this.clusterFeatureFormComponentStore.update({ postgres: postgres.form });
    }

    changeRedisForm(redis: ClusterCreateFormOutputs<ClusterDataBaseForm>) {
        this.formValidations.set('redis', redis.isValid);
        this.clusterFeatureFormComponentStore.update({ redis: redis.form });
    }

    changeRabbitForm(rabbit: ClusterCreateFormOutputs<ClusterRabbitForm>) {
        this.formValidations.set('rabbit', rabbit.isValid);
        this.clusterFeatureFormComponentStore.update({ rabbit: rabbit.form });
    }

    changeMetricForm(metric: ClusterCreateFormOutputs<ClusterMetricForm>) {
        this.formValidations.set('metric', metric.isValid);
        this.clusterFeatureFormComponentStore.update({ metric: metric.form });
    }

    changeIngressForm(ingress: ClusterCreateFormOutputs<ClusterIngressForm>) {
        this.formValidations.set('ingress', ingress.isValid);
        this.clusterFeatureFormComponentStore.update({ ingress: ingress.form });
    }

    changeAccessForm(access: ClusterCreateFormOutputs<ClusterAccessForm>) {
        this.formValidations.set('access', access.isValid);
        this.clusterFeatureFormComponentStore.update({ access: access.form });
    }

    stepChange(step: ClusterStepName) {
        this.clusterFeatureFormComponentStore.update({ step });
    }

    completeChange() {
        const isFormValid = Array.from(this.formValidations, ([_, value]) => value).every((item) => item);
        if (isFormValid && this.clusterFeatureFormComponentStore.form) {
            this.clusterStoreService.createCluster(
                clusterRequestUtils.toRequest(this.clusterFeatureFormComponentStore.form)
            );
            this.clusterFeatureFormComponentStore.resetState();
        } else {
            console.warn('form must be valid');
        }
    }
}
