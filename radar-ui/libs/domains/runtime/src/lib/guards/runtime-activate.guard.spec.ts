import { ReplaySubject } from 'rxjs';
import { Store } from '@ngrx/store';
import { provideAutoSpy } from 'jest-auto-spies';
import { render } from '@testing-library/angular';
import { EnvironmentInjector, runInInjectionContext } from '@angular/core';
import { Router, UrlTree } from '@angular/router';

import { LoadStatus, RouterName } from '@cs/core';

import { LOAD_RUNTIME_CONFIG_TODO_ACTION } from '../stores/runtime-action.store';
import { RuntimeState } from '../interfaces';
import { runtimeActivateGuard } from './runtime-activate.guard';

describe('runtimeActivateGuard', () => {
    let injector: EnvironmentInjector;
    let loadStatus$: ReplaySubject<LoadStatus>;
    let store: jest.Mocked<Store<RuntimeState>>;
    let router: jest.Mocked<Router>;

    beforeEach(async () => {
        loadStatus$ = new ReplaySubject<LoadStatus>(LoadStatus.IN_PROGRESS);

        const { fixture } = await render('<div></div>', {
            providers: [provideAutoSpy(Store<RuntimeState>), provideAutoSpy(Router)]
        });

        injector = fixture.componentRef.injector.get(EnvironmentInjector);
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<RuntimeState>>;
        store.select.mockReturnValue(loadStatus$.asObservable());
        router = fixture.debugElement.injector.get(Router) as jest.Mocked<Router>;
        router.createUrlTree.mockReturnValue({} as UrlTree);
    });

    afterEach(() => {
        jest.clearAllMocks();
    });

    it('should emit LOAD_RUNTIME_CONFIG_TODO_ACTION when status is INIT', (done) => {
        runInInjectionContext(injector, () => {
            runtimeActivateGuard().subscribe(() => {
                expect(store.dispatch).toHaveBeenCalledWith(LOAD_RUNTIME_CONFIG_TODO_ACTION());
                done();
            });
        });

        loadStatus$.next(LoadStatus.INIT);
        loadStatus$.next(LoadStatus.LOADED);
    });

    it('should return UrlTree when status is ERROR', (done) => {
        runInInjectionContext(injector, () => {
            runtimeActivateGuard().subscribe(() => {
                expect(router.createUrlTree).toHaveBeenCalledWith([RouterName.ERROR]);
                done();
            });
        });

        loadStatus$.next(LoadStatus.ERROR);
    });
});
