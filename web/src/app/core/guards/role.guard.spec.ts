import { TestBed } from '@angular/core/testing';
import { Router, UrlTree, provideRouter } from '@angular/router';

import { AuthService } from '../services/auth.service';
import { roleGuard } from './role.guard';

describe('roleGuard', () => {
  let authService: AuthService;
  let router: Router;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideRouter([])],
    });

    authService = TestBed.inject(AuthService);
    router = TestBed.inject(Router);
  });

  it('blocks navigation for wrong role', () => {
    spyOn(authService, 'currentRole').and.returnValue('practitioner');

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['manager'] } } as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/staff/practitioner/attendance-children');
  });

  it('allows navigation for matching role', () => {
    spyOn(authService, 'currentRole').and.returnValue('manager');

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['manager'] } } as never, {} as never),
    );

    expect(result).toBeTrue();
  });

  it('redirects unknown role state to sign-in', () => {
    spyOn(authService, 'currentRole').and.returnValue(null);

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['manager'] } } as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/signin');
  });
});
