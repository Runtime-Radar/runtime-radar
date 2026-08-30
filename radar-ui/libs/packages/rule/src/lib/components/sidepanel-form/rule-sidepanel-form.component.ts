import { Router } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AfterViewInit, ChangeDetectionStrategy, Component, DestroyRef, OnInit, inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, debounceTime, distinctUntilChanged, map, startWith, switchMap } from 'rxjs';

import { DetectorExtended, DetectorStoreService, DetectorType } from '@cs/domains/detector';
import { FormScheme, RouterName, CoreUtilsService as utils } from '@cs/core';
import { Notification, NotificationStoreService } from '@cs/domains/notification';
import { RuleSeverity, RuleType } from '@cs/domains/rule';

import { RuleForm } from '../../interfaces/rule-form.interface';
import { RuleSidepanelFormProps } from '../../interfaces/rule-sidepanel.interface';

@Component({
    templateUrl: './rule-sidepanel-form.component.html',
    styleUrl: './rule-sidepanel-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class RulePackageSidepanelFormComponent implements OnInit, AfterViewInit {
    private readonly destroyRef = inject(DestroyRef);
    private readonly formBuilder = inject(FormBuilder);
    private readonly router = inject(Router);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    private readonly detectorStoreService = inject(DetectorStoreService);
    private readonly notificationStoreService = inject(NotificationStoreService);

    readonly props = inject<Partial<RuleSidepanelFormProps>>(KBQ_SIDEPANEL_DATA);

    readonly form: FormGroup<FormScheme<RuleForm>> = this.formBuilder.group({
        name: ['', Validators.required],
        namespaces: [[] as string[], Validators.required],
        pods: [[] as string[], Validators.required],
        containers: [[] as string[], Validators.required],
        nodes: [[] as string[], Validators.required],
        imageNames: [[] as string[], Validators.required],
        registries: [[] as string[], Validators.required],
        binaries: [[] as string[]],
        notifySeverity: [RuleSeverity.NONE],
        blockSeverity: [RuleSeverity.NONE],
        mailIds: [[] as string[]],
        detectors: [[] as string[]]
    });

    readonly detectors$: Observable<DetectorExtended[]> = this.detectorStoreService.detectors$([DetectorType.RUNTIME]);

    readonly notifications$: Observable<Notification[]> = this.notificationStoreService.notificationsByEventType$(
        RuleType.TYPE_RUNTIME
    );

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => {
            const values = utils.getFormValues<RuleForm>(this.form.controls);
            const hasSeverity = values.blockSeverity !== RuleSeverity.NONE || values.notifySeverity !== RuleSeverity.NONE;
            return utils.isFormValid(this.form.controls) && hasSeverity;
        })
    );

    private readonly isMailIdsControlEnable$: Observable<void> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.getFormValues<RuleForm>(this.form.controls).notifySeverity !== RuleSeverity.NONE),
        switchMap((hasSeverity) =>
            this.notificationStoreService.notificationsByEventType$(RuleType.TYPE_RUNTIME).pipe(
                map((notifications) => {
                    utils.toggleControlEnable(this.form.get('mailIds'), !!notifications.length && hasSeverity, []);
                })
            )
        )
    );

    ngOnInit() {
        this.isMailIdsControlEnable$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.props.rule) {
            this.form.patchValue({
                name: this.props.rule.name || '',
                namespaces: this.props.rule.scope?.namespaces || [],
                pods: this.props.rule.scope?.pods || [],
                nodes: this.props.rule.scope?.nodes || [],
                containers: this.props.rule.scope?.containers || [],
                binaries: this.props.rule.rule?.whitelist.binaries || [],
                imageNames: this.props.rule.scope?.image_names || [],
                registries: this.props.rule?.scope?.registries || [],
                notifySeverity: this.props.rule.rule?.notify?.severity || RuleSeverity.NONE,
                blockSeverity: this.props.rule.rule?.block?.severity || RuleSeverity.NONE,
                mailIds: this.props.rule.rule?.notify?.targets || [],
                detectors: this.props.rule.rule?.whitelist.threats || []
            });
        }
    }

    goToIntegrationPage() {
        this.router.navigate([RouterName.SETTINGS, RouterName.INTEGRATIONS]);
        this.sidepanelRef.close(undefined);
    }

    confirm() {
        const formValues = utils.getFormValues<RuleForm>(this.form.controls);
        this.sidepanelRef.close(utils.getTrimmedFormValues<RuleForm>(formValues));
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }
}
