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
    let store: jest.Mocked<Store<RuntimeState>>;
    let router: jest.Mocked<Router>;
    let status$: ReplaySubject<LoadStatus>;
    let injector: EnvironmentInjector;

    beforeEach(async () => {
        status$ = new ReplaySubject<LoadStatus>(LoadStatus.IN_PROGRESS);

        const { fixture } = await render('<div></div>', {
            providers: [provideAutoSpy(Store<RuntimeState>), provideAutoSpy(Router)]
        });

        injector = fixture.componentRef.injector.get(EnvironmentInjector);
        store = fixture.debugElement.injector.get(Store) as jest.Mocked<Store<RuntimeState>>;
        router = fixture.debugElement.injector.get(Router) as jest.Mocked<Router>;

        store.select.mockReturnValue(status$.asObservable());
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

        status$.next(LoadStatus.INIT);
        status$.next(LoadStatus.LOADED);
    });

    it('should return UrlTree when status is ERROR', (done) => {
        runInInjectionContext(injector, () => {
            runtimeActivateGuard().subscribe(() => {
                expect(router.createUrlTree).toHaveBeenCalledWith([RouterName.ERROR]);
                done();
            });
        });

        status$.next(LoadStatus.ERROR);
    });
});
