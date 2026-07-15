import { CommonModule } from '@angular/common';
import { Component, inject, OnInit, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
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
import { Subject, debounceTime, takeUntil } from 'rxjs';

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
  ManagerInvoiceStatus,
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

const STATUS_OPTIONS: { value: ManagerInvoiceStatus; label: string }[] = [
  { value: 'draft', label: 'Draft' },
  { value: 'issued', label: 'Issued' },
  { value: 'payment_failed', label: 'Payment failed' },
  { value: 'overdue', label: 'Overdue' },
  { value: 'paid', label: 'Paid' },
];

const LIMIT = 50;
const LS_KEY = 'nursery.invoice_filters';

interface FilterState {
  statuses: ManagerInvoiceStatus[];
  billingMonthFrom: string;
  billingMonthTo: string;
  activePreset: string;
  q: string;
}

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

const VALID_STATUSES = new Set<string>(['draft', 'issued', 'payment_failed', 'paid', 'overdue']);

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
export class ManagerInvoicesComponent implements OnInit, OnDestroy {
  private readonly apiService = inject(ManagerInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly toast = inject(ToastService);
  private readonly destroy$ = new Subject<void>();
  private readonly searchChanged$ = new Subject<string>();

  readonly statusOptions = STATUS_OPTIONS;
  readonly rangePresets = RANGE_PRESETS;
  readonly invoicesRoute = ROLE_ROUTES.managerInvoices;

  selectedBillingMonthFrom: string;
  selectedBillingMonthTo: string;
  activePreset = 'this';
  selectedStatuses: ManagerInvoiceStatus[] = [];
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

  get draftItems(): ManagerInvoiceListItem[] {
    return this.items.filter((i) => i.status === 'draft');
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
    this.searchChanged$.pipe(debounceTime(300), takeUntil(this.destroy$)).subscribe(() => {
      this.offset = 0;
      this.loadList();
    });

    this.route.queryParams.pipe(takeUntil(this.destroy$)).subscribe((params) => {
      const hasUrlParams = Object.keys(params).length > 0;

      if (hasUrlParams) {
        this.applyQueryParams(params);
      } else {
        this.restoreFromLocalStorage();
      }
      this.loadList();
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
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
    this.offset = 0;
    this.onFilterChange();
  }

  onCustomRangeChange(): void {
    this.activePreset = 'custom';
    this.offset = 0;
    this.onFilterChange();
  }

  onStatusToggle(status: ManagerInvoiceStatus): void {
    const idx = this.selectedStatuses.indexOf(status);
    if (idx >= 0) {
      this.selectedStatuses = this.selectedStatuses.filter((s) => s !== status);
    } else {
      this.selectedStatuses = [...this.selectedStatuses, status];
    }
    this.offset = 0;
    this.onFilterChange();
  }

  isStatusSelected(status: ManagerInvoiceStatus): boolean {
    return this.selectedStatuses.includes(status);
  }

  onSearchInput(value: string): void {
    this.searchQuery = value;
    this.searchChanged$.next(value);
  }

  clearSearch(): void {
    this.searchQuery = '';
    this.offset = 0;
    this.onFilterChange();
  }

  clearAllFilters(): void {
    const now = new Date();
    this.selectedStatuses = [];
    this.activePreset = 'this';
    this.selectedBillingMonthFrom = formatBillingMonth(now);
    this.selectedBillingMonthTo = formatBillingMonth(now);
    this.searchQuery = '';
    this.offset = 0;
    localStorage.removeItem(LS_KEY);
    this.router.navigate([], { queryParams: {} });
    this.loadList();
  }

  get hasActiveFilters(): boolean {
    const now = new Date();
    const defaultFrom = formatBillingMonth(now);
    const defaultTo = formatBillingMonth(now);
    return (
      this.selectedStatuses.length > 0 ||
      this.searchQuery.trim() !== '' ||
      this.activePreset !== 'this' ||
      this.selectedBillingMonthFrom !== defaultFrom ||
      this.selectedBillingMonthTo !== defaultTo
    );
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

  daysOverdue(dueAt: string | null): number | null {
    if (!dueAt) return null;
    const due = new Date(dueAt);
    const now = new Date();
    const diffMs = now.getTime() - due.getTime();
    if (diffMs <= 0) return null;
    return Math.ceil(diffMs / (1000 * 60 * 60 * 24));
  }

  attemptStatusLabel(status: string | null): string {
    if (!status) return '';
    const labels: Record<string, string> = {
      pending: 'Pending',
      processing: 'Processing',
      succeeded: 'Succeeded',
      failed: 'Failed',
      cancelled: 'Cancelled',
      expired: 'Expired',
    };
    return labels[status] ?? status;
  }

  attemptTooltipText(item: ManagerInvoiceListItem): string | null {
    if (!item.latestPaymentAttemptStatus) return null;
    const parts: string[] = [];
    parts.push(this.attemptStatusLabel(item.latestPaymentAttemptStatus));
    if (item.latestPaymentAttemptCreatedAt) {
      parts.push(formatInstant(item.latestPaymentAttemptCreatedAt));
    }
    return parts.join(' · ');
  }

  previousPage(): void {
    this.offset = Math.max(0, this.offset - LIMIT);
    this.loadList();
  }

  nextPage(): void {
    this.offset += LIMIT;
    this.loadList();
  }

  private onFilterChange(): void {
    this.saveToLocalStorage();
    this.syncUrlParams();
    this.loadList();
  }

  private applyQueryParams(params: Record<string, string>): void {
    if (params['status']) {
      const statuses = params['status']
        .split(',')
        .filter((s): s is ManagerInvoiceStatus => VALID_STATUSES.has(s));
      this.selectedStatuses = statuses;
    } else {
      this.selectedStatuses = [];
    }

    if (params['billing_month_from']) {
      this.selectedBillingMonthFrom = params['billing_month_from'];
    }
    if (params['billing_month_to']) {
      this.selectedBillingMonthTo = params['billing_month_to'];
    }

    if (params['billing_month_from'] && params['billing_month_to']) {
      this.activePreset = 'custom';
    }

    if (params['q']) {
      this.searchQuery = params['q'];
    }

    if (params['offset']) {
      const o = parseInt(params['offset'], 10);
      if (!isNaN(o) && o >= 0) {
        this.offset = o;
      }
    }
  }

  private restoreFromLocalStorage(): void {
    try {
      const raw = localStorage.getItem(LS_KEY);
      if (!raw) return;
      const state: FilterState = JSON.parse(raw);

      if (state.statuses && Array.isArray(state.statuses)) {
        this.selectedStatuses = state.statuses.filter((s) => VALID_STATUSES.has(s));
      }
      if (state.billingMonthFrom) {
        this.selectedBillingMonthFrom = state.billingMonthFrom;
      }
      if (state.billingMonthTo) {
        this.selectedBillingMonthTo = state.billingMonthTo;
      }
      if (state.activePreset) {
        this.activePreset = state.activePreset;
      }
      if (state.q) {
        this.searchQuery = state.q;
      }
    } catch {
      // corrupted localStorage — ignore
    }
  }

  private saveToLocalStorage(): void {
    const state: FilterState = {
      statuses: this.selectedStatuses,
      billingMonthFrom: this.selectedBillingMonthFrom,
      billingMonthTo: this.selectedBillingMonthTo,
      activePreset: this.activePreset,
      q: this.searchQuery,
    };
    try {
      localStorage.setItem(LS_KEY, JSON.stringify(state));
    } catch {
      // localStorage full or unavailable — ignore
    }
  }

  private syncUrlParams(): void {
    const queryParams: Record<string, string> = {};

    if (this.selectedStatuses.length > 0) {
      queryParams['status'] = this.selectedStatuses.join(',');
    }
    if (this.selectedBillingMonthFrom) {
      queryParams['billing_month_from'] = this.selectedBillingMonthFrom;
    }
    if (this.selectedBillingMonthTo) {
      queryParams['billing_month_to'] = this.selectedBillingMonthTo;
    }
    if (this.searchQuery.trim()) {
      queryParams['q'] = this.searchQuery.trim();
    }
    if (this.offset > 0) {
      queryParams['offset'] = String(this.offset);
    }

    this.router.navigate([], { queryParams, queryParamsHandling: 'merge' });
  }

  private loadList(): void {
    this.isLoading = true;
    this.errorMessage = null;
    this.selectedIds = new Set();

    this.apiService
      .listInvoices({
        billingMonthFrom: this.selectedBillingMonthFrom,
        billingMonthTo: this.selectedBillingMonthTo,
        status: this.selectedStatuses.length > 0 ? this.selectedStatuses.join(',') : undefined,
        q: this.searchQuery.trim() || undefined,
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
