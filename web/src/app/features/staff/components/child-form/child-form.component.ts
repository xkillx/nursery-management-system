import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { ChildRecord, ChildWritePayload } from '../../models/children.models';
import { minorToPounds, poundsToMinor } from '../../utils/manager-list-formatters';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';

type ChildFormValue = {
  full_name: string;
  date_of_birth: string;
  start_date: string;
  core_hourly_rate_gbp: number | null;
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

    if (this.form.core_hourly_rate_gbp !== null && this.form.core_hourly_rate_gbp !== undefined) {
      payload.core_hourly_rate_minor = poundsToMinor(this.form.core_hourly_rate_gbp);
    }

    this.saved.emit(payload);
  }

  private fromChild(child: ChildRecord): ChildFormValue {
    return {
      full_name: child.fullName,
      date_of_birth: child.dateOfBirth,
      start_date: child.startDate,
      core_hourly_rate_gbp: minorToPounds(child.coreHourlyRateMinor),
      end_date: child.endDate ?? '',
      notes: child.notes ?? '',
    };
  }

  private emptyForm(): ChildFormValue {
    return {
      full_name: '',
      date_of_birth: '',
      start_date: '',
      core_hourly_rate_gbp: null,
      end_date: '',
      notes: '',
    };
  }
}
