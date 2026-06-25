import { CommonModule } from '@angular/common';
import { Component, DestroyRef, inject, OnDestroy } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { interval, Subject, switchMap, takeUntil, tap } from 'rxjs';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowPath,
  heroCheckBadge,
  heroClock,
      heroFunnel,
      heroMagnifyingGlass,
      heroUsers,
  heroXCircle,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { StaffApiService } from '../../data/staff-api.service';
import { AttendanceChildRecord, AttendanceState } from '../../models/attendance-child.models';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

type StatusFilter = 'all' | 'not_checked_in' | 'checked_in' | 'absent';
type LoadSource = 'initial' | 'manual' | 'mutation' | 'poll';

@Component({
  selector: 'app-practitioner-attendance-children',
  imports: [
    CommonModule,
    NgIcon,
    ButtonComponent,
    AlertComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  providers: [
    provideIcons({
      heroArrowPath,
      heroCheckBadge,
      heroClock,
      heroFunnel,
  heroMagnifyingGlass,
      heroUsers,
      heroXCircle,
    }),
  ],
  templateUrl: './practitioner-attendance-children.component.html',
})
export class PractitionerAttendanceChildrenComponent implements OnDestroy {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly destroyRef = inject(DestroyRef);

  private readonly pollIntervalMs = 30000;
  private pollSubscription: import('rxjs').Subscription | null = null;
  private listRequestInFlight = false;

  children: AttendanceChildRecord[] = [];
  isLoading = false;
  isBackgroundRefreshing = false;
  errorMessage: string | null = null;
  autoRefreshEnabled = true;
  lastUpdatedAt: Date | null = null;

  searchTerm = '';
  statusFilter: StatusFilter = 'all';
  rowErrors: Record<string, string> = {};
  pendingChildIds = new Set<string>();

  clockTime = '';
  currentDate = '';
  private clockInterval: ReturnType<typeof setInterval> | null = null;

  get checkedInCount(): number {
    return this.children.filter((c) => this.isCheckedIn(c)).length;
  }

  get notInCount(): number {
    return this.children.filter((c) => !this.isCheckedIn(c)).length;
  }

  get absentCount(): number {
    return this.children.filter((c) => this.isAbsent(c)).length;
  }

  get expectedSoonCount(): number {
    return this.children.filter(c => !this.isCheckedIn(c) && !this.isAbsent(c)).length;
  }

  get filteredChildren(): AttendanceChildRecord[] {
    return this.children.filter((child) => {
      if (this.statusFilter === 'checked_in' && !this.isCheckedIn(child)) return false;
      if (this.statusFilter === 'not_checked_in' && this.isCheckedIn(child)) return false;
      if (this.statusFilter === 'absent' && !this.isAbsent(child)) return false;
      if (this.searchTerm) {
        const term = this.searchTerm.toLowerCase();
        if (!child.fullName.toLowerCase().includes(term)) return false;
      }
      return true;
    });
  }

  get lastUpdatedLabel(): string {
    if (!this.lastUpdatedAt) return '';
    return new Intl.DateTimeFormat('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'Europe/London',
    }).format(this.lastUpdatedAt);
  }

  ngOnInit(): void {
    this.loadChildren('initial');
    this.startPolling();
    this.updateClock();
    this.clockInterval = setInterval(() => this.updateClock(), 1000);
  }

  ngOnDestroy(): void {
    this.stopPolling();
    if (this.clockInterval) {
      clearInterval(this.clockInterval);
    }
  }

  toggleAutoRefresh(): void {
    this.autoRefreshEnabled = !this.autoRefreshEnabled;
    if (this.autoRefreshEnabled) {
      this.loadChildren('manual');
      this.startPolling();
    } else {
      this.stopPolling();
    }
  }

  loadChildren(source: LoadSource = 'manual'): void {
    if (this.listRequestInFlight && source === 'poll') return;

    const isBackground = source === 'poll';
    const showForegroundLoading = !isBackground && this.children.length === 0;

    if (isBackground) {
      this.isBackgroundRefreshing = true;
    } else {
      this.isLoading = true;
    }
    this.errorMessage = null;
    this.listRequestInFlight = true;

    this.staffApi.listAttendanceChildren().subscribe({
      next: (children) => {
        this.children = children;
        this.pruneStaleRowErrors(children);
        this.lastUpdatedAt = new Date();
        this.isLoading = false;
        this.isBackgroundRefreshing = false;
        this.listRequestInFlight = false;
      },
      error: (error) => {
        this.isLoading = false;
        this.isBackgroundRefreshing = false;
        this.listRequestInFlight = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'attendance.list'));
      },
    });
  }

  onSearchChange(value: string): void {
    this.searchTerm = value;
  }

  setStatusFilter(filter: StatusFilter): void {
    this.statusFilter = filter;
  }

  isCheckedIn(child: AttendanceChildRecord): boolean {
    return child.attendanceState === 'checked_in' || !!child.openSessionId;
  }

  isAbsent(child: AttendanceChildRecord): boolean {
    return child.attendanceState === 'absent';
  }

  isPending(childId: string): boolean {
    return this.pendingChildIds.has(childId);
  }

  showIncompleteSessionWarning(child: AttendanceChildRecord): boolean {
    return child.hasIncompleteSession && !this.isCheckedIn(child);
  }

  canCheckIn(child: AttendanceChildRecord): boolean {
    return (
      !this.isCheckedIn(child) &&
      !this.isAbsent(child) &&
      child.attendanceState === 'not_checked_in' &&
      child.enrollmentComplete &&
      !this.isForegroundLoading() &&
      !this.isPending(child.id)
    );
  }

  canCheckOut(child: AttendanceChildRecord): boolean {
    return this.isCheckedIn(child) && !this.isForegroundLoading() && !this.isPending(child.id);
  }

  canMarkAbsent(child: AttendanceChildRecord): boolean {
    return (
      !this.isCheckedIn(child) &&
      !this.isAbsent(child) &&
      child.enrollmentComplete &&
      !this.isForegroundLoading() &&
      !this.isPending(child.id)
    );
  }

  canClearAbsence(child: AttendanceChildRecord): boolean {
    return (
      this.isAbsent(child) &&
      !!child.absenceMarkerId &&
      !this.isForegroundLoading() &&
      !this.isPending(child.id)
    );
  }

  checkIn(child: AttendanceChildRecord): void {
    if (!this.canCheckIn(child)) return;
    this.executeMutation(child.id, () => this.staffApi.checkInChild(child.id));
  }

  checkOut(child: AttendanceChildRecord): void {
    if (!this.canCheckOut(child)) return;
    this.executeMutation(child.id, () => this.staffApi.checkOutChild(child.id));
  }

  markAbsent(child: AttendanceChildRecord): void {
    if (!this.canMarkAbsent(child)) return;
    this.executeMutation(child.id, () => this.staffApi.markChildAbsent(child.id));
  }

  clearAbsence(child: AttendanceChildRecord): void {
    if (!this.canClearAbsence(child)) return;
    this.executeMutation(child.id, () => this.staffApi.clearAbsenceMarker(child.absenceMarkerId!));
  }

  formatLondonTime(iso: string | null): string {
    if (!iso) return '-';
    return new Intl.DateTimeFormat('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'Europe/London',
    }).format(new Date(iso));
  }

  private updateClock(): void {
    const now = new Date();
    this.clockTime = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    this.currentDate = now.toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'short' });
  }

  private isForegroundLoading(): boolean {
    return this.isLoading && !this.isBackgroundRefreshing;
  }

  private startPolling(): void {
    this.stopPolling();
    this.pollSubscription = interval(this.pollIntervalMs)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(() => {
        if (this.autoRefreshEnabled) {
          this.loadChildren('poll');
        }
      });
  }

  private stopPolling(): void {
    if (this.pollSubscription) {
      this.pollSubscription.unsubscribe();
      this.pollSubscription = null;
    }
  }

  private pruneStaleRowErrors(children: AttendanceChildRecord[]): void {
    const currentIds = new Set(children.map((c) => c.id));
    for (const id of Object.keys(this.rowErrors)) {
      if (!currentIds.has(id)) {
        delete this.rowErrors[id];
      }
    }
  }

  private executeMutation(childId: string, mutation: () => unknown): void {
    delete this.rowErrors[childId];
    this.pendingChildIds.add(childId);

    const { next, error, complete } = {
      next: () => {
        this.pendingChildIds.delete(childId);
        this.loadChildren('mutation');
      },
      error: (err: unknown) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.rowErrors[childId] = formatPresentedApiError(presentApiError(mapped, 'attendance.rowAction'));
        this.pendingChildIds.delete(childId);
        this.loadChildren('mutation');
      },
      complete: () => {},
    };

    (mutation() as import('rxjs').Observable<unknown>).subscribe({ next, error, complete });
  }

}
