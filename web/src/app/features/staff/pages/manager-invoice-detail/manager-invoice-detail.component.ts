import { CommonModule } from '@angular/common';
import { Component, inject, OnInit, AfterViewInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowDownTray,
  heroBanknotes,
  heroBuildingOffice2,
  heroCalendarDays,
  heroCheckBadge,
  heroChevronDown,
  heroChevronRight,
  heroClock,
  heroCreditCard,
  heroEnvelope,
  heroExclamationCircle,
  heroExclamationTriangle,
  heroEye,
  heroInformationCircle,
  heroLockClosed,
  heroPencil,
  heroPlus,
  heroReceiptPercent,
  heroShieldCheck,
  heroTrash,
  heroUserCircle,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import {
  ManagerInvoiceDetail,
  ManagerInvoiceLine,
  ManagerPaymentStatus,
  PaymentEvent,
  PaginatedPaymentEvents,
  AddInvoiceLineInput,
  UpdateInvoiceLineInput,
} from '../../models/manager-invoices.models';
import { formatGbp, formatMinutes, formatBillingMonthLabel } from '../../utils/invoice-run-formatters';
import {
  canShowParentRetry,
  getPaymentDisplayState,
  paymentDisplayLabel,
  attemptStatusLabel,
  eventOutcomeLabel,
  webhookStatusLabel,
  retryReasonLabel,
  isOpenPaymentAttempt,
  PaymentDisplayState,
} from '../../utils/manager-payment-formatters';
import { ToastService } from '../../../../shared/services/toast.service';

const IMMUTABLE_STATUSES = new Set(['issued', 'payment_failed', 'paid', 'overdue']);

function formatInstant(iso: string | null): string {
  if (!iso) return '';
  const d = new Date(iso);
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/London',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(d);
}

function formatDate(iso: string | null): string {
  if (!iso) return '';
  const d = new Date(iso);
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: 'Europe/London',
    dateStyle: 'long',
  }).format(d);
}

function lineKindLabel(kind: string): string {
  if (kind === 'ad_hoc') return 'Ad-hoc session';
  return kind
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function invoiceDisplayTitle(detail: ManagerInvoiceDetail): string {
  if (detail.invoiceNumberDisplay) return detail.invoiceNumberDisplay;
  if (detail.invoiceNumber) return detail.invoiceNumber;
  return `Draft invoice — ${detail.childName}`;
}

function lineQuantityLabel(line: ManagerInvoiceLine): string {
  if (line.lineKind === 'funded_deduction') {
    if (line.fundedAllowanceMinutes !== null) {
      return `${formatMinutes(line.fundedAllowanceMinutes)} allowance`;
    }
    if (line.quantityMinutes !== null) {
      return formatMinutes(line.quantityMinutes);
    }
    return '—';
  }
  if (line.sessionCount !== null && line.sessionCount > 0) {
    return `${line.sessionCount} session${line.sessionCount === 1 ? '' : 's'}`;
  }
  if (line.quantityMinutes !== null) {
    return formatMinutes(line.quantityMinutes);
  }
  return '—';
}

function lineKindTone(kind: string): 'primary' | 'accent' | 'success' | 'neutral' {
  if (kind === 'funded_deduction') return 'accent';
  if (kind === 'core_childcare' || kind === 'core_charge') return 'primary';
  if (kind === 'ad_hoc') return 'success';
  if (kind.startsWith('extra_')) return 'success';
  return 'neutral';
}

function fundingModelLabel(model: string | null): string {
  if (model === 'term_time_only') return 'Term-time funding';
  if (model === 'stretched') return 'Stretched funding';
  return 'Funded hours';
}

interface AuditTrailEntry {
  key: string;
  icon: string;
  tone: 'success' | 'primary' | 'warning' | 'error' | 'neutral';
  title: string;
  description: string;
  timestamp: string | null;
}

@Component({
  selector: 'app-manager-invoice-detail',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    LoadingStateComponent,
    AlertComponent,
    StatusBadgeComponent,
    NgIcon,
  ],
  templateUrl: './manager-invoice-detail.component.html',
  providers: [
    provideIcons({
      heroArrowDownTray,
      heroBanknotes,
      heroBuildingOffice2,
      heroCalendarDays,
      heroCheckBadge,
      heroChevronDown,
      heroChevronRight,
      heroClock,
      heroCreditCard,
      heroEnvelope,
      heroExclamationCircle,
      heroExclamationTriangle,
      heroEye,
      heroInformationCircle,
      heroLockClosed,
      heroPencil,
      heroPlus,
      heroReceiptPercent,
      heroShieldCheck,
      heroTrash,
      heroUserCircle,
    }),
  ],
  styles: [
    `
    @keyframes fadeInUp {
      from { opacity: 0; transform: translateY(10px); }
      to   { opacity: 1; transform: translateY(0); }
    }
    .animate-total { animation: fadeInUp 0.6s cubic-bezier(0.16, 1, 0.3, 1) both; }
  `,
  ],
})
export class ManagerInvoiceDetailComponent implements OnInit, AfterViewInit {
  private readonly apiService = inject(ManagerInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly toast = inject(ToastService);

  detail: ManagerInvoiceDetail | null = null;
  isLoading = false;
  errorMessage: string | null = null;

  paymentStatus: ManagerPaymentStatus | null = null;
  paymentEvents: PaymentEvent[] = [];
  isPaymentLoading = false;
  paymentErrorMessage: string | null = null;
  paymentEventsOffset = 0;
  readonly paymentEventsLimit = 50;

  isPaymentReviewCollapsed = true;

  paymentLinkUrl: string | null = null;
  isPaymentLinkLoading = false;
  isDownloadingPdf = false;

  readonly formatGbp = formatGbp;
  readonly formatMinutes = formatMinutes;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly formatDate = formatDate;
  readonly lineKindLabel = lineKindLabel;
  readonly lineQuantityLabel = lineQuantityLabel;
  readonly lineKindTone = lineKindTone;
  readonly fundingModelLabel = fundingModelLabel;

  readonly canShowParentRetry = canShowParentRetry;
  readonly getPaymentDisplayState = getPaymentDisplayState;
  readonly paymentDisplayLabel = paymentDisplayLabel;
  readonly attemptStatusLabel = attemptStatusLabel;
  readonly eventOutcomeLabel = eventOutcomeLabel;
  readonly webhookStatusLabel = webhookStatusLabel;
  readonly retryReasonLabel = retryReasonLabel;
  readonly isOpenPaymentAttempt = isOpenPaymentAttempt;

  totalAnimated = false;

  editingLineId: string | null = null;
  editForm = { description: '', quantityMinutes: 0, unitAmountMinor: 0, lineAmountMinor: 0 };
  isAddingLine = false;
  addForm: AddInvoiceLineInput = { lineKind: 'extra', description: '', quantityMinutes: 0, unitAmountMinor: 0, lineAmountMinor: 0 };
  isSavingLine = false;

  ngOnInit(): void {
    const invoiceId = this.route.snapshot.paramMap.get('invoiceId');
    if (!invoiceId) {
      this.errorMessage = 'Invoice ID is missing.';
      return;
    }
    this.loadDetail(invoiceId);
  }

  get isImmutable(): boolean {
    if (!this.detail) return false;
    return IMMUTABLE_STATUSES.has(this.detail.status) || !!this.detail.lockedAt;
  }

  get displayTitle(): string {
    if (!this.detail) return '';
    return invoiceDisplayTitle(this.detail);
  }

  get balanceDueMinor(): number {
    if (!this.detail) return 0;
    return this.detail.totalDueMinor - this.detail.amountPaidMinor;
  }

  get isPayable(): boolean {
    if (!this.detail) return false;
    return this.detail.status !== 'draft';
  }

  get paymentBalanceDueMinor(): number {
    if (!this.paymentStatus) return 0;
    return this.paymentStatus.totalDueMinor - this.paymentStatus.amountPaidMinor;
  }

  get paymentDisplayStateValue(): PaymentDisplayState {
    if (!this.detail) return 'not_issued';
    return getPaymentDisplayState(
      this.detail.status,
      this.detail.dueStatus,
      this.detail.amountPaidMinor,
      this.paymentStatus?.latestPaymentAttempt?.status ?? null,
    );
  }

  get hasPaymentEventsPrevious(): boolean {
    return this.paymentEventsOffset > 0;
  }

  get hasPaymentEventsNext(): boolean {
    return this.paymentEvents.length === this.paymentEventsLimit;
  }

  get periodLabel(): string {
    if (!this.detail?.period) return '';
    const start = formatDate(this.detail.period.startDate);
    const end = formatDate(this.detail.period.endDate);
    if (!start || !end) return '';
    return `${start} — ${end}`;
  }

  get parentName(): string {
    if (!this.detail) return '';
    if (this.detail.parentContact?.fullName) return this.detail.parentContact.fullName;
    return this.detail.childName;
  }

  get parentAccountId(): string {
    if (!this.detail) return '';
    return this.detail.childId;
  }

  get roomLabel(): string {
    if (!this.detail?.roomName) return '';
    return this.detail.roomName;
  }

  get auditTrail(): AuditTrailEntry[] {
    if (!this.detail) return [];
    const entries: AuditTrailEntry[] = [];

    entries.push({
      key: 'generated',
      icon: 'heroReceiptPercent',
      tone: 'neutral',
      title: 'Invoice generated',
      description: 'Draft prepared by automatic billing',
      timestamp: this.detail.createdAt,
    });

    if (this.detail.generatedRunStartedAt) {
      entries.push({
        key: 'run-started',
        icon: 'heroClock',
        tone: 'primary',
        title: 'Invoice run started',
        description: this.detail.generatedRunStatus
          ? `Run status: ${this.detail.generatedRunStatus}`
          : 'Invoice run started',
        timestamp: this.detail.generatedRunStartedAt,
      });
    }

    if (this.detail.generatedRunCompletedAt) {
      entries.push({
        key: 'run-completed',
        icon: 'heroShieldCheck',
        tone: 'success',
        title: 'Invoice run completed',
        description: this.detail.generatedRunExceptionCount
          ? `Completed with ${this.detail.generatedRunExceptionCount} exception${this.detail.generatedRunExceptionCount === 1 ? '' : 's'}`
          : 'Completed without exceptions',
        timestamp: this.detail.generatedRunCompletedAt,
      });
    }

    if (this.detail.issuedAt) {
      entries.push({
        key: 'issued',
        icon: 'heroCheckBadge',
        tone: 'success',
        title: 'Invoice issued',
        description: 'Sent to the parent and added to the billing cycle',
        timestamp: this.detail.issuedAt,
      });
    }

    if (this.detail.lockedAt && this.detail.lockedAt !== this.detail.issuedAt) {
      entries.push({
        key: 'locked',
        icon: 'heroLockClosed',
        tone: 'neutral',
        title: 'Invoice locked',
        description: 'No further edits — adjustments must use a credit note',
        timestamp: this.detail.lockedAt,
      });
    }

    if (this.detail.paymentFailedAt) {
      entries.push({
        key: 'payment-failed',
        icon: 'heroExclamationCircle',
        tone: 'error',
        title: 'Payment attempt failed',
        description: 'Awaiting retry from the parent portal',
        timestamp: this.detail.paymentFailedAt,
      });
    }

    if (this.detail.paidAt) {
      entries.push({
        key: 'paid',
        icon: 'heroCheckBadge',
        tone: 'success',
        title: 'Payment received',
        description: `Settled for ${formatGbp(this.detail.amountPaidMinor)}`,
        timestamp: this.detail.paidAt,
      });
    }

    return entries;
  }

  auditTrailIconClasses(tone: AuditTrailEntry['tone']): string {
    const map: Record<AuditTrailEntry['tone'], string> = {
      success: 'bg-success-500 text-white',
      primary: 'bg-brand-500 text-white',
      warning: 'bg-warning-500 text-white',
      error: 'bg-error-500 text-white',
      neutral: 'bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-200',
    };
    return map[tone];
  }

  lineToneClasses(kind: string): string {
    const tone = lineKindTone(kind);
    const map: Record<ReturnType<typeof lineKindTone>, string> = {
      primary: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300',
      accent: 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-300',
      success: 'bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-300',
      neutral: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300',
    };
    return map[tone];
  }

  paymentStateBadgeClasses(state: PaymentDisplayState): string {
    switch (state) {
      case 'paid':
        return 'bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-300';
      case 'payment_failed':
        return 'bg-error-50 text-error-600 dark:bg-error-500/15 dark:text-error-300';
      case 'unpaid':
      case 'unpaid_overdue':
        return 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-300';
      case 'awaiting_provider_update':
        return 'bg-blue-light-50 text-blue-light-600 dark:bg-blue-light-500/15 dark:text-blue-light-300';
      default:
        return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300';
    }
  }

  ngAfterViewInit(): void {
    requestAnimationFrame(() => {
      this.totalAnimated = true;
    });
  }

  togglePaymentReview(): void {
    this.isPaymentReviewCollapsed = !this.isPaymentReviewCollapsed;
  }

  generatePaymentLink(): void {
    if (!this.detail || this.isPaymentLinkLoading) return;
    this.isPaymentLinkLoading = true;

    this.apiService.createPaymentLink(this.detail.invoiceId).subscribe({
      next: (result) => {
        this.paymentLinkUrl = result.url;
        this.isPaymentLinkLoading = false;
        if (result.existing) {
          this.toast.success('Payment link copied.');
        } else {
          this.toast.success('Payment link generated.');
        }
        this.copyPaymentLink();
      },
      error: () => {
        this.isPaymentLinkLoading = false;
        this.toast.error('Failed to generate payment link. Please try again.');
      },
    });
  }

  copyPaymentLink(): void {
    if (!this.paymentLinkUrl) return;
    navigator.clipboard.writeText(this.paymentLinkUrl).then(() => {
      this.toast.success('Link copied to clipboard.');
    });
  }

  previousPaymentEventsPage(): void {
    this.paymentEventsOffset = Math.max(0, this.paymentEventsOffset - this.paymentEventsLimit);
    this.loadPaymentEvents(this.detail!.invoiceId);
  }

  nextPaymentEventsPage(): void {
    this.paymentEventsOffset += this.paymentEventsLimit;
    this.loadPaymentEvents(this.detail!.invoiceId);
  }

  downloadPdf(): void {
    if (!this.detail || this.isDownloadingPdf) return;
    this.isDownloadingPdf = true;

    this.apiService.downloadPdf(this.detail.invoiceId).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = this.detail!.invoiceNumber ? `INV-${this.detail!.invoiceNumber}.pdf` : `INV-${this.detail!.invoiceId}.pdf`;
        a.click();
        URL.revokeObjectURL(url);
        this.isDownloadingPdf = false;
      },
      error: () => {
        this.isDownloadingPdf = false;
        this.toast.error('Failed to download PDF. Please try again.');
      },
    });
  }

  isEditableLine(line: ManagerInvoiceLine): boolean {
    return line.lineKind === 'extra' || line.lineKind === 'ad_hoc';
  }

  startEditLine(line: ManagerInvoiceLine): void {
    this.editingLineId = line.lineId;
    this.editForm = {
      description: line.description,
      quantityMinutes: line.quantityMinutes ?? 0,
      unitAmountMinor: line.unitAmountMinor ?? 0,
      lineAmountMinor: line.lineAmountMinor,
    };
  }

  cancelEditLine(): void {
    this.editingLineId = null;
  }

  saveEditLine(lineId: string): void {
    if (!this.detail || this.isSavingLine) return;
    this.isSavingLine = true;

    const input: UpdateInvoiceLineInput = {
      description: this.editForm.description,
      quantityMinutes: this.editForm.quantityMinutes,
      unitAmountMinor: this.editForm.unitAmountMinor,
      lineAmountMinor: this.editForm.lineAmountMinor,
    };

    this.apiService.updateLine(this.detail.invoiceId, lineId, input).subscribe({
      next: () => {
        this.editingLineId = null;
        this.isSavingLine = false;
        this.toast.success('Line updated.');
        this.reloadDetail();
      },
      error: () => {
        this.isSavingLine = false;
        this.toast.error('Failed to update line. Please try again.');
      },
    });
  }

  deleteLine(lineId: string): void {
    if (!this.detail || this.isSavingLine) return;
    this.isSavingLine = true;

    this.apiService.deleteLine(this.detail.invoiceId, lineId).subscribe({
      next: () => {
        this.isSavingLine = false;
        this.toast.success('Line deleted.');
        this.reloadDetail();
      },
      error: () => {
        this.isSavingLine = false;
        this.toast.error('Failed to delete line. Please try again.');
      },
    });
  }

  startAddLine(): void {
    this.isAddingLine = true;
    this.addForm = { lineKind: 'extra', description: '', quantityMinutes: 0, unitAmountMinor: 0, lineAmountMinor: 0 };
  }

  cancelAddLine(): void {
    this.isAddingLine = false;
  }

  saveAddLine(): void {
    if (!this.detail || this.isSavingLine) return;
    this.isSavingLine = true;

    this.apiService.addLine(this.detail.invoiceId, this.addForm).subscribe({
      next: () => {
        this.isAddingLine = false;
        this.isSavingLine = false;
        this.toast.success('Line added.');
        this.reloadDetail();
      },
      error: () => {
        this.isSavingLine = false;
        this.toast.error('Failed to add line. Please try again.');
      },
    });
  }

  private reloadDetail(): void {
    if (!this.detail) return;
    this.apiService.getInvoice(this.detail.invoiceId).subscribe({
      next: (detail) => {
        this.detail = detail;
      },
    });
  }

  private loadDetail(invoiceId: string): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.apiService.getInvoice(invoiceId).subscribe({
      next: (detail) => {
        this.detail = detail;
        this.isLoading = false;
        this.loadPaymentDiagnostics(invoiceId);
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'invoice.managerDetail'));
        this.isLoading = false;
      },
    });
  }

  private loadPaymentDiagnostics(invoiceId: string): void {
    if (!this.isPayable) return;

    this.isPaymentLoading = true;
    this.paymentErrorMessage = null;

    this.apiService.getPaymentStatus(invoiceId).subscribe({
      next: (status) => {
        this.paymentStatus = status;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.paymentErrorMessage = formatPresentedApiError(presentApiError(mapped, 'payment.managerDiagnostics'));
      },
    });

    this.loadPaymentEvents(invoiceId);
  }

  private loadPaymentEvents(invoiceId: string): void {
    this.apiService.listPaymentEvents(invoiceId, { limit: this.paymentEventsLimit, offset: this.paymentEventsOffset }).subscribe({
      next: (result: PaginatedPaymentEvents) => {
        this.paymentEvents = result.items;
        this.isPaymentLoading = false;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.paymentErrorMessage = formatPresentedApiError(presentApiError(mapped, 'payment.managerDiagnostics'));
        this.isPaymentLoading = false;
      },
    });
  }
}
