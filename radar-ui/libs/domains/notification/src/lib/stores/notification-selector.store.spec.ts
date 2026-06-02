import { RuleType } from '@cs/domains/rule';

import { NotificationEntityState } from '../interfaces/state/notification-state.interface';
import { adapter } from './notification-reducer.store';
import { NOTIFICATION_EMAIL, NOTIFICATION_WEBHOOK } from '../mocks/notification.mock';
import { getNotificationsByEventType, getNotificationsByIntegrationId } from './notification-selector.store';

describe('NotificationDomainReducer', () => {
    let entityState: NotificationEntityState = adapter.getInitialState();

    beforeEach(() => {
        entityState = adapter.setAll([NOTIFICATION_EMAIL, NOTIFICATION_WEBHOOK], adapter.getInitialState());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('getNotificationsByIntegrationId', () => {
        it('should return notifications filtered by integrationId', () => {
            const result = getNotificationsByIntegrationId('integrationId1').projector(entityState);

            expect(result).toEqual([NOTIFICATION_EMAIL]);
        });
    });

    describe('getNotificationsByEventType', () => {
        it('should return notifications filtered by ruleType', () => {
            const result = getNotificationsByEventType(RuleType.TYPE_RUNTIME).projector(entityState);

            expect(result).toEqual([NOTIFICATION_EMAIL]);
        });
    });
});
