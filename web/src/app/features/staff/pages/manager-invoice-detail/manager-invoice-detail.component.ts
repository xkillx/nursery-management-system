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
import { ManagerInvoiceDetail, ManagerInvoiceLine } from '../../models/manager-invoices.models';
import { formatGbp, formatMinutes, formatBillingMonthLabel } from '../../utils/invoice-run-formatters';

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

  readonly formatGbp = formatGbp;
  readonly formatMinutes = formatMinutes;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly lineKindLabel = lineKindLabel;

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

  private loadDetail(invoiceId: string): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.apiService.getInvoice(invoiceId).subscribe({
      next: (detail) => {
        this.detail = detail;
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
