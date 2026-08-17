import { isChildCluster } from './argument/child-cluster';

export const environment = {
    api: '/api/v1/',
    availableLocales: ['en-US'], // I18nLocale
    defaultLocale: 'en-US',
    singleTenant: ['signin', 'tokens', 'user', 'role', 'cluster'],
    childCluster: isChildCluster,
    pollingInterval: 120000,
    refreshInterval: 900000,
    production: false
};
