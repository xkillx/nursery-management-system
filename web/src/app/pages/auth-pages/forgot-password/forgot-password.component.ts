import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';

import { ApiErrorMapper } from '../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../core/errors/api-error-presenter';
import { AuthService } from '../../../core/services/auth.service';
import { AuthPageLayoutComponent } from '../../../shared/layout/auth-page-layout/auth-page-layout.component';
import { ButtonComponent } from '../../../shared/components/ui/button/button.component';
import { InputFieldComponent } from '../../../shared/components/form/input/input-field.component';
import { LabelComponent } from '../../../shared/components/form/label/label.component';

@Component({
  selector: 'app-forgot-password',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    AuthPageLayoutComponent,
    ButtonComponent,
    InputFieldComponent,
    LabelComponent,
  ],
  templateUrl: './forgot-password.component.html',
  styles: ``,
})
export class ForgotPasswordComponent {
  private readonly authService = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);

  email = '';
  isSubmitting = false;
  isAccepted = false;

  emailError: string | null = null;
  formError: string | null = null;

  onSubmit() {
    this.clearErrors();

    const trimmed = this.email.trim();

    if (!trimmed) {
      this.emailError = 'Enter your email address.';
      return;
    }

    if (!this.isValidEmail(trimmed)) {
      this.emailError = 'Enter a valid email address.';
      return;
    }

    this.isSubmitting = true;

    this.authService.requestPasswordReset(trimmed).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.isAccepted = true;
      },
      error: (error) => {
        this.isSubmitting = false;
        this.handleError(error);
      },
    });
  }

  private handleError(error: unknown) {
    const mapped = this.errorMapper.map(error);
    const presented = presentApiError(mapped, 'auth.forgotPassword');

    if (mapped.code === 'validation_error' && mapped.fieldErrors['email']) {
      this.emailError = mapped.fieldErrors['email'];
      return;
    }

    if (mapped.code === 'validation_error') {
      this.emailError = mapped.message;
      return;
    }

    if (presented.fieldErrors['email']) {
      this.emailError = presented.fieldErrors['email'];
    }

    this.formError = formatPresentedApiError(presented);
  }

  private clearErrors() {
    this.emailError = null;
    this.formError = null;
  }

  private isValidEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }
}
