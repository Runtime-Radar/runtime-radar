export enum RouterName {
    DEFAULT = '',
    CLUSTERS = 'clusters',
    FORBIDDEN = 'forbidden',
    ERROR = 'error',
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
    COMMON = 'common',
    CLUSTER = 'cluster',
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
