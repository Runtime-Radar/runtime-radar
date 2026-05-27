import { IntegrationType } from '@cs/domains/integration';

import { NotificationEmail, NotificationWebhook } from '../interfaces';

export const NOTIFICATION_EMAIL: NotificationEmail = {
    id: 'id1',
    integration_id: 'integrationId1',
    integration_type: IntegrationType.EMAIL,
    event_type: 'runtime_event',
    name: 'mail',
    recipients: ['recipientId1'],
    template: '',
    central_cs_url: 'csurl',
    cs_cluster_id: 'cid',
    cs_cluster_name: 'name',
    own_cs_url: 'url',
    email: {
        subject_template: ''
    }
};

export const NOTIFICATION_WEBHOOK: NotificationWebhook = {
    id: 'id2',
    integration_id: 'integrationId2',
    integration_type: IntegrationType.WEBHOOK,
    event_type: 'other_event',
    name: 'webhook',
    recipients: ['recipientId1'],
    template: '',
    central_cs_url: 'csurl',
    cs_cluster_id: 'cid',
    cs_cluster_name: 'name',
    own_cs_url: 'url',
    webhook: {
        path: 'path',
        headers: {}
    }
};
