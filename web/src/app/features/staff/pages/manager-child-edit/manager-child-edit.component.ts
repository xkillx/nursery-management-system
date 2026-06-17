import { Component } from '@angular/core';
import { ManagerChildEditStepperComponent } from './manager-child-edit-stepper.component';

@Component({
  selector: 'app-manager-child-edit',
  standalone: true,
  imports: [ManagerChildEditStepperComponent],
  template: `<app-manager-child-edit-stepper />`,
})
export class ManagerChildEditComponent {}
