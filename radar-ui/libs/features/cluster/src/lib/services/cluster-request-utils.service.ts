import { Injectable } from '@angular/core';

import { ClusterIngress, CreateClusterRequest } from '@cs/domains/cluster';

import { ClusterFormState } from '../interfaces/cluster-form-state.interface';
import { ClusterDataBaseForm, ClusterMetricForm } from '../interfaces/cluster-form.interface';

@Injectable({
    providedIn: 'root'
})
export class ClusterFeatureRequestUtilsService {
    static toRequest(state: ClusterFormState): CreateClusterRequest {
        const ingress: ClusterIngress = {
            ingress_class: state.ingress.ingressClass,
            hostname: state.ingress.hostname,
            cert: state.ingress.cert ? state.ingress.cert : undefined,
            cert_key: state.ingress.certKey ? state.ingress.certKey : undefined
        };

        const database = (values: ClusterDataBaseForm) => ({
            user: values.user ? values.user : '',
            password: values.user ? values.password : '',
            use_tls: values.isTls,
            check_cert: values.isTls ? values.hasCheckCert : undefined,
            ca: !values.isInternalCluster && values.hasCheckCert ? values.ca : undefined,
            persistence: values.isInternalCluster ? values.isPersistence : undefined,
            storage_class: values.isInternalCluster && values.isPersistence ? values.storageClass : undefined,
            address: !values.isInternalCluster ? values.address : undefined
        });

        const grafana = (values: ClusterMetricForm) => ({
            user: values.isInternalCluster ? values.user : undefined,
            password: values.isInternalCluster ? values.password : undefined,
            address: !values.isInternalCluster ? values.address : undefined
        });

        const prometheus = (values: ClusterMetricForm) => ({
            deploy: values.isInternalCluster,
            persistence: values.isInternalCluster ? values.isPersistence : false,
            storage_class: values.isInternalCluster && values.isPersistence ? values.storageClass : undefined
        });

        return {
            name: state.access.name,
            config: {
                version: '1', // @todo: create environment constant
                proxy_url: state.access.proxyUrl ? state.access.proxyUrl : undefined,
                own_cs_url: state.access.ownCsUrl,
                central_cs_url: state.access.centralCsUrl,
                namespace: state.access.namespace,
                postgres: {
                    ...database(state.postgres),
                    database: state.postgres.database
                },
                clickhouse: {
                    ...database(state.clickhouse),
                    database: state.clickhouse.database
                },
                redis: database(state.redis),
                enable_metrics: state.metric.isMetricEnabled,
                grafana: state.metric.isMetricEnabled ? grafana(state.metric) : undefined,
                prometheus: state.metric.isMetricEnabled ? prometheus(state.metric) : undefined,
                rabbit: {
                    user: state.rabbit.user,
                    password: state.rabbit.password,
                    persistence: state.rabbit.isInternalCluster ? state.rabbit.isPersistence : undefined,
                    storage_class:
                        state.rabbit.isInternalCluster && state.rabbit.isPersistence
                            ? state.rabbit.storageClass
                            : undefined,
                    address: !state.rabbit.isInternalCluster ? state.rabbit.address : undefined
                },
                registry: {
                    address: state.registry.address,
                    user: state.registry.user,
                    password: state.registry.password,
                    image_short_names: state.registry.isImageShortName
                },
                ingress: state.ingress.isIngressEnabled ? ingress : undefined,
                node_port: state.ingress.isNodePortEnabled ? { port: state.ingress.port } : undefined
            }
        };
    }
}
