import { Injectable, inject } from '@angular/core';
import { Meta, Title } from '@angular/platform-browser';

import { PageMeta } from '../interfaces/core-meta.interface';

@Injectable({
    providedIn: 'root'
})
export class CoreMetaService {
    private readonly metaService = inject(Meta);
    private readonly titleService = inject(Title);

    initPageMetaTags() {
        this.metaService.addTags([
            {
                name: 'description',
                content: ''
            }
        ]);
    }

    setPageMeta(meta: PageMeta) {
        this.titleService.setTitle(meta.title);
        this.metaService.updateTag({
            name: 'description',
            content: meta.description || ''
        });
    }
}
