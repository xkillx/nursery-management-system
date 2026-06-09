import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import {
  ManagerInvoiceDetail,
  ManagerInvoiceLine,
  ManagerPaymentStatus,
  PaymentEvent,
  PaginatedPaymentEvents,
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
} from '../../utils/manager-payment-formatters';

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

function lineKindLabel(kind: string): string {
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

@Component({
  selector: 'app-manager-invoice-detail',
  imports: [
    CommonModule,
    RouterLink,
    PageHeaderComponent,
    LoadingStateComponent,
    AlertComponent,
    TableShellComponent,
    StatusBadgeComponent,
  ],
  templateUrl: './manager-invoice-detail.component.html',
})
export class ManagerInvoiceDetailComponent implements OnInit {
  private readonly apiService = inject(ManagerInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);

  detail: ManagerInvoiceDetail | null = null;
  isLoading = false;
  errorMessage: string | null = null;

  paymentStatus: ManagerPaymentStatus | null = null;
  paymentEvents: PaymentEvent[] = [];
  isPaymentLoading = false;
  paymentErrorMessage: string | null = null;
  paymentEventsOffset = 0;
  readonly paymentEventsLimit = 50;

  readonly formatGbp = formatGbp;
  readonly formatMinutes = formatMinutes;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly lineKindLabel = lineKindLabel;

  readonly canShowParentRetry = canShowParentRetry;
  readonly getPaymentDisplayState = getPaymentDisplayState;
  readonly paymentDisplayLabel = paymentDisplayLabel;
  readonly attemptStatusLabel = attemptStatusLabel;
  readonly eventOutcomeLabel = eventOutcomeLabel;
  readonly webhookStatusLabel = webhookStatusLabel;
  readonly retryReasonLabel = retryReasonLabel;
  readonly isOpenPaymentAttempt = isOpenPaymentAttempt;

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

  get paymentDisplayStateValue(): string {
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

  previousPaymentEventsPage(): void {
    this.paymentEventsOffset = Math.max(0, this.paymentEventsOffset - this.paymentEventsLimit);
    this.loadPaymentEvents(this.detail!.invoiceId);
  }

  nextPaymentEventsPage(): void {
    this.paymentEventsOffset += this.paymentEventsLimit;
    this.loadPaymentEvents(this.detail!.invoiceId);
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
        this.errorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
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
        this.paymentErrorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
      },
    });

    this.loadPaymentEvents(invoiceId);
  }

  private loadPaymentEvents(invoiceId: string): void {
    this.apiService.listPaymentEvents(invoiceId, { limit: this.paymentEventsLimit, offset: this.paymentEventsOffset }).subscribe({
      next: (result) => {
        this.paymentEvents = result.items;
        this.isPaymentLoading = false;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.paymentErrorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
        this.isPaymentLoading = false;
      },
    });
  }
}
