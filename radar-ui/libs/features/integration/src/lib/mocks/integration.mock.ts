import { NotificationWebhookHeadersList } from '@cs/domains/notification';

import { IntegrationRecipientTemplateRecord } from '../interfaces/integration-recipient-form.interace';

export const INTEGRATION_RECIPIENT_HEADERS: NotificationWebhookHeadersList = {
    key1: 'value1',
    key2: 'value2'
};

export const INTEGRATION_RECIPIENT_RECORD: IntegrationRecipientTemplateRecord = {
    uuid1: {
        key: 'key1',
        value: 'value1'
    },
    uuid2: {
        key: 'key2',
        value: 'value2'
    }
};
