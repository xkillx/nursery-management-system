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

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { formatPresentedApiError, presentApiError } from '../../../../core/errors/api-error-presenter';
import { AuthService } from '../../../../core/services/auth.service';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { StaffSessionTemplatesApiService } from '../../data/session-templates-api.service';
import { SessionTemplateListItem } from '../../models/session-template.models';

type TemplateStatusFilter = 'all' | 'active' | 'archived';

@Component({
  selector: 'app-manager-session-templates',
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
  templateUrl: './manager-session-templates.component.html',
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
export class ManagerSessionTemplatesComponent implements OnInit {
  private readonly templatesApi = inject(StaffSessionTemplatesApiService);
  private readonly auth = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);

  readonly newRoute = `${ROLE_ROUTES.managerSessionTemplates}/new`;

  loading = false;
  pageError: string | null = null;
  mutatingId: string | null = null;

  siteId: string | null = null;
  siteName = '';
  searchTerm = '';
  statusFilter: TemplateStatusFilter = 'active';
  templates: SessionTemplateListItem[] = [];

  readonly statusOptions: Option[] = [
    { value: 'all', label: 'All templates' },
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

  get activeTemplates(): SessionTemplateListItem[] {
    return this.templates.filter((t) => t.isActive);
  }

  get archivedTemplates(): SessionTemplateListItem[] {
    return this.templates.filter((t) => !t.isActive);
  }

  get activeCount(): number {
    return this.activeTemplates.length;
  }

  get archivedCount(): number {
    return this.archivedTemplates.length;
  }

  get totalCount(): number {
    return this.templates.length;
  }

  get activePill(): string {
    return this.activeCount === 0 ? 'No active templates' : 'Live snapshot';
  }

  get archivedPill(): string {
    return this.archivedCount === 0 ? 'None archived' : 'Historical';
  }

  get filteredRows(): SessionTemplateListItem[] {
    const term = this.searchTerm.trim().toLowerCase();
    return this.templates.filter((t) => {
      if (this.statusFilter === 'active' && !t.isActive) return false;
      if (this.statusFilter === 'archived' && t.isActive) return false;
      if (term && !t.name.toLowerCase().includes(term)) return false;
      return true;
    });
  }

  get visibleRows(): SessionTemplateListItem[] {
    return this.filteredRows;
  }

  get emptyTitle(): string {
    if (this.searchTerm) return 'No templates match your search';
    if (this.statusFilter === 'archived') return 'No archived templates';
    if (this.statusFilter === 'active') return 'No active templates';
    return 'No templates yet';
  }

  get emptyMessage(): string {
    if (this.searchTerm) return 'Try a different name or adjust the status filter.';
    if (this.statusFilter === 'archived') return 'Archived templates will appear here once you archive an active template.';
    if (this.statusFilter === 'active') return 'Create your first template to start planning booking patterns for this site.';
    return 'Create your first template to start planning booking patterns for this site.';
  }

  editRoute(t: SessionTemplateListItem): string {
    return `${ROLE_ROUTES.managerSessionTemplates}/${t.id}/edit`;
  }

  onSearchChange(event: Event): void {
    this.searchTerm = (event.target as HTMLInputElement).value;
  }

  onStatusFilterChange(value: string): void {
    this.statusFilter = (value as TemplateStatusFilter) || 'all';
  }

  reload(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.pageError = null;
    this.templatesApi.listSessionTemplates(this.siteId, { includeArchived: true }).subscribe({
      next: (templates) => {
        this.templates = templates;
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = this.formatError(err, 'Failed to load session templates.');
      },
    });
  }

  archive(t: SessionTemplateListItem): void {
    if (!this.siteId || !t.isActive) return;
    if (!confirm(`Archive "${t.name}"? It can no longer be used when creating new booking patterns.`)) return;
    this.mutatingId = t.id;
    this.templatesApi.archiveSessionTemplate(this.siteId, t.id).subscribe({
      next: () => {
        this.mutatingId = null;
        this.reload();
      },
      error: (err) => {
        this.mutatingId = null;
        this.pageError = this.formatError(err, 'Failed to archive template.');
      },
    });
  }

  reactivate(t: SessionTemplateListItem): void {
    if (!this.siteId || t.isActive) return;
    this.mutatingId = t.id;
    this.templatesApi.reactivateSessionTemplate(this.siteId, t.id).subscribe({
      next: () => {
        this.mutatingId = null;
        this.reload();
      },
      error: (err) => {
        this.mutatingId = null;
        this.pageError = this.formatError(err, 'Failed to reactivate template.');
      },
    });
  }

  formatError(err: unknown, fallback: string): string {
    const mapped = this.errorMapper.mapAndHandle(err);
    return formatPresentedApiError(presentApiError(mapped, 'sessionTemplates.list')) ?? fallback;
  }
}
