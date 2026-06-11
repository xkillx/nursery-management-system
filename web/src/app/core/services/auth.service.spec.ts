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

const mockOwnerAuthResponse: AuthResponse = {
  access_token: 'owner-access-token',
  token_type: 'Bearer',
  expires_in_seconds: 900,
  user: {
    id: 'owner-user-1',
    email: 'owner@example.com',
  },
  active_membership: {
    membership_id: 'owner-membership-1',
    tenant_id: 'tenant-1',
    tenant_name: 'Little Sprouts Nursery',
    branch_id: null,
    branch_name: null,
    role: 'owner',
  },
  available_memberships: [
    {
      membership_id: 'owner-membership-1',
      tenant_id: 'tenant-1',
      tenant_name: 'Little Sprouts Nursery',
      branch_id: null,
      branch_name: null,
      role: 'owner',
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
      remember_me: true,
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

  it('login applies full session state', () => {
    service.login('manager@example.com', 'password123').subscribe();
    httpMock.expectOne('/api/v1/auth/login').flush(mockAuthResponse);

    expect(service.user()?.id).toBe('user-1');
    expect(service.user()?.email).toBe('manager@example.com');
    expect(service.activeMembership()?.membership_id).toBe('membership-1');
    expect(service.activeMembership()?.role).toBe('manager');
    expect(service.activeMembership()?.tenant_name).toBe('Little Sprouts Nursery');
    const memberships = (service as any).state().availableMemberships;
    expect(memberships.length).toBe(1);
    expect(service.isAuthenticated()).toBeTrue();
    expect(service.currentRole()).toBe('manager');
    expect(service.accessToken()).toBe('access-token');
  });

  it('login without membership sends email, password, and remember_me', () => {
    service.login('manager@example.com', 'password123').subscribe();

    const request = httpMock.expectOne('/api/v1/auth/login');
    expect(request.request.body).toEqual({
      email: 'manager@example.com',
      password: 'password123',
      remember_me: true,
    });
    expect(request.request.body).not.toContain(jasmine.objectContaining({ membership_id: jasmine.anything() }));
    request.flush(mockAuthResponse);
  });

  it('refresh applies replacement session and sends withCredentials', () => {
    service.refresh().subscribe();

    const request = httpMock.expectOne('/api/v1/auth/refresh');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({});
    expect(request.request.withCredentials).toBeTrue();
    request.flush(mockAuthResponse);

    expect(service.isAuthenticated()).toBeTrue();
    expect(service.accessToken()).toBe('access-token');
  });

  it('refresh sends X-CSRF-Token when csrf_token cookie exists', () => {
    document.cookie = 'csrf_token=test-csrf-token;path=/';

    service.refresh().subscribe();

    const request = httpMock.expectOne('/api/v1/auth/refresh');
    expect(request.request.headers.get('X-CSRF-Token')).toBe('test-csrf-token');
    request.flush(mockAuthResponse);
  });

  it('refresh omits X-CSRF-Token when cookie is absent', () => {
    service.refresh().subscribe();

    const request = httpMock.expectOne('/api/v1/auth/refresh');
    expect(request.request.headers.get('X-CSRF-Token')).toBeNull();
    request.flush(mockAuthResponse);
  });

  it('bootstrap failure clears stale session state', async () => {
    service.login('manager@example.com', 'password123').subscribe();
    httpMock.expectOne('/api/v1/auth/login').flush(mockAuthResponse);
    expect(service.isAuthenticated()).toBeTrue();

    const bootstrapPromise = service.bootstrapSession();

    const request = httpMock.expectOne('/api/v1/auth/refresh');
    request.flush('Unauthorized', { status: 401, statusText: 'Unauthorized' });

    await bootstrapPromise;
    expect(service.isAuthenticated()).toBeFalse();
    expect(service.accessToken()).toBeNull();
    expect(service.user()).toBeNull();
  });

  it('logout clears session even when server call fails', () => {
    service.login('manager@example.com', 'password123').subscribe();
    httpMock.expectOne('/api/v1/auth/login').flush(mockAuthResponse);
    expect(service.isAuthenticated()).toBeTrue();

    const errorSpy = jasmine.createSpy('error');
    service.logout().subscribe({ error: errorSpy });

    const request = httpMock.expectOne('/api/v1/auth/logout');
    request.flush('Server Error', { status: 500, statusText: 'Internal Server Error' });

    expect(service.isAuthenticated()).toBeFalse();
    expect(service.accessToken()).toBeNull();
    expect(errorSpy).not.toHaveBeenCalled();
  });

  it('logout observable completes without error when server call fails', () => {
    service.login('manager@example.com', 'password123').subscribe();
    httpMock.expectOne('/api/v1/auth/login').flush(mockAuthResponse);

    const errorSpy = jasmine.createSpy('error');
    const completeSpy = jasmine.createSpy('complete');
    service.logout().subscribe({ error: errorSpy, complete: completeSpy });

    const request = httpMock.expectOne('/api/v1/auth/logout');
    request.flush('Server Error', { status: 500, statusText: 'Internal Server Error' });

    expect(errorSpy).not.toHaveBeenCalled();
    expect(completeSpy).toHaveBeenCalled();
  });

  it('stores owner active membership with nullable branch fields', () => {
    service.login('owner@example.com', 'password123').subscribe();
    httpMock.expectOne('/api/v1/auth/login').flush(mockOwnerAuthResponse);

    expect(service.isAuthenticated()).toBeTrue();
    expect(service.currentRole()).toBe('owner');
    expect(service.activeMembership()?.branch_id).toBeNull();
    expect(service.activeMembership()?.branch_name).toBeNull();
    expect(service.activeMembership()?.tenant_name).toBe('Little Sprouts Nursery');
    expect(service.activeMembership()?.role).toBe('owner');
  });
});
