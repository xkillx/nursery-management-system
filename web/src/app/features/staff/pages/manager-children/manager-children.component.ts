import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowDownTray,
  heroCheckCircle,
  heroChevronLeft,
  heroChevronRight,
  heroClock,
  heroEllipsisVertical,
  heroExclamationCircle,
  heroFunnel,
  heroPlus,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ChildFormComponent } from '../../components/child-form/child-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ChildRecord, ChildWritePayload, StatusFilter } from '../../models/children.models';
import { missingRequirementLabel, statusFilterLabel } from '../../utils/manager-list-formatters';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AvatarTextComponent } from '../../../../shared/components/ui/avatar/avatar-text.component';

@Component({
  selector: 'app-manager-children',
  imports: [
    CommonModule,
    RouterLink,
    SelectComponent,
    ChildFormComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    AvatarTextComponent,
    NgIcon,
  ],
  templateUrl: './manager-children.component.html',
  providers: [
    provideIcons({
      heroArrowDownTray,
      heroCheckCircle,
      heroChevronLeft,
      heroChevronRight,
      heroClock,
      heroEllipsisVertical,
      heroExclamationCircle,
      heroFunnel,
      heroPlus,
      heroUserGroup,
    }),
  ],
})
export class ManagerChildrenComponent {
  private readonly staffApi = inject(StaffApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly auth = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);

  readonly statusOptions: StatusFilter[] = ['active', 'inactive', 'all'];

  readonly statusLabel = statusFilterLabel;
  get statusSelectOptions(): Option[] {
    return this.statusOptions.map(s => ({ value: s, label: statusFilterLabel(s) }));
  }
  readonly requirementLabel = missingRequirementLabel;

  roomOptions: Option[] = [];

  children: ChildRecord[] = [];
  status: StatusFilter = 'active';
  searchTerm = '';
  limit = 10;
  offset = 0;
  isLoading = false;
  isSaving = false;

  selectedChild: ChildRecord | null = null;
  showForm = false;

  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};

  get visibleRangeStart(): number {
    return this.filteredChildren.length === 0 ? 0 : this.offset + 1;
  }

  get visibleRangeEnd(): number {
    return this.offset + this.filteredChildren.length;
  }

  get filteredChildren(): ChildRecord[] {
    const term = this.searchTerm.trim().toLowerCase();
    if (!term) {
      return this.children;
    }

    return this.children.filter((child) => {
      const searchableText = [
        child.fullName,
        child.id,
        child.dateOfBirth,
        child.startDate,
        this.formatMissingRequirements(child),
      ].join(' ').toLowerCase();

      return searchableText.includes(term);
    });
  }

  get activeCount(): number {
    return this.children.filter(child => child.isActive).length;
  }

  get incompleteCount(): number {
    return this.children.filter(child => !child.enrollmentComplete).length;
  }

  get missingRequirementCount(): number {
    return this.children.reduce((total, child) => total + child.missingRequirements.length, 0);
  }

  get canGoNext(): boolean {
    return this.children.length >= this.limit && !this.isLoading;
  }

  get canGoPrevious(): boolean {
    return this.offset > 0 && !this.isLoading;
  }

  ngOnInit(): void {
    this.loadChildren();
    this.loadRoomOptions();
  }

  private loadRoomOptions(): void {
    const branchId = this.auth.activeMembership()?.branch_id;
    if (!branchId) {
      this.roomOptions = [];
      return;
    }
    this.roomsApi
      .listRooms(branchId, { includeArchived: false })
      .subscribe({
        next: (rooms) => {
          this.roomOptions = rooms
            .filter((room) => room.isActive)
            .map((room) => ({ value: room.id, label: room.name }));
        },
        error: () => {
          this.roomOptions = [];
        },
      });
  }

  loadChildren(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi
      .listChildren({
        status: this.status,
        limit: this.limit,
        offset: this.offset,
      })
      .subscribe({
        next: (children) => {
          this.children = children;
          this.isLoading = false;
        },
        error: (error) => {
          this.isLoading = false;
          const mapped = this.errorMapper.mapAndHandle(error);
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        },
      });
  }

  onStatusChange(nextStatus: string): void {
    this.status = nextStatus as StatusFilter;
    this.offset = 0;
    this.searchTerm = '';
    this.loadChildren();
  }

  onSearchChange(event: Event): void {
    this.searchTerm = (event.target as HTMLInputElement).value;
  }

  openEdit(child: ChildRecord): void {
    this.selectedChild = child;
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showForm = true;
  }

  closeForm(): void {
    this.showForm = false;
    this.selectedChild = null;
    this.fieldErrors = {};
    this.errorMessage = null;
  }

  save(payload: ChildWritePayload): void {
    this.isSaving = true;
    this.fieldErrors = {};
    this.errorMessage = null;

    this.staffApi.updateChild(this.selectedChild!.id, payload).subscribe({
      next: () => {
        this.isSaving = false;
        this.closeForm();
        this.loadChildren();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  nextPage(): void {
    if (this.children.length < this.limit) {
      return;
    }

    this.offset += this.limit;
    this.loadChildren();
  }

  previousPage(): void {
    if (this.offset === 0) {
      return;
    }

    this.offset = Math.max(0, this.offset - this.limit);
    this.loadChildren();
  }

  formatMissingRequirements(child: ChildRecord): string {
    if (child.missingRequirements.length === 0) {
      return 'All set';
    }

    return child.missingRequirements.map(this.requirementLabel).join(', ');
  }

  formatAge(dateOfBirth: string): string {
    const birthDate = new Date(dateOfBirth);
    if (Number.isNaN(birthDate.getTime())) {
      return 'Age unavailable';
    }

    const today = new Date();
    let months = (today.getFullYear() - birthDate.getFullYear()) * 12;
    months += today.getMonth() - birthDate.getMonth();
    if (today.getDate() < birthDate.getDate()) {
      months -= 1;
    }

    if (months < 12) {
      return `${Math.max(months, 0)}m`;
    }

    const years = Math.floor(months / 12);
    const remainingMonths = months % 12;
    return remainingMonths === 0 ? `${years}y` : `${years}y ${remainingMonths}m`;
  }

  formatDate(value: string | null): string {
    if (!value) {
      return 'Not set';
    }

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }

    return new Intl.DateTimeFormat('en-GB', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    }).format(date);
  }

  exportChildren(): void {
    const rows = this.filteredChildren.map((child) => [
      child.fullName,
      child.id,
      child.dateOfBirth,
      child.startDate,
      child.isActive ? 'Active' : 'Inactive',
      child.enrollmentComplete ? 'Complete' : 'Incomplete',
      this.formatMissingRequirements(child),
    ]);
    const csv = [
      ['Child name', 'ID', 'Date of birth', 'Start date', 'Active status', 'Enrollment status', 'Missing requirements'],
      ...rows,
    ]
      .map(row => row.map(value => `"${String(value).replace(/"/g, '""')}"`).join(','))
      .join('\n');

    const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8;' }));
    const link = document.createElement('a');
    link.href = url;
    link.download = 'children-directory.csv';
    link.click();
    URL.revokeObjectURL(url);
  }

}
