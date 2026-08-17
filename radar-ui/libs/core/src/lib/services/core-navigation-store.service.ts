import { Store } from '@ngrx/store';
import { BehaviorSubject, Observable } from 'rxjs';
import { Injectable, inject } from '@angular/core';

import {
    CoreNavigationState,
    getCurrentRouteSlug,
    getCurrentRouterName
} from '../stores/navigation/core-navigation-selector.store';
import { LoadStatus, RouterName } from '../constants';

@Injectable({
    providedIn: 'root'
})
export class CoreNavigationStoreService {
    private readonly store = inject<Store<CoreNavigationState>>(Store);

    readonly loadStatus$ = new BehaviorSubject<LoadStatus>(LoadStatus.INIT);

    readonly routeSlug$: Observable<string> = this.store.select(getCurrentRouteSlug);

    readonly routerName$: Observable<RouterName> = this.store.select(getCurrentRouterName);

    setLoadStatus(status: LoadStatus) {
        this.loadStatus$.next(status);
    }
}
