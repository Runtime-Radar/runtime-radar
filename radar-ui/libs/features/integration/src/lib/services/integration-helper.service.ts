import { Injectable } from '@angular/core';

import { NotificationWebhookHeadersList } from '@cs/domains/notification';
import { CoreUtilsService as utils } from '@cs/core';

import { IntegrationRecipientTemplateRecord } from '../interfaces/integration-recipient-form.interace';

@Injectable({
    providedIn: 'root'
})
export class IntegrationFeatureHelperService {
    static convertHeadersToRequestNode(record: IntegrationRecipientTemplateRecord): NotificationWebhookHeadersList {
        return Object.keys(record).reduce<NotificationWebhookHeadersList>((acc, key) => {
            const item = record[key];
            acc[item.key] = item.value;

            return acc;
        }, {});
    }

    static convertResponseNodeToHeaders(list?: NotificationWebhookHeadersList): IntegrationRecipientTemplateRecord {
        const record: IntegrationRecipientTemplateRecord = {};

        if (!list) {
            return record;
        }

        Object.keys(list).forEach((key) => {
            record[utils.generateUuid()] = {
                key,
                value: list[key]
            };
        });

        return record;
    }
}
