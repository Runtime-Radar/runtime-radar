import { Observable } from 'rxjs';
import { Store } from '@ngrx/store';
import { Injectable, inject } from '@angular/core';

import { Role } from '../interfaces';
import { RoleState } from '../interfaces/state/role-state.interface';
import { getRole, getRoles } from '../stores/role-selector.store';

@Injectable({
    providedIn: 'root'
})
export class RoleStoreService {
    private readonly store = inject<Store<RoleState>>(Store);

    readonly roles$: Observable<Role[]> = this.store.select(getRoles);

    readonly role$ = (id: string): Observable<Role | undefined> => this.store.select(getRole(id));
}
