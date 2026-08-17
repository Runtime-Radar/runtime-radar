import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { GetMitreMatrixResponse, MitreTactic } from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class MitreRequestService {
    private readonly http = inject(HttpClient);

    getMitreMatrix(locale: string): Observable<MitreTactic[]> {
        return this.http
            .get<GetMitreMatrixResponse>(`./assets/files/${locale}/mitre-matrix.json`)
            .pipe(map((response) => response.tactics));
    }
}
