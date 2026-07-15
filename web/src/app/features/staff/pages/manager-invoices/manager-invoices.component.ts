import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowDownTray,
  heroCalendarDays,
  heroCheck,
  heroCheckCircle,
  heroChevronLeft,
  heroChevronRight,
  heroClock,
  heroCurrencyPound,
  heroExclamationCircle,
  heroExclamationTriangle,
  heroEye,
  heroFunnel,
  heroMagnifyingGlass,
  heroPencilSquare,
  heroPlus,
  heroReceiptPercent,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ChildAvatarComponent } from '../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { ToastService } from '../../../../shared/services/toast.service';
import {
  ManagerInvoiceStatusFilter,
  ManagerInvoiceListItem,
} from '../../models/manager-invoices.models';
import {
  formatGbp,
  formatBillingMonthLabel,
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

interface RangePreset {
  value: string;
  label: string;
}

interface InvoiceMetric {
  key: string;
  label: string;
  amount: number | null;
  count: number | null;
  tone: 'brand' | 'success' | 'warning' | 'error' | 'neutral';
  pill: string;
}

const RANGE_PRESETS: RangePreset[] = [
  { value: 'this', label: 'This month' },
  { value: '3m', label: 'Last 3 months' },
  { value: '6m', label: 'Last 6 months' },
  { value: 'custom', label: 'Custom' },
];

function formatBillingMonth(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  return `${y}-${m}`;
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
    ChildAvatarComponent,
    ConfirmationDialogComponent,
    NgIcon,
  ],
  templateUrl: './manager-invoices.component.html',
  providers: [
    provideIcons({
      heroArrowDownTray,
      heroCalendarDays,
      heroCheck,
      heroCheckCircle,
      heroChevronLeft,
      heroChevronRight,
      heroClock,
      heroCurrencyPound,
      heroExclamationCircle,
      heroExclamationTriangle,
      heroEye,
      heroFunnel,
      heroMagnifyingGlass,
      heroPencilSquare,
      heroPlus,
      heroReceiptPercent,
      heroXMark,
    }),
  ],
})
export class ManagerInvoicesComponent implements OnInit {
  private readonly apiService = inject(ManagerInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly router = inject(Router);
  private readonly toast = inject(ToastService);

  readonly statusFilters = STATUS_FILTERS;
  readonly rangePresets = RANGE_PRESETS;
  readonly invoicesRoute = ROLE_ROUTES.managerInvoices;

  selectedBillingMonthFrom: string;
  selectedBillingMonthTo: string;
  activePreset = 'this';
  selectedStatus: ManagerInvoiceStatusFilter = 'all';
  searchQuery = '';
  offset = 0;

  items: ManagerInvoiceListItem[] = [];
  isLoading = false;
  errorMessage: string | null = null;

  selectedIds = new Set<string>();

  isConfirmIssueOpen = false;
  isIssuing = false;
  issuingIds = new Set<string>();

  readonly formatGbp = formatGbp;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly invoiceIdentity = invoiceIdentity;

  get filteredItems(): ManagerInvoiceListItem[] {
    const q = this.searchQuery.trim().toLowerCase();
    if (!q) return this.items;
    return this.items.filter(item => {
      const childMatch = item.childName?.toLowerCase().includes(q);
      const numberMatch = item.invoiceNumber?.toLowerCase().includes(q) || 
                          item.invoiceNumberDisplay?.toLowerCase().includes(q);
      return childMatch || numberMatch;
    });
  }

  viewInvoice(invoiceId: string, event: Event): void {
    const target = event.target as HTMLElement;
    if (target.closest('a') || target.closest('button') || target.closest('input') || target.closest('select')) {
      return;
    }
    this.router.navigate(['/manager/invoices', invoiceId]);
  }

  isSelected(invoiceId: string): boolean {
    return this.selectedIds.has(invoiceId);
  }

  get draftItems(): ManagerInvoiceListItem[] {
    return this.filteredItems.filter((i) => i.status === 'draft');
  }

  get isAllDraftsSelected(): boolean {
    const drafts = this.draftItems;
    return drafts.length > 0 && drafts.every((i) => this.selectedIds.has(i.invoiceId));
  }

  toggleRow(invoiceId: string, event: Event): void {
    event.stopPropagation();
    if (this.selectedIds.has(invoiceId)) {
      this.selectedIds.delete(invoiceId);
    } else {
      this.selectedIds.add(invoiceId);
    }
    this.selectedIds = new Set(this.selectedIds);
  }

  toggleAll(): void {
    if (this.isAllDraftsSelected) {
      this.selectedIds = new Set();
    } else {
      this.selectedIds = new Set(this.draftItems.map((i) => i.invoiceId));
    }
  }

  get selectedItems(): ManagerInvoiceListItem[] {
    return this.items.filter((i) => this.selectedIds.has(i.invoiceId));
  }

  get selectedDraftCount(): number {
    return this.selectedItems.filter((i) => i.status === 'draft').length;
  }

  get selectedTotal(): number {
    return this.selectedItems.reduce((sum, i) => sum + i.totalDueMinor, 0);
  }

  openIssueConfirmation(): void {
    this.isConfirmIssueOpen = true;
  }

  cancelIssue(): void {
    this.isConfirmIssueOpen = false;
  }

  clearSelection(): void {
    this.selectedIds = new Set();
  }

  confirmIssue(): void {
    const ids = Array.from(this.selectedIds);
    const billingMonth = this.selectedItems[0]?.billingMonth;
    if (!billingMonth || ids.length === 0) return;

    this.isIssuing = true;
    this.issuingIds = new Set(ids);
    this.isConfirmIssueOpen = false;

    this.apiService.bulkIssueInvoices({ billingMonth, invoiceIds: ids }).subscribe({
      next: (result) => {
        const { successCount, blockedCount } = result.summary;
        if (blockedCount > 0) {
          this.toast.warning(`${successCount} issued, ${blockedCount} blocked`);
        } else {
          this.toast.success(`${successCount} invoice${successCount === 1 ? '' : 's'} issued successfully`);
        }
        this.selectedIds = new Set();
        this.issuingIds = new Set();
        this.isIssuing = false;
        this.loadList();
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.toast.error(formatPresentedApiError(presentApiError(mapped, 'invoice.issue')));
        this.issuingIds = new Set();
        this.isIssuing = false;
      },
    });
  }

  getStatusBorderClass(status: string, dueStatus: string): string {
    if (status === 'draft') return 'border-l-gray-300 dark:border-l-gray-700';
    if (status === 'paid') return 'border-l-success-500';
    if (dueStatus === 'overdue' || status === 'payment_failed') return 'border-l-error-500';
    if (status === 'issued') return 'border-l-brand-500';
    return 'border-l-transparent';
  }

  constructor() {
    const now = new Date();
    this.selectedBillingMonthTo = formatBillingMonth(now);
    this.selectedBillingMonthFrom = formatBillingMonth(now);
  }

  paymentCueState(item: ManagerInvoiceListItem): PaymentDisplayState {
    return getPaymentDisplayState(item.status, item.dueStatus, item.amountPaidMinor, null);
  }

  paymentCueLabel(state: PaymentDisplayState): string {
    return paymentDisplayLabel(state);
  }

  ngOnInit(): void {
    this.loadList();
  }

  onRangePreset(preset: string): void {
    this.activePreset = preset;
    const now = new Date();
    this.selectedBillingMonthTo = formatBillingMonth(now);

    if (preset === 'this') {
      this.selectedBillingMonthFrom = formatBillingMonth(now);
    } else if (preset === '3m') {
      this.selectedBillingMonthFrom = formatBillingMonth(new Date(now.getFullYear(), now.getMonth() - 2, 1));
    } else if (preset === '6m') {
      this.selectedBillingMonthFrom = formatBillingMonth(new Date(now.getFullYear(), now.getMonth() - 5, 1));
    }
    // 'custom' leaves existing from/to values unchanged
    this.offset = 0;
    this.loadList();
  }

  onCustomRangeChange(): void {
    this.activePreset = 'custom';
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

  navigateToNew(): void {
    this.router.navigate(['/manager/invoices/new']);
  }

  get billingMonthLabel(): string {
    if (this.selectedBillingMonthFrom === this.selectedBillingMonthTo) {
      return formatBillingMonthLabel(this.selectedBillingMonthFrom);
    }
    return `${formatBillingMonthLabel(this.selectedBillingMonthFrom)} – ${formatBillingMonthLabel(this.selectedBillingMonthTo)}`;
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
    this.selectedIds = new Set();

    this.apiService
      .listInvoices({
        billingMonthFrom: this.selectedBillingMonthFrom,
        billingMonthTo: this.selectedBillingMonthTo,
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
