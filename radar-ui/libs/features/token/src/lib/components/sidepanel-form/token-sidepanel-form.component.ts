import { DateAdapter } from '@koobiq/components/core';
import { DateTime } from 'luxon';
import { KbqAlertColors } from '@koobiq/components/alert';
import { ChangeDetectionStrategy, Component, OnInit, inject } from '@angular/core';
import { FormBuilder, FormGroup, FormRecord, Validators } from '@angular/forms';
import { KBQ_SIDEPANEL_DATA, KbqSidepanelRef } from '@koobiq/components/sidepanel';
import { Observable, debounceTime, distinctUntilChanged, distinctUntilKeyChanged, map, startWith, tap } from 'rxjs';

import { TokenPermissionName } from '@cs/domains/token';
import { FormScheme, CoreUtilsService as utils } from '@cs/core';
import { PermissionName, PermissionType } from '@cs/domains/role';

import { TokenSidepanelFormProps } from '../../interfaces/token-sidepanel.interface';
import {
    TokenExpiryDatePreset,
    TokenExpiryDatePresetOption,
    TokenForm,
    TokenPermissionRecord,
    TokenPermissionType
} from '../../interfaces/token-form.interface';

const TOKEN_EXPIRY_DATE_PRESET: TokenExpiryDatePresetOption[] = [
    {
        id: TokenExpiryDatePreset.WEEK,
        localizationKey: 'Token.Pseudo.ExpiryDatePreset.Week'
    },
    {
        id: TokenExpiryDatePreset.MONTH,
        localizationKey: 'Token.Pseudo.ExpiryDatePreset.Month'
    },
    {
        id: TokenExpiryDatePreset.QUARTER,
        localizationKey: 'Token.Pseudo.ExpiryDatePreset.Quarter'
    },
    {
        id: TokenExpiryDatePreset.INDEFINITELY,
        localizationKey: 'Token.Pseudo.ExpiryDatePreset.Indefinitely'
    },
    {
        id: TokenExpiryDatePreset.CUSTOM,
        localizationKey: 'Token.Pseudo.ExpiryDatePreset.Custom'
    }
];

@Component({
    templateUrl: './token-sidepanel-form.component.html',
    styleUrl: './token-sidepanel-form.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class TokenFeatureSidepanelFormComponent implements OnInit {
    private readonly dateAdapter = inject<DateAdapter<DateTime>>(DateAdapter);
    private readonly formBuilder = inject(FormBuilder);
    private readonly sidepanelRef = inject(KbqSidepanelRef);

    readonly props = inject<TokenSidepanelFormProps>(KBQ_SIDEPANEL_DATA);

    readonly form: FormGroup<FormScheme<TokenForm, 'permissions'>> = this.formBuilder.group({
        name: ['', Validators.required],
        date: [this.dateAdapter.today()],
        preset: [TokenExpiryDatePreset.WEEK],
        permissions: this.formBuilder.record<TokenPermissionRecord>({})
    });

    readonly expiryDatePreset$: Observable<TokenExpiryDatePreset> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        distinctUntilKeyChanged('preset'),
        map(() => utils.getFormValues<TokenForm>(this.form.controls).preset),
        tap((preset) => {
            const control = this.form.get('date');
            utils.toggleControlEnable(control, preset === TokenExpiryDatePreset.CUSTOM);

            switch (preset) {
                case TokenExpiryDatePreset.WEEK:
                    control?.setValue(this.dateAdapter.today().plus({ days: 7 }), { onlySelf: true });
                    break;
                case TokenExpiryDatePreset.MONTH:
                    control?.setValue(this.dateAdapter.today().plus({ months: 1 }), { onlySelf: true });
                    break;
                case TokenExpiryDatePreset.QUARTER:
                    control?.setValue(this.dateAdapter.today().plus({ months: 3 }), { onlySelf: true });
                    break;
                case TokenExpiryDatePreset.INDEFINITELY:
                    control?.setValue(null, { onlySelf: true });
                    break;
                case TokenExpiryDatePreset.CUSTOM:
                    control?.setValue(this.dateAdapter.today().plus({ days: 1 }), { onlySelf: true });
                    break;
            }
        })
    );

    readonly isFormValid$: Observable<boolean> = this.form.valueChanges.pipe(
        startWith(this.form.value),
        /* eslint @typescript-eslint/no-magic-numbers: "off" */
        debounceTime(250),
        distinctUntilChanged(),
        map(() => {
            const permissions = utils.getFormValues<TokenForm>(this.form.controls).permissions;
            const arePermissionsValid = Object.values(permissions)
                .reduce<boolean[]>((acc, item) => acc.concat(Object.values(item)), [])
                .some((item) => !!item);

            return utils.isFormValid(this.form.controls) && arePermissionsValid;
        })
    );

    get permissionsFormGroup(): FormRecord {
        return this.form.get('permissions') as FormRecord;
    }

    readonly minDate = this.dateAdapter.today().plus({ days: 1 });

    readonly dateTimeFullFormat = DateTime.DATETIME_FULL;

    readonly tokenExpiryDatePresetOptions: TokenExpiryDatePresetOption[] = TOKEN_EXPIRY_DATE_PRESET;

    readonly tokenExpiryDatePreset = TokenExpiryDatePreset;

    readonly tokenPermissionType = TokenPermissionType;

    readonly permissionType = PermissionType;

    readonly permissionName = PermissionName;

    readonly alertColors = KbqAlertColors;

    ngOnInit() {
        const clusterPermissionsForm = this.formBuilder.group({
            [TokenPermissionType.READ]: [
                { value: false, disabled: !this.props.permissions[PermissionName.CLUSTERS].has(PermissionType.READ) }
            ]
        });

        const rulePermissionsForm = this.formBuilder.group({
            [TokenPermissionType.CREATE]: [
                { value: false, disabled: !this.props.permissions[PermissionName.RULES].has(PermissionType.CREATE) }
            ],
            [TokenPermissionType.READ]: [
                { value: false, disabled: !this.props.permissions[PermissionName.RULES].has(PermissionType.READ) }
            ],
            [TokenPermissionType.UPDATE]: [
                { value: false, disabled: !this.props.permissions[PermissionName.RULES].has(PermissionType.UPDATE) }
            ],
            [TokenPermissionType.DELETE]: [
                { value: false, disabled: !this.props.permissions[PermissionName.RULES].has(PermissionType.DELETE) }
            ]
        });

        const eventPermissionsForm = this.formBuilder.group({
            [TokenPermissionType.READ]: [
                { value: false, disabled: !this.props.permissions[PermissionName.EVENTS].has(PermissionType.READ) }
            ]
        });

        this.permissionsFormGroup.addControl(TokenPermissionName.CLUSTERS, clusterPermissionsForm);
        this.permissionsFormGroup.addControl(TokenPermissionName.RULES, rulePermissionsForm);
        this.permissionsFormGroup.addControl(TokenPermissionName.EVENTS, eventPermissionsForm);
    }

    confirm() {
        const formValues = utils.getFormValues<TokenForm>(this.form.controls);
        this.sidepanelRef.close(utils.getTrimmedFormValues<TokenForm>(formValues));
    }

    cancel() {
        this.sidepanelRef.close(undefined);
    }
}
