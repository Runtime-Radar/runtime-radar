export enum ClusterStepName {
    REGISTRY = 'registry',
    DATABASE = 'database',
    METRIC = 'metric',
    INGRESS = 'ingress',
    ACCESS = 'access'
}

export interface ClusterStepperTab {
    id: ClusterStepName;
    title: string;
    description: string;
}
