import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCheck, heroPencilSquare, heroXMark } from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { StaffApiService } from '../../data/staff-api.service';

@Component({
  selector: 'app-manager-billing-setup',
  standalone: true,
  imports: [CommonModule, FormsModule, NgIcon, AlertComponent, LoadingStateComponent],
  templateUrl: './manager-billing-setup.component.html',
  providers: [
    provideIcons({ heroCheck, heroPencilSquare, heroXMark }),
  ],
})
export class ManagerBillingSetupComponent implements OnInit {
  private readonly api = inject(StaffApiService);

  loading = true;
  error: string | null = null;
  rateMinor: number | null = null;
  hasRate = false;

  editing = false;
  editValue = '';
  editError: string | null = null;
  saving = false;

  gracePeriod = 3;
  reminderDays = 3;
  editingGrace = false;
  graceEditValue = '3';
  reminderEditValue = '3';
  graceEditError: string | null = null;
  savingGrace = false;

  ngOnInit(): void {
    this.loadRate();
    this.loadBranchSettings();
  }

  get displayRate(): string {
    if (!this.hasRate || this.rateMinor === null) {
      return 'Not set';
    }
    return '£' + (this.rateMinor / 100).toFixed(2) + '/hr';
  }

  startEdit(): void {
    this.editing = true;
    this.editValue = this.rateMinor !== null ? (this.rateMinor / 100).toFixed(2) : '';
    this.editError = null;
  }

  cancelEdit(): void {
    this.editing = false;
    this.editValue = '';
    this.editError = null;
  }

  save(): void {
    const pounds = parseFloat(this.editValue);
    if (isNaN(pounds) || pounds <= 0) {
      this.editError = 'Enter a positive rate.';
      return;
    }
    const minor = Math.round(pounds * 100);
    this.editError = null;
    this.saving = true;

    this.api.updateSiteRate(minor).subscribe({
      next: () => {
        this.saving = false;
        this.editing = false;
        this.editValue = '';
        this.loadRate();
      },
      error: () => {
        this.saving = false;
        this.editError = 'Failed to save rate. Please try again.';
      },
    });
  }

  private loadRate(): void {
    this.loading = true;
    this.error = null;

    this.api.getSiteRate().subscribe({
      next: (res) => {
        this.rateMinor = res.core_hourly_rate_minor;
        this.hasRate = res.has_rate;
        this.loading = false;
      },
      error: () => {
        this.error = 'Failed to load billing setup. Please try again.';
        this.loading = false;
      },
    });
  }

  startGraceEdit(): void {
    this.editingGrace = true;
    this.graceEditValue = String(this.gracePeriod);
    this.reminderEditValue = String(this.reminderDays);
    this.graceEditError = null;
  }

  cancelGraceEdit(): void {
    this.editingGrace = false;
    this.graceEditValue = '3';
    this.reminderEditValue = '3';
    this.graceEditError = null;
  }

  saveGrace(): void {
    const graceDays = parseInt(this.graceEditValue, 10);
    const reminderDays = parseInt(this.reminderEditValue, 10);

    if (isNaN(graceDays) || graceDays < 0 || graceDays > 30) {
      this.graceEditError = 'Grace period must be between 0 and 30 days.';
      return;
    }

    if (isNaN(reminderDays) || reminderDays < 1 || reminderDays > 30) {
      this.graceEditError = 'Reminder days must be between 1 and 30.';
      return;
    }

    this.graceEditError = null;
    this.savingGrace = true;

    this.api.updateBranchSettings(graceDays, reminderDays).subscribe({
      next: () => {
        this.savingGrace = false;
        this.editingGrace = false;
        this.gracePeriod = graceDays;
        this.reminderDays = reminderDays;
      },
      error: () => {
        this.savingGrace = false;
        this.graceEditError = 'Failed to save settings. Please try again.';
      },
    });
  }

  private loadBranchSettings(): void {
    this.api.getBranchSettings().subscribe({
      next: (res) => {
        this.gracePeriod = res.overdue_grace_days;
        this.reminderDays = res.reminder_days_before;
      },
      error: () => {
        // Silently fail - use defaults
      },
    });
  }
}
