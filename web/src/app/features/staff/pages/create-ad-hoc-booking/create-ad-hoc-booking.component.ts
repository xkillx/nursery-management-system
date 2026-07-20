import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendarDays,
  heroArrowLeft,
  heroUser,
  heroClock,
  heroSparkles,
  heroCheckCircle,
  heroInformationCircle,
  heroCake,
  heroHomeModern,
  heroCalendar,
  heroCheck,
  heroShieldCheck,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { ChildAvatarComponent } from '../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { SearchAutocompleteComponent } from '../../../../shared/components/form/search-autocomplete/search-autocomplete.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { RadioCardGroupComponent, RadioCardOption } from '../../../../shared/components/form/radio-card-group/radio-card-group.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffRoomsApiService, StaffRoom } from '../../data/staff-rooms-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-create-ad-hoc-booking',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    BadgeComponent,
    ChildAvatarComponent,
    FormFieldComponent,
    SearchAutocompleteComponent,
    DatePickerComponent,
    RadioCardGroupComponent,
    NgIcon,
  ],
  templateUrl: './create-ad-hoc-booking.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroArrowLeft,
      heroUser,
      heroClock,
      heroSparkles,
      heroCheckCircle,
      heroInformationCircle,
      heroCake,
      heroHomeModern,
      heroCalendar,
      heroCheck,
      heroShieldCheck,
    }),
  ],
})
export class CreateAdHocBookingComponent implements OnInit {
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

  childId = '';
  selectedChild: ChildRecord | null = null;
  date = '';
  sessionTypeId = '';
  hourlyRateMinor: number | null = null;

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  childLabelFn = (child: ChildRecord): string => {
    if (!child) return '';
    const name = `${child.firstName || ''} ${child.lastName || ''}`.trim();
    return name || child.fullName;
  };

  get selectedSessionType(): StaffSessionType | undefined {
    return this.sessionTypes.find((s) => s.id === this.sessionTypeId);
  }

  get sessionTypeOptions(): RadioCardOption[] {
    return this.sessionTypes.map((st) => ({
      value: st.id,
      label: st.name,
      description: `${st.startTime} - ${st.endTime}`,
    }));
  }

  get childRoomName(): string {
    if (!this.selectedChild?.primaryRoomId) return '';
    const room = this.rooms.find((r) => r.id === this.selectedChild!.primaryRoomId);
    return room?.name ?? '';
  }

  get selectedRoom(): StaffRoom | undefined {
    if (!this.selectedChild?.primaryRoomId) return undefined;
    return this.rooms.find((r) => r.id === this.selectedChild!.primaryRoomId);
  }

  get occupancyPercentage(): number {
    const room = this.selectedRoom;
    if (!room || !room.capacity) return 0;
    return Math.round(((room.assignedCount || 0) / room.capacity) * 100);
  }

  get computedDurationHours(): number {
    const st = this.selectedSessionType;
    if (!st || !st.startTime || !st.endTime) return 0;
    const [sh, sm] = st.startTime.split(':').map(Number);
    const [eh, em] = st.endTime.split(':').map(Number);
    return ((eh * 60 + em) - (sh * 60 + sm)) / 60;
  }

  get estimatedCostMinor(): number {
    if (this.hourlyRateMinor === null || this.hourlyRateMinor <= 0) return 0;
    return Math.round(this.computedDurationHours * this.hourlyRateMinor);
  }

  get canSubmit(): boolean {
    return !!this.childId && !!this.date && !!this.sessionTypeId;
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
    if (this.formFieldErrors['child_id']) {
      delete this.formFieldErrors['child_id'];
    }
  }

  selectSessionType(typeId: string): void {
    this.sessionTypeId = typeId;
    if (this.formFieldErrors['session_type_id']) {
      delete this.formFieldErrors['session_type_id'];
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

  formatGbp(minor: number): string {
    return `£${(minor / 100).toFixed(2)}`;
  }

  submit(): void {
    if (!this.siteId || !this.canSubmit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi
      .createAdHocBooking(this.siteId, {
        child_id: this.childId,
        calendar_date: this.date,
        session_type_id: this.sessionTypeId,
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

  private loadData(): void {
    if (!this.siteId) return;

    this.roomsApi.listRooms(this.siteId, { includeArchived: false, includeOccupancy: true }).subscribe({
      next: (rooms) => (this.rooms = rooms.filter((r) => r.isActive)),
      error: () => {
        /* Room load failure handled by template defaults */
      },
    });

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => (this.sessionTypes = types.filter((t) => t.isActive)),
      error: () => {
        /* Session types load failure handled by template defaults */
      },
    });

    this.staffApi.listChildren({ status: 'active', limit: 200, offset: 0 }).subscribe({
      next: (result) => (this.children = result.items),
      error: () => {
        /* Children load failure handled by template defaults */
      },
    });

    this.staffApi.getSiteRate().subscribe({
      next: (res) => (this.hourlyRateMinor = res.core_hourly_rate_minor),
      error: () => {
        /* Rate load failure */
      },
    });
  }
}
