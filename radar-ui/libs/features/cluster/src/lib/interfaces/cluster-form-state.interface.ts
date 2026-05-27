import { ClusterStepName } from './cluster-stepper.interface';
import {
    ClusterAccessForm,
    ClusterDataBaseForm,
    ClusterIngressForm,
    ClusterMetricForm,
    ClusterRabbitForm,
    ClusterRegistryForm
} from './cluster-form.interface';

export interface ClusterFormState {
    id: number;
    step: ClusterStepName;
    registry: ClusterRegistryForm;
    clickhouse: ClusterDataBaseForm;
    postgres: ClusterDataBaseForm;
    redis: ClusterDataBaseForm;
    rabbit: ClusterRabbitForm;
    metric: ClusterMetricForm;
    ingress: ClusterIngressForm;
    access: ClusterAccessForm;
}

export type ClusterFormType = keyof Omit<ClusterFormState, 'id' | 'step'>;
