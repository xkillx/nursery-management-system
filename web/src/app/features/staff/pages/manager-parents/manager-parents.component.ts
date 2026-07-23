import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NavigationEnd, Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroChevronDown,
  heroChevronLeft,
  heroChevronRight,
  heroChevronUp,
  heroChevronUpDown,
  heroEnvelope,
  heroExclamationCircle,
  heroEye,
  heroMagnifyingGlass,
  heroPhone,
  heroPlus,
  heroShieldCheck,
  heroUser,
  heroUserGroup,
  heroUsers,
  heroXMark,
} from '@ng-icons/heroicons/outline';
import { Subject } from 'rxjs';
import { debounceTime, filter, takeUntil } from 'rxjs/operators';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ParentsApiService } from '../../data/parents-api.service';
import { ParentRecord, ParentStatusFilter } from '../../models/parents.models';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ToastService } from '../../../../shared/services/toast.service';

type SortColumn = 'name' | 'email' | 'created_at';
type SortDirection = 'asc' | 'desc';

interface ParentMetric {
  key: string;
  label: string;
  count: number;
  tone: 'brand' | 'success' | 'neutral';
  pill: string;
}

@Component({
  selector: 'app-manager-parents',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    AlertComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    StatusBadgeComponent,
    NgIcon,
  ],
  templateUrl: './manager-parents.component.html',
  providers: [
    provideIcons({
      heroChevronDown,
      heroChevronLeft,
      heroChevronRight,
      heroChevronUp,
      heroChevronUpDown,
      heroEnvelope,
      heroExclamationCircle,
      heroEye,
      heroMagnifyingGlass,
      heroPhone,
      heroPlus,
      heroShieldCheck,
      heroUser,
      heroUserGroup,
      heroUsers,
      heroXMark,
    }),
  ],
})
export class ManagerParentsComponent implements OnInit, OnDestroy {
  private readonly parentsApi = inject(ParentsApiService);
  private readonly router = inject(Router);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);
  private readonly searchSubject = new Subject<string>();
  private readonly destroy$ = new Subject<void>();

  readonly statusChipOptions: { value: ParentStatusFilter; label: string }[] = [
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'all', label: 'All' },
  ];

  parents: ParentRecord[] = [];
  totalCount = 0;
  status: ParentStatusFilter = 'active';
  searchTerm = '';
  page = 1;
  pageSize = 25;
  isLoading = false;

  sortColumn: SortColumn | null = null;
  sortDirection: SortDirection = 'asc';

  errorMessage: string | null = null;

  private activeCount = 0;
  private inactiveCount = 0;

  get metricCards(): ParentMetric[] {
    return [
      {
        key: 'active',
        label: 'Active parents',
        count: this.activeCount,
        tone: 'success',
        pill: 'Current',
      },
      {
        key: 'inactive',
        label: 'Inactive parents',
        count: this.inactiveCount,
        tone: 'neutral',
        pill: 'Archived',
      },
      {
        key: 'total',
        label: 'Total parents',
        count: this.totalCount,
        tone: 'brand',
        pill: 'All records',
      },
    ];
  }

  metricIconToneClasses(tone: ParentMetric['tone']): string {
    const map: Record<ParentMetric['tone'], string> = {
      brand: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300',
      success: 'bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-300',
      neutral: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300',
    };
    return map[tone];
  }

  metricPillToneClasses(tone: ParentMetric['tone']): string {
    const map: Record<ParentMetric['tone'], string> = {
      brand: 'text-brand-600 dark:text-brand-300',
      success: 'text-success-600 dark:text-success-300',
      neutral: 'text-gray-600 dark:text-gray-300',
    };
    return map[tone];
  }

  metricProgressToneClasses(tone: ParentMetric['tone']): string {
    const map: Record<ParentMetric['tone'], string> = {
      brand: 'bg-brand-500',
      success: 'bg-success-500',
      neutral: 'bg-gray-400',
    };
    return map[tone];
  }

  metricProgressWidth(count: number): number {
    if (count === 0) return 0;
    const maxCount = Math.max(this.activeCount, this.inactiveCount, this.totalCount, 1);
    return Math.max(6, Math.min(100, Math.round((count / maxCount) * 100)));
  }

  get totalPages(): number {
    return Math.max(1, Math.ceil(this.totalCount / this.pageSize));
  }

  get hasActiveFilters(): boolean {
    return this.status !== 'active' || this.searchTerm.trim().length > 0;
  }

  get isSearchActive(): boolean {
    return this.searchTerm.trim().length > 0;
  }

  ngOnInit(): void {
    this.searchSubject.pipe(
      debounceTime(200),
      takeUntil(this.destroy$),
    ).subscribe(term => {
      this.searchTerm = term;
      this.page = 1;
      this.loadParents();
    });

    this.loadParents();

    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd),
      takeUntil(this.destroy$),
    ).subscribe((event) => {
      if (event.urlAfterRedirects.split('?')[0] === '/manager/parents') {
        this.loadParents();
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadParents(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.parentsApi.list(this.page, this.pageSize, this.status, this.searchTerm || undefined).subscribe({
      next: ({ parents, total_count }) => {
        this.parents = parents;
        this.totalCount = total_count;
        this.isLoading = false;
        this.loadMetricCounts();
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  private loadMetricCounts(): void {
    this.parentsApi.list(1, 1, 'active').subscribe({
      next: ({ total_count }) => (this.activeCount = total_count),
    });
    this.parentsApi.list(1, 1, 'inactive').subscribe({
      next: ({ total_count }) => (this.inactiveCount = total_count),
    });
  }

  previousPage(): void {
    if (this.page <= 1) return;
    this.page--;
    this.loadParents();
  }

  nextPage(): void {
    const maxPage = Math.ceil(this.totalCount / this.pageSize);
    if (this.page >= maxPage) return;
    this.page++;
    this.loadParents();
  }

  get hasNextPage(): boolean {
    return this.page * this.pageSize < this.totalCount;
  }

  get hasPreviousPage(): boolean {
    return this.page > 1;
  }

  onStatusToggle(nextStatus: ParentStatusFilter): void {
    this.status = this.status === nextStatus ? 'active' : nextStatus;
    this.page = 1;
    this.loadParents();
  }

  onSearchInput(value: string): void {
    this.searchSubject.next(value);
  }

  clearSearch(): void {
    this.searchTerm = '';
    this.page = 1;
    this.loadParents();
  }

  clearAllFilters(): void {
    this.status = 'active';
    this.searchTerm = '';
    this.page = 1;
    this.loadParents();
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

  get sortedParents(): ParentRecord[] {
    if (!this.sortColumn) return this.parents;

    return [...this.parents].sort((a, b) => {
      let cmp = 0;
      if (this.sortColumn === 'name') {
        cmp = this.parentName(a).localeCompare(this.parentName(b));
      } else if (this.sortColumn === 'email') {
        cmp = (a.email || '').localeCompare(b.email || '');
      } else if (this.sortColumn === 'created_at') {
        cmp = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      }
      return this.sortDirection === 'asc' ? cmp : -cmp;
    });
  }

  parentName(parent: ParentRecord): string {
    return [parent.first_name, parent.last_name].filter(Boolean).join(' ');
  }

  formatPhone(parent: ParentRecord): string {
    return parent.phone || '—';
  }

  formatDate(value: string | null): string {
    if (!value) return '—';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value || '—';
    return new Intl.DateTimeFormat('en-GB', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    }).format(date);
  }

  viewParent(parent: ParentRecord): void {
    this.router.navigate(['/manager/parents', parent.id]);
  }
}
