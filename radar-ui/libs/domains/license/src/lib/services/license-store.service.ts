import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Store } from '@ngrx/store';

import { LicenseState } from '../interfaces';
import { getAppVersion, getCentralUrl, getHostAppVersion } from '../stores/license-selector.store';

@Injectable({
    providedIn: 'root'
})
export class LicenseStoreService {
    readonly appVersion$: Observable<string> = this.store.select(getAppVersion);

    readonly hostAppVersion$: Observable<string> = this.store.select(getHostAppVersion);

    readonly centralUrl$: Observable<string> = this.store.select(getCentralUrl);

    constructor(private readonly store: Store<LicenseState>) {}
}
