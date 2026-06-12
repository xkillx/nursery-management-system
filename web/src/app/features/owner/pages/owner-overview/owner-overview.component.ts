import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';

import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { OwnerApiService } from '../../data/owner-api.service';
import {
  OwnerSiteSummariesResponse,
  OwnerSiteSummary,
  OwnerSiteSummaryTotals,
} from '../../models/owner.models';
import { formatGbp, formatSetupStatus, isExceptionSite } from '../../utils/owner-formatters';

@Component({
  selector: 'app-owner-overview',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    PageHeaderComponent,
    SelectComponent,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
  ],
  templateUrl: './owner-overview.component.html',
})
export class OwnerOverviewComponent implements OnInit {
  private readonly api = inject(OwnerApiService);

  loading = true;
  error: string | null = null;
  billingMonthControl = '';
  selectedSiteId: string | null = null;
  billingMonthError: string | null = null;

  response: OwnerSiteSummariesResponse | null = null;

  ngOnInit(): void {
    this.loadSummaries();
  }

  get sites(): OwnerSiteSummary[] {
    return this.response?.sites ?? [];
  }

  get totals(): OwnerSiteSummaryTotals | null {
    return this.response?.totals ?? null;
  }

  get sortedSites(): OwnerSiteSummary[] {
    return [...this.sites].sort((a, b) => {
      const aEx = isExceptionSite(a) ? 0 : 1;
      const bEx = isExceptionSite(b) ? 0 : 1;
      return aEx - bEx;
    });
  }

  formatGbp = formatGbp;
  formatSetupStatus = formatSetupStatus;
  isExceptionSite = isExceptionSite;

  get siteFilterOptions(): Option[] {
    return [
      { value: '', label: 'All sites' },
      ...this.sites.map(s => ({ value: s.siteId, label: s.siteName })),
    ];
  }

  onSiteFilterChange(value: string): void {
    this.selectedSiteId = value || null;
    this.onSiteFocus(this.selectedSiteId);
  }

  onBillingMonthChange(): void {
    this.billingMonthError = null;
    if (this.billingMonthControl && !/^\d{4}-\d{2}$/.test(this.billingMonthControl)) {
      this.billingMonthError = 'Enter a valid month in YYYY-MM format.';
      return;
    }
    this.loadSummaries();
  }

  onSiteFocus(siteId: string | null): void {
    this.selectedSiteId = siteId;
    this.loadSummaries();
  }

  private loadSummaries(): void {
    this.loading = true;
    this.error = null;

    const params: { billingMonth?: string; siteId?: string } = {};
    if (this.billingMonthControl && /^\d{4}-\d{2}$/.test(this.billingMonthControl)) {
      params.billingMonth = this.billingMonthControl;
    }
    if (this.selectedSiteId) {
      params.siteId = this.selectedSiteId;
    }

    this.api.getSiteSummaries(params).subscribe({
      next: (res) => {
        this.response = res;
        this.loading = false;
      },
      error: () => {
        this.error = 'Failed to load site summaries. Please try again.';
        this.loading = false;
      },
    });
  }
}
