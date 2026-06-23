import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroEnvelope, heroLockClosed, heroEye, heroEyeSlash } from '@ng-icons/heroicons/outline';
import { LabelComponent } from '../../form/label/label.component';
import { CheckboxComponent } from '../../form/input/checkbox.component';
import { ButtonComponent } from '../../ui/button/button.component';
import { InputFieldComponent } from '../../form/input/input-field.component';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';

import { defaultRouteForRole } from '../../../../core/constants/roles';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { AuthService } from '../../../../core/services/auth.service';
import { MembershipModel } from '../../../../core/models/auth.models';
import { isMembershipSelectionRequired } from '../../../../core/models/api-error.models';

@Component({
  selector: 'app-signin-form',
  imports: [
    CommonModule,
    NgIcon,
    LabelComponent,
    CheckboxComponent,
    ButtonComponent,
    InputFieldComponent,
    RouterModule,
    FormsModule
],
  providers: [provideIcons({ heroEnvelope, heroLockClosed, heroEye, heroEyeSlash })],
  templateUrl: './signin-form.component.html',
  styles: ``
})
export class SigninFormComponent {
  private readonly authService = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly router = inject(Router);

  showPassword = false;
  isChecked = true;

  email = '';
  password = '';
  isSubmitting = false;

  formError: string | null = null;
  emailError: string | null = null;
  passwordError: string | null = null;

  membershipChoices: MembershipModel[] = [];
  selectedMembershipId: string | null = null;
  membershipChallengeMessage: string | null = null;

  get isMembershipChallenge(): boolean {
    return this.membershipChoices.length > 0;
  }

  togglePasswordVisibility() {
    this.showPassword = !this.showPassword;
  }

  onSignIn() {
    this.clearErrors();

    if (this.isMembershipChallenge) {
      this.onSelectMembership();
      return;
    }

    let hasError = false;
    if (!this.email.trim()) {
      this.emailError = 'Email is required.';
      hasError = true;
    }
    if (!this.password) {
      this.passwordError = 'Password is required.';
      hasError = true;
    }
    if (hasError) {
      return;
    }

    this.isSubmitting = true;
    this.authService.login(this.email.trim(), this.password, undefined, this.isChecked).subscribe({
      next: () => this.handleSuccess(),
      error: (error) => this.handleError(error),
    });
  }

  onSelectMembership() {
    if (!this.selectedMembershipId) return;
    this.clearErrors();
    this.isSubmitting = true;

    this.authService.login(this.email.trim(), this.password, this.selectedMembershipId, this.isChecked).subscribe({
      next: () => this.handleSuccess(),
      error: (error) => this.handleError(error),
    });
  }

  selectMembership(membershipId: string) {
    this.selectedMembershipId = membershipId;
  }

  onEmailChange(value: string) {
    this.email = value;
    this.clearChallenge();
  }

  onPasswordChange(value: string) {
    this.password = value;
    this.clearChallenge();
  }

  private handleSuccess() {
    this.isSubmitting = false;
    const role = this.authService.currentRole();
    this.router.navigateByUrl(defaultRouteForRole(role));
  }

  private handleError(error: any) {
    this.isSubmitting = false;

    const body = error?.error;
    if (isMembershipSelectionRequired(body)) {
      this.membershipChoices = body.available_memberships;
      this.membershipChallengeMessage = body.message;
      this.selectedMembershipId = null;
      return;
    }

    this.membershipChoices = [];
    this.membershipChallengeMessage = null;
    this.selectedMembershipId = null;

    const mapped = this.errorMapper.mapAndHandle(error);
    const presented = presentApiError(mapped, 'auth.signin');

    if (presented.fieldErrors['email']) {
      this.emailError = presented.fieldErrors['email'];
    }
    if (presented.fieldErrors['password']) {
      this.passwordError = presented.fieldErrors['password'];
    }

    this.formError = formatPresentedApiError(presented);
  }

  private clearErrors() {
    this.formError = null;
    this.emailError = null;
    this.passwordError = null;
  }

  private clearChallenge() {
    this.membershipChoices = [];
    this.selectedMembershipId = null;
    this.membershipChallengeMessage = null;
  }
}
