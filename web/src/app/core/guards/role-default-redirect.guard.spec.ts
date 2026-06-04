import { TestBed } from '@angular/core/testing';
import { Router, UrlTree, provideRouter } from '@angular/router';

import { ROLES } from '../constants/roles';
import { AuthService } from '../services/auth.service';
import { roleDefaultRedirectGuard } from './role-default-redirect.guard';

describe('roleDefaultRedirectGuard', () => {
  let authService: AuthService;
  let router: Router;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideRouter([])],
    });

    authService = TestBed.inject(AuthService);
    router = TestBed.inject(Router);
  });

  it('redirects manager to /staff/manager/dashboard', () => {
    spyOn(authService, 'currentRole').and.returnValue(ROLES.manager);

    const result = TestBed.runInInjectionContext(() =>
      roleDefaultRedirectGuard({} as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/staff/manager/dashboard');
  });

  it('redirects practitioner to /staff/practitioner/attendance', () => {
    spyOn(authService, 'currentRole').and.returnValue(ROLES.practitioner);

    const result = TestBed.runInInjectionContext(() =>
      roleDefaultRedirectGuard({} as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/staff/practitioner/attendance');
  });

  it('redirects parent to /parent/invoices', () => {
    spyOn(authService, 'currentRole').and.returnValue(ROLES.parent);

    const result = TestBed.runInInjectionContext(() =>
      roleDefaultRedirectGuard({} as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/parent/invoices');
  });

  it('redirects null role to /signin', () => {
    spyOn(authService, 'currentRole').and.returnValue(null);

    const result = TestBed.runInInjectionContext(() =>
      roleDefaultRedirectGuard({} as never, {} as never),
    );

    expect(result instanceof UrlTree).toBeTrue();
    expect(router.serializeUrl(result as UrlTree)).toBe('/signin');
  });
});
