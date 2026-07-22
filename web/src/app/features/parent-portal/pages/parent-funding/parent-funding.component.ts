import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';

import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ParentFundingApiService } from '../../data/parent-funding-api.service';
import { ParentFundingEntitlement } from '../../models/parent-funding.models';

function formatAllowance(minutes: number): string {
  if (minutes === 0) return '0h 0m';
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

function childName(item: ParentFundingEntitlement): string {
  const parts = [item.childFirstName, item.childMiddleName, item.childLastName].filter(Boolean);
  return parts.join(' ');
}

@Component({
  selector: 'app-parent-funding',
  imports: [
    CommonModule,
    PageHeaderComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    AlertComponent,
  ],
  templateUrl: './parent-funding.component.html',
})
export class ParentFundingComponent implements OnInit {
  private readonly fundingApi = inject(ParentFundingApiService);

  items: ParentFundingEntitlement[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  readonly formatAllowance = formatAllowance;
  readonly childName = childName;

  ngOnInit(): void {
    this.loadFunding();
  }

  private loadFunding(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.fundingApi.getFunding().subscribe({
      next: (items) => {
        this.items = items;
        this.isLoading = false;
      },
      error: (err) => {
        this.isLoading = false;
        this.errorMessage = err?.message ?? 'Failed to load funding information.';
      },
    });
  }
}
