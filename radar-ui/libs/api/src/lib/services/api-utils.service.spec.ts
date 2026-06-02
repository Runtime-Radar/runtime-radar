import { HttpErrorResponse } from '@angular/common/http';

import { ApiUtilsService } from './api-utils.service';
import { ApiErrorCode, ApiProtoCode } from '../interfaces/contract/api-error-contract.interface';

describe('ApiUtilsService', () => {
    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should return UNKNOWN_ISSUE when structure is invalid', () => {
        const testCases = [{ error: { details: [] } }, { error: {} }, { error: 'error' }, { error: undefined }];

        testCases.forEach((testCase) => {
            expect(ApiUtilsService.getReasonCode(new HttpErrorResponse(testCase))).toEqual(ApiErrorCode.UNKNOWN_ISSUE);
        });
    });

    it('should return UNKNOWN_ISSUE when there is non-http error', () => {
        const response = new Error('Network error') as any;

        expect(ApiUtilsService.getReasonCode(response)).toEqual(ApiErrorCode.UNKNOWN_ISSUE);
    });

    it('should return correct reasons', () => {
        const response1 = new HttpErrorResponse({
            error: {
                code: ApiProtoCode.NOT_FOUND,
                details: [{ reason: ApiErrorCode.ACCESS_TOKEN_EXPIRED }, { reason: ApiErrorCode.REFRESH_TOKEN_EXPIRED }]
            }
        });
        const response2 = new HttpErrorResponse({
            error: {
                code: ApiProtoCode.NOT_FOUND,
                details: [{ reason: ApiErrorCode.REGISTRIES_UNAVAILABLE }]
            }
        });

        expect(ApiUtilsService.getReasonCode(response1)).toEqual(ApiErrorCode.ACCESS_TOKEN_EXPIRED);
        expect(ApiUtilsService.getReasonCode(response2)).toEqual(ApiErrorCode.REGISTRIES_UNAVAILABLE);
    });
});
