import { FormControl, ValidatorFn } from '@angular/forms';

import { CoreValidators } from '../validators/core.validators';
import { FORM_VALIDATION_DENIED_IP } from '../constants';

describe('CoreValidators', () => {
    let segmentValidator: ValidatorFn;

    beforeEach(() => {
        segmentValidator = CoreValidators.isIpSegmentAllowed(FORM_VALIDATION_DENIED_IP.LOCALHOST);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('isIpSegmentAllowed', () => {
        it('should return segment validation errors', () => {
            expect(segmentValidator(new FormControl('127.10.0.10'))).toEqual({
                segment: {
                    allowedValue: '127.0.0.0',
                    actualValue: '127.10.0.10'
                }
            });
            expect(segmentValidator(new FormControl('127.10.100.0'))).toEqual({
                segment: {
                    allowedValue: '127.0.0.0',
                    actualValue: '127.10.100.0'
                }
            });
        });

        it('should return nullable values', () => {
            expect(segmentValidator(new FormControl('155.145.35.0'))).toBeNull();
            expect(segmentValidator(new FormControl(''))).toBeNull();
        });
    });
});
