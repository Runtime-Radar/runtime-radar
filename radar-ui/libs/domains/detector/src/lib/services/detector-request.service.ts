import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { ApiService } from '@cs/api';

import {
    CreateDetectorRequest,
    CreateDetectorResponse,
    DeleteDetectorRequest,
    Detector,
    EmptyDetectorResponse,
    GetDetectorsRequest,
    GetDetectorsResponse
} from '../interfaces';

@Injectable({
    providedIn: 'root'
})
export class DetectorRequestService {
    private readonly apiService = inject(ApiService);

    getRuntimeDetectors(request: GetDetectorsRequest, page = 1): Observable<GetDetectorsResponse> {
        return this.apiService.get<GetDetectorsRequest, GetDetectorsResponse>(
            `detector/event-processor/page/${page}`,
            request
        );
    }

    createRuntimeDetector(request: CreateDetectorRequest): Observable<Detector> {
        return this.apiService
            .post<CreateDetectorRequest, CreateDetectorResponse>('detector/event-processor', request)
            .pipe(map((response) => response.detector));
    }

    deleteRuntimeDetector(request: DeleteDetectorRequest): Observable<string> {
        return this.apiService
            .deleteWithRequest<DeleteDetectorRequest, EmptyDetectorResponse>('detector/event-processor', request)
            .pipe(
                map((response) => (response && !Object.keys(response).length ? `${request.id}${request.version}` : ''))
            );
    }
}
