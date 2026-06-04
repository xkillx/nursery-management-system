import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { StaffApiService } from '../../data/staff-api.service';
import { AttendanceChildRecord, AttendanceState } from '../../models/attendance-child.models';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

type StatusFilter = 'all' | 'not_checked_in' | 'checked_in';

@Component({
  selector: 'app-practitioner-attendance-children',
  imports: [
    CommonModule,
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  templateUrl: './practitioner-attendance-children.component.html',
})
export class PractitionerAttendanceChildrenComponent {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  children: AttendanceChildRecord[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  searchTerm = '';
  statusFilter: StatusFilter = 'all';
  rowErrors: Record<string, string> = {};
  pendingChildIds = new Set<string>();

  get checkedInCount(): number {
    return this.children.filter((c) => this.isCheckedIn(c)).length;
  }

  get notInCount(): number {
    return this.children.filter((c) => !this.isCheckedIn(c)).length;
  }

  get filteredChildren(): AttendanceChildRecord[] {
    return this.children.filter((child) => {
      if (this.statusFilter === 'checked_in' && !this.isCheckedIn(child)) return false;
      if (this.statusFilter === 'not_checked_in' && this.isCheckedIn(child)) return false;
      if (this.searchTerm) {
        const term = this.searchTerm.toLowerCase();
        if (!child.fullName.toLowerCase().includes(term)) return false;
      }
      return true;
    });
  }

  ngOnInit(): void {
    this.loadChildren();
  }

  loadChildren(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi.listAttendanceChildren().subscribe({
      next: (children) => {
        this.children = children;
        this.isLoading = false;
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
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

  isPending(childId: string): boolean {
    return this.pendingChildIds.has(childId);
  }

  canCheckIn(child: AttendanceChildRecord): boolean {
    return !this.isCheckedIn(child) && child.enrollmentComplete && !this.isLoading && !this.isPending(child.id);
  }

  canCheckOut(child: AttendanceChildRecord): boolean {
    return this.isCheckedIn(child) && !this.isLoading && !this.isPending(child.id);
  }

  checkIn(child: AttendanceChildRecord): void {
    this.executeMutation(child.id, () => this.staffApi.checkInChild(child.id));
  }

  checkOut(child: AttendanceChildRecord): void {
    this.executeMutation(child.id, () => this.staffApi.checkOutChild(child.id));
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

  private executeMutation(childId: string, mutation: () => unknown): void {
    delete this.rowErrors[childId];
    this.pendingChildIds.add(childId);

    const { next, error, complete } = {
      next: () => {
        this.pendingChildIds.delete(childId);
        this.loadChildren();
      },
      error: (err: unknown) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.rowErrors[childId] = this.messageWithRequestId(mapped.message, mapped.requestId);
        this.pendingChildIds.delete(childId);
        this.loadChildren();
      },
      complete: () => {},
    };

    (mutation() as import('rxjs').Observable<unknown>).subscribe({ next, error, complete });
  }

  private messageWithRequestId(message: string, requestId: string | null): string {
    return requestId ? `${message} (Request: ${requestId})` : message;
  }
}
