import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowDownTray,
  heroCheckCircle,
  heroClock,
  heroCurrencyPound,
  heroExclamationCircle,
  heroExclamationTriangle,
  heroEye,
  heroFunnel,
  heroMagnifyingGlass,
  heroPencilSquare,
  heroReceiptPercent,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
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

interface InvoiceMetric {
  key: string;
  label: string;
  amount: number | null;
  count: number | null;
  tone: 'brand' | 'success' | 'warning' | 'error' | 'neutral';
  pill: string;
}

@Component({
  selector: 'app-manager-invoices',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    EmptyStateComponent,
    LoadingStateComponent,
    AlertComponent,
    StatusBadgeComponent,
    NgIcon,
  ],
  templateUrl: './manager-invoices.component.html',
  providers: [
    provideIcons({
      heroArrowDownTray,
      heroCheckCircle,
      heroClock,
      heroCurrencyPound,
      heroExclamationCircle,
      heroExclamationTriangle,
      heroEye,
      heroFunnel,
      heroMagnifyingGlass,
      heroPencilSquare,
      heroReceiptPercent,
    }),
  ],
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

  get billingMonthLabel(): string {
    return formatBillingMonthLabel(this.selectedBillingMonth);
  }

  get hasInvoices(): boolean {
    return this.items.length > 0;
  }

  get currentPage(): number {
    return Math.floor(this.offset / LIMIT) + 1;
  }

  get metricCards(): InvoiceMetric[] {
    const issued = this.items.filter((i) => i.status === 'issued');
    const paid = this.items.filter((i) => i.status === 'paid');
    const overdue = this.items.filter((i) => i.dueStatus === 'overdue');
    const drafts = this.items.filter((i) => i.status === 'draft');

    const outstanding = issued
      .filter((i) => i.amountPaidMinor < i.totalDueMinor)
      .reduce((sum, i) => sum + Math.max(i.totalDueMinor - i.amountPaidMinor, 0), 0);
    const paidTotal = paid.reduce((sum, i) => sum + i.totalDueMinor, 0);
    const overdueTotal = overdue.reduce((sum, i) => sum + Math.max(i.totalDueMinor - i.amountPaidMinor, 0), 0);
    const draftTotal = drafts.reduce((sum, i) => sum + i.totalDueMinor, 0);

    return [
      {
        key: 'outstanding',
        label: 'Total outstanding',
        amount: outstanding,
        count: null,
        tone: 'brand',
        pill: 'Issued & overdue',
      },
      {
        key: 'paid',
        label: 'Paid this month',
        amount: paidTotal,
        count: paid.length,
        tone: 'success',
        pill: 'Settled invoices',
      },
      {
        key: 'overdue',
        label: 'Overdue invoices',
        amount: overdueTotal,
        count: overdue.length,
        tone: 'error',
        pill: 'Action needed',
      },
      {
        key: 'drafts',
        label: 'Draft invoices',
        amount: draftTotal,
        count: drafts.length,
        tone: 'neutral',
        pill: 'Awaiting issue',
      },
    ];
  }

  metricIconToneClasses(tone: InvoiceMetric['tone']): string {
    const map: Record<InvoiceMetric['tone'], string> = {
      brand: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300',
      success: 'bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-300',
      warning: 'bg-warning-50 text-warning-600 dark:bg-warning-500/15 dark:text-warning-300',
      error: 'bg-error-50 text-error-600 dark:bg-error-500/15 dark:text-error-300',
      neutral: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300',
    };
    return map[tone];
  }

  metricPillToneClasses(tone: InvoiceMetric['tone']): string {
    const map: Record<InvoiceMetric['tone'], string> = {
      brand: 'text-brand-600 dark:text-brand-300',
      success: 'text-success-600 dark:text-success-300',
      warning: 'text-warning-700 dark:text-warning-300',
      error: 'text-error-600 dark:text-error-300',
      neutral: 'text-gray-600 dark:text-gray-300',
    };
    return map[tone];
  }

  metricProgressToneClasses(tone: InvoiceMetric['tone']): string {
    const map: Record<InvoiceMetric['tone'], string> = {
      brand: 'bg-brand-500',
      success: 'bg-success-500',
      warning: 'bg-warning-500',
      error: 'bg-error-500',
      neutral: 'bg-gray-400',
    };
    return map[tone];
  }

  metricProgressWidth(amount: number | null): number {
    if (amount === null) return 0;
    const all = this.metricCards.reduce((sum, m) => sum + (m.amount ?? 0), 0);
    if (all <= 0) return 0;
    return Math.max(6, Math.min(100, Math.round((amount / all) * 100)));
  }

  metricDisplayValue(metric: InvoiceMetric): string {
    if (metric.amount === null) return '—';
    return formatGbp(metric.amount);
  }

  paymentCueRowLabel(item: ManagerInvoiceListItem): string {
    return paymentDisplayLabel(this.paymentCueState(item));
  }

  previousPage(): void {
    this.offset = Math.max(0, this.offset - LIMIT);
    this.loadList();
  }

  nextPage(): void {
    this.offset += LIMIT;
    this.loadList();
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
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'invoice.managerList'));
          this.isLoading = false;
        },
      });
  }
}
