import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { AuthPageLayoutComponent } from '../../../shared/layout/auth-page-layout/auth-page-layout.component';

@Component({
  selector: 'app-sign-up',
  imports: [
    AuthPageLayoutComponent,
    RouterLink,
  ],
  templateUrl: './sign-up.component.html',
  styles: ``
})
export class SignUpComponent {

}
