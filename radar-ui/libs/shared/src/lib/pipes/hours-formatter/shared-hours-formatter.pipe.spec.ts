import { SharedHoursFormatterPipe } from './shared-hours-formatter.pipe';

describe('SharedHoursFormatterPipe', () => {
    let pipe: SharedHoursFormatterPipe;

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    jest.spyOn(console, 'warn').mockImplementation(() => {});

    beforeEach(() => {
        pipe = new SharedHoursFormatterPipe();
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should return default value', () => {
        expect(pipe.transform(null)).toBe(0);
    });

    it('should return hours to days correctly', () => {
        expect(pipe.transform(22)).toBe(1);
        expect(pipe.transform(23.9)).toBe(1);
        expect(pipe.transform(24)).toBe(1);
        expect(pipe.transform(26)).toBe(1);
        expect(pipe.transform(46)).toBe(2);
        expect(pipe.transform(48)).toBe(2);
    });
});
