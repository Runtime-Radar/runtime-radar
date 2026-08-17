import { GridPackageColumnVisibleDirective } from './grid-column-visible.directive';

describe('GridPackageColumnVisibleDirective', () => {
    let directive: GridPackageColumnVisibleDirective;

    beforeEach(() => {
        directive = new GridPackageColumnVisibleDirective();
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should hide container', () => {
        directive.columnVisible = false;

        expect(directive.display).toEqual('none');
    });

    it('should show container', () => {
        directive.columnVisible = true;

        expect(directive.display).toEqual('');
    });

    it('should show container when value is null', () => {
        directive.columnVisible = null;

        expect(directive.display).toEqual('');
    });
});
