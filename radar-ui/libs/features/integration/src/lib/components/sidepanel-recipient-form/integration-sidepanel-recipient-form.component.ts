import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AfterViewInit, ChangeDetectionStrategy, Component, DestroyRef, OnInit, inject } from '@angular/core';
import { FormBuilder, FormGroup, FormRecord, Validators } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import {
    Observable,
    debounceTime,
    distinctUntilChanged,
    distinctUntilKeyChanged,
    map,
    of,
    startWith,
    switchMap,
    take,
    tap
} from 'rxjs';

import { I18nService } from '@cs/i18n';
import { IntegrationType } from '@cs/domains/integration';
import { RegisteredCluster } from '@cs/domains/cluster';
import { FORM_VALIDATION_REG_EXP, FormScheme, CoreUtilsService as utils } from '@cs/core';
import { NotificationRequestService, NotificationWebhookHeadersList } from '@cs/domains/notification';

import { IntegrationSidepanelRecipientFormProps } from '../../interfaces/integration-sidepanel.interface';
import { IntegrationFeatureHelperService as integrationHelper } from '../../services/integration-helper.service';
import {
    IntegrationRecipientForm,
    IntegrationRecipientTemplateHeaderForm,
    IntegrationRecipientTemplateRecord
} from '../../interfaces/integration-recipient-form.interace';

const DEFAULT_WEBHOOK_HEADERS: NotificationWebhookHeadersList = {
    ['Content-type']: 'application/json'
};

const INTEGRATION_SUBJECT_TEMPLATES = new Map<string, string>([
    ['runtime_event', 'Integration.RecipientForm.Value.Subject.Runtime']
]);

@Component({
    templateUrl: './integration-sidepanel-recipient-form.component.html',
    styleUrl: './integration-sidepanel-recipient-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class IntegrationFeatureSidepanelRecipientFormComponent implements OnInit, AfterViewInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    private readonly i18nService = inject(I18nService);
    private readonly notificationRequestService = inject(NotificationRequestService);

    readonly props = inject<IntegrationSidepanelRecipientFormProps>(KBQ_SIDEPANEL_DATA);

    readonly form: FormGroup<FormScheme<IntegrationRecipientForm, 'header'>> = this.formBuilder.group({
        name: ['', Validators.required],
        recipients: [[] as string[]],
        eventType: ['runtime_event'],
        clusterId: [''],
        clusterUrl: [''],
        clusterName: [{ value: '', disabled: true }, Validators.required],
        centralUrl: ['', [Validators.required, Validators.pattern(FORM_VALIDATION_REG_EXP.IP_DOMAIN_SCHEME)]],
        template: [''],
        isTemplateDefault: [true],
        subjectTemplate: [''],
        path: [''],
        header: this.formBuilder.record<IntegrationRecipientTemplateRecord>({})
    });

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.isFormValid(this.form.controls))
    );

    readonly isTemplateDefault$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilChanged((a, b) => a.eventType === b.eventType && a.isTemplateDefault === b.isTemplateDefault),
        map(() => utils.getFormValues<IntegrationRecipientForm>(this.form.controls)),
        switchMap((form) => {
            const integrationType = this.props.integration?.type;
            if (!form.isTemplateDefault && integrationType) {
                return this.notificationRequestService.getNotificationTemplate(form.eventType, integrationType).pipe(
                    tap((template) => {
                        if (!this.props.isEdit || (this.props.isEdit && !this.props.notification?.template)) {
                            this.form.get('template')?.setValue(template, { onlySelf: true });
                        } else if (this.props.isEdit && this.props.notification?.template) {
                            this.form.get('template')?.setValue(this.props.notification.template, { onlySelf: true });
                        }
                    }),
                    map(() => form.isTemplateDefault)
                );
            }

            return of(form.isTemplateDefault);
        }),
        tap((isDefault) => {
            utils.toggleControlEnable(this.form.get('template'), !isDefault);
        })
    );

    private readonly formEventTypeSubjectTemplate$: Observable<string> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('eventType'),
        map(() => utils.getFormValues<IntegrationRecipientForm>(this.form.controls).eventType),
        tap((type) => {
            if (!this.props.isEdit && this.props.integration?.type === IntegrationType.EMAIL) {
                const localizationKey = INTEGRATION_SUBJECT_TEMPLATES.get(type);
                if (localizationKey) {
                    this.form
                        .get('subjectTemplate')
                        ?.setValue(this.i18nService.translate(localizationKey), { onlySelf: true });
                } else {
                    console.warn('eventType must be provided');
                }
            }
        })
    );

    private readonly defaultCentralUrl$: Observable<string> = this.props.centralUrl$.pipe(
        tap((centralUrl) => {
            if (centralUrl && !this.props.isEdit) {
                this.form.get('centralUrl')?.setValue(centralUrl, { onlySelf: true });
            }
        })
    );

    private readonly clusterName$: Observable<RegisteredCluster | undefined> = this.props.activeRegisteredCluster$.pipe(
        switchMap((cluster) =>
            this.props.centralUrl$.pipe(
                take(1),
                map((centralUrl) => ({
                    cluster,
                    centralUrl
                }))
            )
        ),
        tap(({ cluster, centralUrl }) => {
            if (cluster && !this.props.isEdit) {
                this.form.get('clusterId')?.setValue(cluster.id, { onlySelf: true });
                this.form.get('clusterUrl')?.setValue(cluster.own_cs_url || centralUrl, { onlySelf: true });
                this.form
                    .get('clusterName')
                    ?.setValue(cluster.name || this.i18nService.translate('Common.Pseudo.Label.CentralCluster'), {
                        onlySelf: true
                    });
            }
        }),
        map(({ cluster }) => cluster)
    );

    readonly integrationType = IntegrationType;

    readonly formValidationRegExp = FORM_VALIDATION_REG_EXP;

    get headerFormGroup(): FormRecord {
        return this.form.get('header') as FormRecord;
    }

    ngOnInit() {
        this.formEventTypeSubjectTemplate$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
        this.defaultCentralUrl$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
        this.clusterName$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.props.integration?.type === IntegrationType.EMAIL) {
            this.form.get('recipients')?.addValidators(Validators.required);
            this.form.get('subjectTemplate')?.addValidators(Validators.required);
        }

        if (this.props.integration?.type === IntegrationType.WEBHOOK && !this.props.isEdit) {
            this.setHeaderForm(DEFAULT_WEBHOOK_HEADERS);
        }

        if (this.props.notification && this.props.isEdit) {
            this.form.get('eventType')?.disable({ onlySelf: true });

            this.form.patchValue({
                name: this.props.notification.name,
                eventType: this.props.notification.event_type,
                clusterId: this.props.notification.cs_cluster_id,
                clusterUrl: this.props.notification.own_cs_url,
                clusterName: this.props.notification.cs_cluster_name,
                centralUrl: this.props.notification.central_cs_url,
                template: this.props.notification.template,
                isTemplateDefault: !this.props.notification.template,
                recipients: this.props.notification.recipients,
                subjectTemplate:
                    this.props.notification.integration_type === IntegrationType.EMAIL
                        ? this.props.notification.email.subject_template
                        : '',
                path:
                    this.props.notification.integration_type === IntegrationType.WEBHOOK
                        ? this.props.notification.webhook.path
                        : ''
            });

            if (this.props.notification.integration_type === IntegrationType.WEBHOOK) {
                this.setHeaderForm(this.props.notification.webhook.headers);
            }
        }
    }

    addHeaderItem() {
        const headerForm: FormGroup<FormScheme<IntegrationRecipientTemplateHeaderForm>> = this.formBuilder.group({
            key: ['', Validators.required],
            value: ['', Validators.required]
        });

        this.headerFormGroup.addControl(utils.generateUuid(), headerForm);
    }

    removeHeaderItem(id: string) {
        this.headerFormGroup.removeControl(id);
    }

    confirm() {
        if (this.form.valid) {
            const formValues = utils.getFormValues<IntegrationRecipientForm>(this.form.controls);
            this.sidepanelRef.close(utils.getTrimmedFormValues<IntegrationRecipientForm>(formValues));
        }
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }

    private setHeaderForm(list: NotificationWebhookHeadersList) {
        const headers = integrationHelper.convertResponseNodeToHeaders(list);

        Object.keys(headers).forEach((key) => {
            const control: IntegrationRecipientTemplateHeaderForm = headers[key];

            if (control) {
                // All default values should be disabled
                const isDisabled = Object.entries(DEFAULT_WEBHOOK_HEADERS)
                    .map(([id, value]) => id + value)
                    .includes(control.key + control.value);

                this.headerFormGroup.addControl(
                    key,
                    this.formBuilder.group({
                        key: { value: control.key, disabled: isDisabled },
                        value: { value: control.value, disabled: isDisabled }
                    })
                );
            }
        });
    }
}
