import { Component } from '@angular/core';
import { ManagerRegistrationIntakeComponent } from '../manager-registration-intake/manager-registration-intake.component';

@Component({
  selector: 'app-manager-child-edit',
  standalone: true,
  imports: [ManagerRegistrationIntakeComponent],
  template: `<app-manager-registration-intake />`,
})
export class ManagerChildEditComponent {}
