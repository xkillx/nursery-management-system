import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output, inject, OnChanges, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroClock, heroXMark } from '@ng-icons/heroicons/outline';

import { DrawerComponent } from '../../../../../shared/components/ui/modal/drawer.component';
import { AlertComponent } from '../../../../../shared/components/ui/alert/alert.component';
import { DaySelectorComponent } from '../../../../../shared/components/form/day-selector/day-selector.component';
import { BookingsApiService } from '../../../data/bookings-api.service';
import { StaffRoomsApiService, StaffRoom } from '../../../data/staff-rooms-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../../data/session-types-api.service';
import { StaffSessionTemplatesApiService } from '../../../data/session-templates-api.service';
import { SessionTemplateListItem } from '../../../models/session-template.models';
import { StaffApiService } from '../../../data/staff-api.service';
import { ChildRecord } from '../../../models/children.models';
import { BookingType } from '../../../models/booking.models';

interface BookingTypeOption {
  value: BookingType;
  label: string;
  description: string;
}

@Component({
  selector: 'app-create-booking-drawer',
  imports: [
    CommonModule,
    FormsModule,
    DrawerComponent,
    AlertComponent,
    DaySelectorComponent,
    NgIcon,
  ],
  templateUrl: './create-booking-drawer.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroClock,
      heroXMark,
    }),
  ],
})
export class CreateBookingDrawerComponent implements OnChanges {
  @Input() isOpen = false;
  @Input() siteId: string | null = null;
  @Output() closed = new EventEmitter<void>();
  @Output() created = new EventEmitter<void>();

  private readonly bookingsApi = inject(BookingsApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly sessionTemplatesApi = inject(StaffSessionTemplatesApiService);
  private readonly staffApi = inject(StaffApiService);

  readonly typeOptions: BookingTypeOption[] = [
    { value: 'recurring', label: 'Recurring', description: 'Weekly recurring booking with a session template' },
    { value: 'ad_hoc', label: 'Ad-hoc', description: 'One-off booking for a specific date' },
    { value: 'hourly', label: 'Hourly', description: 'Hourly booking for a specific date and time' },
  ];

  selectedType: BookingType | '' = '';

  rooms: StaffRoom[] = [];
  sessionTypes: StaffSessionType[] = [];
  sessionTemplates: SessionTemplateListItem[] = [];
  children: ChildRecord[] = [];

  // Recurring form
  recurringChildId = '';
  recurringSessionTemplateId = '';
  recurringRoomId = '';
  recurringDaysOfWeek: number[] = [];
  recurringStartDate = '';
  recurringEndDate = '';
  recurringFundingType = '';
  recurringFundingHours: number | null = null;
  recurringLaReference = '';

  // Ad-hoc form
  adHocChildId = '';
  adHocDate = '';
  adHocSessionTypeId = '';

  // Hourly form
  hourlyChildId = '';
  hourlyDate = '';
  hourlyStartTime = '';
  hourlyDuration = 60;
  hourlySessionTypeId = '';

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['isOpen'] && this.isOpen) {
      this.reset();
      this.loadData();
    }
  }

  get canSubmit(): boolean {
    if (this.selectedType === 'recurring') {
      return !!this.recurringChildId && !!this.recurringSessionTemplateId && !!this.recurringRoomId &&
        this.recurringDaysOfWeek.length > 0 && !!this.recurringStartDate;
    }
    if (this.selectedType === 'ad_hoc') {
      return !!this.adHocChildId && !!this.adHocDate && !!this.adHocSessionTypeId;
    }
    if (this.selectedType === 'hourly') {
      return !!this.hourlyChildId && !!this.hourlyDate && !!this.hourlyStartTime && this.hourlyDuration > 0;
    }
    return false;
  }

  selectType(type: BookingType): void {
    this.selectedType = type;
    this.formError = null;
    this.formFieldErrors = {};
  }

  isSelectedType(type: BookingType): boolean {
    return this.selectedType === type;
  }

  submit(): void {
    if (!this.siteId || !this.selectedType || !this.canSubmit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    let request$: import('rxjs').Observable<unknown>;

    switch (this.selectedType) {
      case 'recurring':
        request$ = this.bookingsApi.createRecurringBooking(this.siteId, {
          child_id: this.recurringChildId,
          session_template_id: this.recurringSessionTemplateId,
          room_id: this.recurringRoomId,
          days_of_week: this.recurringDaysOfWeek,
          effective_start_date: this.recurringStartDate,
          effective_end_date: this.recurringEndDate || undefined,
          funding_type: this.recurringFundingType || undefined,
          funding_hours_per_week: this.recurringFundingHours ?? undefined,
          la_reference: this.recurringLaReference || undefined,
        });
        break;
      case 'ad_hoc':
        request$ = this.bookingsApi.createAdHocBooking(this.siteId, {
          child_id: this.adHocChildId,
          calendar_date: this.adHocDate,
          session_type_id: this.adHocSessionTypeId,
        });
        break;
      case 'hourly':
        request$ = this.bookingsApi.createHourlyBooking(this.siteId, {
          child_id: this.hourlyChildId,
          calendar_date: this.hourlyDate,
          start_time_minutes: this.parseTimeToMinutes(this.hourlyStartTime),
          duration_minutes: this.hourlyDuration,
          session_type_id: this.hourlySessionTypeId || undefined,
        });
        break;
      default:
        return;
    }

    request$.subscribe({
      next: () => {
        this.isSaving = false;
        this.created.emit();
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

  onClose(): void {
    this.reset();
    this.closed.emit();
  }

  private reset(): void {
    this.selectedType = '';
    this.recurringChildId = '';
    this.recurringSessionTemplateId = '';
    this.recurringRoomId = '';
    this.recurringDaysOfWeek = [];
    this.recurringStartDate = '';
    this.recurringEndDate = '';
    this.recurringFundingType = '';
    this.recurringFundingHours = null;
    this.recurringLaReference = '';
    this.adHocChildId = '';
    this.adHocDate = '';
    this.adHocSessionTypeId = '';
    this.hourlyChildId = '';
    this.hourlyDate = '';
    this.hourlyStartTime = '';
    this.hourlyDuration = 60;
    this.hourlySessionTypeId = '';
    this.isSaving = false;
    this.formError = null;
    this.formFieldErrors = {};
  }

  private loadData(): void {
    if (!this.siteId) return;

    this.roomsApi.listRooms(this.siteId, { includeArchived: false }).subscribe({
      next: (rooms) => this.rooms = rooms.filter((r) => r.isActive),
      error: () => { /* Room load failure handled by template defaults */ },
    });

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => this.sessionTypes = types.filter((t) => t.isActive),
      error: () => { /* Session types load failure handled by template defaults */ },
    });

    this.sessionTemplatesApi.listSessionTemplates(this.siteId, { includeArchived: false }).subscribe({
      next: (templates) => this.sessionTemplates = templates.filter((t) => t.isActive),
      error: () => { /* Session templates load failure handled by template defaults */ },
    });

    this.staffApi.listChildren({ status: 'active', limit: 200, offset: 0 }).subscribe({
      next: (result) => this.children = result.items,
      error: () => { /* Children load failure handled by template defaults */ },
    });
  }

  private parseTimeToMinutes(time: string): number {
    const [h, m] = time.split(':').map(Number);
    return h * 60 + m;
  }
}
