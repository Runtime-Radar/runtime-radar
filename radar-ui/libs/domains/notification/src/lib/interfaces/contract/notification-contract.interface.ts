import { IntegrationType } from '@cs/domains/integration';

interface AbstractNotification {
    id: string;
    integration_id: string;
    integration_type: IntegrationType;
    event_type: string;
    name: string;
    recipients: string[];
    template: string;
    central_cs_url: string;
    cs_cluster_id: string;
    cs_cluster_name: string;
    own_cs_url: string;
}

export interface NotificationEmailEntity {
    subject_template: string;
}

export interface NotificationEmail extends AbstractNotification {
    integration_type: IntegrationType.EMAIL;
    email: NotificationEmailEntity;
}

export interface NotificationSyslog extends AbstractNotification {
    integration_type: IntegrationType.SYSLOG;
    syslog: Record<string, never>;
}

export interface NotificationWebhookHeadersList {
    [key: string]: string;
}

export interface NotificationWebhookEntity {
    path: string;
    headers: NotificationWebhookHeadersList;
}

export interface NotificationWebhook extends AbstractNotification {
    integration_type: IntegrationType.WEBHOOK;
    webhook: NotificationWebhookEntity;
}

export type Notification = NotificationEmail | NotificationSyslog | NotificationWebhook;
