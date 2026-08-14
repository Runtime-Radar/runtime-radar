import { ApplicationRef, Injectable, TemplateRef, inject } from '@angular/core';
import { BlockTranspiler, TranslationMarkupRenderer } from 'ngx-transloco-markup';

interface I18nTemplateTranspilerCreateRendererParams {
    template: TemplateRef<unknown>;
    context: unknown;
}

@Injectable({
    providedIn: 'root'
})
export class I18nTemplateTranspiler extends BlockTranspiler {
    private readonly applicationRef = inject(ApplicationRef);

    constructor() {
        super('[template]', '[/template]');
    }

    public createRenderer(childRenderers: TranslationMarkupRenderer[] = []): TranslationMarkupRenderer {
        const appRef = this.applicationRef;

        return function renderer(params): HTMLElement {
            const { template, context } = params as I18nTemplateTranspilerCreateRendererParams;
            const viewRef = template.createEmbeddedView(context);
            const el = viewRef.rootNodes[0] as HTMLElement;

            for (const child of childRenderers) {
                (el.querySelector('#cs-transloco-template') || el).appendChild(child(params));
            }

            appRef.attachView(viewRef);

            return el;
        };
    }
}
