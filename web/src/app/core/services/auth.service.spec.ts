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
    tenant_name: 'Little Sprouts Nursery',
    branch_id: 'branch-1',
    branch_name: 'Main Branch',
    role: 'manager',
  },
  available_memberships: [
    {
      membership_id: 'membership-1',
      tenant_id: 'tenant-1',
      tenant_name: 'Little Sprouts Nursery',
      branch_id: 'branch-1',
      branch_name: 'Main Branch',
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

  it('sends membership_id in login request body when provided', () => {
    service.login('manager@example.com', 'password123', 'membership-1').subscribe();

    const request = httpMock.expectOne('/api/v1/auth/login');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({
      email: 'manager@example.com',
      password: 'password123',
      membership_id: 'membership-1',
    });
    request.flush(mockAuthResponse);
  });

  it('requestPasswordReset posts email to password-reset-requests', () => {
    service.requestPasswordReset('user@example.com').subscribe();

    const request = httpMock.expectOne('/api/v1/auth/password-reset-requests');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ email: 'user@example.com' });
    request.flush({ status: 'accepted' });
  });

  it('resetPassword posts token and new_password to password-resets', () => {
    service.resetPassword('test-token', 'newpassword1').subscribe();

    const request = httpMock.expectOne('/api/v1/auth/password-resets');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ token: 'test-token', new_password: 'newpassword1' });
    request.flush(null, { status: 204, statusText: 'No Content' });
  });

  it('acceptInvite posts token and new_password to invites/accept', () => {
    service.acceptInvite('invite-token', 'newpassword1').subscribe();

    const request = httpMock.expectOne('/api/v1/invites/accept');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({ token: 'invite-token', new_password: 'newpassword1' });
    request.flush(null, { status: 204, statusText: 'No Content' });
  });
});
