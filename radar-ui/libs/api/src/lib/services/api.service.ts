import { Observable } from 'rxjs';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';

import { ApiPathService } from './api-path.service';
import { ApiRequestParams } from '../interfaces/api-request.interface';

@Injectable({
    providedIn: 'root'
})
export class ApiService {
    private readonly apiPathService = inject(ApiPathService);
    private readonly http = inject(HttpClient);

    private readonly apiPath = (path: string): string => this.apiPathService.get(path);

    get<R, T>(path: string, request: R = {} as R, headers?: HttpHeaders): Observable<T> {
        return this.http.get<T>(this.apiPath(path), {
            params: { ...request } as unknown as ApiRequestParams,
            headers: headers || undefined
        });
    }

    post<R extends object, T>(path: string, request: R): Observable<T> {
        return this.http.post<T>(this.apiPath(path), request);
    }

    patch<R, T>(path: string, request: R): Observable<T> {
        return this.http.patch<T>(this.apiPath(path), request);
    }

    delete<T>(path: string): Observable<T> {
        return this.http.delete<T>(this.apiPath(path));
    }

    deleteWithRequest<R, T>(path: string, request: R): Observable<T> {
        return this.http.delete<T>(this.apiPath(path), { params: request as unknown as ApiRequestParams });
    }
}
