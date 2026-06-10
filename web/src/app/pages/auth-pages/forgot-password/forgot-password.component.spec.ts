import { provideHttpClient } from '@angular/common/http';
import { HttpErrorResponse } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of, throwError } from 'rxjs';

import { ForgotPasswordComponent } from './forgot-password.component';
import { AuthService } from '../../../core/services/auth.service';

describe('ForgotPasswordComponent', () => {
  let component: ForgotPasswordComponent;
  let fixture: ComponentFixture<ForgotPasswordComponent>;
  let authServiceMock: jasmine.SpyObj<AuthService>;

  beforeEach(async () => {
    authServiceMock = jasmine.createSpyObj('AuthService', ['requestPasswordReset']);

    await TestBed.configureTestingModule({
      imports: [ForgotPasswordComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ForgotPasswordComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('shows email form initially', () => {
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Forgot password');
    expect(compiled.querySelector('input[name="email"]')).toBeTruthy();
    expect(component.isAccepted).toBeFalse();
  });

  it('shows email required error for empty email', () => {
    component.email = '  ';
    component.onSubmit();
    expect(component.emailError).toBe('Enter your email address.');
    expect(authServiceMock.requestPasswordReset).not.toHaveBeenCalled();
  });

  it('shows invalid email error for malformed email', () => {
    component.email = 'not-an-email';
    component.onSubmit();
    expect(component.emailError).toBe('Enter a valid email address.');
    expect(authServiceMock.requestPasswordReset).not.toHaveBeenCalled();
  });

  it('posts trimmed email and shows accepted state on success', fakeAsync(() => {
    authServiceMock.requestPasswordReset.and.returnValue(of({ status: 'accepted' as const }));

    component.email = '  user@example.com  ';
    component.onSubmit();
    tick();

    expect(authServiceMock.requestPasswordReset).toHaveBeenCalledWith('user@example.com');
    expect(component.isAccepted).toBeTrue();
    expect(component.isSubmitting).toBeFalse();

    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Check your email');
    expect(compiled.textContent).toContain('If an account exists for that email');
  }));

  it('shows rate_limited error message', fakeAsync(() => {
    authServiceMock.requestPasswordReset.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'rate_limited', message: 'Too many requests.' },
        status: 429,
      }),
    ));

    component.email = 'user@example.com';
    component.onSubmit();
    tick();

    expect(component.formError).toBe('Too many attempts. Wait a moment and try again.');
  }));

  it('shows request id for unknown errors', fakeAsync(() => {
    authServiceMock.requestPasswordReset.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'internal_error', message: 'Something went wrong.', request_id: 'req-123' },
        status: 500,
      }),
    ));

    component.email = 'user@example.com';
    component.onSubmit();
    tick();

    expect(component.formError).toContain('Something went wrong.');
    expect(component.formError).toContain('Request: req-123');
  }));

  it('shows validation_error with email field as email error', fakeAsync(() => {
    authServiceMock.requestPasswordReset.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'validation_error', message: 'Invalid email format.', details: { field: 'email' } },
        status: 400,
      }),
    ));

    component.email = 'user@example.com';
    component.onSubmit();
    tick();

    expect(component.emailError).toBe('Invalid email format.');
  }));

  it('shows validation_error without field as email error', fakeAsync(() => {
    authServiceMock.requestPasswordReset.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        error: { code: 'validation_error', message: 'Email is required.' },
        status: 400,
      }),
    ));

    component.email = 'user@example.com';
    component.onSubmit();
    tick();

    expect(component.emailError).toBe('Email is required.');
  }));
});
