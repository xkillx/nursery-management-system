import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArchiveBox,
  heroArrowPath,
  heroClock,
  heroPlus,
  heroRectangleStack,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import {
  StaffSessionType,
  StaffSessionTypesApiService,
} from '../../data/session-types-api.service';

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
  ],
  templateUrl: './manager-session-types.component.html',
  providers: [
    provideIcons({
      heroArchiveBox,
      heroArrowPath,
      heroClock,
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
  includeArchived = false;
  types: StaffSessionType[] = [];

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

  editRoute(t: StaffSessionType): string {
    return `${ROLE_ROUTES.managerSessionTypes}/${t.id}/edit`;
  }

  reload(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.pageError = null;
    this.api.listSessionTypes(this.siteId, { includeArchived: this.includeArchived }).subscribe({
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

  toggleArchived(): void {
    this.includeArchived = !this.includeArchived;
    this.reload();
  }

  archive(t: StaffSessionType): void {
    if (!this.siteId || !t.isActive) return;
    if (!confirm(`Archive "${t.name}"? It can no longer be used for new booking patterns.`)) return;
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
