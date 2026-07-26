import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendarDays,
  heroCheck,
  heroClock,
  heroInformationCircle,
  heroPencilSquare,
  heroPlus,
  heroTrash,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { SelectComponent } from '../../../../shared/components/form/select/select.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { HolidayPeriodsApiService } from '../../data/holiday-periods-api.service';
import {
  HolidayPeriod,
  HolidayPeriodType,
  HOLIDAY_PERIOD_TYPE_OPTIONS,
  HOLIDAY_PERIOD_TYPE_LABELS,
} from '../../models/holiday-period.models';

interface HolidayPeriodForm {
  name: string;
  type: HolidayPeriodType | '';
  start_date: string;
  end_date: string;
}

@Component({
  selector: 'app-manager-holiday-periods',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    BadgeComponent,
    SelectComponent,
    DatePickerComponent,
    ConfirmationDialogComponent,
  ],
  templateUrl: './manager-holiday-periods.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroCheck,
      heroClock,
      heroInformationCircle,
      heroPencilSquare,
      heroPlus,
      heroTrash,
      heroXMark,
    }),
  ],
})
export class ManagerHolidayPeriodsComponent implements OnInit {
  private readonly api = inject(HolidayPeriodsApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  siteName = '';
  loading = false;
  periods = signal<HolidayPeriod[]>([]);

  totalPeriods = computed(() => this.periods().length);
  activePeriods = computed(() => {
    const todayStr = new Date().toISOString().split('T')[0];
    return this.periods().filter((p) => p.end_date >= todayStr).length;
  });
  pastPeriods = computed(() => {
    const todayStr = new Date().toISOString().split('T')[0];
    return this.periods().filter((p) => p.end_date < todayStr).length;
  });

  readonly typeOptions = HOLIDAY_PERIOD_TYPE_OPTIONS;

  editorMode: 'closed' | 'create' | 'edit' = 'closed';
  editingPeriodId: string | null = null;
  form: HolidayPeriodForm = { name: '', type: '', start_date: '', end_date: '' };
  formSaving = false;
  formError: string | null = null;
  formFieldErrors: { name?: string; type?: string; start_date?: string; end_date?: string } = {};

  isConfirmDeleteOpen = false;
  periodToDelete: HolidayPeriod | null = null;
  deleteSaving = false;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.toast.error('No site is attached to this manager session.');
      return;
    }
    this.siteId = membership.branch_id;
    this.siteName = membership.branch_name ?? 'Assigned site';
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

  openCreate(): void {
    this.editorMode = 'create';
    this.editingPeriodId = null;
    this.form = { name: '', type: '', start_date: '', end_date: '' };
    this.formError = null;
    this.formFieldErrors = {};
  }

  openEdit(period: HolidayPeriod): void {
    this.editorMode = 'edit';
    this.editingPeriodId = period.id;
    this.form = { name: period.name, type: period.type, start_date: period.start_date, end_date: period.end_date };
    this.formError = null;
    this.formFieldErrors = {};
  }

  closeEditor(): void {
    this.editorMode = 'closed';
    this.editingPeriodId = null;
    this.formError = null;
    this.formFieldErrors = {};
  }

  save(): void {
    if (!this.siteId || !this.form.name || !this.form.type || !this.form.start_date || !this.form.end_date) return;

    this.formSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    const body = {
      name: this.form.name.trim(),
      type: this.form.type as HolidayPeriodType,
      start_date: this.form.start_date,
      end_date: this.form.end_date,
    };

    const req$ = this.editorMode === 'edit' && this.editingPeriodId
      ? this.api.update(this.siteId, this.editingPeriodId, body)
      : this.api.create(this.siteId, body);

    req$.subscribe({
      next: () => {
        this.formSaving = false;
        this.toast.success(this.editorMode === 'edit' ? 'Holiday period updated.' : 'Holiday period created.');
        this.closeEditor();
        this.loadPeriods();
      },
      error: (err) => {
        this.formSaving = false;
        const body = err?.error;
        if (body?.code === 'validation_error' && body?.fields) {
          const fields = body.fields as Record<string, string>;
          this.formFieldErrors = {
            name: fields['name'],
            type: fields['type'],
            start_date: fields['start_date'],
            end_date: fields['end_date'],
          };
          this.formError = 'Please correct the highlighted fields.';
        } else {
          this.formError = body?.message ?? 'Failed to save holiday period.';
        }
      },
    });
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
