import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { NavigationEnd, Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroChevronDown,
  heroChevronUp,
  heroChevronUpDown,
  heroMagnifyingGlass,
  heroPlus,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';
import { Subject } from 'rxjs';
import { debounceTime, filter, takeUntil } from 'rxjs/operators';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ParentsApiService } from '../../data/parents-api.service';
import { ParentRecord, ParentStatusFilter } from '../../models/parents.models';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { TablePaginationComponent } from '../../../../shared/components/ui/table/table-pagination.component';
import { ToastService } from '../../../../shared/services/toast.service';

type SortColumn = 'name' | 'email' | 'created_at';
type SortDirection = 'asc' | 'desc';

@Component({
  selector: 'app-manager-parents',
  imports: [
    CommonModule,
    RouterLink,
    SelectComponent,
    AlertComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    TablePaginationComponent,
    NgIcon,
  ],
  templateUrl: './manager-parents.component.html',
  providers: [
    provideIcons({
      heroChevronDown,
      heroChevronUp,
      heroChevronUpDown,
      heroMagnifyingGlass,
      heroPlus,
      heroUserGroup,
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

  readonly statusOptions: ParentStatusFilter[] = ['active', 'inactive', 'all'];
  readonly statusLabel = (s: ParentStatusFilter) => s === 'active' ? 'Active' : s === 'inactive' ? 'Inactive' : 'All';
  get statusSelectOptions(): Option[] {
    return this.statusOptions.map(s => ({ value: s, label: this.statusLabel(s) }));
  }

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
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
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

  onStatusChange(nextStatus: string): void {
    this.status = nextStatus as ParentStatusFilter;
    this.page = 1;
    this.loadParents();
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
