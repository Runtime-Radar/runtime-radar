import { ReplaySubject } from 'rxjs';
import { Store } from '@ngrx/store';
import { provideAutoSpy } from 'jest-auto-spies';
import { render } from '@testing-library/angular';
import { EnvironmentInjector, runInInjectionContext } from '@angular/core';

import { LoadStatus } from '@cs/core';

import { DEACTIVATE_RUNTIME_CONFIG_TODO_ACTION } from '../stores/runtime-action.store';
import { RuntimeState } from '../interfaces';
import { runtimeDeactivateGuard } from './runtime-deactivate.guard';

describe('runtimeDeactivateGuard', () => {
    let injector: EnvironmentInjector;
    let loadStatus$: ReplaySubject<LoadStatus>;
    let store: jest.Mocked<Store<RuntimeState>>;

    beforeEach(async () => {
        loadStatus$ = new ReplaySubject<LoadStatus>(LoadStatus.IN_PROGRESS);

        const { fixture } = await render('<div></div>', {
            providers: [provideAutoSpy(Store<RuntimeState>)]
        });

        injector = fixture.componentRef.injector.get(EnvironmentInjector);
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<RuntimeState>>;
        store.select.mockReturnValue(loadStatus$.asObservable());
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should emit DEACTIVATE_RUNTIME_CONFIG_TODO_ACTION', (done) => {
        runInInjectionContext(injector, () => {
            runtimeDeactivateGuard().subscribe(() => {
                expect(store.dispatch).toHaveBeenCalledWith(DEACTIVATE_RUNTIME_CONFIG_TODO_ACTION());
                done();
            });
        });

        loadStatus$.next(LoadStatus.LOADED);
        loadStatus$.next(LoadStatus.INIT);
    });
});
