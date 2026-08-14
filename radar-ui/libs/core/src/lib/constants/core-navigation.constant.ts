import { NavigationMenu } from '../interfaces/core-navigation.interface';
import { RouterName } from './core-router.constant';

// If navigation items are going to be changed, it needs to update $menu-height into navbar component.
export const NAVIGATION: NavigationMenu[] = [
    {
        path: RouterName.DEFAULT,
        children: [
            {
                path: RouterName.INVENTORY,
                localizationKey: 'Common.Pseudo.Menu.Inventory',
                testId: 'inventory-navbar-link',
                icon: 'kbq-map_16'
            },
            {
                path: RouterName.CLUSTERS,
                localizationKey: 'Common.Pseudo.Menu.Cluster',
                testId: 'cluster-navbar-link',
                icon: 'kbq-box-open_16'
            },
            {
                path: RouterName.RUNTIME,
                localizationKey: 'Common.Pseudo.Menu.Runtime',
                testId: 'runtime-navbar-link',
                icon: 'kbq-play-rewind_16'
            }
        ]
    },
    {
        path: RouterName.SETTINGS,
        localizationKey: 'Common.Pseudo.Menu.Management',
        children: [
            {
                path: RouterName.USERS,
                localizationKey: 'Common.Pseudo.Menu.User',
                testId: 'user-navbar-link',
                icon: 'kbq-user-multiple_16'
            },
            {
                path: RouterName.RULES,
                localizationKey: 'Common.Pseudo.Menu.Rule',
                testId: 'rule-navbar-link',
                icon: 'kbq-shield-check_16'
            },
            {
                path: RouterName.INTEGRATIONS,
                localizationKey: 'Common.Pseudo.Menu.Integration',
                testId: 'integration-navbar-link',
                icon: 'kbq-bell_16'
            }
        ]
    }
];
