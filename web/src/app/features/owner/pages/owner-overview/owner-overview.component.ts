import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';

import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { OwnerApiService } from '../../data/owner-api.service';
import {
  OwnerSiteSummariesResponse,
  OwnerSiteSummary,
  OwnerSiteSummaryTotals,
} from '../../models/owner.models';
import { formatGbp, formatSetupStatus, isExceptionSite, formatSiteRate } from '../../utils/owner-formatters';

@Component({
  selector: 'app-owner-overview',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
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

  editingRateSiteId: string | null = null;
  rateEditValue = '';
  rateEditError: string | null = null;

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
  formatSiteRate = formatSiteRate;

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

  startRateEdit(site: OwnerSiteSummary): void {
    this.editingRateSiteId = site.siteId;
    this.rateEditValue = site.siteCoreHourlyRateMinor !== null && site.siteCoreHourlyRateMinor !== undefined
      ? (site.siteCoreHourlyRateMinor / 100).toFixed(2)
      : '';
    this.rateEditError = null;
  }

  cancelRateEdit(): void {
    this.editingRateSiteId = null;
    this.rateEditValue = '';
    this.rateEditError = null;
  }

  saveSiteRate(siteId: string): void {
    const pounds = parseFloat(this.rateEditValue);
    if (isNaN(pounds) || pounds <= 0) {
      this.rateEditError = 'Enter a positive rate.';
      return;
    }
    const minor = Math.round(pounds * 100);
    this.rateEditError = null;

    this.api.updateSiteBillingSetup(siteId, minor).subscribe({
      next: () => {
        this.editingRateSiteId = null;
        this.rateEditValue = '';
        this.loadSummaries();
      },
      error: () => {
        this.rateEditError = 'Failed to save rate. Please try again.';
      },
    });
  }

  updateSiteRate(siteId: string, rateMinor: number): void {
    this.api.updateSiteBillingSetup(siteId, rateMinor).subscribe({
      next: () => this.loadSummaries(),
    });
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
