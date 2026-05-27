import { CreateClusterRequest } from '@cs/domains/cluster';

import { ClusterFeatureRequestUtilsService } from './cluster-request-utils.service';
import { ClusterFormState } from '../interfaces/cluster-form-state.interface';
import { ClusterStepName } from '../interfaces/cluster-stepper.interface';

describe('ClusterFeatureRequestUtilsService', () => {
    const state: ClusterFormState = {
        id: 1,
        step: ClusterStepName.ACCESS,
        registry: {
            address: 'raddress',
            user: 'ruser',
            password: 'rpassword',
            isImageShortName: true
        },
        clickhouse: {
            isInternalCluster: true,
            user: 'cuser',
            password: 'cpassword',
            database: 'cdatabase',
            isTls: true,
            hasCheckCert: true,
            ca: 'ca',
            isPersistence: true,
            storageClass: 'cclass',
            address: 'caddress'
        },
        postgres: {
            isInternalCluster: true,
            user: 'puser',
            password: 'ppassword',
            database: 'pdatabase',
            isTls: true,
            hasCheckCert: true,
            ca: 'ca',
            isPersistence: true,
            storageClass: 'pclass',
            address: 'paddress'
        },
        redis: {
            isInternalCluster: true,
            user: 'ruser',
            password: 'rpassword',
            database: 'rdatabase',
            isTls: true,
            hasCheckCert: true,
            ca: 'ca',
            isPersistence: true,
            storageClass: 'rclass',
            address: 'raddress'
        },
        rabbit: {
            isInternalCluster: true,
            user: 'auser',
            password: 'apassword',
            isPersistence: true,
            storageClass: 'aclass',
            address: 'aaddress'
        },
        metric: {
            isMetricEnabled: true,
            isInternalCluster: true,
            user: 'muser',
            password: 'mpassword',
            isPersistence: true,
            storageClass: 'mclass',
            address: 'maddress'
        },
        ingress: {
            isIngressEnabled: true,
            ingressClass: 'iclass',
            hostname: 'ihost',
            cert: 'icert',
            certKey: 'icertkey',
            isNodePortEnabled: true,
            port: 'iport'
        },
        access: {
            proxyUrl: 'purl',
            ownCsUrl: 'csurl',
            centralCsUrl: 'curl',
            namespace: 'namespace',
            name: 'name'
        }
    };

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('toRequest', () => {
        it('should return correct request', () => {
            const request: CreateClusterRequest = {
                name: 'name',
                config: {
                    version: '1',
                    proxy_url: 'purl',
                    own_cs_url: 'csurl',
                    central_cs_url: 'curl',
                    namespace: 'namespace',
                    enable_metrics: true,
                    clickhouse: {
                        user: 'cuser',
                        password: 'cpassword',
                        use_tls: true,
                        check_cert: true,
                        ca: undefined,
                        persistence: true,
                        storage_class: 'cclass',
                        address: undefined,
                        database: 'cdatabase'
                    },
                    postgres: {
                        user: 'puser',
                        password: 'ppassword',
                        use_tls: true,
                        check_cert: true,
                        ca: undefined,
                        persistence: true,
                        storage_class: 'pclass',
                        address: undefined,
                        database: 'pdatabase'
                    },
                    redis: {
                        user: 'ruser',
                        password: 'rpassword',
                        use_tls: true,
                        check_cert: true,
                        ca: undefined,
                        persistence: true,
                        storage_class: 'rclass',
                        address: undefined
                    },
                    rabbit: {
                        user: 'auser',
                        password: 'apassword',
                        persistence: true,
                        storage_class: 'aclass',
                        address: undefined
                    },
                    grafana: {
                        user: 'muser',
                        password: 'mpassword',
                        address: undefined
                    },
                    prometheus: {
                        deploy: true,
                        persistence: true,
                        storage_class: 'mclass'
                    },
                    ingress: {
                        ingress_class: 'iclass',
                        hostname: 'ihost',
                        cert: 'icert',
                        cert_key: 'icertkey'
                    },
                    node_port: {
                        port: 'iport'
                    },
                    registry: {
                        address: 'raddress',
                        user: 'ruser',
                        password: 'rpassword',
                        image_short_names: true
                    }
                }
            };

            expect(ClusterFeatureRequestUtilsService.toRequest(state)).toEqual(request);
        });
    });
});
