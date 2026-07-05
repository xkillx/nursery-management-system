import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendarDays,
  heroPlus,
  heroTrash,
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
import { ClosureDaysApiService } from '../../data/closure-days-api.service';
import { ClosureDay } from '../../models/closure-day.models';

@Component({
  selector: 'app-manager-closure-days',
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
  templateUrl: './manager-closure-days.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroPlus,
      heroTrash,
      heroXMark,
    }),
  ],
})
export class ManagerClosureDaysComponent implements OnInit {
  private readonly api = inject(ClosureDaysApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  loading = false;
  closureDays = signal<ClosureDay[]>([]);

  currentYear = new Date().getFullYear();

  formDate = '';
  selectedReasonOption = '';
  customReason = '';
  formSaving = false;
  formError: string | null = null;

  reasonOptions = [
    { value: 'Bank Holiday', label: 'Bank Holiday' },
    { value: 'Inset Day', label: 'Inset Day / Staff Training' },
    { value: 'Christmas Closure', label: 'Christmas Holiday Closure' },
    { value: 'Easter Closure', label: 'Easter Holiday Closure' },
    { value: 'Other', label: 'Other (specify reason)' },
  ];

  isConfirmDeleteOpen = false;
  dayToDelete: ClosureDay | null = null;
  deleteSaving = false;

  totalClosures = computed(() => this.closureDays().length);

  nextClosure = computed(() => {
    const todayStr = new Date().toISOString().split('T')[0];
    const upcoming = this.closureDays()
      .filter((d) => d.date >= todayStr)
      .sort((a, b) => a.date.localeCompare(b.date));
    return upcoming.length > 0 ? upcoming[0] : null;
  });

  nextClosureDaysRemaining = computed(() => {
    const next = this.nextClosure();
    if (!next) return null;
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const [y, m, d] = next.date.split('-').map(Number);
    const target = new Date(y, m - 1, d);
    const diffTime = target.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    if (diffDays === 0) return 'today';
    if (diffDays === 1) return 'tomorrow';
    return `in ${diffDays} days`;
  });

  groupedClosureDays = computed(() => {
    const sorted = [...this.closureDays()].sort((a, b) => a.date.localeCompare(b.date));
    const groups: { monthKey: string; days: ClosureDay[] }[] = [];
    const monthNames = [
      'January', 'February', 'March', 'April', 'May', 'June',
      'July', 'August', 'September', 'October', 'November', 'December'
    ];

    for (const d of sorted) {
      const [yearStr, monthStr, dayStr] = d.date.split('-');
      const year = parseInt(yearStr, 10);
      const month = parseInt(monthStr, 10) - 1; // 0-indexed
      const monthName = `${monthNames[month]} ${year}`;

      let group = groups.find((g) => g.monthKey === monthName);
      if (!group) {
        group = { monthKey: monthName, days: [] };
        groups.push(group);
      }
      group.days.push(d);
    }
    return groups;
  });

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.toast.error('No site is attached to this manager session.');
      return;
    }
    this.siteId = membership.branch_id;
    this.loadClosureDays();
  }

  loadClosureDays(): void {
    if (!this.siteId) return;
    this.loading = true;
    const year = new Date().getFullYear();
    const from = `${year}-01-01`;
    const to = `${year}-12-31`;
    this.api.list(this.siteId, from, to).subscribe({
      next: (days) => {
        this.closureDays.set(days);
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.toast.error('Failed to load closure days.');
      },
    });
  }

  addClosureDay(form: NgForm): void {
    if (!this.siteId || !this.formDate) return;

    let reason = '';
    if (this.selectedReasonOption === 'Other') {
      reason = this.customReason.trim();
    } else {
      reason = this.selectedReasonOption;
    }

    this.formSaving = true;
    this.formError = null;
    this.api.create(this.siteId, this.formDate, reason || undefined).subscribe({
      next: () => {
        this.formSaving = false;
        this.toast.success('Closure day added.');
        this.formDate = '';
        this.selectedReasonOption = '';
        this.customReason = '';
        form.resetForm();
        this.loadClosureDays();
      },
      error: (err) => {
        this.formSaving = false;
        const body = err?.error;
        this.formError = body?.message ?? 'Failed to add closure day.';
      },
    });
  }

  getDateStatus(dateStr: string): 'today' | 'upcoming' | 'past' {
    const todayStr = new Date().toISOString().split('T')[0];
    if (dateStr === todayStr) return 'today';
    if (dateStr > todayStr) return 'upcoming';
    return 'past';
  }

  getReasonColor(reason: string): 'warning' | 'info' | 'error' | 'success' | 'light' {
    const clean = reason.toLowerCase();
    if (clean.includes('bank holiday')) return 'warning';
    if (clean.includes('inset') || clean.includes('training') || clean.includes('staff')) return 'info';
    if (clean.includes('christmas')) return 'error';
    if (clean.includes('easter')) return 'success';
    return 'light';
  }

  confirmDelete(day: ClosureDay): void {
    this.dayToDelete = day;
    this.isConfirmDeleteOpen = true;
  }

  doDelete(): void {
    if (!this.siteId || !this.dayToDelete) return;
    this.deleteSaving = true;
    this.api.delete(this.siteId, this.dayToDelete.id).subscribe({
      next: () => {
        this.deleteSaving = false;
        this.isConfirmDeleteOpen = false;
        this.dayToDelete = null;
        this.toast.success('Closure day removed.');
        this.loadClosureDays();
      },
      error: () => {
        this.deleteSaving = false;
        this.isConfirmDeleteOpen = false;
        this.dayToDelete = null;
        this.toast.error('Failed to remove closure day.');
      },
    });
  }

  cancelDelete(): void {
    this.isConfirmDeleteOpen = false;
    this.dayToDelete = null;
  }
}
