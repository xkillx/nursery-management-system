import { provideHttpClient } from '@angular/common/http';
import { HttpErrorResponse } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';

import { InviteAcceptComponent } from './invite-accept.component';
import { AuthService } from '../../../core/services/auth.service';

describe('InviteAcceptComponent', () => {
  let component: InviteAcceptComponent;
  let fixture: ComponentFixture<InviteAcceptComponent>;
  let authServiceMock: jasmine.SpyObj<AuthService>;

  function createComponent(token: string | null) {
    authServiceMock = jasmine.createSpyObj('AuthService', ['acceptInvite', 'clearSession']);

    TestBed.configureTestingModule({
      imports: [InviteAcceptComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authServiceMock },
        {
          provide: ActivatedRoute,
          useValue: {
            snapshot: { queryParamMap: { get: () => token } },
          },
        },
      ],
    });

    fixture = TestBed.createComponent(InviteAcceptComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  it('shows invalid_link state when token is missing', () => {
    createComponent(null);
    expect(component.state).toBe('invalid_link');
    expect(component.token).toBeNull();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Invitation link unusable');
    expect(compiled.querySelector('input')).toBeNull();
  });

  it('shows invalid_link state when token is whitespace only', () => {
    createComponent('   ');
    expect(component.state).toBe('invalid_link');
  });

  it('shows form state when token is present', () => {
    createComponent('valid-token');
    expect(component.state).toBe('form');
    expect(component.token).toBe('valid-token');

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Accept invitation');
    expect(compiled.querySelector('input[name="new_password"]')).toBeTruthy();
  });

  it('shows error for empty new password', () => {
    createComponent('valid-token');
    component.onSubmit();
    expect(component.newPasswordError).toBe('Enter a new password.');
    expect(authServiceMock.acceptInvite).not.toHaveBeenCalled();
  });

  it('shows error for short password', () => {
    createComponent('valid-token');
    component.newPassword = 'short';
    component.confirmPassword = 'short';
    component.onSubmit();
    expect(component.newPasswordError).toBe('Password must be at least 8 characters.');
  });

  it('shows error when confirm password is empty', () => {
    createComponent('valid-token');
    component.newPassword = 'password1';
    component.onSubmit();
    expect(component.confirmPasswordError).toBe('Confirm your new password.');
  });

  it('shows error when passwords do not match', () => {
    createComponent('valid-token');
    component.newPassword = 'password1';
    component.confirmPassword = 'password2';
    component.onSubmit();
    expect(component.confirmPasswordError).toBe('Passwords do not match.');
  });

  it('submits and shows complete state on success', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(of(undefined));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(authServiceMock.acceptInvite).toHaveBeenCalledWith('valid-token', 'newpassword1');
    expect(authServiceMock.clearSession).toHaveBeenCalled();
    expect(component.state).toBe('complete');
    expect(component.newPassword).toBe('');
    expect(component.confirmPassword).toBe('');

    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Invitation accepted');
    expect(compiled.textContent).toContain('Sign in with your email and new password');
  }));

  it('maps invite_token_invalid to invalid_link state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'invite_token_invalid', message: 'Invalid token.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('invalid_link');
  }));

  it('maps invite_token_expired to expired_link state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'invite_token_expired', message: 'Token expired.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('expired_link');
  }));

  it('maps invite_token_accepted to already_accepted state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'invite_token_accepted', message: 'Already accepted.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('already_accepted');
  }));

  it('maps invite_token_revoked to revoked_link state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'invite_token_revoked', message: 'Revoked.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('revoked_link');
  }));

  it('maps backend validation_error with field to new_password error', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'validation_error', message: 'Password too weak.', details: { field: 'new_password' } },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.newPasswordError).toBe('Password too weak.');
  }));

  it('maps backend validation_error without field to new_password error', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'validation_error', message: 'Password too weak.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.newPasswordError).toBe('Password too weak.');
  }));

  it('shows rate-limited form error without changing state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'rate_limited', message: 'Slow down.' },
        status: 429,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('form');
    expect(component.formError).toBe('Too many requests. Please try again later.');
  }));

  it('shows request id for unknown errors', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'internal_error', message: 'Something went wrong.', request_id: 'req-789' },
        status: 500,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.formError).toContain('Something went wrong.');
    expect(component.formError).toContain('Request: req-789');
  }));

  it('clears password fields on token error', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.acceptInvite.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'invite_token_invalid', message: 'Invalid token.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.newPassword).toBe('');
    expect(component.confirmPassword).toBe('');
  }));
});
