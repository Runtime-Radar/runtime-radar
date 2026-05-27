import { ReplaySubject } from 'rxjs';
import { Store } from '@ngrx/store';
import { provideAutoSpy } from 'jest-auto-spies';
import { render } from '@testing-library/angular';
import { EnvironmentInjector, runInInjectionContext } from '@angular/core';

import { UPDATE_RUNTIME_STATE_DOC_ACTION } from '../stores/runtime-action.store';
import { runtimeConfigModifyDeactivateGuard } from './runtime-config-modify-deactivate.guard';
import { RuntimeConfigStatus, RuntimeState } from '../interfaces';

describe('runtimeConfigModifyDeactivateGuard', () => {
    let store: jest.Mocked<Store<RuntimeState>>;
    let status$: ReplaySubject<RuntimeConfigStatus>;
    let injector: EnvironmentInjector;

    beforeEach(async () => {
        status$ = new ReplaySubject(RuntimeConfigStatus.MODIFY);

        const { fixture } = await render('<div></div>', {
            providers: [provideAutoSpy(Store<RuntimeState>)]
        });

        injector = fixture.componentRef.injector.get(EnvironmentInjector);
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<RuntimeState>>;

        store.select.mockReturnValue(status$.asObservable());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION when status is MODIFY', (done) => {
        runInInjectionContext(injector, () => {
            runtimeConfigModifyDeactivateGuard().subscribe(() => {
                expect(store.dispatch).toHaveBeenCalledWith(
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        isOverlayed: true
                    })
                );
                done();
            });
        });

        status$.next(RuntimeConfigStatus.MODIFY);
        status$.next(RuntimeConfigStatus.STAY);
    });

    it('should emit UPDATE_RUNTIME_STATE_DOC_ACTION when status is STAY', (done) => {
        runInInjectionContext(injector, () => {
            runtimeConfigModifyDeactivateGuard().subscribe(() => {
                expect(store.dispatch).toHaveBeenCalledWith(
                    UPDATE_RUNTIME_STATE_DOC_ACTION({
                        configStatus: RuntimeConfigStatus.INIT,
                        isOverlayed: false
                    })
                );
                done();
            });
        });

        status$.next(RuntimeConfigStatus.STAY);
    });
});
