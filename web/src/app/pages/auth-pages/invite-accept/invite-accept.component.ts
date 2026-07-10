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

type ScreenState = 'form' | 'invalid_link' | 'expired_link' | 'already_accepted' | 'revoked_link' | 'complete';

@Component({
  selector: 'app-invite-accept',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    AuthPageLayoutComponent,
    ButtonComponent,
    InputFieldComponent,
    LabelComponent,
  ],
  templateUrl: './invite-accept.component.html',
  styles: ``,
})
export class InviteAcceptComponent {
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
      this.state = 'invalid_link';
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

    this.authService.acceptInvite(this.token!, this.newPassword).subscribe({
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

    const tokenStateMap: Record<string, ScreenState> = {
      invite_token_invalid: 'invalid_link',
      invite_token_expired: 'expired_link',
      invite_token_accepted: 'already_accepted',
      invite_token_revoked: 'revoked_link',
    };

    if (tokenStateMap[mapped.code]) {
      this.newPassword = '';
      this.confirmPassword = '';
      this.state = tokenStateMap[mapped.code];
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

    if (mapped.code === 'rate_limited') {
      this.formError = 'Too many requests. Please try again later.';
      return;
    }

    this.formError = formatPresentedApiError(presentApiError(mapped, 'auth.inviteAccept'));
  }

  private clearErrors() {
    this.newPasswordError = null;
    this.confirmPasswordError = null;
    this.formError = null;
  }
}
