import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroBanknotes,
  heroCheck,
  heroClock,
  heroInformationCircle,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { StaffApiService } from '../../data/staff-api.service';

@Component({
  selector: 'app-manager-billing-setup',
  standalone: true,
  imports: [CommonModule, FormsModule, NgIcon, AlertComponent, LoadingStateComponent],
  templateUrl: './manager-billing-setup.component.html',
  providers: [
    provideIcons({ heroBanknotes, heroCheck, heroClock, heroInformationCircle }),
  ],
})
export class ManagerBillingSetupComponent implements OnInit {
  private readonly api = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteName = '';
  loading = true;
  error: string | null = null;

  ratePounds = '';
  gracePeriod = '3';
  reminderDays = '3';

  formSaving = false;
  formError: string | null = null;
  formFieldErrors: { rate?: string; gracePeriod?: string; reminderDays?: string } = {};

  private originalRateMinor: number | null = null;
  private originalGrace = 3;
  private originalReminder = 3;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    this.siteName = membership?.branch_name ?? 'Your site';
    this.loadData();
  }

  get hasChanges(): boolean {
    const currentRate = this.ratePounds ? Math.round(parseFloat(this.ratePounds) * 100) : 0;
    const currentGrace = parseInt(this.gracePeriod, 10) || 0;
    const currentReminder = parseInt(this.reminderDays, 10) || 0;
    return (
      currentRate !== (this.originalRateMinor ?? 0) ||
      currentGrace !== this.originalGrace ||
      currentReminder !== this.originalReminder
    );
  }

  save(): void {
    this.formError = null;
    this.formFieldErrors = {};

    const pounds = parseFloat(this.ratePounds);
    if (isNaN(pounds) || pounds <= 0) {
      this.formFieldErrors.rate = 'Enter a positive hourly rate.';
      this.formError = 'Please correct the highlighted fields.';
      return;
    }

    const graceDays = parseInt(this.gracePeriod, 10);
    if (isNaN(graceDays) || graceDays < 0 || graceDays > 30) {
      this.formFieldErrors.gracePeriod = 'Grace period must be between 0 and 30 days.';
      this.formError = 'Please correct the highlighted fields.';
      return;
    }

    const reminder = parseInt(this.reminderDays, 10);
    if (isNaN(reminder) || reminder < 1 || reminder > 30) {
      this.formFieldErrors.reminderDays = 'Reminder days must be between 1 and 30.';
      this.formError = 'Please correct the highlighted fields.';
      return;
    }

    const rateMinor = Math.round(pounds * 100);
    this.formSaving = true;

    this.api.updateSiteRate(rateMinor).subscribe({
      next: () => {
        this.api.updateBranchSettings(graceDays, reminder).subscribe({
          next: () => {
            this.formSaving = false;
            this.originalRateMinor = rateMinor;
            this.originalGrace = graceDays;
            this.originalReminder = reminder;
            this.toast.success('Billing settings saved.');
          },
          error: () => {
            this.formSaving = false;
            this.toast.error('Rate saved but overdue settings failed. Please retry.');
          },
        });
      },
      error: () => {
        this.formSaving = false;
        this.toast.error('Failed to save billing settings. Please try again.');
      },
    });
  }

  private loadData(): void {
    this.loading = true;
    this.error = null;

    this.api.getSiteRate().subscribe({
      next: (rateRes) => {
        this.originalRateMinor = rateRes.core_hourly_rate_minor;
        this.ratePounds = rateRes.has_rate ? (rateRes.core_hourly_rate_minor / 100).toFixed(2) : '';

        this.api.getBranchSettings().subscribe({
          next: (settingsRes) => {
            this.originalGrace = settingsRes.overdue_grace_days;
            this.originalReminder = settingsRes.reminder_days_before;
            this.gracePeriod = String(settingsRes.overdue_grace_days);
            this.reminderDays = String(settingsRes.reminder_days_before);
            this.loading = false;
          },
          error: () => {
            this.loading = false;
          },
        });
      },
      error: () => {
        this.error = 'Failed to load billing setup. Please try again.';
        this.loading = false;
      },
    });
  }
}
