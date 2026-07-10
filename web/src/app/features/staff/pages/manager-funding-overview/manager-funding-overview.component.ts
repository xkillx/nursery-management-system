import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroEye } from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { StaffApiService } from '../../data/staff-api.service';
import { FundingOverviewFlag, FundingOverviewRecord } from '../../models/funding.models';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ChildAvatarComponent } from '../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';

const FLAG_LABELS: Record<FundingOverviewFlag, string> = {
  missing_profile: 'Missing allowance',
  explicit_zero_allowance: 'Zero allowance',
  under_one_hour_allowance: 'Under one hour',
  above_160_hours_allowance: 'Above 160 hours',
};

function currentBillingMonth(): string {
  const now = new Date();
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, '0');
  return `${y}-${m}`;
}

function formatAllowance(minutes: number | null): string {
  if (minutes === null) return 'Not set';
  if (minutes === 0) return '0h 0m';
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

function flagColor(flag: FundingOverviewFlag): 'warning' | 'danger' | 'info' {
  switch (flag) {
    case 'missing_profile': return 'danger';
    case 'explicit_zero_allowance': return 'warning';
    case 'under_one_hour_allowance': return 'warning';
    case 'above_160_hours_allowance': return 'info';
  }
}

@Component({
  selector: 'app-manager-funding-overview',
  imports: [
    CommonModule,
    RouterLink,
    NgIcon,
    PageHeaderComponent,
    AlertComponent,
    TableShellComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    ChildAvatarComponent,
  ],
  templateUrl: './manager-funding-overview.component.html',
  providers: [
    provideIcons({ heroEye }),
  ],
})
export class ManagerFundingOverviewComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  selectedBillingMonth = currentBillingMonth();
  overview: FundingOverviewRecord | null = null;
  isLoading = false;
  errorMessage: string | null = null;

  readonly flagLabel = (flag: FundingOverviewFlag) => FLAG_LABELS[flag];
  readonly formatAllowance = formatAllowance;
  readonly flagColor = flagColor;

  ngOnInit(): void {
    this.loadOverview();
  }

  onMonthChange(month: string): void {
    this.selectedBillingMonth = month;
    this.loadOverview();
  }

  reviewLink(childId: string): string[] {
    return ['/manager/children', childId];
  }

  reviewQueryParams(): Record<string, string> {
    return { billing_month: this.selectedBillingMonth };
  }

  private loadOverview(): void {
    this.isLoading = true;
    this.errorMessage = null;
    this.overview = null;

    this.staffApi.getFundingOverview(this.selectedBillingMonth).subscribe({
      next: (result) => {
        this.overview = result;
        this.isLoading = false;
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = mapped.requestId
          ? `${mapped.message} (Request ID: ${mapped.requestId})`
          : mapped.message;
      },
    });
  }
}
