import { KbqTagInputEvent } from '@koobiq/components/tags';
import { PopUpPlacements } from '@koobiq/components/core';
import { Router } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AfterViewInit, ChangeDetectionStrategy, Component, DestroyRef, Inject, OnInit } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, debounceTime, distinctUntilChanged, map, startWith, switchMap } from 'rxjs';

import { DetectorExtended, DetectorStoreService, DetectorType } from '@cs/domains/detector';
import {
    FORM_SEPARATOR_KEY_CODES,
    FORM_VALIDATION_REG_EXP,
    FormScheme,
    RouterName,
    CoreUtilsService as utils
} from '@cs/core';
import { Notification, NotificationStoreService } from '@cs/domains/notification';
import { RuleSeverity, RuleType } from '@cs/domains/rule';

import { RuleForm, SharedRuleSidepanelFormProps } from './shared-rule-sidepanel-form.interface';

@Component({
    templateUrl: './shared-rule-sidepanel-form.component.html',
    styleUrl: './shared-rule-sidepanel-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class SharedRuleSidepanelFormComponent implements AfterViewInit, OnInit {
    readonly form: FormGroup<
        FormScheme<
            RuleForm,
            never,
            'imageNames' | 'registries' | 'namespaces' | 'pods' | 'containers' | 'nodes' | 'binaries'
        >
    > = this.formBuilder.group({
        name: ['', Validators.required],
        namespaces: this.formBuilder.array<string>([], Validators.required),
        pods: this.formBuilder.array<string>([], Validators.required),
        containers: this.formBuilder.array<string>([], Validators.required),
        nodes: this.formBuilder.array<string>([], Validators.required),
        imageNames: this.formBuilder.array<string>([], Validators.required),
        registries: this.formBuilder.array<string>([], Validators.required),
        binaries: this.formBuilder.array<string>([]),
        notifySeverity: [RuleSeverity.NONE],
        mailIds: [[] as string[]],
        detectors: [[] as string[]]
    });

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

            return utils.isFormValid(this.form.controls) && values.notifySeverity !== RuleSeverity.NONE;
        })
    );

    private readonly isMailIdsControlEnable$: Observable<void> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => utils.getFormValues<RuleForm>(this.form.controls)),
        switchMap((form) =>
            this.notificationStoreService.notificationsByEventType$(RuleType.TYPE_RUNTIME).pipe(
                map((notifications) => {
                    const mailIdsControlCondition = !!notifications.length && form.notifySeverity !== RuleSeverity.NONE;
                    utils.toggleControlEnable(this.form.get('mailIds'), mailIdsControlCondition, []);
                })
            )
        )
    );

    readonly detectors$: Observable<DetectorExtended[]> = this.detectorStoreService.detectors$([DetectorType.RUNTIME]);

    readonly tooltipPlacements = PopUpPlacements;

    readonly separatorKeyCodes = FORM_SEPARATOR_KEY_CODES;

    get imageNamesControl(): FormArray {
        return this.form.get('imageNames') as FormArray;
    }

    get registriesControl(): FormArray {
        return this.form.get('registries') as FormArray;
    }

    get namespacesControl(): FormArray {
        return this.form.get('namespaces') as FormArray;
    }

    get podsControl(): FormArray {
        return this.form.get('pods') as FormArray;
    }

    get nodesControl(): FormArray {
        return this.form.get('nodes') as FormArray;
    }

    get containersControl(): FormArray {
        return this.form.get('containers') as FormArray;
    }

    get binariesControl(): FormArray {
        return this.form.get('binaries') as FormArray;
    }

    constructor(
        private readonly destroyRef: DestroyRef,
        private readonly formBuilder: FormBuilder,
        private readonly notificationStoreService: NotificationStoreService,
        private readonly router: Router,
        private readonly sidepanelRef: KbqSidepanelRef,
        private readonly detectorStoreService: DetectorStoreService,
        @Inject(KBQ_SIDEPANEL_DATA) public readonly props: Partial<SharedRuleSidepanelFormProps>
    ) {}

    ngOnInit() {
        this.isMailIdsControlEnable$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe();
    }

    ngAfterViewInit() {
        if (this.props.rule) {
            this.form.patchValue({
                name: this.props.rule.name || '',
                notifySeverity: this.props.rule.rule?.notify?.severity || RuleSeverity.NONE,
                mailIds: this.props.rule.rule?.notify?.targets || [],
                detectors: this.props.rule.rule?.whitelist.threats || []
            });
        }

        const registriesSub: string[] = [];
        const imageNamesSub: string[] = [];
        this.props.rule?.scope?.image_names.forEach((item) => {
            if (item.indexOf('/') !== -1) {
                const [r, i] = item.split('/');
                registriesSub.push(r);
                imageNamesSub.push(i);
            } else {
                imageNamesSub.push(item);
            }
        });

        utils.setArrayControlValue(this.namespacesControl, this.props.rule?.scope?.namespaces, this.formBuilder);
        utils.setArrayControlValue(this.podsControl, this.props.rule?.scope?.pods, this.formBuilder);
        utils.setArrayControlValue(this.nodesControl, this.props.rule?.scope?.nodes, this.formBuilder);
        utils.setArrayControlValue(this.containersControl, this.props.rule?.scope?.containers, this.formBuilder);
        utils.setArrayControlValue(this.binariesControl, this.props.rule?.rule?.whitelist?.binaries, this.formBuilder);
        utils.setArrayControlValue(this.imageNamesControl, imageNamesSub, this.formBuilder);
        utils.setArrayControlValue(
            this.registriesControl,
            registriesSub.concat(this.props.rule?.scope?.registries || []),
            this.formBuilder
        );

        if (this.props.isEdit) {
            this.form.get('blockSeverity')?.disable({ onlySelf: true });
            utils.toggleArrayControlEnable(this.nodesControl, true);
            utils.toggleArrayControlEnable(this.imageNamesControl, true);
            utils.toggleArrayControlEnable(this.registriesControl, true);
            utils.toggleArrayControlEnable(this.namespacesControl, true);
            utils.toggleArrayControlEnable(this.podsControl, true);
            utils.toggleArrayControlEnable(this.containersControl, true);
        }
    }

    addEntity(event: KbqTagInputEvent, control: FormArray) {
        const value = event.value.trim();

        if (value) {
            control.push(this.formBuilder.control(value, Validators.pattern(FORM_VALIDATION_REG_EXP.TEXT_SYMBOLS)));
            event.input.value = '';
        }
    }

    removeEntity(control: FormArray, id: number) {
        control.removeAt(id);
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
