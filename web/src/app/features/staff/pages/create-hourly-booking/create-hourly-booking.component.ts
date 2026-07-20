import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroClock,
  heroArrowLeft,
  heroUser,
  heroCalendarDays,
  heroSparkles,
  heroCheckCircle,
  heroInformationCircle,
  heroCake,
  heroCalendar,
  heroCheck,
  heroCurrencyPound,
  heroUsers,
  heroXMark,
  heroHomeModern,
  heroShieldCheck,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { SearchAutocompleteComponent } from '../../../../shared/components/form/search-autocomplete/search-autocomplete.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { TimePickerComponent } from '../../../../shared/components/form/time-picker/time-picker.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { UnifiedBooking } from '../../models/booking.models';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-create-hourly-booking',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    BadgeComponent,
    FormFieldComponent,
    SearchAutocompleteComponent,
    DatePickerComponent,
    TimePickerComponent,
    NgIcon,
  ],
  templateUrl: './create-hourly-booking.component.html',
  providers: [
    provideIcons({
      heroClock,
      heroArrowLeft,
      heroUser,
      heroCalendarDays,
      heroSparkles,
      heroCheckCircle,
      heroInformationCircle,
      heroCake,
      heroCalendar,
      heroCheck,
      heroCurrencyPound,
      heroUsers,
      heroXMark,
      heroHomeModern,
      heroShieldCheck,
    }),
  ],
})
export class CreateHourlyBookingComponent implements OnInit {
  private readonly bookingsApi = inject(BookingsApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  siteId: string | null = null;

  children: ChildRecord[] = [];
  recentBookings: UnifiedBooking[] = [];

  childId = '';
  selectedChild: ChildRecord | null = null;
  date = '';
  startTime = '08:00';
  endTime = '09:00';
  hourlyRateMinor = 0;
  rateLoaded = false;
  rateError: string | null = null;
  readonly fundingApplied = 0;

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  readonly childLabelFn = (child: ChildRecord): string => {
    if (!child) return '';
    const name = `${child.firstName || ''} ${child.lastName || ''}`.trim();
    return name || child.fullName;
  };

  get canSubmit(): boolean {
    return !!this.childId && !!this.date && !!this.startTime && !!this.endTime && this.computedDuration > 0 && this.rateLoaded && !this.rateError && this.hourlyRateMinor > 0;
  }

  get computedDuration(): number {
    return this.parseTimeToMinutes(this.endTime) - this.parseTimeToMinutes(this.startTime);
  }

  get endTimeError(): string | null {
    if (!this.endTime || !this.startTime) return null;
    if (this.computedDuration <= 0) return 'End time must be after start time.';
    return null;
  }

  get standardRateAmount(): number {
    return (this.computedDuration / 60) * (this.hourlyRateMinor / 100);
  }

  get totalChargeAmount(): number {
    return this.standardRateAmount - this.fundingApplied;
  }

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.formError = 'No site is attached to this session.';
      return;
    }
    this.siteId = membership.branch_id;
    this.loadData();
    this.loadSiteRate();
  }

  onStartTimeChange(time: string): void {
    this.startTime = time;
    if (this.formFieldErrors['start_time_minutes']) {
      delete this.formFieldErrors['start_time_minutes'];
    }
  }

  onEndTimeChange(time: string): void {
    this.endTime = time;
    if (this.formFieldErrors['end_time']) {
      delete this.formFieldErrors['end_time'];
    }
  }

  onChildSelected(child: ChildRecord | null): void {
    this.selectedChild = child;
    this.childId = child?.id ?? '';
    if (this.formFieldErrors['child_id']) {
      delete this.formFieldErrors['child_id'];
    }
    if (child && this.siteId) {
      this.loadChildRecentBookings(child.id);
    } else {
      this.recentBookings = [];
    }
  }

  setQuickDate(preset: 'today' | 'tomorrow' | 'next_monday'): void {
    const d = new Date();
    if (preset === 'tomorrow') {
      d.setDate(d.getDate() + 1);
    } else if (preset === 'next_monday') {
      const day = d.getDay();
      const diff = day === 0 ? 1 : 8 - day;
      d.setDate(d.getDate() + diff);
    }
    const year = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const dateStr = String(d.getDate()).padStart(2, '0');
    this.date = `${year}-${month}-${dateStr}`;

    if (this.formFieldErrors['calendar_date']) {
      delete this.formFieldErrors['calendar_date'];
    }
  }

  calculateAge(dobStr?: string): string {
    if (!dobStr) return '';
    const dob = new Date(dobStr);
    if (isNaN(dob.getTime())) return '';
    const now = new Date();
    let years = now.getFullYear() - dob.getFullYear();
    let months = now.getMonth() - dob.getMonth();
    if (months < 0 || (months === 0 && now.getDate() < dob.getDate())) {
      years--;
      months += 12;
    }
    if (years > 0) {
      return `${years} yr${years > 1 ? 's' : ''}${months > 0 ? ` ${months} mo${months > 1 ? 's' : ''}` : ''}`;
    }
    return `${months} mo${months > 1 ? 's' : ''}`;
  }

  formatDateDisplay(dateStr: string): string {
    if (!dateStr) return '';
    const parts = dateStr.split('-');
    if (parts.length !== 3) return dateStr;
    const d = new Date(parseInt(parts[0]), parseInt(parts[1]) - 1, parseInt(parts[2]));
    if (isNaN(d.getTime())) return dateStr;
    return d.toLocaleDateString('en-GB', {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  }

  submit(): void {
    if (!this.siteId || !this.canSubmit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi
      .createHourlyBooking(this.siteId, {
        child_id: this.childId,
        calendar_date: this.date,
        start_time_minutes: this.parseTimeToMinutes(this.startTime),
        duration_minutes: this.computedDuration,
      })
      .subscribe({
        next: () => {
          this.isSaving = false;
          this.router.navigate(['/manager/bookings']);
        },
        error: (err) => {
          this.isSaving = false;
          const body = err?.error;
          if (body?.code === 'validation_error' && body?.fields) {
            this.formFieldErrors = body.fields as Record<string, string>;
            this.formError = 'Please correct the highlighted fields.';
          } else {
            this.formError = body?.message ?? 'Failed to create booking.';
          }
        },
      });
  }

  cancel(): void {
    this.router.navigate(['/manager/bookings']);
  }

  private parseTimeToMinutes(time: string): number {
    if (!time) return 0;
    const [h, m] = time.split(':').map(Number);
    return (h || 0) * 60 + (m || 0);
  }

  private loadData(): void {
    if (!this.siteId) return;

    this.staffApi.listChildren({ status: 'active', limit: 200, offset: 0 }).subscribe({
      next: (result) => (this.children = result.items),
      error: () => {
        /* Handled gracefully */
      },
    });
  }

  private loadSiteRate(): void {
    this.staffApi.getSiteRate().subscribe({
      next: (res) => {
        this.hourlyRateMinor = res.core_hourly_rate_minor;
        this.rateLoaded = true;
        if (!res.has_rate || res.core_hourly_rate_minor <= 0) {
          this.rateError = 'No hourly rate configured. Please set up the rate in billing settings.';
        }
      },
      error: () => {
        this.rateLoaded = true;
        this.rateError = 'Failed to load hourly rate. Please try again.';
      },
    });
  }

  private loadChildRecentBookings(childId: string): void {
    if (!this.siteId) return;
    this.bookingsApi.listBookings(this.siteId, { childId }, 1, 5).subscribe({
      next: (res) => (this.recentBookings = res.items),
      error: () => (this.recentBookings = []),
    });
  }
}

