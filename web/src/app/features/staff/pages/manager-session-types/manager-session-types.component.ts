import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArchiveBox,
  heroArchiveBoxXMark,
  heroArrowPath,
  heroArrowUturnLeft,
  heroClock,
  heroFunnel,
  heroPencilSquare,
  heroPlus,
  heroRectangleStack,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import {
  StaffSessionType,
  StaffSessionTypesApiService,
} from '../../data/session-types-api.service';

type SessionTypeStatusFilter = 'all' | 'active' | 'archived';

@Component({
  selector: 'app-manager-session-types',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    SelectComponent,
  ],
  templateUrl: './manager-session-types.component.html',
  providers: [
    provideIcons({
      heroArchiveBox,
      heroArchiveBoxXMark,
      heroArrowPath,
      heroArrowUturnLeft,
      heroClock,
      heroFunnel,
      heroPencilSquare,
      heroPlus,
      heroRectangleStack,
    }),
  ],
})
export class ManagerSessionTypesComponent implements OnInit {
  private readonly api = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);

  readonly newRoute = `${ROLE_ROUTES.managerSessionTypes}/new`;

  loading = false;
  pageError: string | null = null;
  mutatingId: string | null = null;

  siteId: string | null = null;
  siteName = '';
  searchTerm = '';
  statusFilter: SessionTypeStatusFilter = 'active';
  types: StaffSessionType[] = [];

  readonly statusOptions: Option[] = [
    { value: 'all', label: 'All session types' },
    { value: 'active', label: 'Active only' },
    { value: 'archived', label: 'Archived only' },
  ];

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.pageError = 'No site is attached to this manager session.';
      return;
    }
    this.siteId = membership.branch_id;
    this.siteName = membership.branch_name ?? 'Assigned site';
    this.reload();
  }

  get activeTypes(): StaffSessionType[] {
    return this.types.filter((t) => t.isActive);
  }

  get archivedTypes(): StaffSessionType[] {
    return this.types.filter((t) => !t.isActive);
  }

  get activeCount(): number {
    return this.activeTypes.length;
  }

  get archivedCount(): number {
    return this.archivedTypes.length;
  }

  get totalCount(): number {
    return this.types.length;
  }

  get activePill(): string {
    return this.activeCount === 0 ? 'No active types' : 'Live snapshot';
  }

  get archivedPill(): string {
    return this.archivedCount === 0 ? 'None archived' : 'Historical';
  }

  get filteredRows(): StaffSessionType[] {
    const term = this.searchTerm.trim().toLowerCase();
    return this.types.filter((t) => {
      if (this.statusFilter === 'active' && !t.isActive) return false;
      if (this.statusFilter === 'archived' && t.isActive) return false;
      if (term && !t.name.toLowerCase().includes(term)) return false;
      return true;
    });
  }

  get visibleRows(): StaffSessionType[] {
    return this.filteredRows;
  }

  get emptyTitle(): string {
    if (this.searchTerm) return 'No session types match your search';
    if (this.statusFilter === 'archived') return 'No archived session types';
    if (this.statusFilter === 'active') return 'No active session types';
    return 'No session types yet';
  }

  get emptyMessage(): string {
    if (this.searchTerm) return 'Try a different name or adjust the status filter.';
    if (this.statusFilter === 'archived') return 'Archived types will appear here once you archive an active type.';
    if (this.statusFilter === 'active') return 'Create your first session type to start planning bookings for this site.';
    return 'Create your first session type to start planning bookings for this site.';
  }

  editRoute(t: StaffSessionType): string {
    return `${ROLE_ROUTES.managerSessionTypes}/${t.id}/edit`;
  }

  onSearchChange(event: Event): void {
    this.searchTerm = (event.target as HTMLInputElement).value;
  }

  onStatusFilterChange(value: string): void {
    this.statusFilter = (value as SessionTypeStatusFilter) || 'all';
  }

  reload(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.pageError = null;
    this.api.listSessionTypes(this.siteId, { includeArchived: true }).subscribe({
      next: (types) => {
        this.types = types;
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = err?.error?.message ?? 'Failed to load session types.';
      },
    });
  }

  archive(t: StaffSessionType): void {
    if (!this.siteId || !t.isActive) return;
    if (!confirm(`Archive "${t.name}"? It can no longer be used for new bookings.`)) return;
    this.mutatingId = t.id;
    this.api.archiveSessionType(this.siteId, t.id).subscribe({
      next: () => {
        this.mutatingId = null;
        this.reload();
      },
      error: (err) => {
        this.mutatingId = null;
        this.pageError = err?.error?.message ?? 'Failed to archive session type.';
      },
    });
  }

  reactivate(t: StaffSessionType): void {
    if (!this.siteId || t.isActive) return;
    this.mutatingId = t.id;
    this.api.reactivateSessionType(this.siteId, t.id).subscribe({
      next: () => {
        this.mutatingId = null;
        this.reload();
      },
      error: (err) => {
        this.mutatingId = null;
        this.pageError = err?.error?.message ?? 'Failed to reactivate session type.';
      },
    });
  }
}
