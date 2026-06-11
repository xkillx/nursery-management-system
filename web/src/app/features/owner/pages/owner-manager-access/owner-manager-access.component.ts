import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule } from '@angular/router';

import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { LabelComponent } from '../../../../shared/components/form/label/label.component';
import { OwnerApiService } from '../../data/owner-api.service';
import {
  OwnerManagerAccessRecord,
  OwnerSiteSummary,
  OwnerGrantManagerAccessResult,
} from '../../models/owner.models';
import { formatGrantOutcome } from '../../utils/owner-formatters';

@Component({
  selector: 'app-owner-manager-access',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    PageHeaderComponent,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    ButtonComponent,
    InputFieldComponent,
    LabelComponent,
  ],
  templateUrl: './owner-manager-access.component.html',
})
export class OwnerManagerAccessComponent implements OnInit {
  private readonly api = inject(OwnerApiService);
  private readonly route = inject(ActivatedRoute);

  loadingSites = true;
  loadingAccess = false;
  submitting = false;
  pageError: string | null = null;
  successMessage: string | null = null;

  sites: OwnerSiteSummary[] = [];
  selectedSiteId: string | null = null;
  statusFilter: 'active' | 'inactive' | 'all' = 'active';
  accessRecords: OwnerManagerAccessRecord[] = [];

  grantEmail = '';
  grantEmailError: string | null = null;

  formatGrantOutcome = formatGrantOutcome;

  ngOnInit(): void {
    this.loadSites();
  }

  get canSubmitGrant(): boolean {
    return !!this.selectedSiteId && !!this.grantEmail.trim() && !this.submitting;
  }

  onSiteSelect(): void {
    this.successMessage = null;
    this.pageError = null;
    if (this.selectedSiteId) {
      this.loadAccess();
    } else {
      this.accessRecords = [];
    }
  }

  onStatusFilterChange(): void {
    if (this.selectedSiteId) {
      this.loadAccess();
    }
  }

  onGrantSubmit(): void {
    this.grantEmailError = null;
    this.successMessage = null;

    if (!this.grantEmail.trim()) {
      this.grantEmailError = 'Email is required.';
      return;
    }

    if (!this.selectedSiteId) return;

    this.submitting = true;
    this.api.grantManagerAccess(this.selectedSiteId, this.grantEmail.trim()).subscribe({
      next: (result) => {
        this.submitting = false;
        this.successMessage = formatGrantOutcome(result.outcome);
        this.grantEmail = '';
        this.loadAccess();
      },
      error: (err) => {
        this.submitting = false;
        const code = err?.error?.code;
        if (code === 'validation_error' && err?.error?.field === 'email') {
          this.grantEmailError = err.error.message;
        } else {
          this.pageError = this.mapError(err);
        }
      },
    });
  }

  onDeactivate(record: OwnerManagerAccessRecord): void {
    if (!this.selectedSiteId) return;
    if (!confirm(`Deactivate manager access for ${record.email}?`)) return;

    this.pageError = null;
    this.api.deactivateManagerAccess(this.selectedSiteId, record.membershipId).subscribe({
      next: () => {
        this.successMessage = `${record.email} deactivated.`;
        this.loadAccess();
      },
      error: (err) => {
        this.pageError = this.mapError(err);
      },
    });
  }

  onReactivate(record: OwnerManagerAccessRecord): void {
    if (!this.selectedSiteId) return;

    this.pageError = null;
    this.api.reactivateManagerAccess(this.selectedSiteId, record.membershipId).subscribe({
      next: () => {
        this.successMessage = `${record.email} reactivated.`;
        this.loadAccess();
      },
      error: (err) => {
        this.pageError = this.mapError(err);
      },
    });
  }

  private loadSites(): void {
    this.api.getSiteSummaries().subscribe({
      next: (res) => {
        this.sites = res.sites;
        this.loadingSites = false;
        this.applyQueryParam();
      },
      error: () => {
        this.pageError = 'Failed to load sites.';
        this.loadingSites = false;
      },
    });
  }

  private applyQueryParam(): void {
    const qSiteId = this.route.snapshot.queryParamMap.get('site_id');
    if (qSiteId && this.sites.some((s) => s.siteId === qSiteId)) {
      this.selectedSiteId = qSiteId;
      this.loadAccess();
    }
  }

  private loadAccess(): void {
    if (!this.selectedSiteId) return;

    this.loadingAccess = true;
    this.pageError = null;
    this.api.listManagerAccess(this.selectedSiteId, this.statusFilter).subscribe({
      next: (records) => {
        this.accessRecords = records;
        this.loadingAccess = false;
      },
      error: () => {
        this.pageError = 'Failed to load manager access.';
        this.loadingAccess = false;
      },
    });
  }

  private mapError(err: any): string {
    const code = err?.error?.code;
    if (code === 'site_not_found') return 'Site not found or no longer active.';
    if (code === 'manager_membership_not_found') return 'Manager membership not found. The list has been refreshed.';
    return 'An unexpected error occurred. Please try again.';
  }
}
