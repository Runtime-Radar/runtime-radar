import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Store } from '@ngrx/store';

import { LoadStatus } from '@cs/core';

import {
    APPLY_AUTH_TOKENS_TODO_ACTION,
    EXPIRE_AUTH_TOKENS_TODO_ACTION,
    GET_LOCATION_PATH_TODO_ACTION,
    REDIRECT_TO_SWITCH_ROUTE_TODO_ACTION,
    SIGN_IN_TODO_ACTION,
    SIGN_OUT_TODO_ACTION
} from '../stores/auth-action.store';
import { AuthCredentials, AuthState } from '../interfaces';
import { getAuthCredentials, getAuthLoadStatus } from '../stores/auth-selector.store';

@Injectable({
    providedIn: 'root'
})
export class AuthStoreService {
    readonly credentials$: Observable<AuthCredentials> = this.store.select(getAuthCredentials);

    readonly loadStatus$: Observable<LoadStatus> = this.store.select(getAuthLoadStatus);

    constructor(private readonly store: Store<AuthState>) {}

    getLocationPath(location: Location) {
        this.store.dispatch(
            GET_LOCATION_PATH_TODO_ACTION({
                location: JSON.parse(JSON.stringify(location))
            })
        );
    }

    applyTokens() {
        this.store.dispatch(APPLY_AUTH_TOKENS_TODO_ACTION());
    }

    destroyTokens() {
        this.store.dispatch(EXPIRE_AUTH_TOKENS_TODO_ACTION());
    }

    redirectToSwitchRoute() {
        this.store.dispatch(REDIRECT_TO_SWITCH_ROUTE_TODO_ACTION());
    }

    signIn(username: string, password: string) {
        this.store.dispatch(SIGN_IN_TODO_ACTION({ username, password }));
    }

    signOut() {
        this.store.dispatch(SIGN_OUT_TODO_ACTION());
    }
}
