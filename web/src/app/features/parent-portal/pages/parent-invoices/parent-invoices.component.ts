import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ParentInvoicesApiService } from '../../data/parent-invoices-api.service';
import { ParentInvoiceListItem, ChildInvoiceGroup } from '../../models/parent-invoices.models';
import {
  formatGbp,
  formatBillingMonthLabel,
  formatInstant,
  isPayableInvoice,
  isAttentionInvoice,
  sortAttentionInvoices,
  groupHistoryByChild,
  balanceDueMinor,
} from '../../utils/parent-invoice-formatters';

const LIMIT = 200;

@Component({
  selector: 'app-parent-invoices',
  imports: [
    CommonModule,
    RouterLink,
    PageHeaderComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    AlertComponent,
    StatusBadgeComponent,
  ],
  templateUrl: './parent-invoices.component.html',
})
export class ParentInvoicesComponent implements OnInit {
  private readonly apiService = inject(ParentInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  items: ParentInvoiceListItem[] = [];
  offset = 0;
  isLoading = false;
  isLoadingMore = false;
  errorMessage: string | null = null;
  payingInvoiceIds = new Set<string>();

  readonly formatGbp = formatGbp;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly balanceDueMinor = balanceDueMinor;

  get attentionInvoices(): ParentInvoiceListItem[] {
    return this.items
      .filter(isAttentionInvoice)
      .sort(sortAttentionInvoices);
  }

  get historyGroups(): ChildInvoiceGroup[] {
    const attentionIds = new Set(this.attentionInvoices.map((i) => i.invoiceId));
    const historyItems = this.items.filter((i) => !attentionIds.has(i.invoiceId));
    return groupHistoryByChild(historyItems);
  }

  get hasMore(): boolean {
    return this.items.length > 0 && this.items.length % LIMIT === 0;
  }

  isPayable = isPayableInvoice;

  ngOnInit(): void {
    this.loadInvoices();
  }

  loadMore(): void {
    this.isLoadingMore = true;
    this.offset += LIMIT;
    this.fetchInvoices(true);
  }

  startPayment(invoiceId: string): void {
    if (this.payingInvoiceIds.has(invoiceId)) return;
    this.payingInvoiceIds.add(invoiceId);

    this.apiService.createCheckoutSession(invoiceId).subscribe({
      next: (result) => {
        this.redirectTo(result.checkoutUrl);
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
        this.payingInvoiceIds.delete(invoiceId);
      },
    });
  }

  isPaying(invoiceId: string): boolean {
    return this.payingInvoiceIds.has(invoiceId);
  }

  private loadInvoices(): void {
    this.isLoading = true;
    this.errorMessage = null;
    this.offset = 0;
    this.fetchInvoices(false);
  }

  redirectTo(url: string): void {
    window.location.href = url;
  }

  private fetchInvoices(append: boolean): void {
    this.apiService
      .listInvoices({ limit: LIMIT, offset: this.offset })
      .subscribe({
        next: (result) => {
          if (append) {
            const existingIds = new Set(this.items.map((i) => i.invoiceId));
            const newItems = result.items.filter((i) => !existingIds.has(i.invoiceId));
            this.items = [...this.items, ...newItems];
          } else {
            this.items = result.items;
          }
          this.isLoading = false;
          this.isLoadingMore = false;
        },
        error: (err) => {
          const mapped = this.errorMapper.mapAndHandle(err);
          this.errorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
          this.isLoading = false;
          this.isLoadingMore = false;
        },
      });
  }
}
