import { KbqSidepanelService } from '@koobiq/components/sidepanel';
import { Spy, createSpyFromClass } from 'jest-auto-spies';

import { INVENTORY_SIDEPANEL_CONTEXT_NODE } from '../mocks/inventory-context.mock';
import { DEFAULT_CONTEXT, InventoryFeatureSidepanelContextService } from './inventory-sidepanel-context.service';
import { InventorySidepanelContext, InventorySidepanelContextType } from '../interfaces/inventory-sidepanel.interface';

const fakeSidepanelContext: InventorySidepanelContext = {
    id: 'id',
    sidepanelId: 'sidepanelId',
    path: 'path',
    type: InventorySidepanelContextType.NONE
};

const context = (path: string): InventorySidepanelContext => {
    return INVENTORY_SIDEPANEL_CONTEXT_NODE.find((item) => item.path === path) || fakeSidepanelContext;
};

describe('InventoryFeatureSidepanelContextService', () => {
    let service: InventoryFeatureSidepanelContextService;
    let sidepanelService: Spy<KbqSidepanelService>;

    beforeEach(() => {
        sidepanelService = createSpyFromClass(KbqSidepanelService, {
            methodsToSpyOn: ['getSidepanelById']
        });
        sidepanelService.getSidepanelById.mockReturnValue({
            close: jest.fn()
        } as any);

        service = new InventoryFeatureSidepanelContextService(sidepanelService);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    describe('set', () => {
        it('should add correct context', () => {
            const node = context('nodeId1');

            service.set(node);

            expect(service.get()).toEqual([node]);
        });
    });

    describe('remove', () => {
        const node = context('nodeId1');

        beforeEach(() => {
            service.set(node);
        });

        afterEach(() => {
            service.remove('nodeId1');
        });

        it('should not modify context', () => {
            service.remove('fakeId');

            expect(service.get()).toEqual([node]);
        });

        it('should remove selected context', () => {
            service.set(context('nodeId1:namespaceId1'));

            service.remove('namespaceId1');

            expect(service.get()).toEqual([node]);
        });

        it('should remove last context', () => {
            service.remove('nodeId1');

            expect(service.get()).toEqual([]);
            expect(service.context$.getValue()).toEqual(DEFAULT_CONTEXT);
        });
    });

    describe('slice', () => {
        describe('node', () => {
            it('should return context when open another node', () => {
                service.set(context('nodeId1'));

                service.slice(context('nodeId2'));

                expect(service.get()).toEqual([context('nodeId2')]);
            });

            it('should return context when open the same node in node/namespace/pod/container', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1'));

                expect(service.get()).toEqual([context('nodeId1')]);
            });

            it('should return context when open another node in node/pod/container', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId2'));

                expect(service.get()).toEqual([context('nodeId2')]);
            });
        });

        describe('namespace', () => {
            it('should return context when open another namespace in another node', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));

                service.slice(context('nodeId2:namespaceId3'));

                expect(service.get()).toEqual([context('nodeId2:namespaceId3')]);
            });

            it('should return context when open another namespace in another node', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));

                service.slice(context('nodeId2:namespaceId3'));

                expect(service.get()).toEqual([context('nodeId2:namespaceId3')]);
            });

            it('should return context when open another namespace in node/namespace', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));

                service.slice(context('nodeId1:namespaceId2'));

                expect(service.get()).toEqual([context('nodeId1'), context('nodeId1:namespaceId2')]);
            });

            it('should return context when open another namespace in namespace/pod', () => {
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));

                service.slice(context('nodeId1:namespaceId2'));

                expect(service.get()).toEqual([context('nodeId1:namespaceId2')]);
            });

            it('should return context when open another namespace in pod/container', () => {
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId2'));

                expect(service.get()).toEqual([context('nodeId1:namespaceId2')]);
            });

            it('should return context when open another namespace in node/pod/container #1', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId2'));

                expect(service.get()).toEqual([context('nodeId1'), context('nodeId1:namespaceId2')]);
            });

            it('should return context when open another namespace in node/pod/container #2', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId2:namespaceId3'));

                expect(service.get()).toEqual([context('nodeId2:namespaceId3')]);
            });

            it('should return context when open the same namespace in namespace/pod', () => {
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));

                service.slice(context('nodeId1:namespaceId1'));

                expect(service.get()).toEqual([context('nodeId1:namespaceId1')]);
            });

            it('should return context when open the same namespace in node/namespace/pod', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));

                service.slice(context('nodeId1:namespaceId1'));

                expect(service.get()).toEqual([context('nodeId1'), context('nodeId1:namespaceId1')]);
            });
        });

        describe('pod', () => {
            it('should return context when open another pod in another node', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId2:namespaceId3:podId4'));

                expect(service.get()).toEqual([context('nodeId2:namespaceId3:podId4')]);
            });

            it('should return context when open another pod in the same node', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId1:podId2'));

                expect(service.get()).toEqual([context('nodeId1'), context('nodeId1:namespaceId1:podId2')]);
            });

            it('should return context when open another pod in node/namespace', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId1:podId2'));

                expect(service.get()).toEqual([
                    context('nodeId1'),
                    context('nodeId1:namespaceId1'),
                    context('nodeId1:namespaceId1:podId2')
                ]);
            });

            it('should return context when open another pod in node/namespace/pod/container', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId2:podId3'));

                expect(service.get()).toEqual([context('nodeId1'), context('nodeId1:namespaceId2:podId3')]);
            });

            it('should return context when open another pod in namespace/pod', () => {
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));

                service.slice(context('nodeId1:namespaceId1:podId2'));

                expect(service.get()).toEqual([
                    context('nodeId1:namespaceId1'),
                    context('nodeId1:namespaceId1:podId2')
                ]);
            });

            it('should return context when open the same pod in namespace/pod/container', () => {
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId1:podId1'));

                expect(service.get()).toEqual([
                    context('nodeId1:namespaceId1'),
                    context('nodeId1:namespaceId1:podId1')
                ]);
            });
        });

        describe('container', () => {
            it('should return context when open another container', () => {
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId1:podId1:containerId2'));

                expect(service.get()).toEqual([context('nodeId1:namespaceId1:podId1:containerId2')]);
            });

            it('should return context when open another container in another namespace', () => {
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId2:podId3:containerId4'));

                expect(service.get()).toEqual([context('nodeId1:namespaceId2:podId3:containerId4')]);
            });

            it('should return context when open another container in another pod', () => {
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId1:podId2:containerId3'));

                expect(service.get()).toEqual([
                    context('nodeId1:namespaceId1'),
                    context('nodeId1:namespaceId1:podId2:containerId3')
                ]);
            });

            it('should return context when open another container in the same pod', () => {
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId1:podId1:containerId2'));

                expect(service.get()).toEqual([
                    context('nodeId1:namespaceId1:podId1'),
                    context('nodeId1:namespaceId1:podId1:containerId2')
                ]);
            });

            it('should return context when open another container in another node', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId2:namespaceId3:podId4:containerId5'));

                expect(service.get()).toEqual([context('nodeId2:namespaceId3:podId4:containerId5')]);
            });

            it('should return context when open another container in the same node', () => {
                service.set(context('nodeId1'));
                service.set(context('nodeId1:namespaceId1'));
                service.set(context('nodeId1:namespaceId1:podId1'));
                service.set(context('nodeId1:namespaceId1:podId1:containerId1'));

                service.slice(context('nodeId1:namespaceId2:podId3:containerId4'));

                expect(service.get()).toEqual([
                    context('nodeId1'),
                    context('nodeId1:namespaceId2:podId3:containerId4')
                ]);
            });
        });
    });
});
