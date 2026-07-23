import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroCheck,
  heroEnvelope,
  heroExclamationCircle,
  heroHome,
  heroMapPin,
  heroPhone,
  heroShieldCheck,
  heroUser,
} from '@ng-icons/heroicons/outline';

import { ParentRecord } from '../../models/parents.models';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { SwitchComponent } from '../../../../shared/components/form/input/switch.component';
import { TextAreaComponent } from '../../../../shared/components/form/input/text-area.component';

export interface ParentFormData {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  address_line1: string;
  address_line2: string;
  address_city: string;
  address_postcode: string;
  has_parental_responsibility: boolean;
  can_pick_up: boolean;
  is_emergency_contact: boolean;
  notes: string;
}

@Component({
  selector: 'app-manager-parent-form',
  imports: [
    FormsModule,
    RouterLink,
    AlertComponent,
    ButtonComponent,
    FormFieldComponent,
    InputFieldComponent,
    SwitchComponent,
    TextAreaComponent,
    NgIcon,
  ],
  templateUrl: './manager-parent-form.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroCheck,
      heroEnvelope,
      heroExclamationCircle,
      heroHome,
      heroMapPin,
      heroPhone,
      heroShieldCheck,
      heroUser,
    }),
  ],
})
export class ManagerParentFormComponent {
  @Input() set parent(value: ParentRecord | null) {
    if (value) {
      this.patchForm(value);
    }
  }
  @Input() isEditMode = false;
  @Input() isSubmitting = false;
  @Input() errorMessage: string | null = null;

  @Output() saved = new EventEmitter<ParentFormData>();

  form: ParentFormData = {
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    address_line1: '',
    address_line2: '',
    address_city: '',
    address_postcode: '',
    has_parental_responsibility: false,
    can_pick_up: false,
    is_emergency_contact: false,
    notes: '',
  };

  submitted = false;

  private patchForm(parent: ParentRecord): void {
    this.form = {
      first_name: parent.first_name ?? '',
      last_name: parent.last_name ?? '',
      email: parent.email ?? '',
      phone: parent.phone ?? '',
      address_line1: parent.address_line1 ?? '',
      address_line2: parent.address_line2 ?? '',
      address_city: parent.address_city ?? '',
      address_postcode: parent.address_postcode ?? '',
      has_parental_responsibility: parent.has_parental_responsibility ?? false,
      can_pick_up: parent.can_pick_up ?? false,
      is_emergency_contact: parent.is_emergency_contact ?? false,
      notes: parent.notes ?? '',
    };
  }

  get isFirstNameInvalid(): boolean {
    return this.submitted && !this.form.first_name.trim();
  }

  get pageTitle(): string {
    return this.isEditMode ? 'Edit Parent' : 'Add New Parent';
  }

  get submitLabel(): string {
    return this.isEditMode ? 'Save Changes' : 'Create Parent';
  }

  onSubmit(): void {
    this.submitted = true;
    if (!this.form.first_name.trim()) {
      return;
    }
    this.saved.emit({ ...this.form });
  }
}
