import { BehaviorSubject } from 'rxjs';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';

import { INTEGRATION_TYPE, IntegrationType } from '@cs/domains/integration';

import { IntegrationSidepanelFormProps } from '../../interfaces/integration-sidepanel.interface';
import {
    IntegrationEmailForm,
    IntegrationSyslogForm,
    IntegrationWebhookForm
} from '../../interfaces/integration-form.interface';

@Component({
    templateUrl: './integration-sidepanel-form.component.html',
    styleUrl: './integration-sidepanel-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureSidepanelFormComponent {
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    readonly props = inject<Partial<IntegrationSidepanelFormProps>>(KBQ_SIDEPANEL_DATA);

    readonly isFormValid$ = new BehaviorSubject(false);

    readonly integrationType = IntegrationType;

    readonly integrationTypeOptions = INTEGRATION_TYPE;

    private emailFormValues?: IntegrationEmailForm;

    private syslogFormValues?: IntegrationSyslogForm;

    private webhookFormValues?: IntegrationWebhookForm;

    integrationTypeValue = this.props.type || IntegrationType.EMAIL;

    changeType() {
        this.isFormValid$.next(false);
        this.emailFormValues = undefined;
        this.syslogFormValues = undefined;
        this.webhookFormValues = undefined;
    }

    changeEmailForm(form?: IntegrationEmailForm) {
        this.isFormValid$.next(form === undefined ? false : true);
        this.emailFormValues = form;
    }

    changeSyslogForm(form?: IntegrationSyslogForm) {
        this.isFormValid$.next(form === undefined ? false : true);
        this.syslogFormValues = form;
    }

    changeWebhookForm(form?: IntegrationWebhookForm) {
        this.isFormValid$.next(form === undefined ? false : true);
        this.webhookFormValues = form;
    }

    confirmWithoutCheck(hasSkipCheck: boolean) {
        if (this.integrationTypeValue === IntegrationType.EMAIL && this.emailFormValues) {
            this.sidepanelRef.close({
                type: this.integrationTypeValue,
                email: this.emailFormValues,
                hasSkipCheck
            });
        } else if (this.integrationTypeValue === IntegrationType.SYSLOG && this.syslogFormValues) {
            this.sidepanelRef.close({
                type: this.integrationTypeValue,
                syslog: this.syslogFormValues,
                hasSkipCheck
            });
        } else if (this.integrationTypeValue === IntegrationType.WEBHOOK && this.webhookFormValues) {
            this.sidepanelRef.close({
                type: this.integrationTypeValue,
                webhook: this.webhookFormValues,
                hasSkipCheck
            });
        }
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }
}
