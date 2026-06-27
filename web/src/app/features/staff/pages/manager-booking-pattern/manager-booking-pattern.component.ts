import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendar,
  heroCheck,
  heroClock,
  heroPencilSquare,
  heroPlus,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AuthService } from '../../../../core/services/auth.service';
import {
  StaffSessionType,
  StaffSessionTypesApiService,
} from '../../data/session-types-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { BookingPattern, BookingPatternInput } from '../../models/booking-pattern.models';

const DAYS = [
  { value: 1, label: 'Monday' },
  { value: 2, label: 'Tuesday' },
  { value: 3, label: 'Wednesday' },
  { value: 4, label: 'Thursday' },
  { value: 5, label: 'Friday' },
  { value: 6, label: 'Saturday' },
  { value: 7, label: 'Sunday' },
];

@Component({
  selector: 'app-manager-booking-pattern',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
  ],
  templateUrl: './manager-booking-pattern.component.html',
  providers: [
    provideIcons({
      heroCalendar,
      heroCheck,
      heroClock,
      heroPencilSquare,
      heroPlus,
      heroXMark,
    }),
  ],
})
export class ManagerBookingPatternComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly auth = inject(AuthService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);

  readonly days = DAYS;
  childId = '';
  loading = false;
  saving = false;
  pageError: string | null = null;
  fieldError: string | null = null;

  siteId: string | null = null;
  siteName = '';
  patterns: BookingPattern[] = [];
  current: BookingPattern | null = null;
  editablePattern: BookingPattern | null = null;

  sessionTypes: StaffSessionType[] = [];

  // Edit form state
  isCreating = false;
  isEditing = false;
  effectiveFrom = '';
  entries: { dayOfWeek: number; sessionTypeId: string }[] = [];

  ngOnInit(): void {
    this.childId = this.route.snapshot.paramMap.get('childId') ?? '';
    const membership = this.auth.activeMembership();
    if (membership?.branch_id) {
      this.siteId = membership.branch_id;
      this.siteName = membership.branch_name ?? 'Assigned site';
    }
    this.effectiveFrom = this.todayString;
    this.load();
  }

  isManager(): boolean {
    return this.auth.activeMembership()?.role === 'manager';
  }

  get todayString(): string {
    const d = new Date();
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    const dd = String(d.getDate()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd}`;
  }

  get isEditable(): boolean {
    if (!this.editablePattern) return false;
    if (!this.editablePattern.is_current) return false;
    return this.editablePattern.effective_from >= this.todayString;
  }

  startCreate(): void {
    this.isCreating = true;
    this.isEditing = false;
    this.editablePattern = null;
    this.effectiveFrom = this.todayString;
    this.entries = [];
    this.fieldError = null;
  }

  startEdit(p: BookingPattern): void {
    this.isCreating = false;
    this.isEditing = true;
    this.editablePattern = p;
    this.effectiveFrom = p.effective_from;
    this.entries = p.entries.map((e) => ({
      dayOfWeek: e.day_of_week,
      sessionTypeId: e.session_type.id,
    }));
    this.fieldError = null;
  }

  cancelEdit(): void {
    this.isCreating = false;
    this.isEditing = false;
    this.editablePattern = null;
    this.entries = [];
    this.fieldError = null;
  }

  toggleEntry(day: number, stID: string): void {
    const idx = this.entries.findIndex((e) => e.dayOfWeek === day && e.sessionTypeId === stID);
    if (idx >= 0) {
      this.entries.splice(idx, 1);
    } else {
      this.entries.push({ dayOfWeek: day, sessionTypeId: stID });
    }
  }

  isEntrySelected(day: number, stID: string): boolean {
    return this.entries.some((e) => e.dayOfWeek === day && e.sessionTypeId === stID);
  }

  entriesByDay(day: number) {
    return this.entries
      .filter((e) => e.dayOfWeek === day)
      .map((e) => {
        const st = this.sessionTypes.find((s) => s.id === e.sessionTypeId);
        return { sessionType: st, dayOfWeek: e.dayOfWeek, sessionTypeId: e.sessionTypeId };
      });
  }

  save(): void {
    if (!this.childId) return;
    if (this.entries.length === 0) {
      this.fieldError = 'Add at least one booked session.';
      return;
    }
    if (!/^\d{4}-\d{2}-\d{2}$/.test(this.effectiveFrom)) {
      this.fieldError = 'Effective date is required (YYYY-MM-DD).';
      return;
    }
    this.saving = true;
    this.fieldError = null;
    this.pageError = null;

    const payload: BookingPatternInput = {
      effective_from: this.effectiveFrom,
      entries: this.entries.map((e) => ({ day_of_week: e.dayOfWeek, session_type_id: e.sessionTypeId })),
    };

    const op = this.isCreating
      ? this.staffApi.createChildBookingPattern(this.childId, payload)
      : this.staffApi.updateChildBookingPattern(
          this.childId,
          this.editablePattern!.id,
          payload,
        );

    op.subscribe({
      next: () => {
        this.saving = false;
        this.cancelEdit();
        this.load();
      },
      error: (err) => {
        this.saving = false;
        this.fieldError = err?.error?.message ?? 'Failed to save booking pattern.';
      },
    });
  }

  load(): void {
    if (!this.childId) return;
    this.loading = true;
    this.pageError = null;

    if (this.siteId) {
      this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
        next: (types) => {
          this.sessionTypes = types;
        },
        error: () => {
          this.sessionTypes = [];
        },
      });
    }

    this.staffApi.listChildBookingPatterns(this.childId).subscribe({
      next: (patterns) => {
        this.patterns = patterns;
        this.current = patterns.find((p) => p.is_current) ?? null;
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = err?.error?.message ?? 'Failed to load booking patterns.';
      },
    });
  }
}
