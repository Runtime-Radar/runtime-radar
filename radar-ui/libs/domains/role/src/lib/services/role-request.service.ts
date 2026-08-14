import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { ApiEmptyRequest, ApiService } from '@cs/api';

import { GetRolesResponse, Role } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class RoleRequestService {
    private readonly apiService = inject(ApiService);

    getRoles(): Observable<Role[]> {
        return this.apiService.get<ApiEmptyRequest, GetRolesResponse>('role').pipe(map((response) => response.roles));
    }
}
