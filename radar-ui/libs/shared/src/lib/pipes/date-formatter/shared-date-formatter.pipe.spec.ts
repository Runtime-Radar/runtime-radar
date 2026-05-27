import { DateAdapter } from '@koobiq/date-adapter';
import { DateTime } from 'luxon';

import { I18nService } from '@cs/i18n';

import { SharedDateFormatterPipe } from './shared-date-formatter.pipe';

describe('SharedDateFormatterPipe', () => {
    let pipe: SharedDateFormatterPipe;
    let i18nService: jest.Mocked<I18nService>;
    let dateAdapter: any;

    const dateStr = '2025-09-18T12:00:00.000000Z';
    const dateTime = DateTime.fromISO(dateStr);

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    jest.spyOn(console, 'warn').mockImplementation(() => {});

    beforeEach(() => {
        i18nService = {
            getLocale: jest.fn(() => 'en-US')
        } as unknown as jest.Mocked<I18nService>;
        dateAdapter = {
            deserialize: jest.fn(),
            setLocale: jest.fn(),
            toFormat: jest.fn()
        } as unknown as DateAdapter<DateTime>;

        pipe = new SharedDateFormatterPipe(dateAdapter, i18nService);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should return default value', () => {
        dateAdapter.deserialize.mockReturnValue(null);

        expect(pipe.transform('')).toEqual('');
        expect(pipe.transform('invalid')).toEqual('');
        expect(pipe.transform('2015-09-18')).toEqual('');
    });

    it('should return formatted date', () => {
        dateAdapter.deserialize.mockReturnValue(dateTime);
        dateAdapter.setLocale.mockReturnValue(dateTime);

        expect(pipe.transform(dateStr, DateTime.DATE_HUGE)).toEqual('Thursday, September 18, 2025');
        expect(pipe.transform(dateStr, DateTime.DATETIME_MED)).toEqual('Sep 18, 2025, 3:00 PM');
    });

    it('should return formatted date without specified format', () => {
        dateAdapter.deserialize.mockReturnValue(dateTime);
        dateAdapter.setLocale.mockReturnValue(dateTime);

        expect(pipe.transform(dateStr)).toEqual('September 18, 2025, 3:00 PM');
    });
});
