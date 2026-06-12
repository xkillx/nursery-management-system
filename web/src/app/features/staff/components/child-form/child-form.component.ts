import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { ChildRecord, ChildWritePayload } from '../../models/children.models';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';

type ChildFormValue = {
  full_name: string;
  date_of_birth: string;
  start_date: string;
  end_date: string;
  notes: string;
};

@Component({
  selector: 'app-child-form',
  imports: [
    CommonModule,
    FormsModule,
    FormFieldComponent,
    ButtonComponent,
    AlertComponent,
    DatePickerComponent,
  ],
  templateUrl: './child-form.component.html',
})
export class ChildFormComponent {
  @Input() mode: 'create' | 'edit' = 'create';
  @Input() selectedChild: ChildRecord | null = null;
  @Input() submitting = false;
  @Input() fieldErrors: Record<string, string> = {};
  @Input() serverError: string | null = null;

  @Output() saved = new EventEmitter<ChildWritePayload>();
  @Output() cancelled = new EventEmitter<void>();

  form: ChildFormValue = this.emptyForm();

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['selectedChild']) {
      this.form = this.selectedChild ? this.fromChild(this.selectedChild) : this.emptyForm();
    }
  }

  submit(): void {
    const payload: ChildWritePayload = {
      full_name: this.form.full_name.trim(),
      date_of_birth: this.form.date_of_birth,
      start_date: this.form.start_date,
      end_date: this.form.end_date.trim(),
      notes: this.form.notes.trim(),
    };

    this.saved.emit(payload);
  }

  private fromChild(child: ChildRecord): ChildFormValue {
    return {
      full_name: child.fullName,
      date_of_birth: child.dateOfBirth,
      start_date: child.startDate,
      end_date: child.endDate ?? '',
      notes: child.notes ?? '',
    };
  }

  private emptyForm(): ChildFormValue {
    return {
      full_name: '',
      date_of_birth: '',
      start_date: '',
      end_date: '',
      notes: '',
    };
  }
}
