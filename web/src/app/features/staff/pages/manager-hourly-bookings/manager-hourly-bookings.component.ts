import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroClock,
  heroPlus,
  heroXMark,
  heroXCircle,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { HourlyBookingsApiService } from '../../data/hourly-bookings-api.service';
import { HourlyBooking, CreateHourlyBookingRequest } from '../../models/hourly-booking.models';

function formatTime(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`;
}

@Component({
  selector: 'app-manager-hourly-bookings',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
  ],
  templateUrl: './manager-hourly-bookings.component.html',
  providers: [
    provideIcons({
      heroClock,
      heroPlus,
      heroXMark,
      heroXCircle,
    }),
  ],
})
export class ManagerHourlyBookingsComponent implements OnInit {
  private readonly api = inject(HourlyBookingsApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  siteName = '';
  loading = false;
  pageError: string | null = null;
  bookings = signal<HourlyBooking[]>([]);
  filterChildId = '';

  showCreateForm = false;
  form: CreateHourlyBookingRequest = { child_id: '', calendar_date: '', start_time_minutes: 540, duration_minutes: 60 };
  formSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  readonly formatTime = formatTime;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.pageError = 'No site is attached to this manager session.';
      return;
    }
    this.siteId = membership.branch_id;
    this.siteName = membership.branch_name ?? 'Assigned site';
    this.loadData();
  }

  loadData(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.pageError = null;
    const from = new Date();
    from.setDate(1);
    const to = new Date();
    to.setMonth(to.getMonth() + 2, 0);
    const fromStr = from.toISOString().split('T')[0];
    const toStr = to.toISOString().split('T')[0];

    this.api.listBookings(this.siteId, { from: fromStr, to: toStr }).subscribe({
      next: (bookings) => {
        this.bookings.set(bookings);
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.pageError = 'Failed to load hourly bookings.';
      },
    });
  }

  openCreate(): void {
    this.showCreateForm = true;
    this.form = { child_id: '', calendar_date: '', start_time_minutes: 540, duration_minutes: 60 };
    this.formError = null;
    this.formFieldErrors = {};
  }

  closeCreate(): void {
    this.showCreateForm = false;
    this.formError = null;
    this.formFieldErrors = {};
  }

  save(): void {
    if (!this.siteId) return;
    this.formSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.api.createBooking(this.siteId, this.form).subscribe({
      next: () => {
        this.formSaving = false;
        this.toast.success('Hourly booking created.');
        this.closeCreate();
        this.loadData();
      },
      error: (err) => {
        this.formSaving = false;
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

  cancel(booking: HourlyBooking): void {
    if (!this.siteId) return;
    this.api.cancelBooking(this.siteId, booking.id).subscribe({
      next: () => {
        this.toast.success('Booking cancelled.');
        this.loadData();
      },
      error: () => {
        this.toast.error('Failed to cancel booking.');
      },
    });
  }
}
