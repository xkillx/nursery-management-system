import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Subject, switchMap, takeUntil, tap } from 'rxjs';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { MappedApiError } from '../../../../core/models/api-error.models';
import { ChildRecord, StatusFilter } from '../../../staff/models/children.models';
import { StaffApiService } from '../../../staff/data/staff-api.service';
import {
  AttendanceCorrectionReasonCode,
  AttendanceSessionRecord,
  CorrectionHistory,
  CorrectionHistoryEvent,
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

@Component({
  selector: 'app-manager-attendance-corrections',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    LoadingStateComponent,
    EmptyStateComponent,
  ],
  templateUrl: './manager-attendance-corrections.component.html',
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
  formError: MappedApiError | null = null;

  readonly reasonOptions = REASON_OPTIONS;

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
    this.loadSessions();
  }

  onDateChange(): void {
    this.selectedSessionId = null;
    this.history = null;
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
          this.formError = this.errorMapper.mapAndHandle(err);
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
      .subscribe((children) => {
        this.children = children;
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
          this.formError = this.errorMapper.mapAndHandle(err);
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
    this.formError = null;
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
