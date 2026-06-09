import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import {
  ManagerInvoiceStatusFilter,
  ManagerInvoiceListItem,
} from '../../models/manager-invoices.models';
import {
  formatGbp,
  formatBillingMonthLabel,
  defaultCompletedBillingMonth,
} from '../../utils/invoice-run-formatters';
import {
  PaymentDisplayState,
  getPaymentDisplayState,
  paymentDisplayLabel,
} from '../../utils/manager-payment-formatters';

const STATUS_FILTERS: { value: ManagerInvoiceStatusFilter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'draft', label: 'Draft' },
  { value: 'issued', label: 'Issued' },
  { value: 'payment_failed', label: 'Payment failed' },
  { value: 'overdue', label: 'Overdue' },
  { value: 'paid', label: 'Paid' },
];

const LIMIT = 50;

function formatInstant(iso: string | null): string {
  if (!iso) return '';
  const d = new Date(iso);
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/London',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(d);
}

function invoiceIdentity(item: ManagerInvoiceListItem): string {
  if (item.invoiceNumberDisplay) return item.invoiceNumberDisplay;
  if (item.invoiceNumber) return item.invoiceNumber;
  return `Draft invoice — ${item.childName}`;
}

@Component({
  selector: 'app-manager-invoices',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    PageHeaderComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    AlertComponent,
    TableShellComponent,
    StatusBadgeComponent,
  ],
  templateUrl: './manager-invoices.component.html',
})
export class ManagerInvoicesComponent implements OnInit {
  private readonly apiService = inject(ManagerInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  readonly statusFilters = STATUS_FILTERS;
  readonly invoicesRoute = ROLE_ROUTES.managerInvoices;

  selectedBillingMonth = defaultCompletedBillingMonth();
  selectedStatus: ManagerInvoiceStatusFilter = 'all';
  offset = 0;

  items: ManagerInvoiceListItem[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  readonly formatGbp = formatGbp;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly invoiceIdentity = invoiceIdentity;

  paymentCueState(item: ManagerInvoiceListItem): PaymentDisplayState {
    return getPaymentDisplayState(item.status, item.dueStatus, item.amountPaidMinor, null);
  }

  paymentCueLabel(state: PaymentDisplayState): string {
    return paymentDisplayLabel(state);
  }

  ngOnInit(): void {
    this.loadList();
  }

  onMonthChange(month: string): void {
    this.selectedBillingMonth = month;
    this.offset = 0;
    this.loadList();
  }

  onStatusChange(status: ManagerInvoiceStatusFilter): void {
    this.selectedStatus = status;
    this.offset = 0;
    this.loadList();
  }

  get hasPrevious(): boolean {
    return this.offset > 0;
  }

  get hasNext(): boolean {
    return this.items.length === LIMIT;
  }

  previousPage(): void {
    this.offset = Math.max(0, this.offset - LIMIT);
    this.loadList();
  }

  nextPage(): void {
    this.offset += LIMIT;
    this.loadList();
  }

  get billingMonthLabel(): string {
    return formatBillingMonthLabel(this.selectedBillingMonth);
  }

  private loadList(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.apiService
      .listInvoices({
        billingMonth: this.selectedBillingMonth,
        status: this.selectedStatus,
        limit: LIMIT,
        offset: this.offset,
      })
      .subscribe({
        next: (result) => {
          this.items = result.items;
          this.isLoading = false;
        },
        error: (err) => {
          const mapped = this.errorMapper.mapAndHandle(err);
          this.errorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
          this.isLoading = false;
        },
      });
  }
}
