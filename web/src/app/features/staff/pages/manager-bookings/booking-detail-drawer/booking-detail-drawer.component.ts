import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output, inject, OnChanges, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroPencilSquare, heroXMark, heroXCircle } from '@ng-icons/heroicons/outline';

import { DrawerComponent } from '../../../../../shared/components/ui/modal/drawer.component';
import { AlertComponent } from '../../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../../shared/components/ui/badge/status-badge.component';
import { ConfirmationDialogComponent } from '../../../../../shared/components/ui/modal/confirmation-dialog.component';
import { DaySelectorComponent } from '../../../../../shared/components/form/day-selector/day-selector.component';
import { BookingsApiService } from '../../../data/bookings-api.service';
import { UnifiedBooking, BookingType } from '../../../models/booking.models';

@Component({
  selector: 'app-booking-detail-drawer',
  imports: [
    CommonModule,
    FormsModule,
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

  // Edit form fields
  editDaysOfWeek: number[] = [];
  editStartDate = '';
  editEndDate = '';
  editFundingType = '';
  editFundingHours: number | null = null;
  editLaReference = '';

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['isOpen'] && this.isOpen) {
      this.isEditMode = false;
      this.formError = null;
      this.formFieldErrors = {};
    }
    if (changes['booking'] && this.booking) {
      this.populateEditForm();
    }
  }

  get isRecurring(): boolean {
    return this.booking?.bookingType === 'recurring';
  }

  get canEdit(): boolean {
    return this.isRecurring;
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

  openEdit(): void {
    if (!this.siteId || !this.booking) return;
    this.isLoadingDetail = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi.getBooking(this.siteId, this.booking.id).subscribe({
      next: (full) => {
        this.isEditMode = true;
        this.isLoadingDetail = false;
        this.editStartDate = full.startDate;
        this.editEndDate = full.endDate ?? '';
        // The full booking detail may include additional fields not in the list response
        // For now, use what's available; fields not in UnifiedBooking default to empty
        this.editDaysOfWeek = [];
        this.editFundingType = '';
        this.editFundingHours = null;
        this.editLaReference = '';
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
    if (!this.siteId || !this.booking || !this.isRecurring) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi.updateRecurringBooking(this.siteId, this.booking.id, {
      days_of_week: this.editDaysOfWeek.length > 0 ? this.editDaysOfWeek : undefined,
      effective_start_date: this.editStartDate || undefined,
      effective_end_date: this.editEndDate || undefined,
      funding_type: this.editFundingType || undefined,
      funding_hours_per_week: this.editFundingHours ?? undefined,
      la_reference: this.editLaReference || undefined,
    }).subscribe({
      next: () => {
        this.isSaving = false;
        this.isEditMode = false;
        this.updated.emit();
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

  onClose(): void {
    this.isEditMode = false;
    this.closed.emit();
  }

  private populateEditForm(): void {
    if (!this.booking) return;
    this.editStartDate = this.booking.startDate;
    this.editEndDate = this.booking.endDate ?? '';
    // days_of_week, funding_type, funding_hours, la_reference are not in the unified list response
    // These would need a detail API call to populate fully
    this.editDaysOfWeek = [];
    this.editFundingType = '';
    this.editFundingHours = null;
    this.editLaReference = '';
  }
}
