import {
    ChangeDetectionStrategy,
    Component,
    EventEmitter,
    Input,
    OnChanges,
    Output,
    booleanAttribute
} from '@angular/core';

import { DEFAULT_PAGINATOR_PAGE_INDEX, DEFAULT_PAGINATOR_PAGE_SIZE } from './shared-paginator.constant';

const PAGINATOR_SEPARATOR_KEY = Infinity;

const PAGINATOR_INDEX_INDENT = 2;

@Component({
    selector: 'cs-paginator-component',
    templateUrl: './shared-paginator.component.html',
    styleUrl: './shared-paginator.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class SharedPaginatorComponent implements OnChanges {
    // Optional prop for e2e framework.
    @Input() testLocator?: string;

    // The length of the total number of items that are being paginated.
    @Input() count = 0;

    // The number of items to display on a page.
    @Input() pageSize = DEFAULT_PAGINATOR_PAGE_SIZE;

    // The zero-based page index of the displayed list of items.
    @Input() pageIndex = DEFAULT_PAGINATOR_PAGE_INDEX;

    // Whether to show the first/last buttons UI to the user.
    @Input({ transform: booleanAttribute }) isShowFirstLastButtons = false;

    // Hide paginator if there is only one page.
    @Input({ transform: booleanAttribute }) isShowSinglePagination = false;

    // The event emitted when the paginator changes the page index.
    @Output() pageChange = new EventEmitter<number>();

    pageActiveIndex = 0;

    pageLastIndex = 0;

    pagesTemplateIterator: number[] = [];

    private pagesIterator: number[] = [];

    readonly separator = PAGINATOR_SEPARATOR_KEY;

    ngOnChanges() {
        this.pageActiveIndex = this.pageIndex - 1;

        if (this.count) {
            const count = Math.ceil(this.count / this.pageSize);
            this.pagesIterator = [...Array(count).keys()];
            this.pageLastIndex = count - 1;
            this.pagesTemplateIterator = this.generatePagination(this.pagesIterator);
        }
    }

    prev() {
        const index = this.pageActiveIndex < 1 ? 0 : --this.pageActiveIndex;
        this.navigate(index);
    }

    next() {
        const index = this.pageActiveIndex > this.pageLastIndex ? this.pageLastIndex : ++this.pageActiveIndex;
        this.navigate(index);
    }

    navigate(index: number) {
        this.pageActiveIndex = index;
        this.pagesTemplateIterator = this.generatePagination(this.pagesIterator);
        this.pageChange.emit(index + 1);
    }

    private generatePagination(pages: number[]): number[] {
        return pages
            .reduce<number[]>((acc, item) => {
                let isItemPushed = false;

                if (item < PAGINATOR_INDEX_INDENT) {
                    isItemPushed = true;
                    acc.push(item);
                }

                if (
                    !acc.includes(item) &&
                    item >= this.pageActiveIndex - PAGINATOR_INDEX_INDENT &&
                    item <= this.pageActiveIndex + PAGINATOR_INDEX_INDENT
                ) {
                    isItemPushed = true;
                    acc.push(item);
                }

                if (!acc.includes(item) && item > this.pageLastIndex - PAGINATOR_INDEX_INDENT) {
                    isItemPushed = true;
                    acc.push(item);
                }

                if (!isItemPushed) {
                    acc.push(PAGINATOR_SEPARATOR_KEY);
                }

                return acc;
            }, [])
            .filter((item, i, self) => item !== self[i + 1]);
    }
}
