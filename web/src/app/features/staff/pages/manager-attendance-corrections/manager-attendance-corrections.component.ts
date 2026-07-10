import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroCheck } from '@ng-icons/heroicons/outline';
import { Subject, takeUntil } from 'rxjs';

import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ChildRecord, StatusFilter } from '../../../staff/models/children.models';
import { StaffApiService } from '../../../staff/data/staff-api.service';
import {
  AttendanceCorrectionReasonCode,
  AttendanceSessionRecord,
  CorrectionHistory,
  CorrectionSessionContext,
  IssuedInvoiceWarning,
} from '../../models/attendance-child.models';

const REASON_OPTIONS: { code: AttendanceCorrectionReasonCode; label: string }[] = [
  { code: 'missed_check_in', label: 'Missed check-in' },
  { code: 'missed_check_out', label: 'Missed check-out' },
  { code: 'incorrect_time', label: 'Incorrect time' },
  { code: 'duplicate_entry', label: 'Duplicate entry' },
  { code: 'other', label: 'Other' },
];

type CorrectionErrorKind =
  | 'overlap'
  | 'enrollment_window'
  | 'reason'
  | 'note'
  | 'time_order'
  | 'future_time'
  | 'authorization'
  | 'not_found'
  | 'generic';

interface CorrectionError {
  kind: CorrectionErrorKind;
  title: string;
  message: string;
  field?: 'checkInTime' | 'checkOutTime' | 'reasonCode' | 'reasonNote';
}

@Component({
  selector: 'app-manager-attendance-corrections',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    SelectComponent,
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    LoadingStateComponent,
    EmptyStateComponent,
    DatePickerComponent,
  ],
  templateUrl: './manager-attendance-corrections.component.html',
  providers: [
    provideIcons({ heroCheck }),
  ],
})
export class ManagerAttendanceCorrectionsComponent implements OnInit, OnDestroy {
  private readonly api = inject(StaffApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly destroy$ = new Subject<void>();

  children: ChildRecord[] = [];
  sessions: AttendanceSessionRecord[] = [];
  invoiceWarning: IssuedInvoiceWarning | null = null;
  history: CorrectionHistory | null = null;

  selectedChildId: string | null = null;
  selectedLocalDate = '';
  selectedSessionId: string | null = null;

  checkInTime = '';
  checkOutTime = '';
  reasonCode: AttendanceCorrectionReasonCode | '' = '';
  reasonNote = '';

  loading = false;
  loadingSessions = false;
  loadingHistory = false;
  submitting = false;

  successMessage: string | null = null;
  correctionError: CorrectionError | null = null;

  readonly reasonOptions = REASON_OPTIONS;

  get childSelectOptions(): Option[] {
    return this.isActiveChildrenFirst.map(c => ({
      value: c.id,
      label: c.fullName + (!c.isActive ? ' (inactive)' : ''),
    }));
  }

  get reasonSelectOptions(): Option[] {
    return this.reasonOptions.map(r => ({ value: r.code, label: r.label }));
  }

  onChildSelect(value: string): void {
    this.selectedChildId = value || null;
    this.onChildChange();
  }

  get isMissedSessionMode(): boolean {
    return this.selectedChildId != null && this.selectedSessionId === null;
  }

  get canSubmit(): boolean {
    if (!this.selectedChildId || !this.selectedLocalDate || !this.checkInTime || !this.checkOutTime || !this.reasonCode) {
      return false;
    }
    if (this.reasonCode === 'other' && !this.reasonNote.trim()) {
      return false;
    }
    return true;
  }

  get selectedChild(): ChildRecord | undefined {
    return this.children.find((c) => c.id === this.selectedChildId);
  }

  get selectedSession(): AttendanceSessionRecord | undefined {
    if (!this.selectedSessionId) return undefined;
    return this.sessions.find((s) => s.id === this.selectedSessionId);
  }

  get isSessionIncomplete(): boolean {
    return !!this.selectedSession && !this.selectedSession.checkOutAt;
  }

  get showCorrectionForm(): boolean {
    return !!this.selectedChildId && !!this.selectedLocalDate && (!!this.selectedSessionId || this.sessions.length === 0);
  }

  get localFieldErrors(): Record<string, string> {
    if (!this.showCorrectionForm) return {};
    const errors: Record<string, string> = {};
    if (!this.reasonCode) {
      errors['reasonCode'] = 'Select a correction reason.';
    } else if (this.reasonCode === 'other' && !this.reasonNote.trim()) {
      errors['reasonNote'] = 'Add a note when the reason is Other.';
    }
    if (this.checkInTime && this.checkOutTime && this.checkOutTime <= this.checkInTime) {
      errors['timeOrder'] = 'Set check-out after check-in.';
    }
    return errors;
  }

  getFieldError(fieldName: string): string | null {
    if (this.correctionError?.field === fieldName) return this.correctionError.message;
    return this.localFieldErrors[fieldName] ?? null;
  }

  ngOnInit(): void {
    this.loadChildren();
    this.applyQueryParams();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onChildChange(): void {
    this.selectedSessionId = null;
    this.history = null;
    this.clearCorrectionError();
    this.loadSessions();
  }

  onDateChange(): void {
    this.selectedSessionId = null;
    this.history = null;
    this.clearCorrectionError();
    this.loadSessions();
  }

  onSessionSelect(sessionId: string | null): void {
    this.selectedSessionId = sessionId;
    this.clearMessages();
    if (sessionId) {
      this.loadHistory(sessionId);
    } else {
      this.history = null;
    }
  }

  onSubmit(): void {
    if (!this.canSubmit || this.submitting) return;
    this.clearMessages();
    this.submitting = true;

    const checkInAt = this.localDateTimeToRfc3339(this.selectedLocalDate, this.checkInTime);
    const checkOutAt = this.localDateTimeToRfc3339(this.selectedLocalDate, this.checkOutTime);

    this.api
      .correctAttendance({
        sessionId: this.selectedSessionId ?? undefined,
        childId: this.selectedSessionId ? undefined : this.selectedChildId ?? undefined,
        checkInAt,
        checkOutAt,
        reasonCode: this.reasonCode as AttendanceCorrectionReasonCode,
        reasonNote: this.reasonNote || undefined,
      })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.submitting = false;
          this.successMessage = 'Correction saved successfully.';
          this.clearFormFields();
          this.loadSessions();
          if (this.selectedSessionId) {
            this.loadHistory(this.selectedSessionId);
          }
        },
        error: (err) => {
          this.submitting = false;
          this.handleCorrectionError(err);
        },
      });
  }

  private applyQueryParams(): void {
    this.route.queryParams.pipe(takeUntil(this.destroy$)).subscribe((params) => {
      if (params['child_id']) {
        this.selectedChildId = params['child_id'];
      }
      if (params['local_date']) {
        this.selectedLocalDate = params['local_date'];
      }
      if (params['session_id']) {
        this.selectedSessionId = params['session_id'];
      }
      if (this.selectedChildId && this.selectedLocalDate) {
        this.loadSessions();
        if (this.selectedSessionId) {
          this.loadHistory(this.selectedSessionId);
        }
      }
    });
  }

  private loadChildren(): void {
    this.api
      .listChildren({ status: 'all' as StatusFilter, limit: 200, offset: 0 })
      .pipe(takeUntil(this.destroy$))
      .subscribe(({ items }) => {
        this.children = items;
      });
  }

  private loadSessions(): void {
    if (!this.selectedChildId || !this.selectedLocalDate) return;
    this.loadingSessions = true;
    this.clearMessages();

    this.api
      .listCorrectionSessions(this.selectedChildId, this.selectedLocalDate)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (ctx: CorrectionSessionContext) => {
          this.loadingSessions = false;
          this.sessions = ctx.items;
          this.invoiceWarning = ctx.invoiceWarning;
        },
        error: (err) => {
          this.loadingSessions = false;
          this.handleCorrectionError(err);
        },
      });
  }

  private loadHistory(sessionId: string): void {
    this.loadingHistory = true;
    this.api
      .getCorrectionHistory(sessionId)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (h: CorrectionHistory) => {
          this.loadingHistory = false;
          this.history = h;
        },
        error: () => {
          this.loadingHistory = false;
        },
      });
  }

  private clearFormFields(): void {
    this.checkInTime = '';
    this.checkOutTime = '';
    this.reasonCode = '';
    this.reasonNote = '';
  }

  private clearMessages(): void {
    this.successMessage = null;
    this.clearCorrectionError();
  }

  private clearCorrectionError(): void {
    this.correctionError = null;
  }

  onTimeChange(): void {
    if (this.correctionError?.field === 'checkInTime' || this.correctionError?.field === 'checkOutTime') {
      this.clearCorrectionError();
    }
  }

  onReasonChange(): void {
    if (this.correctionError?.field === 'reasonCode') {
      this.clearCorrectionError();
    }
  }

  onNoteChange(): void {
    if (this.correctionError?.field === 'reasonNote') {
      this.clearCorrectionError();
    }
  }

  private handleCorrectionError(err: unknown): void {
    const mapped = this.errorMapper.mapAndHandle(err);

    const forbiddenCodes = ['forbidden_role', 'forbidden_role_unknown', 'forbidden_scope_selection', 'forbidden_scope'];

    switch (mapped.code) {
      case 'attendance_session_overlap':
        this.correctionError = {
          kind: 'overlap',
          title: 'Session overlap',
          message: 'Change the check-in or check-out time so this correction does not overlap another session for this child.',
          field: 'checkInTime',
        };
        break;
      case 'attendance_outside_enrollment_window':
        this.correctionError = {
          kind: 'enrollment_window',
          title: 'Outside enrollment window',
          message: this.enrollmentWindowMessage(),
        };
        break;
      case 'attendance_correction_reason_required':
      case 'attendance_correction_reason_invalid':
        this.correctionError = {
          kind: 'reason',
          title: 'Missing reason',
          message: 'Select a correction reason.',
          field: 'reasonCode',
        };
        break;
      case 'reason_note_required_for_other':
        this.correctionError = {
          kind: 'note',
          title: 'Note required',
          message: 'Add a note when the reason is Other.',
          field: 'reasonNote',
        };
        break;
      case 'attendance_invalid_time_order':
        this.correctionError = {
          kind: 'time_order',
          title: 'Invalid times',
          message: 'Set check-out after check-in.',
          field: 'checkInTime',
        };
        break;
      case 'attendance_correction_future_time':
        this.correctionError = {
          kind: 'future_time',
          title: 'Future time',
          message: 'Use a check-in and check-out time that has already happened.',
          field: 'checkInTime',
        };
        break;
      case 'attendance_session_not_found':
        this.correctionError = {
          kind: 'not_found',
          title: 'Session not found',
          message: 'Select the session again. It may have changed since this page loaded.',
        };
        break;
      case 'child_not_found':
        this.correctionError = {
          kind: 'not_found',
          title: 'Child not found',
          message: 'Select the child again. This child is not available in the current branch.',
        };
        break;
      default:
        if (forbiddenCodes.includes(mapped.code)) {
          this.correctionError = {
            kind: 'authorization',
            title: 'No access',
            message: 'Sign in as a manager for this branch, or switch to a manager membership before correcting attendance.',
          };
        } else {
          const presented = presentApiError(mapped, 'attendance.correction');
          this.correctionError = {
            kind: 'generic',
            title: '',
            message: formatPresentedApiError(presented),
          };
        }
    }
  }

  private enrollmentWindowMessage(): string {
    const child = this.selectedChild;
    if (!child) return "Choose a date within the child's enrollment window.";
    const start = this.formatDateShort(child.startDate);
    if (child.endDate) {
      return `Choose a date from ${start} to ${this.formatDateShort(child.endDate)}.`;
    }
    return `Choose a date on or after ${start}.`;
  }

  private formatDateShort(iso: string): string {
    return new Date(iso).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  private localDateTimeToRfc3339(date: string, time: string): string {
    return `${date}T${time}:00Z`;
  }

  getReasonLabel(code: string | null): string {
    if (!code) return '';
    const option = REASON_OPTIONS.find((r) => r.code === code);
    return option ? option.label : code;
  }

  formatTime(iso: string | null): string {
    if (!iso) return '';
    const d = new Date(iso);
    return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
  }

  formatDateTime(iso: string | null): string {
    if (!iso) return '';
    const d = new Date(iso);
    return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) +
      ' ' + d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
  }

  getStatusVariant(status: string): string {
    switch (status) {
      case 'open': return 'warning';
      case 'complete': return 'success';
      case 'corrected': return 'info';
      default: return 'light';
    }
  }

  get isActiveChildrenFirst(): ChildRecord[] {
    return [...this.children].sort((a, b) => {
      if (a.isActive === b.isActive) return a.fullName.localeCompare(b.fullName);
      return a.isActive ? -1 : 1;
    });
  }
}
