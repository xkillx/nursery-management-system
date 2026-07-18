import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroClock, heroArrowLeft } from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { AuthService } from '../../../../core/services/auth.service';

@Component({
  selector: 'app-create-hourly-booking',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    NgIcon,
  ],
  templateUrl: './create-hourly-booking.component.html',
  providers: [
    provideIcons({
      heroClock,
      heroArrowLeft,
    }),
  ],
})
export class CreateHourlyBookingComponent implements OnInit {
  private readonly bookingsApi = inject(BookingsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  siteId: string | null = null;

  sessionTypes: StaffSessionType[] = [];
  children: ChildRecord[] = [];

  childId = '';
  date = '';
  startTime = '';
  duration = 60;
  sessionTypeId = '';

  isSaving = false;
  formError: string | null = null;
  formFieldErrors: Record<string, string> = {};

  get canSubmit(): boolean {
    return !!this.childId && !!this.date && !!this.startTime && this.duration > 0;
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

    this.bookingsApi.createHourlyBooking(this.siteId, {
      child_id: this.childId,
      calendar_date: this.date,
      start_time_minutes: this.parseTimeToMinutes(this.startTime),
      duration_minutes: this.duration,
      session_type_id: this.sessionTypeId || undefined,
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

  private parseTimeToMinutes(time: string): number {
    const [h, m] = time.split(':').map(Number);
    return h * 60 + m;
  }

  private loadData(): void {
    if (!this.siteId) return;

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => this.sessionTypes = types.filter((t) => t.isActive),
      error: () => { /* Session types load failure handled by template defaults */ },
    });

    this.staffApi.listChildren({ status: 'active', limit: 200, offset: 0 }).subscribe({
      next: (result) => this.children = result.items,
      error: () => { /* Children load failure handled by template defaults */ },
    });
  }
}
