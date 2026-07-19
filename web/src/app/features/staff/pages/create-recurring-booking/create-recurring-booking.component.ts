import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { ChildAvatarComponent } from '../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';
import { RadioCardGroupComponent, RadioCardOption } from '../../../../shared/components/form/radio-card-group/radio-card-group.component';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroArrowLeft, heroUser, heroCreditCard, heroShieldCheck, heroEllipsisHorizontal, heroCake, heroHomeModern } from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { SearchAutocompleteComponent } from '../../../../shared/components/form/search-autocomplete/search-autocomplete.component';
import { SessionGridComponent } from '../../../../shared/components/form/session-grid/session-grid.component';
import { BookingSummarySidebarComponent } from './booking-summary-sidebar/booking-summary-sidebar.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { SelectComponent, type Option } from '../../../../shared/components/form/select/select.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffRoomsApiService, StaffRoom } from '../../data/staff-rooms-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { SessionEntry } from '../../models/booking.models';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-create-recurring-booking',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    SearchAutocompleteComponent,
    SessionGridComponent,
    BookingSummarySidebarComponent,
    FormFieldComponent,
    InputFieldComponent,
    SelectComponent,
    DatePickerComponent,
    ChildAvatarComponent,
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
      heroShieldCheck,
      heroEllipsisHorizontal,
      heroCake,
      heroHomeModern,
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
  roomId = '';
  sessionEntries: SessionEntry[] = [];
  startDate = '';
  endDate = '';
  fundingType = '';
  fundingHours: number | null = null;
  laReference = '';
  hourlyRateMinor: number | null = null;

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  readonly fundingOptions: RadioCardOption[] = [
    { value: 'fifteen_hours', label: 'Universal (15h)', description: 'Available for 3-4 year olds' },
    { value: 'thirty_hours', label: 'Extended (30h)', description: 'Working families grant' },
    { value: 'none', label: 'None / Private', description: 'Fully chargeable rate' },
  ];

  get canSubmit(): boolean {
    return !!this.childId && !!this.roomId && this.sessionEntries.length > 0 && !!this.startDate;
  }

  get childDisplayName(): string {
    if (!this.selectedChild) return '';
    return `${this.selectedChild.firstName} ${this.selectedChild.lastName}`;
  }

  get selectedRoom(): StaffRoom | undefined {
    return this.rooms.find((r) => r.id === this.roomId);
  }

  get occupancyPercentage(): number {
    const room = this.selectedRoom;
    if (!room || !room.capacity) return 0;
    return Math.round(((room.assignedCount || 0) / room.capacity) * 100);
  }

  get roomOptions(): Option[] {
    return this.rooms.map((r) => ({ value: r.id, label: r.name }));
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
  }

  childLabelFn(child: ChildRecord): string {
    return `${child.firstName} ${child.lastName}`;
  }

  submit(): void {
    if (!this.siteId || !this.canSubmit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi.createRecurringBooking(this.siteId, {
      child_id: this.childId,
      room_id: this.roomId,
      session_entries: this.sessionEntries,
      effective_start_date: this.startDate,
      effective_end_date: this.endDate || undefined,
      funding_type: this.fundingType || undefined,
      funding_hours_per_week: this.fundingHours ?? undefined,
      la_reference: this.laReference || undefined,
    }).subscribe({
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

  private loadData(): void {
    if (!this.siteId) return;

    this.roomsApi.listRooms(this.siteId, { includeArchived: false, includeOccupancy: true }).subscribe({
      next: (rooms) => this.rooms = rooms.filter((r) => r.isActive),
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
