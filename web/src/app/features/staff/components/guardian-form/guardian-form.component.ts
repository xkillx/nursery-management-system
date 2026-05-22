import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { GuardianRecord, GuardianWritePayload } from '../../models/guardians.models';

type GuardianFormValue = {
  full_name: string;
  email: string;
  phone: string;
  notes: string;
};

@Component({
  selector: 'app-guardian-form',
  imports: [CommonModule, FormsModule],
  templateUrl: './guardian-form.component.html',
})
export class GuardianFormComponent {
  @Input() mode: 'create' | 'edit' = 'create';
  @Input() selectedGuardian: GuardianRecord | null = null;
  @Input() submitting = false;
  @Input() fieldErrors: Record<string, string> = {};
  @Input() serverError: string | null = null;

  @Output() saved = new EventEmitter<GuardianWritePayload>();
  @Output() cancelled = new EventEmitter<void>();

  form: GuardianFormValue = this.emptyForm();

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['selectedGuardian']) {
      this.form = this.selectedGuardian ? this.fromGuardian(this.selectedGuardian) : this.emptyForm();
    }
  }

  submit(): void {
    const payload: GuardianWritePayload = {
      full_name: this.form.full_name.trim(),
      email: this.form.email.trim(),
      phone: this.form.phone.trim(),
      notes: this.form.notes.trim(),
    };

    this.saved.emit(payload);
  }

  private fromGuardian(guardian: GuardianRecord): GuardianFormValue {
    return {
      full_name: guardian.fullName,
      email: guardian.email ?? '',
      phone: guardian.phone ?? '',
      notes: guardian.notes ?? '',
    };
  }

  private emptyForm(): GuardianFormValue {
    return {
      full_name: '',
      email: '',
      phone: '',
      notes: '',
    };
  }
}
