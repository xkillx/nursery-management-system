import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { LabelComponent } from '../../form/label/label.component';
import { CheckboxComponent } from '../../form/input/checkbox.component';
import { ButtonComponent } from '../../ui/button/button.component';
import { InputFieldComponent } from '../../form/input/input-field.component';
import { Router, RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';

import { defaultRouteForRole } from '../../../../core/constants/roles';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-signin-form',
  imports: [
    CommonModule,
    LabelComponent,
    CheckboxComponent,
    ButtonComponent,
    InputFieldComponent,
    RouterModule,
    FormsModule
],
  templateUrl: './signin-form.component.html',
  styles: ``
})
export class SigninFormComponent {
  private readonly authService = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly router = inject(Router);

  showPassword = false;
  isChecked = false;

  email = '';
  password = '';
  isSubmitting = false;

  formError: string | null = null;
  emailError: string | null = null;
  passwordError: string | null = null;

  togglePasswordVisibility() {
    this.showPassword = !this.showPassword;
  }

  onSignIn() {
    this.formError = null;
    this.emailError = null;
    this.passwordError = null;
    this.isSubmitting = true;

    this.authService.login(this.email.trim(), this.password).subscribe({
      next: () => {
        this.isSubmitting = false;
        const role = this.authService.currentRole();
        this.router.navigateByUrl(defaultRouteForRole(role));
      },
      error: (error) => {
        this.isSubmitting = false;
        const mapped = this.errorMapper.mapAndHandle(error);

        if (mapped.fieldErrors['email']) {
          this.emailError = mapped.fieldErrors['email'];
        }
        if (mapped.fieldErrors['password']) {
          this.passwordError = mapped.fieldErrors['password'];
        }

        this.formError = mapped.requestId
          ? `${mapped.message} (Request: ${mapped.requestId})`
          : mapped.message;
      },
    });
  }
}
