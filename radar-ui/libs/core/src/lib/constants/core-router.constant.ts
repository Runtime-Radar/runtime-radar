export enum RouterName {
    CLUSTERS = 'clusters',
    DEFAULT = '',
    ERROR = 'error',
    FORBIDDEN = 'forbidden',
    INTEGRATIONS = 'integrations',
    INVENTORY = 'inventory',
    RULES = 'rules',
    RUNTIME = 'runtime',
    SETTINGS = 'settings',
    SIGN_IN = 'sign-in',
    SWITCH = 'switch',
    TOKENS = 'tokens',
    USERS = 'users'
}

export enum TranslationDict {
    AUTH = 'auth',
    CLUSTER = 'cluster',
    COMMON = 'common',
    INTEGRATION = 'integration',
    INVENTORY = 'inventory',
    REPORT = 'report',
    RULE = 'rule',
    RUNTIME = 'runtime',
    TOKEN = 'token',
    USER = 'user'
}

export const DEFAULT_ROUTER_NAME = RouterName.RUNTIME;

export const DEFAULT_TRANSLATION_DICTS = [TranslationDict.COMMON, TranslationDict.USER];
