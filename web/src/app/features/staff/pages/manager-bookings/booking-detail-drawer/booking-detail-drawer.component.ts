import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output, inject, OnChanges, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroPencilSquare, heroXMark, heroXCircle } from '@ng-icons/heroicons/outline';

import { DrawerComponent } from '../../../../../shared/components/ui/modal/drawer.component';
import { AlertComponent } from '../../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../../shared/components/ui/badge/status-badge.component';
import { ConfirmationDialogComponent } from '../../../../../shared/components/ui/modal/confirmation-dialog.component';
import { DaySelectorComponent } from '../../../../../shared/components/form/day-selector/day-selector.component';
import { BookingsApiService } from '../../../data/bookings-api.service';
import { UnifiedBooking, BookingType, BookingDetail } from '../../../models/booking.models';

@Component({
  selector: 'app-booking-detail-drawer',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    DrawerComponent,
    AlertComponent,
    StatusBadgeComponent,
    ConfirmationDialogComponent,
    DaySelectorComponent,
    NgIcon,
  ],
  templateUrl: './booking-detail-drawer.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroPencilSquare,
      heroXMark,
      heroXCircle,
    }),
  ],
})
export class BookingDetailDrawerComponent implements OnChanges {
  @Input() isOpen = false;
  @Input() booking: UnifiedBooking | null = null;
  @Input() siteId: string | null = null;
  @Input() sessionLookup: Record<string, string> = {};
  @Output() closed = new EventEmitter<void>();
  @Output() cancelled = new EventEmitter<void>();
  @Output() updated = new EventEmitter<void>();

  private readonly bookingsApi = inject(BookingsApiService);

  isEditMode = false;
  isConfirmCancelOpen = false;
  isCancelling = false;
  isSaving = false;
  isLoadingDetail = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};
  bookingDetail: BookingDetail | null = null;

  // Edit form fields
  editDaysOfWeek: number[] = [];
  editStartDate = '';
  editEndDate = '';
  editFundingType = '';
  editFundingHours: number | null = null;
  editLaReference = '';
  editSessionTypeId = '';
  editStartTimeMinutes: number | null = null;
  editDurationMinutes: number | null = null;

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['isOpen'] && this.isOpen) {
      this.isEditMode = false;
      this.formError = null;
      this.formFieldErrors = {};
      this.bookingDetail = null;
    }
    if (changes['booking'] && this.booking) {
      this.populateEditForm();
    }
  }

  get isRecurring(): boolean {
    return this.booking?.bookingType === 'recurring';
  }

  get isAdhoc(): boolean {
    return this.booking?.bookingType === 'ad_hoc';
  }

  get isHourly(): boolean {
    return this.booking?.bookingType === 'hourly';
  }

  get canEdit(): boolean {
    return this.booking?.status === 'active';
  }

  get computedEndTime(): string {
    if (this.editStartTimeMinutes == null || this.editDurationMinutes == null) return '';
    const endMinutes = this.editStartTimeMinutes + this.editDurationMinutes;
    const hours = Math.floor(endMinutes / 60);
    const mins = endMinutes % 60;
    return `${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}`;
  }

  formatTime(minutes: number | undefined): string {
    if (minutes == null) return '—';
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    return `${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}`;
  }

  formatDuration(minutes: number | undefined): string {
    if (minutes == null) return '—';
    const hours = Math.floor(minutes / 60);
    const mins = minutes % 60;
    if (hours === 0) return `${mins} min`;
    if (mins === 0) return `${hours} hr`;
    return `${hours} hr ${mins} min`;
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
        return 'bg-brand-50 text-brand-700 dark:bg-brand-500/15 dark:text-brand-300';
      case 'ad_hoc':
        return 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-300';
      case 'hourly':
        return 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-300';
    }
  }

  formatDate(iso: string | null): string {
    if (!iso) return '—';
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

  sessionName(id: string | undefined): string {
    if (!id) return '—';
    return this.sessionLookup[id] ?? id ?? '—';
  }

  detailSessionName(): string {
    if (this.bookingDetail?.sessionTypeName) return this.bookingDetail.sessionTypeName;
    if (this.bookingDetail?.sessionTypeId) return this.sessionName(this.bookingDetail.sessionTypeId);
    return '—';
  }

  fundingTypeLabel(type: string | null | undefined): string {
    switch (type) {
      case 'universal_15': return 'Universal (15h)';
      case 'working_parent': return 'Working Parent (30h)';
      case 'working_parent_under_3': return 'Working Parent Under 3 (15h)';
      case 'disadvantaged_2yo': return 'Disadvantaged 2yr (15h)';
      case 'none': return 'None';
      default: return type ?? '—';
    }
  }

  openEdit(): void {
    if (!this.siteId || !this.booking) return;
    this.isLoadingDetail = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi.getBookingDetail(this.siteId, this.booking.id).subscribe({
      next: (detail) => {
        this.bookingDetail = detail;
        this.isEditMode = true;
        this.isLoadingDetail = false;
        this.editStartDate = detail.startDate;
        this.editEndDate = detail.endDate ?? '';
        this.editSessionTypeId = detail.sessionTypeId ?? '';
        this.editStartTimeMinutes = detail.startTimeMinutes ?? null;
        this.editDurationMinutes = detail.durationMinutes ?? null;
        this.editDaysOfWeek = [];
        this.editFundingType = detail.fundingType ?? '';
        this.editFundingHours = detail.fundingHoursPerWeek ?? null;
        this.editLaReference = detail.laReference ?? '';
      },
      error: () => {
        this.isLoadingDetail = false;
        this.isEditMode = true;
        this.populateEditForm();
      },
    });
  }

  cancelEdit(): void {
    this.isEditMode = false;
    this.formError = null;
    this.formFieldErrors = {};
  }

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
        this.cancelled.emit();
      },
      error: () => {
        this.isCancelling = false;
        this.formError = 'Failed to cancel booking.';
      },
    });
  }

  submitEdit(): void {
    if (!this.siteId || !this.booking || !this.canEdit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    if (this.isRecurring) {
      this.bookingsApi.updateRecurringBooking(this.siteId, this.booking.id, {
        effective_start_date: this.editStartDate || undefined,
        effective_end_date: this.editEndDate || undefined,
        funding_type: this.editFundingType || undefined,
        funding_hours_per_week: this.editFundingHours ?? undefined,
        la_reference: this.editLaReference || undefined,
      }).subscribe({
        next: () => this.onSaveSuccess(),
        error: (err) => this.onSaveError(err),
      });
    } else if (this.isAdhoc) {
      this.bookingsApi.updateAdHocBooking(this.siteId, this.booking.id, {
        calendar_date: this.editStartDate || undefined,
        session_type_id: this.editSessionTypeId || undefined,
      }).subscribe({
        next: () => this.onSaveSuccess(),
        error: (err) => this.onSaveError(err),
      });
    } else if (this.isHourly) {
      this.bookingsApi.updateHourlyBooking(this.siteId, this.booking.id, {
        calendar_date: this.editStartDate || undefined,
        start_time_minutes: this.editStartTimeMinutes ?? undefined,
        duration_minutes: this.editDurationMinutes ?? undefined,
        session_type_id: this.editSessionTypeId || undefined,
      }).subscribe({
        next: () => this.onSaveSuccess(),
        error: (err) => this.onSaveError(err),
      });
    }
  }

  onClose(): void {
    this.isEditMode = false;
    this.bookingDetail = null;
    this.closed.emit();
  }

  private onSaveSuccess(): void {
    this.isSaving = false;
    this.isEditMode = false;
    this.bookingDetail = null;
    this.updated.emit();
  }

  private onSaveError(err: unknown): void {
    this.isSaving = false;
    const httpErr = err as { error?: { code?: string; fields?: Record<string, string>; message?: string } };
    const body = httpErr?.error;
    if (body?.code === 'validation_error' && body?.fields) {
      this.formFieldErrors = body.fields;
      this.formError = 'Please correct the highlighted fields.';
    } else {
      this.formError = body?.message ?? 'Failed to update booking.';
    }
  }

  private populateEditForm(): void {
    if (!this.booking) return;
    this.editStartDate = this.booking.startDate;
    this.editEndDate = this.booking.endDate ?? '';
    this.editDaysOfWeek = [];
    this.editFundingType = '';
    this.editFundingHours = null;
    this.editLaReference = '';
    this.editSessionTypeId = '';
    this.editStartTimeMinutes = null;
    this.editDurationMinutes = null;
  }
}
