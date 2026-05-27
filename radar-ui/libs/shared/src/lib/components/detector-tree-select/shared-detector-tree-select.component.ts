import {
    AbstractControl,
    ControlValueAccessor,
    NG_VALIDATORS,
    NG_VALUE_ACCESSOR,
    ValidationErrors,
    Validators
} from '@angular/forms';
import { ChangeDetectionStrategy, Component, Input, OnChanges, SimpleChanges, ViewChild } from '@angular/core';
import {
    FlatTreeControl,
    KbqTreeFlatDataSource,
    KbqTreeFlattener,
    KbqTreeOption,
    KbqTreeSelection
} from '@koobiq/components/tree';
import { KbqTreeSelect, KbqTreeSelectChange } from '@koobiq/components/tree-select';

import { I18nService } from '@cs/i18n';
import { CoreUtilsService as utils } from '@cs/core';
import { DETECTOR_TYPE, DetectorExtended, DetectorType } from '@cs/domains/detector';

interface SharedDetectorTreeNode {
    id: string;
    name: string;
    type: DetectorType;
    isExtra: boolean;
    children: SharedDetectorTreeNode[];
}

interface SharedDetectorTreeFlatNode {
    id: string;
    name: string;
    type: DetectorType;
    level: number;
    isExtra: boolean;
    isExpandable: boolean;
    parent: SharedDetectorTreeFlatNode | null;
}

@Component({
    selector: 'cs-detector-tree-select-component',
    templateUrl: './shared-detector-tree-select.component.html',
    styleUrl: './shared-detector-tree-select.component.scss',
    changeDetection: ChangeDetectionStrategy.OnPush,
    providers: [
        {
            provide: NG_VALUE_ACCESSOR,
            useExisting: SharedDetectorTreeSelectComponent,
            multi: true
        },
        {
            provide: NG_VALIDATORS,
            useExisting: SharedDetectorTreeSelectComponent,
            multi: true
        }
    ]
})
export class SharedDetectorTreeSelectComponent implements ControlValueAccessor, OnChanges {
    @ViewChild(KbqTreeSelection) tree!: KbqTreeSelection;

    @ViewChild(KbqTreeSelect) select!: KbqTreeSelect;

    @Input() id?: string;

    @Input() testLocator?: string;

    @Input() detectors?: DetectorExtended[] | null;

    // Additional options for runtime type.
    @Input() runtimeExtras?: string[] | null;

    @Input() titleLocalizationKeyCollection?: Map<DetectorType, string>;

    readonly treeControl = new FlatTreeControl<SharedDetectorTreeFlatNode>(
        this.getLevel,
        this.isExpandable,
        this.getValue,
        this.getViewValue
    );

    readonly treeFlattener = new KbqTreeFlattener<SharedDetectorTreeNode, SharedDetectorTreeFlatNode>(
        this.flattenerTransform,
        this.getLevel,
        this.isExpandable,
        this.getChildren
    );

    readonly dataSource = new KbqTreeFlatDataSource<SharedDetectorTreeNode, SharedDetectorTreeFlatNode>(
        this.treeControl,
        this.treeFlattener
    );

    isTouched = false;

    isDisabled = false;

    detectorIds: string[] = [];

    private selectedIds: string[] = [];

    private detectorTypes: string[] = DETECTOR_TYPE.map((item) => item.id.toString());

    // RuntimeExtras input clone which is used for removeOption functionality.
    private runtimeExtrasClone: string[] = [];

    constructor(private readonly i18nService: I18nService) {}

    ngOnChanges(changes: SimpleChanges) {
        /* eslint @typescript-eslint/dot-notation: "off" */
        const detectors = changes['detectors'];
        const runtimeExtras = changes['runtimeExtras'];

        if (detectors && !utils.isEqual(detectors.currentValue, detectors.previousValue)) {
            const extras = runtimeExtras ? (runtimeExtras.currentValue as string[]) : undefined;
            this.detectorIds = [];
            this.dataSource.data = this.buildTreeSourceData(this.removeIdDuplicate(this.detectors), extras);
        }
    }

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onChange = (detectorIds: string[]) => {};

    /* eslint @typescript-eslint/no-empty-function: "off" */
    onTouched = () => {};

    registerOnChange(fn: any) {
        this.onChange = fn;
    }

    registerOnTouched(fn: any) {
        this.onTouched = fn;
    }

    markAsTouched() {
        if (!this.isTouched) {
            this.isTouched = true;
            this.onTouched();
        }
    }

    setDisabledState(isDisabled: boolean) {
        this.isDisabled = isDisabled;
    }

    writeValue(detectorIds?: string[] | null) {
        if (detectorIds) {
            this.detectorIds = detectorIds;
            this.selectedIds = [...detectorIds];
        }
    }

    validate(control: AbstractControl): ValidationErrors | null {
        return control.hasValidator(Validators.required) && !control.value ? { required: true } : null;
    }

    changeSelection(event: KbqTreeSelectChange) {
        const option: KbqTreeOption = event.value;
        if (option.isExpandable) {
            this.tree.setStateChildren(option, !option.selected);
        }

        this.toggleParents(event.value.data.parent as SharedDetectorTreeFlatNode | null);
        this.markAsTouched();
    }

    changeDetectorIds(detectorIds: string[]) {
        const ids = detectorIds.filter((item) => !this.detectorTypes.includes(item));
        if (!utils.isEqual(ids, this.selectedIds)) {
            this.selectedIds = [...ids];

            if (!this.isDisabled) {
                this.onChange(this.selectedIds);
            }
        }
    }

    removeOption(node: SharedDetectorTreeFlatNode) {
        const selectedIds = this.select
            ? (this.select.selected as SharedDetectorTreeFlatNode[]).map((item) => item.id)
            : [];
        if (selectedIds.includes(node.id)) {
            this.runtimeExtrasClone.push(node.id);
            this.select.selectionModel.deselect(node);
            this.changeDetectorIds(this.selectedIds.filter((id) => id !== node.id));
        }
    }

    // @todo: it needs to fix extra calculations which are increase rerenders
    isExtraOptionRemoved(node: SharedDetectorTreeFlatNode): boolean {
        return this.runtimeExtrasClone.includes(node.id);
    }

    // @todo: create directive to reduce the count of rerenders
    isExtraOptionDisabled(node: SharedDetectorTreeFlatNode): boolean {
        if (this.select) {
            const selectedIds = (this.select.selected as SharedDetectorTreeFlatNode[]).map((item) => item.id);

            return node.isExtra && !selectedIds.includes(node.id);
        }

        return false;
    }

    hasChild(_: number, node: SharedDetectorTreeFlatNode) {
        return node.isExpandable;
    }

    searchChange(value: string) {
        this.treeControl.filterNodes(value);
    }

    private toggleParents(parent: SharedDetectorTreeFlatNode | null) {
        if (parent) {
            const descendants: SharedDetectorTreeFlatNode[] = this.treeControl.getDescendants(parent);
            const isParentSelected = this.select.selectionModel.isSelected(parent);
            const isAllDescendantsSelected = descendants.every((descendant: SharedDetectorTreeFlatNode) =>
                this.select.selectionModel.isSelected(descendant)
            );

            if (!isParentSelected && isAllDescendantsSelected) {
                this.select.selectionModel.select(parent);
                this.toggleParents(parent.parent);
            } else if (isParentSelected) {
                this.select.selectionModel.deselect(parent);
                this.toggleParents(parent.parent);
            }
        }
    }

    private flattenerTransform(
        node: SharedDetectorTreeNode,
        level: number,
        parent: SharedDetectorTreeFlatNode | null
    ): SharedDetectorTreeFlatNode {
        return {
            id: node.id,
            name: node.name,
            type: node.type,
            isExtra: node.isExtra,
            isExpandable: !!node.children.length,
            level,
            parent
        };
    }

    private getLevel(node: SharedDetectorTreeFlatNode): number {
        return node.level;
    }

    private isExpandable(node: SharedDetectorTreeFlatNode): boolean {
        return node.isExpandable;
    }

    private getChildren(node: SharedDetectorTreeNode): SharedDetectorTreeNode[] {
        return node.children;
    }

    private getValue(node: SharedDetectorTreeFlatNode): string {
        return node.id;
    }

    private getViewValue(node: SharedDetectorTreeFlatNode): string {
        return node.id;
    }

    private getTreeChildren(detectors: DetectorExtended[], extras: string[]): SharedDetectorTreeNode[] {
        const defaultVersion = 1;
        const extraDetectors: DetectorExtended[] = extras.map((id) => ({
            id: `${id}${defaultVersion}`,
            key: id,
            name: id,
            type: DetectorType.RUNTIME,
            description: '',
            version: defaultVersion
        }));

        return [...detectors, ...extraDetectors]
            .sort((a, b) => (a.key.toLowerCase() < b.key.toLowerCase() ? -1 : 1))
            .map((item) => ({
                id: item.key,
                name: item.name,
                type: item.type,
                isExtra: item.name === item.key,
                children: []
            }));
    }

    private removeIdDuplicate(detectors?: DetectorExtended[] | null): DetectorExtended[] {
        if (!detectors || !detectors.length) {
            return [];
        }

        const dtctrs = new Map<string, DetectorExtended>();
        detectors.forEach((item) => {
            const obj = dtctrs.get(item.key);
            if (!obj || obj.version < item.version) {
                dtctrs.set(item.key, item);
            }
        });

        return Array.from(dtctrs.values());
    }

    private buildTreeSourceData(detectors?: DetectorExtended[] | null, extras?: string[]): SharedDetectorTreeNode[] {
        const dtctrs = detectors || [];
        if (extras && !dtctrs.length) {
            return extras.map((id) => ({
                id,
                name: id,
                type: DetectorType.RUNTIME,
                isExtra: true,
                children: []
            }));
        }

        const tree: SharedDetectorTreeNode[] = dtctrs
            .map((item) => item.type)
            .filter((item, i, self) => self.indexOf(item) === i)
            .map((type) => {
                const option = DETECTOR_TYPE.find((item) => item.id === type);
                const localizationKey = this.titleLocalizationKeyCollection?.get(type) || option?.localizationKey;
                const extraChildren =
                    type === DetectorType.RUNTIME
                        ? (extras || []).filter((item) => !dtctrs.map((dtc) => dtc.key).includes(item))
                        : [];

                return {
                    id: type.toString(),
                    type,
                    name: localizationKey ? this.i18nService.translate(localizationKey) : '',
                    isExtra: false,
                    children: this.getTreeChildren(
                        dtctrs.filter((item) => item.type === type),
                        extraChildren
                    )
                };
            });

        return tree;
    }
}
