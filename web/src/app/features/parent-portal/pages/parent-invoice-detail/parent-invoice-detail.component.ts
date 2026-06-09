import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ParentInvoicesApiService } from '../../data/parent-invoices-api.service';
import { ParentInvoiceDetail } from '../../models/parent-invoices.models';
import {
  formatGbp,
  formatBillingMonthLabel,
  formatInstant,
  formatMinutes,
  lineKindLabel,
  balanceDueMinor,
  isPayableInvoice,
} from '../../utils/parent-invoice-formatters';

@Component({
  selector: 'app-parent-invoice-detail',
  imports: [
    CommonModule,
    RouterLink,
    PageHeaderComponent,
    LoadingStateComponent,
    AlertComponent,
    TableShellComponent,
    StatusBadgeComponent,
  ],
  templateUrl: './parent-invoice-detail.component.html',
})
export class ParentInvoiceDetailComponent implements OnInit {
  private readonly apiService = inject(ParentInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);

  detail: ParentInvoiceDetail | null = null;
  isLoading = false;
  errorMessage: string | null = null;
  isPaying = false;

  readonly formatGbp = formatGbp;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly formatMinutes = formatMinutes;
  readonly lineKindLabel = lineKindLabel;
  readonly balanceDueMinor = balanceDueMinor;

  ngOnInit(): void {
    const invoiceId = this.route.snapshot.paramMap.get('invoiceId');
    if (!invoiceId) {
      this.errorMessage = 'Invoice ID is missing.';
      return;
    }
    this.loadDetail(invoiceId);
  }

  get displayTitle(): string {
    if (!this.detail) return '';
    if (this.detail.invoiceNumberDisplay) return this.detail.invoiceNumberDisplay;
    if (this.detail.invoiceNumber) return this.detail.invoiceNumber;
    return `Invoice — ${this.detail.childName}`;
  }

  get isPayable(): boolean {
    if (!this.detail) return false;
    return isPayableInvoice(this.detail);
  }

  get balanceDue(): number {
    if (!this.detail) return 0;
    return balanceDueMinor(this.detail);
  }

  startPayment(): void {
    if (!this.detail || this.isPaying) return;
    this.isPaying = true;

    this.apiService.createCheckoutSession(this.detail.invoiceId).subscribe({
      next: (result) => {
        this.redirectTo(result.checkoutUrl);
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = mapped.message + (mapped.requestId ? ` (Request: ${mapped.requestId})` : '');
        this.isPaying = false;
      },
    });
  }

  redirectTo(url: string): void {
    window.location.href = url;
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
