import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { RadioCardGroupComponent, RadioCardOption } from '../../../../shared/components/form/radio-card-group/radio-card-group.component';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroArrowLeft, heroUser, heroCreditCard, heroCake, heroHomeModern, heroCheckCircle, heroExclamationTriangle, heroBanknotes, heroPencilSquare, heroArrowUturnLeft, heroClock } from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { SearchAutocompleteComponent } from '../../../../shared/components/form/search-autocomplete/search-autocomplete.component';
import { SessionGridComponent } from '../../../../shared/components/form/session-grid/session-grid.component';
import { BookingSummarySidebarComponent } from './booking-summary-sidebar/booking-summary-sidebar.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { SwitchComponent } from '../../../../shared/components/form/input/switch.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffRoomsApiService, StaffRoom } from '../../data/staff-rooms-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { SessionEntry, CreateRecurringBookingRequest } from '../../models/booking.models';
import { FundingRecordDetail } from '../../models/funding.models';
import { AuthService } from '../../../../core/services/auth.service';

interface AutoFundingInfo {
  fundingType: string;
  fundingHours: number | null;
  laReference: string;
  termTimeOnly: boolean;
  fundingEndDate: string | null;
}

@Component({
  selector: 'app-create-recurring-booking',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    BadgeComponent,
    SearchAutocompleteComponent,
    SessionGridComponent,
    BookingSummarySidebarComponent,
    FormFieldComponent,
    InputFieldComponent,
    SwitchComponent,
    DatePickerComponent,
    RadioCardGroupComponent,
    NgIcon,
  ],
  templateUrl: './create-recurring-booking.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroArrowLeft,
      heroUser,
      heroCreditCard,
      heroCake,
      heroHomeModern,
      heroCheckCircle,
      heroExclamationTriangle,
      heroBanknotes,
      heroPencilSquare,
      heroArrowUturnLeft,
      heroClock,
    }),
  ],
})
export class CreateRecurringBookingComponent implements OnInit {
  private readonly bookingsApi = inject(BookingsApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  siteId: string | null = null;

  rooms: StaffRoom[] = [];
  sessionTypes: StaffSessionType[] = [];
  children: ChildRecord[] = [];

  selectedChild: ChildRecord | null = null;
  childId = '';
  sessionEntries: SessionEntry[] = [];
  startDate = '';
  endDate = '';
  fundingType = '';
  fundingHours: number | null = null;
  laReference = '';
  hourlyRateMinor: number | null = null;
  termTimeOnly = false;

  autoFunding: AutoFundingInfo | null = null;
  hasFundingRecord = false;
  isLoadingFunding = false;
  isOverrideActive = false;

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  readonly fundingOptions: RadioCardOption[] = [
    { value: 'universal_15', label: 'Universal (15h)', description: 'Available for 3-4 year olds' },
    { value: 'working_parent', label: 'Working Parent (30h)', description: 'Working families grant' },
    { value: 'working_parent_under_3', label: 'Working Parent Under 3 (15h)', description: 'Under 3 working parent' },
    { value: 'disadvantaged_2yo', label: 'Disadvantaged 2yr (15h)', description: 'Two-year-old funding' },
    { value: 'none', label: 'None / Private', description: 'Fully chargeable rate' },
  ];

  onFundingTypeChange(): void {
    if (this.fundingType === 'universal_15' || this.fundingType === 'disadvantaged_2yo' || this.fundingType === 'working_parent_under_3') {
      this.fundingHours = 15;
    } else if (this.fundingType === 'working_parent') {
      this.fundingHours = 30;
    } else {
      this.fundingHours = null;
    }
  }

  get canSubmit(): boolean {
    return !!this.childId && this.sessionEntries.length > 0 && !!this.startDate;
  }

  get childDisplayName(): string {
    if (!this.selectedChild) return '';
    return `${this.selectedChild.firstName || ''} ${this.selectedChild.lastName || ''}`.trim();
  }

  get childRoomName(): string {
    if (!this.selectedChild?.primaryRoomId) return '';
    const room = this.rooms.find((r) => r.id === this.selectedChild!.primaryRoomId);
    return room?.name ?? '';
  }

  get effectiveFundingType(): string {
    return this.isOverrideActive ? this.fundingType : (this.autoFunding?.fundingType ?? '');
  }

  get effectiveFundingHours(): number | null {
    return this.isOverrideActive ? this.fundingHours : (this.autoFunding?.fundingHours ?? null);
  }

  get effectiveLaReference(): string {
    return this.isOverrideActive ? this.laReference : (this.autoFunding?.laReference ?? '');
  }

  get effectiveTermTimeOnly(): boolean {
    return this.isOverrideActive ? this.termTimeOnly : (this.autoFunding?.termTimeOnly ?? false);
  }

  get fundingTypeLabel(): string {
    const type = this.effectiveFundingType;
    const opt = this.fundingOptions.find((o) => o.value === type);
    return opt?.label ?? type;
  }

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.formError = 'No site is attached to this session.';
      return;
    }
    this.siteId = membership.branch_id;
    this.loadData();
  }

  onChildSelected(child: ChildRecord | null): void {
    this.selectedChild = child;
    this.childId = child?.id ?? '';
    this.autoFunding = null;
    this.hasFundingRecord = false;
    this.isOverrideActive = false;
    this.fundingType = '';
    this.fundingHours = null;
    this.laReference = '';
    this.termTimeOnly = false;

    if (child && this.siteId) {
      this.loadChildFunding(child.id);
    }
  }

  toggleOverride(): void {
    this.isOverrideActive = !this.isOverrideActive;
    if (this.isOverrideActive && this.autoFunding) {
      this.fundingType = this.autoFunding.fundingType;
      this.fundingHours = this.autoFunding.fundingHours;
      this.laReference = this.autoFunding.laReference;
      this.termTimeOnly = this.autoFunding.termTimeOnly;
    }
  }

  clearOverride(): void {
    this.isOverrideActive = false;
    this.fundingType = '';
    this.fundingHours = null;
    this.laReference = '';
    this.termTimeOnly = false;
  }

  childLabelFn(child: ChildRecord): string {
    return `${child.firstName || ''} ${child.lastName || ''}`.trim();
  }

  calculateAge(dob: string): string {
    if (!dob) return '';
    const birthDate = new Date(dob);
    const today = new Date();
    let years = today.getFullYear() - birthDate.getFullYear();
    let months = today.getMonth() - birthDate.getMonth();
    if (months < 0) {
      years--;
      months += 12;
    }
    if (years === 0) {
      return `${months}mo`;
    }
    return `${years}y ${months}mo`;
  }

  submit(): void {
    if (!this.siteId || !this.canSubmit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    const payload: CreateRecurringBookingRequest = {
      child_id: this.childId,
      session_entries: this.sessionEntries,
      effective_start_date: this.startDate,
      effective_end_date: this.endDate || undefined,
      term_time_only: this.effectiveTermTimeOnly,
    };

    if (this.isOverrideActive) {
      payload.funding_type = this.fundingType || undefined;
      payload.funding_hours_per_week = this.fundingHours ?? undefined;
      payload.la_reference = this.laReference || undefined;
    }

    this.bookingsApi.createRecurringBooking(this.siteId, payload).subscribe({
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

  private loadChildFunding(childId: string): void {
    this.isLoadingFunding = true;
    const billingMonth = new Date().toISOString().slice(0, 7);
    this.staffApi.getFundingRecord(childId, billingMonth).subscribe({
      next: (detail: FundingRecordDetail) => {
        this.isLoadingFunding = false;
        if (detail.record.fundingEnabled) {
          this.hasFundingRecord = true;
          this.autoFunding = {
            fundingType: detail.record.fundingType,
            fundingHours: detail.record.fundedHoursPerWeek,
            laReference: detail.record.eligibilityCode ?? '',
            termTimeOnly: detail.record.fundingModel === 'term_time_only',
            fundingEndDate: detail.record.fundingEndDate,
          };
        } else {
          this.hasFundingRecord = false;
          this.autoFunding = null;
        }
      },
      error: () => {
        this.isLoadingFunding = false;
        this.hasFundingRecord = false;
        this.autoFunding = null;
      },
    });
  }

  private loadData(): void {
    if (!this.siteId) return;

    this.roomsApi.listRooms(this.siteId, { includeArchived: false, includeOccupancy: false }).subscribe({
      next: (rooms) => (this.rooms = rooms.filter((r) => r.isActive)),
      error: () => { /* Room load failure handled by template defaults */ },
    });

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => this.sessionTypes = types.filter((t) => t.isActive),
      error: () => { /* Session types load failure handled by template defaults */ },
    });

    this.staffApi.listChildren({ status: 'active', limit: 200, offset: 0 }).subscribe({
      next: (result) => this.children = result.items,
      error: () => { /* Children load failure handled by template defaults */ },
    });

    this.staffApi.getSiteRate().subscribe({
      next: (res) => this.hourlyRateMinor = res.core_hourly_rate_minor,
      error: () => { /* Rate load failure — sidebar shows setup prompt */ },
    });
  }
}
