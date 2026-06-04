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

  it('redirects practitioner accessing manager-only route to /staff/practitioner/attendance', () => {
    spyOn(authService, 'currentRole').and.returnValue('practitioner');

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['manager'] } } as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/staff/practitioner/attendance');
  });

  it('redirects parent accessing manager-only route to /parent/invoices', () => {
    spyOn(authService, 'currentRole').and.returnValue('parent');

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['manager'] } } as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/parent/invoices');
  });

  it('redirects manager accessing parent-only route to /staff/manager/dashboard', () => {
    spyOn(authService, 'currentRole').and.returnValue('manager');

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['parent'] } } as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/staff/manager/dashboard');
  });

  it('allows manager on manager+practitioner route', () => {
    spyOn(authService, 'currentRole').and.returnValue('manager');

    const result = TestBed.runInInjectionContext(() =>
      roleGuard({ data: { roles: ['manager', 'practitioner'] } } as never, {} as never),
    );

    expect(result).toBeTrue();
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
