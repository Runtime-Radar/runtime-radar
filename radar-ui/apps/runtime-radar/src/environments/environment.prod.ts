import { isChildCluster } from './argument/child-cluster';

export const environment = {
    api: '/api/v1/',
    availableLocales: ['en-US'], // I18nLocale
    defaultLocale: 'en-US',
    singleTenant: ['signin', 'tokens', 'user', 'role', 'cluster'],
    childCluster: isChildCluster,
    pollingInterval: 60000,
    refreshInterval: 300000,
    production: true
};
