import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroCalendarDays,
  heroClock,
  heroPencilSquare,
  heroXCircle,
  heroUser,
  heroCheckCircle,
  heroCurrencyPound,
  heroDocumentText,
  heroInformationCircle,
  heroArrowTopRightOnSquare,
  heroChevronRight,
  heroSparkles,
  heroShieldCheck,
  heroBuildingStorefront,
  heroAcademicCap,
  heroQueueList,
  heroPrinter,
  heroCheck,
  heroDocumentDuplicate,
  heroArrowPath,
} from '@ng-icons/heroicons/outline';

import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { BookingDetail, BookingType, BookingStatus } from '../../models/booking.models';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { DaySelectorComponent } from '../../../../shared/components/form/day-selector/day-selector.component';

export type BookingDetailTab = 'overview' | 'audit' | 'edit';

const DAY_NAMES = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

@Component({
  selector: 'app-manager-booking-detail',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    NgIcon,
    AlertComponent,
    LoadingStateComponent,
    ConfirmationDialogComponent,
    DaySelectorComponent,
  ],
  templateUrl: './manager-booking-detail.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroCalendarDays,
      heroClock,
      heroPencilSquare,
      heroXCircle,
      heroUser,
      heroCheckCircle,
      heroCurrencyPound,
      heroDocumentText,
      heroInformationCircle,
      heroArrowTopRightOnSquare,
      heroChevronRight,
      heroSparkles,
      heroShieldCheck,
      heroBuildingStorefront,
      heroAcademicCap,
      heroQueueList,
      heroPrinter,
      heroCheck,
      heroDocumentDuplicate,
      heroArrowPath,
    }),
  ],
})
export class ManagerBookingDetailComponent implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);
  private readonly bookingsApi = inject(BookingsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly destroy$ = new Subject<void>();

  activeTab: BookingDetailTab = 'overview';

  bookingId = '';
  siteId: string | null = null;
  booking: BookingDetail | null = null;
  sessionTypes: StaffSessionType[] = [];
  sessionLookup: Record<string, string> = {};

  childAge = '';
  childPhotoUrl: string | null = null;
  assignedRoomName = '';
  roomsLookup: Record<string, string> = {};

  isLoading = false;
  errorMessage: string | null = null;

  // Edit form state
  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};
  editDaysOfWeek: number[] = [];
  editStartDate = '';
  editEndDate = '';
  editFundingType = '';
  editFundingHours: number | null = null;
  editLaReference = '';

  // Cancel state
  isConfirmCancelOpen = false;
  isCancelling = false;

  readonly dayNames = DAY_NAMES;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    this.siteId = membership?.branch_id ?? null;

    this.route.paramMap.pipe(takeUntil(this.destroy$)).subscribe((params) => {
      this.bookingId = params.get('bookingId') ?? '';
      if (this.bookingId && this.siteId) {
        this.loadBooking();
      }
    });

    if (this.siteId) {
      this.loadSessionTypes();
      this.loadRooms();
    }
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  get isRecurring(): boolean {
    return this.booking?.bookingType === 'recurring';
  }

  get canEdit(): boolean {
    return this.isRecurring && this.booking?.status === 'active';
  }

  get scheduledDays(): { name: string; dayIndex: number; sessionName: string }[] {
    if (!this.booking?.sessionEntries?.length) return [];
    return this.booking.sessionEntries
      .sort((a, b) => a.day_of_week - b.day_of_week)
      .map((e) => ({
        name: DAY_NAMES[e.day_of_week] ?? `Day ${e.day_of_week}`,
        dayIndex: e.day_of_week,
        sessionName: this.sessionLookup[e.session_type_id] ?? 'Session',
      }));
  }

  get weeklyDaysCount(): number {
    return this.scheduledDays.length;
  }

  printPage(): void {
    window.print();
  }

  copyBookingRef(): void {
    if (!this.booking) return;
    const ref = this.shortRef(this.booking.id);
    navigator.clipboard.writeText(ref).then(() => {
      this.toast.success(`Booking reference ${ref} copied to clipboard.`);
    }).catch(() => {
      this.toast.success(`Booking reference: ${ref}`);
    });
  }

  duplicateBooking(): void {
    if (!this.booking) return;
    if (this.booking.bookingType === 'recurring') {
      this.router.navigate(['/manager/bookings/new/recurring'], { queryParams: { childId: this.booking.childId } });
    } else if (this.booking.bookingType === 'ad_hoc') {
      this.router.navigate(['/manager/bookings/new/ad_hoc'], { queryParams: { childId: this.booking.childId } });
    } else {
      this.router.navigate(['/manager/bookings/new/hourly'], { queryParams: { childId: this.booking.childId } });
    }
  }

  selectTab(tab: BookingDetailTab): void {
    this.activeTab = tab;
    this.formError = null;
    this.formFieldErrors = {};
  }

  bookingTypeLabel(type: BookingType): string {
    switch (type) {
      case 'recurring': return 'Recurring';
      case 'ad_hoc': return 'Ad-hoc';
      case 'hourly': return 'Hourly';
    }
  }

  bookingTypeClasses(type: BookingType): string {
    switch (type) {
      case 'recurring':
        return 'bg-brand-50 text-brand-700 border-brand-100 dark:bg-brand-500/15 dark:text-brand-300 dark:border-brand-500/20';
      case 'ad_hoc':
        return 'bg-warning-50 text-warning-700 border-warning-100 dark:bg-warning-500/15 dark:text-warning-300 dark:border-warning-500/20';
      case 'hourly':
        return 'bg-success-50 text-success-700 border-success-100 dark:bg-success-500/15 dark:text-success-300 dark:border-success-500/20';
    }
  }

  statusClasses(status: BookingStatus): string {
    switch (status) {
      case 'active':
        return 'bg-success-50 text-success-700 border-success-100 dark:bg-success-500/15 dark:text-success-300';
      case 'paused':
        return 'bg-warning-50 text-warning-700 border-warning-100 dark:bg-warning-500/15 dark:text-warning-300';
      case 'cancelled':
        return 'bg-error-50 text-error-700 border-error-100 dark:bg-error-500/15 dark:text-error-300';
    }
  }

  formatDate(iso: string | null): string {
    if (!iso) return 'Ongoing';
    const d = new Date(iso);
    return new Intl.DateTimeFormat('en-GB', {
      timeZone: 'Europe/London',
      dateStyle: 'medium',
    }).format(d);
  }

  formatDateTime(iso: string): string {
    if (!iso) return '';
    const d = new Date(iso);
    return new Intl.DateTimeFormat('en-GB', {
      timeZone: 'Europe/London',
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(d);
  }

  childFullName(): string {
    if (!this.booking) return '';
    return `${this.booking.childFirstName} ${this.booking.childLastName}`.trim();
  }

  childInitials(): string {
    if (!this.booking) return '';
    const f = this.booking.childFirstName?.[0] ?? '';
    const l = this.booking.childLastName?.[0] ?? '';
    return (f + l).toUpperCase();
  }

  fundingTypeLabel(type: string | null | undefined): string {
    switch (type) {
      case 'universal_15': return 'Universal 15h Entitlement';
      case 'working_parent': return 'Working Parent 30h Entitlement';
      case 'working_parent_under_3': return 'Working Parent Under 3 (15h)';
      case 'disadvantaged_2yo': return 'Disadvantaged 2yr (15h)';
      case 'none': return 'None (Fully Chargeable)';
      default: return type ?? 'None';
    }
  }

  shortRef(id: string): string {
    return 'BK-' + id.slice(0, 8).toUpperCase();
  }

  computeAge(dateOfBirthIso?: string): string {
    if (!dateOfBirthIso) return '';
    const dob = new Date(dateOfBirthIso);
    const now = new Date();
    let years = now.getFullYear() - dob.getFullYear();
    let months = now.getMonth() - dob.getMonth();
    if (months < 0) {
      years--;
      months += 12;
    }
    if (now.getDate() < dob.getDate()) {
      months--;
      if (months < 0) {
        years--;
        months += 12;
      }
    }
    if (years > 0) {
      return `${years} yr${years > 1 ? 's' : ''}${months > 0 ? `, ${months} mo${months > 1 ? 's' : ''}` : ''}`;
    }
    return `${months} mo${months > 1 ? 's' : ''}`;
  }

  isDayActive(dayIndex: number): boolean {
    if (!this.booking?.sessionEntries) return false;
    return this.booking.sessionEntries.some((e) => e.day_of_week === dayIndex);
  }

  isSlotActive(dayIndex: number, slot: 'morning' | 'afternoon' | 'evening'): boolean {
    if (!this.booking?.sessionEntries) return false;
    const entry = this.booking.sessionEntries.find((e) => e.day_of_week === dayIndex);
    if (!entry) return false;

    const sName = (this.sessionLookup[entry.session_type_id] ?? '').toLowerCase();
    if (sName.includes('morning')) return slot === 'morning';
    if (sName.includes('afternoon')) return slot === 'afternoon';
    if (sName.includes('evening')) return slot === 'evening';
    // Default full day / standard session: active morning and afternoon
    return slot === 'morning' || slot === 'afternoon';
  }

  // Edit form
  openEdit(): void {
    if (!this.booking) return;
    this.activeTab = 'edit';
    this.editStartDate = this.booking.startDate;
    this.editEndDate = this.booking.endDate ?? '';
    this.editFundingType = this.booking.fundingType ?? '';
    this.editFundingHours = this.booking.fundingHoursPerWeek ?? null;
    this.editLaReference = this.booking.laReference ?? '';
    this.editDaysOfWeek = this.booking.sessionEntries.map((e) => e.day_of_week);
    this.formError = null;
    this.formFieldErrors = {};
  }

  cancelEdit(): void {
    this.activeTab = 'overview';
    this.formError = null;
    this.formFieldErrors = {};
  }

  submitEdit(): void {
    if (!this.siteId || !this.booking || !this.isRecurring) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi.updateRecurringBooking(this.siteId, this.booking.id, {
      effective_start_date: this.editStartDate || undefined,
      effective_end_date: this.editEndDate || undefined,
      funding_type: this.editFundingType || undefined,
      funding_hours_per_week: this.editFundingHours ?? undefined,
      la_reference: this.editLaReference || undefined,
    }).subscribe({
      next: () => {
        this.isSaving = false;
        this.activeTab = 'overview';
        this.toast.success('The booking has been updated successfully.');
        this.loadBooking();
      },
      error: (err) => {
        this.isSaving = false;
        const body = err?.error;
        if (body?.code === 'validation_error' && body?.fields) {
          this.formFieldErrors = body.fields as Record<string, string>;
          this.formError = 'Please correct the highlighted fields.';
        } else {
          this.formError = body?.message ?? 'Failed to update booking.';
        }
      },
    });
  }

  // Cancel
  openCancelConfirm(): void {
    this.isConfirmCancelOpen = true;
  }

  closeCancelConfirm(): void {
    this.isConfirmCancelOpen = false;
  }

  confirmCancel(): void {
    if (!this.siteId || !this.booking) return;
    this.isCancelling = true;

    this.bookingsApi.cancelBooking(this.siteId, this.booking.bookingType, this.booking.id).subscribe({
      next: () => {
        this.isCancelling = false;
        this.isConfirmCancelOpen = false;
        this.toast.success('The booking has been cancelled.');
        this.router.navigate(['/manager/bookings']);
      },
      error: () => {
        this.isCancelling = false;
        this.formError = 'Failed to cancel booking.';
      },
    });
  }

  private loadBooking(): void {
    if (!this.siteId || !this.bookingId) return;
    this.isLoading = true;
    this.errorMessage = null;

    this.bookingsApi.getBookingDetail(this.siteId, this.bookingId).subscribe({
      next: (booking) => {
        this.booking = booking;
        this.isLoading = false;
        if (booking.childId) {
          this.loadChildDetails(booking.childId);
        }
      },
      error: () => {
        this.errorMessage = 'Failed to load booking details.';
        this.isLoading = false;
      },
    });
  }

  private loadChildDetails(childId: string): void {
    this.staffApi.getChild(childId).subscribe({
      next: (child) => {
        this.childAge = this.computeAge(child.dateOfBirth);
        this.childPhotoUrl = child.photoUrl ?? null;
        if (child.primaryRoomId && this.roomsLookup[child.primaryRoomId]) {
          this.assignedRoomName = this.roomsLookup[child.primaryRoomId];
        }
      },
      error: () => {
        // Soft fallback
      },
    });
  }

  private loadRooms(): void {
    if (!this.siteId) return;
    this.roomsApi.listRooms(this.siteId).subscribe({
      next: (rooms) => {
        this.roomsLookup = {};
        for (const r of rooms) {
          this.roomsLookup[r.id] = r.name;
        }
      },
    });
  }

  private loadSessionTypes(): void {
    if (!this.siteId) return;
    this.sessionTypesApi.listSessionTypes(this.siteId).subscribe({
      next: (types) => {
        this.sessionTypes = types;
        this.sessionLookup = {};
        for (const t of types) {
          this.sessionLookup[t.id] = t.name;
        }
      },
    });
  }
}

