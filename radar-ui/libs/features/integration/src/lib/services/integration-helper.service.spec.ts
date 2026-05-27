import { CoreUtilsService as utils } from '@cs/core';

import { IntegrationFeatureHelperService } from './integration-helper.service';
import { INTEGRATION_RECIPIENT_HEADERS, INTEGRATION_RECIPIENT_RECORD } from '../mocks/integration.mock';

describe('IntegrationFeatureHelperService', () => {
    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('convertHeadersToRequestNode', () => {
        it('should return converted headers', () => {
            expect(IntegrationFeatureHelperService.convertHeadersToRequestNode(INTEGRATION_RECIPIENT_RECORD)).toEqual(
                INTEGRATION_RECIPIENT_HEADERS
            );
        });
    });

    describe('convertResponseNodeToHeaders', () => {
        it('should return empty record', () => {
            expect(IntegrationFeatureHelperService.convertResponseNodeToHeaders()).toEqual({});
        });

        it('should return converted records', () => {
            jest.spyOn(utils, 'generateUuid').mockReturnValueOnce('uuid1').mockReturnValueOnce('uuid2');

            expect(IntegrationFeatureHelperService.convertResponseNodeToHeaders(INTEGRATION_RECIPIENT_HEADERS)).toEqual(
                INTEGRATION_RECIPIENT_RECORD
            );
        });
    });
});
