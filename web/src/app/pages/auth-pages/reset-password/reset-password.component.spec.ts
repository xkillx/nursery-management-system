import { provideHttpClient } from '@angular/common/http';
import { HttpErrorResponse } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';

import { ResetPasswordComponent } from './reset-password.component';
import { AuthService } from '../../../core/services/auth.service';

describe('ResetPasswordComponent', () => {
  let component: ResetPasswordComponent;
  let fixture: ComponentFixture<ResetPasswordComponent>;
  let authServiceMock: jasmine.SpyObj<AuthService>;

  function createComponent(token: string | null) {
    authServiceMock = jasmine.createSpyObj('AuthService', ['resetPassword', 'clearSession']);

    TestBed.configureTestingModule({
      imports: [ResetPasswordComponent],
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

    fixture = TestBed.createComponent(ResetPasswordComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  }

  it('shows unusable link state when token is missing', () => {
    createComponent(null);
    expect(component.state).toBe('unusable_link');
    expect(component.token).toBeNull();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Reset link unusable');
    expect(compiled.querySelector('input')).toBeNull();
  });

  it('shows unusable link state when token is whitespace only', () => {
    createComponent('   ');
    expect(component.state).toBe('unusable_link');
  });

  it('shows password form when token is present', () => {
    createComponent('valid-token');
    expect(component.state).toBe('form');
    expect(component.token).toBe('valid-token');

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Reset password');
    expect(compiled.querySelector('input[name="new_password"]')).toBeTruthy();
  });

  it('shows error for empty new password', () => {
    createComponent('valid-token');
    component.onSubmit();
    expect(component.newPasswordError).toBe('Enter a new password.');
    expect(authServiceMock.resetPassword).not.toHaveBeenCalled();
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
    authServiceMock.resetPassword.and.returnValue(of(undefined));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(authServiceMock.resetPassword).toHaveBeenCalledWith('valid-token', 'newpassword1');
    expect(authServiceMock.clearSession).toHaveBeenCalled();
    expect(component.state).toBe('complete');

    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Password reset');
    expect(compiled.textContent).toContain('sign in with your new password');
  }));

  it('maps password_reset_token_invalid to unusable link state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.resetPassword.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'password_reset_token_invalid', message: 'Invalid token.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('unusable_link');
  }));

  it('maps password_reset_token_expired to unusable link state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.resetPassword.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'password_reset_token_expired', message: 'Token expired.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('unusable_link');
  }));

  it('maps password_reset_token_used to unusable link state', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.resetPassword.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'password_reset_token_used', message: 'Token already used.' },
        status: 400,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.state).toBe('unusable_link');
  }));

  it('maps backend validation_error without field to new_password', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.resetPassword.and.returnValue(throwError(() =>
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

  it('shows request id for unknown errors', fakeAsync(() => {
    createComponent('valid-token');
    authServiceMock.resetPassword.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'internal_error', message: 'Something went wrong.', request_id: 'req-456' },
        status: 500,
      }),
    ));

    component.newPassword = 'newpassword1';
    component.confirmPassword = 'newpassword1';
    component.onSubmit();
    tick();

    expect(component.formError).toContain('Something went wrong.');
    expect(component.formError).toContain('Request: req-456');
  }));
});
