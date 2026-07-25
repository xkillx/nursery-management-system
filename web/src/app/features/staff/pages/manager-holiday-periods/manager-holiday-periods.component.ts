import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendarDays,
  heroPlus,
  heroTrash,
  heroPencil,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { SelectComponent } from '../../../../shared/components/form/select/select.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { ComponentCardComponent } from '../../../../shared/components/common/component-card/component-card.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { HolidayPeriodsApiService } from '../../data/holiday-periods-api.service';
import {
  HolidayPeriod,
  HolidayPeriodType,
  HOLIDAY_PERIOD_TYPE_OPTIONS,
  HOLIDAY_PERIOD_TYPE_LABELS,
} from '../../models/holiday-period.models';

@Component({
  selector: 'app-manager-holiday-periods',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    ConfirmationDialogComponent,
    PageHeaderComponent,
    ButtonComponent,
    InputFieldComponent,
    DatePickerComponent,
    SelectComponent,
    AlertComponent,
    FormFieldComponent,
    ComponentCardComponent,
    BadgeComponent,
  ],
  templateUrl: './manager-holiday-periods.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroPlus,
      heroTrash,
      heroPencil,
      heroXMark,
    }),
  ],
})
export class ManagerHolidayPeriodsComponent implements OnInit {
  private readonly api = inject(HolidayPeriodsApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  loading = false;
  periods = signal<HolidayPeriod[]>([]);

  formName = '';
  formType: HolidayPeriodType | '' = '';
  formStartDate = '';
  formEndDate = '';
  formSaving = false;
  formError: string | null = null;

  typeOptions = HOLIDAY_PERIOD_TYPE_OPTIONS;

  isConfirmDeleteOpen = false;
  periodToDelete: HolidayPeriod | null = null;
  deleteSaving = false;

  editingPeriod: HolidayPeriod | null = null;

  totalPeriods = computed(() => this.periods().length);

  upcomingPeriods = computed(() => {
    const todayStr = new Date().toISOString().split('T')[0];
    return this.periods().filter((p) => p.end_date >= todayStr).length;
  });

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.toast.error('No site is attached to this manager session.');
      return;
    }
    this.siteId = membership.branch_id;
    this.loadPeriods();
  }

  loadPeriods(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.api.list(this.siteId).subscribe({
      next: (periods) => {
        this.periods.set(periods);
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.toast.error('Failed to load holiday periods.');
      },
    });
  }

  onSubmit(form: NgForm): void {
    if (!this.siteId || !this.formName || !this.formType || !this.formStartDate || !this.formEndDate) return;

    this.formSaving = true;
    this.formError = null;

    const body = {
      name: this.formName.trim(),
      type: this.formType,
      start_date: this.formStartDate,
      end_date: this.formEndDate,
    };

    const req = this.editingPeriod
      ? this.api.update(this.siteId, this.editingPeriod.id, body)
      : this.api.create(this.siteId, body);

    req.subscribe({
      next: () => {
        this.formSaving = false;
        this.toast.success(this.editingPeriod ? 'Holiday period updated.' : 'Holiday period added.');
        this.resetForm(form);
        this.loadPeriods();
      },
      error: (err) => {
        this.formSaving = false;
        const body = err?.error;
        this.formError = body?.message ?? 'Failed to save holiday period.';
      },
    });
  }

  startEdit(period: HolidayPeriod): void {
    this.editingPeriod = period;
    this.formName = period.name;
    this.formType = period.type;
    this.formStartDate = period.start_date;
    this.formEndDate = period.end_date;
  }

  cancelEdit(): void {
    this.editingPeriod = null;
    this.formName = '';
    this.formType = '';
    this.formStartDate = '';
    this.formEndDate = '';
    this.formError = null;
  }

  resetForm(form: NgForm): void {
    this.editingPeriod = null;
    this.formName = '';
    this.formType = '';
    this.formStartDate = '';
    this.formEndDate = '';
    this.formError = null;
    form.resetForm();
  }

  confirmDelete(period: HolidayPeriod): void {
    this.periodToDelete = period;
    this.isConfirmDeleteOpen = true;
  }

  doDelete(): void {
    if (!this.siteId || !this.periodToDelete) return;
    this.deleteSaving = true;
    this.api.delete(this.siteId, this.periodToDelete.id).subscribe({
      next: () => {
        this.deleteSaving = false;
        this.isConfirmDeleteOpen = false;
        this.periodToDelete = null;
        this.toast.success('Holiday period removed.');
        this.loadPeriods();
      },
      error: () => {
        this.deleteSaving = false;
        this.isConfirmDeleteOpen = false;
        this.periodToDelete = null;
        this.toast.error('Failed to remove holiday period.');
      },
    });
  }

  cancelDelete(): void {
    this.isConfirmDeleteOpen = false;
    this.periodToDelete = null;
  }

  getTypeLabel(type: string): string {
    return HOLIDAY_PERIOD_TYPE_LABELS[type as HolidayPeriodType] ?? type;
  }

  getTypeColor(type: string): 'warning' | 'info' | 'error' | 'success' | 'light' {
    switch (type) {
      case 'half_term': return 'warning';
      case 'christmas': return 'error';
      case 'easter': return 'success';
      case 'summer': return 'info';
      default: return 'light';
    }
  }

  getDurationDays(startDate: string, endDate: string): number {
    const start = new Date(startDate);
    const end = new Date(endDate);
    return Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) + 1;
  }
}
