import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCalendarDays, heroArrowLeft } from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { DaySelectorComponent } from '../../../../shared/components/form/day-selector/day-selector.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffRoomsApiService, StaffRoom } from '../../data/staff-rooms-api.service';
import { StaffSessionTemplatesApiService } from '../../data/session-templates-api.service';
import { SessionTemplateListItem } from '../../models/session-template.models';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-create-recurring-booking',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    DaySelectorComponent,
    NgIcon,
  ],
  templateUrl: './create-recurring-booking.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroArrowLeft,
    }),
  ],
})
export class CreateRecurringBookingComponent implements OnInit {
  private readonly bookingsApi = inject(BookingsApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly sessionTemplatesApi = inject(StaffSessionTemplatesApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  siteId: string | null = null;

  rooms: StaffRoom[] = [];
  sessionTemplates: SessionTemplateListItem[] = [];
  children: ChildRecord[] = [];

  childId = '';
  sessionTemplateId = '';
  roomId = '';
  daysOfWeek: number[] = [];
  startDate = '';
  endDate = '';
  fundingType = '';
  fundingHours: number | null = null;
  laReference = '';

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  get canSubmit(): boolean {
    return !!this.childId && !!this.sessionTemplateId && !!this.roomId &&
      this.daysOfWeek.length > 0 && !!this.startDate;
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

  submit(): void {
    if (!this.siteId || !this.canSubmit) return;
    this.isSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    this.bookingsApi.createRecurringBooking(this.siteId, {
      child_id: this.childId,
      session_template_id: this.sessionTemplateId,
      room_id: this.roomId,
      days_of_week: this.daysOfWeek,
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

  private loadData(): void {
    if (!this.siteId) return;

    this.roomsApi.listRooms(this.siteId, { includeArchived: false }).subscribe({
      next: (rooms) => this.rooms = rooms.filter((r) => r.isActive),
      error: () => { /* Room load failure handled by template defaults */ },
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
}
