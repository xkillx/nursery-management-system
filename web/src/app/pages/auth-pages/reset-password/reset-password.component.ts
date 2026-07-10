import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule } from '@angular/router';

import { ApiErrorMapper } from '../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../core/errors/api-error-presenter';
import { AuthService } from '../../../core/services/auth.service';
import { AuthPageLayoutComponent } from '../../../shared/layout/auth-page-layout/auth-page-layout.component';
import { ButtonComponent } from '../../../shared/components/ui/button/button.component';
import { InputFieldComponent } from '../../../shared/components/form/input/input-field.component';
import { LabelComponent } from '../../../shared/components/form/label/label.component';

type ScreenState = 'form' | 'unusable_link' | 'complete';

@Component({
  selector: 'app-reset-password',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    AuthPageLayoutComponent,
    ButtonComponent,
    InputFieldComponent,
    LabelComponent,
  ],
  templateUrl: './reset-password.component.html',
  styles: ``,
})
export class ResetPasswordComponent {
  private readonly authService = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);

  readonly token: string | null;

  state: ScreenState = 'form';
  newPassword = '';
  confirmPassword = '';
  isSubmitting = false;

  newPasswordError: string | null = null;
  confirmPasswordError: string | null = null;
  formError: string | null = null;

  constructor() {
    const rawToken = this.route.snapshot.queryParamMap.get('token');
    this.token = rawToken?.trim() || null;

    if (!this.token) {
      this.state = 'unusable_link';
    }
  }

  onSubmit() {
    this.clearErrors();

    if (!this.newPassword) {
      this.newPasswordError = 'Enter a new password.';
      return;
    }

    if (this.newPassword.length < 8) {
      this.newPasswordError = 'Password must be at least 8 characters.';
      return;
    }

    if (!this.confirmPassword) {
      this.confirmPasswordError = 'Confirm your new password.';
      return;
    }

    if (this.newPassword !== this.confirmPassword) {
      this.confirmPasswordError = 'Passwords do not match.';
      return;
    }

    this.isSubmitting = true;

    this.authService.resetPassword(this.token!, this.newPassword).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.newPassword = '';
        this.confirmPassword = '';
        this.authService.clearSession();
        this.state = 'complete';
      },
      error: (error) => {
        this.isSubmitting = false;
        this.handleError(error);
      },
    });
  }

  private handleError(error: unknown) {
    const mapped = this.errorMapper.map(error);
    const tokenErrorCodes = [
      'password_reset_token_invalid',
      'password_reset_token_expired',
      'password_reset_token_used',
    ];

    if (tokenErrorCodes.includes(mapped.code)) {
      this.newPassword = '';
      this.confirmPassword = '';
      this.state = 'unusable_link';
      return;
    }

    if (mapped.code === 'validation_error' && mapped.fieldErrors['new_password']) {
      this.newPasswordError = mapped.fieldErrors['new_password'];
      return;
    }

    if (mapped.code === 'validation_error') {
      this.newPasswordError = mapped.message;
      return;
    }

    this.formError = formatPresentedApiError(presentApiError(mapped, 'auth.resetPassword'));
  }

  private clearErrors() {
    this.newPasswordError = null;
    this.confirmPasswordError = null;
    this.formError = null;
  }
}
