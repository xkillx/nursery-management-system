import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendar,
  heroPlus,
  heroXMark,
  heroXCircle,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { AdHocBookingsApiService } from '../../data/ad-hoc-bookings-api.service';
import { AdHocBooking, AdHocBookingInput } from '../../models/ad-hoc-booking.models';
import { StaffSessionType, StaffSessionTypesApiService } from '../../data/session-types-api.service';

@Component({
  selector: 'app-manager-ad-hoc-booking',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
  ],
  templateUrl: './manager-ad-hoc-booking.component.html',
  providers: [
    provideIcons({
      heroCalendar,
      heroPlus,
      heroXMark,
      heroXCircle,
    }),
  ],
})
export class ManagerAdHocBookingComponent implements OnInit {
  private readonly api = inject(AdHocBookingsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  siteName = '';
  loading = false;
  pageError: string | null = null;
  bookings = signal<AdHocBooking[]>([]);
  sessionTypes = signal<StaffSessionType[]>([]);
  filterChildId = '';

  showCreateForm = false;
  form: AdHocBookingInput = { child_id: '', calendar_date: '', session_type_id: '' };
  formSaving = false;
  formError: string | null = null;
  formFieldErrors: { child_id?: string; calendar_date?: string; session_type_id?: string } = {};

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
        this.pageError = 'Failed to load ad-hoc bookings.';
      },
    });

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => this.sessionTypes.set(types.filter((t) => t.isActive)),
      error: () => { /* Session types load failure handled by template defaults */ },
    });
  }

  openCreate(): void {
    this.showCreateForm = true;
    this.form = { child_id: '', calendar_date: '', session_type_id: '' };
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
        this.toast.success('Ad-hoc booking created.');
        this.closeCreate();
        this.loadData();
      },
      error: (err) => {
        this.formSaving = false;
        const body = err?.error;
        if (body?.code === 'validation_error' && body?.fields) {
          const fields = body.fields as Record<string, string>;
          this.formFieldErrors = {
            child_id: fields['child_id'],
            calendar_date: fields['calendar_date'],
            session_type_id: fields['session_type_id'],
          };
          this.formError = 'Please correct the highlighted fields.';
        } else {
          this.formError = body?.message ?? 'Failed to create booking.';
        }
      },
    });
  }

  cancel(booking: AdHocBooking): void {
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
