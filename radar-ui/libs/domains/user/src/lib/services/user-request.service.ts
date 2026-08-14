import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { GetTokenResponse } from '@cs/domains/auth';
import { ApiEmptyRequest, ApiService } from '@cs/api';

import {
    CreateUserRequest,
    DeleteUserResponse,
    GetUsersResponse,
    UpdatePasswordRequest,
    UpdateUserRequest,
    User
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class UserRequestService {
    private readonly apiService = inject(ApiService);

    getUsers(): Observable<User[]> {
        return this.apiService.get<ApiEmptyRequest, GetUsersResponse>('user').pipe(map((response) => response.users));
    }

    createUser(request: CreateUserRequest): Observable<User> {
        return this.apiService.post<CreateUserRequest, User>('user', request);
    }

    updateUser(id: string, request: UpdateUserRequest): Observable<User> {
        return this.apiService.patch<UpdateUserRequest, User>(`user/${id}`, request);
    }

    updatePassword(id: string, password: string): Observable<GetTokenResponse> {
        return this.apiService.patch<UpdatePasswordRequest, GetTokenResponse>(`user/${id}/password`, { password });
    }

    deleteUser(id: string): Observable<string> {
        return this.apiService.delete<DeleteUserResponse>(`user/${id}`).pipe(map((response) => response.id));
    }
}
