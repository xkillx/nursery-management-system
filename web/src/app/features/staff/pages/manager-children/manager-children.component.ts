import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCheckCircle,
  heroChevronDown,
  heroChevronUp,
  heroChevronUpDown,
  heroClock,
  heroExclamationCircle,
  heroMagnifyingGlass,
  heroPlus,
  heroUserGroup,
  heroPencilSquare,
} from '@ng-icons/heroicons/outline';
import { Subject } from 'rxjs';
import { debounceTime, filter, takeUntil } from 'rxjs/operators';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { StaffApiService } from '../../data/staff-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ChildRecord, StatusFilter } from '../../models/children.models';
import { missingRequirementLabel, statusFilterLabel } from '../../utils/manager-list-formatters';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AvatarTextComponent } from '../../../../shared/components/ui/avatar/avatar-text.component';
import { TablePaginationComponent } from '../../../../shared/components/ui/table/table-pagination.component';
import { ToastService } from '../../../../shared/services/toast.service';

type SortColumn = 'name' | 'age';
type SortDirection = 'asc' | 'desc';

@Component({
  selector: 'app-manager-children',
  imports: [
    CommonModule,
    RouterLink,
    SelectComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    AvatarTextComponent,
    TablePaginationComponent,
    NgIcon,
  ],
  templateUrl: './manager-children.component.html',
  providers: [
    provideIcons({
      heroCheckCircle,
      heroChevronDown,
      heroChevronUp,
      heroChevronUpDown,
      heroClock,
      heroExclamationCircle,
      heroMagnifyingGlass,
      heroPlus,
      heroUserGroup,
      heroPencilSquare,
    }),
  ],
})
export class ManagerChildrenComponent implements OnInit, OnDestroy {
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);
  private readonly searchSubject = new Subject<string>();
  private readonly destroy$ = new Subject<void>();

  readonly statusOptions: StatusFilter[] = ['active', 'inactive', 'all'];

  readonly statusLabel = statusFilterLabel;
  get statusSelectOptions(): Option[] {
    return this.statusOptions.map(s => ({ value: s, label: statusFilterLabel(s) }));
  }
  readonly requirementLabel = missingRequirementLabel;

  children: ChildRecord[] = [];
  totalCount = 0;
  status: StatusFilter = 'active';
  searchTerm = '';
  limit = 25;
  offset = 0;
  isLoading = false;
  isLoadingCards = true;

  sortColumn: SortColumn | null = null;
  sortDirection: SortDirection = 'asc';

  errorMessage: string | null = null;

  get isSearchActive(): boolean {
    return this.searchTerm.trim().length > 0;
  }

  get filteredChildren(): ChildRecord[] {
    const term = this.searchTerm.trim().toLowerCase();
    let result = this.children;

    if (term) {
      result = result.filter((child) => {
        const searchableText = [
          child.fullName,
          child.dateOfBirth,
          child.startDate,
          this.formatMissingRequirements(child),
        ].join(' ').toLowerCase();

        return searchableText.includes(term);
      });
    }

    result = this.sortData(result);
    return result;
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

  ngOnInit(): void {
    this.searchSubject.pipe(
      debounceTime(200),
      takeUntil(this.destroy$),
    ).subscribe(term => {
      this.searchTerm = term;
    });

    this.loadChildren();

    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd),
      takeUntil(this.destroy$),
    ).subscribe((event) => {
      if (event.urlAfterRedirects.split('?')[0] === '/manager/children') {
        this.loadChildren();
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
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
        next: ({ items, total }) => {
          this.children = items;
          this.totalCount = total;
          this.isLoading = false;
          this.isLoadingCards = false;
        },
        error: (error) => {
          this.isLoading = false;
          this.isLoadingCards = false;
          const mapped = this.errorMapper.mapAndHandle(error);
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        },
      });
  }

  previousPage(): void {
    if (this.offset === 0) return;
    this.offset -= this.limit;
    this.loadChildren();
  }

  nextPage(): void {
    if (this.offset + this.limit >= this.totalCount) return;
    this.offset += this.limit;
    this.loadChildren();
  }

  onStatusChange(nextStatus: string): void {
    this.status = nextStatus as StatusFilter;
    this.offset = 0;
    this.loadChildren();
  }

  onSearchChange(event: Event): void {
    this.searchSubject.next((event.target as HTMLInputElement).value);
  }

  toggleSort(column: SortColumn): void {
    if (this.sortColumn === column) {
      this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortColumn = column;
      this.sortDirection = 'asc';
    }
  }

  getSortIcon(column: SortColumn): string {
    if (this.sortColumn !== column) return 'heroChevronUpDown';
    return this.sortDirection === 'asc' ? 'heroChevronUp' : 'heroChevronDown';
  }

  private sortData(data: ChildRecord[]): ChildRecord[] {
    if (!this.sortColumn) return data;

    return [...data].sort((a, b) => {
      let cmp = 0;
      if (this.sortColumn === 'name') {
        cmp = a.fullName.localeCompare(b.fullName);
      } else if (this.sortColumn === 'age') {
        cmp = new Date(a.dateOfBirth).getTime() - new Date(b.dateOfBirth).getTime();
      }
      return this.sortDirection === 'asc' ? cmp : -cmp;
    });
  }

  activeFilterByStatus(status: 'incomplete' | 'requirements'): void {
    this.status = 'all';
    this.offset = 0;
    this.searchTerm = '';
    this.loadChildren();
  }

  openEdit(child: ChildRecord): void {
    this.router.navigate(['/manager/children', child.id, 'edit']);
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

  formatRoomLabel(child: ChildRecord): string {
    return child.hasCurrentRoom ? 'Assigned' : 'Not assigned';
  }

  childOverviewStatus(child: ChildRecord): string {
    if (!child.isActive) return 'inactive';
    if (!child.enrollmentComplete) return 'incomplete';
    if (!child.hasCurrentRoom || child.hasBookingPattern === false) return 'incomplete';
    return 'enrolled';
  }

  childOverStatus(child: ChildRecord): string {
    return child.enrollmentComplete ? 'complete' : 'incomplete';
  }
}
