import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AuthResponse } from '../models/auth.models';
import { AuthService } from './auth.service';

const mockAuthResponse: AuthResponse = {
  access_token: 'access-token',
  token_type: 'Bearer',
  expires_in_seconds: 900,
  user: {
    id: 'user-1',
    email: 'manager@example.com',
  },
  active_membership: {
    membership_id: 'membership-1',
    tenant_id: 'tenant-1',
    branch_id: 'branch-1',
    role: 'manager',
  },
  available_memberships: [
    {
      membership_id: 'membership-1',
      tenant_id: 'tenant-1',
      branch_id: 'branch-1',
      role: 'manager',
    },
  ],
};

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
    document.cookie = 'csrf_token=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/';
  });

  it('sets login state on successful login', () => {
    service.login('manager@example.com', 'password123').subscribe();

    const request = httpMock.expectOne('/api/v1/auth/login');
    expect(request.request.method).toBe('POST');
    request.flush(mockAuthResponse);

    expect(service.isAuthenticated()).toBeTrue();
    expect(service.accessToken()).toBe('access-token');
    expect(service.currentRole()).toBe('manager');
  });

  it('bootstraps session from refresh and resolves', async () => {
    document.cookie = 'csrf_token=test-csrf-token;path=/';

    const bootstrapPromise = service.bootstrapSession();

    const request = httpMock.expectOne('/api/v1/auth/refresh');
    expect(request.request.method).toBe('POST');
    expect(request.request.headers.get('X-CSRF-Token')).toBe('test-csrf-token');
    request.flush(mockAuthResponse);

    await bootstrapPromise;
    expect(service.isAuthenticated()).toBeTrue();
  });

  it('clears state on logout', () => {
    service.login('manager@example.com', 'password123').subscribe();
    httpMock.expectOne('/api/v1/auth/login').flush(mockAuthResponse);
    expect(service.isAuthenticated()).toBeTrue();

    document.cookie = 'csrf_token=test-csrf-token;path=/';
    service.logout().subscribe();

    const request = httpMock.expectOne('/api/v1/auth/logout');
    expect(request.request.method).toBe('POST');
    expect(request.request.headers.get('X-CSRF-Token')).toBe('test-csrf-token');
    request.flush(null);

    expect(service.isAuthenticated()).toBeFalse();
    expect(service.accessToken()).toBeNull();
  });
});
