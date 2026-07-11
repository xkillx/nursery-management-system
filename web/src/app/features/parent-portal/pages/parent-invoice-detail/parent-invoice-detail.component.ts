import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Subject, Subscription, timer } from 'rxjs';
import { switchMap, takeUntil, takeWhile } from 'rxjs/operators';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroArrowDownTray } from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ParentInvoicesApiService } from '../../data/parent-invoices-api.service';
import { ParentInvoiceDetail, ParentInvoiceStatus } from '../../models/parent-invoices.models';
import {
  formatGbp,
  formatBillingMonthLabel,
  formatInstant,
  formatMinutes,
  lineKindLabel,
  balanceDueMinor,
  isPayableInvoice,
} from '../../utils/parent-invoice-formatters';
import { ToastService } from '../../../../shared/services/toast.service';

export type PaymentReturnKind = 'none' | 'success' | 'cancelled';
export type PaymentReturnDisplayState = 'paid' | 'failed' | 'cancelled' | 'processing' | null;

const POLL_INTERVAL_MS = 2000;
const POLL_MAX_DURATION_MS = 20000;

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
    NgIcon,
  ],
  providers: [
    provideIcons({ heroArrowDownTray }),
  ],
  templateUrl: './parent-invoice-detail.component.html',
})
export class ParentInvoiceDetailComponent implements OnInit, OnDestroy {
  private readonly apiService = inject(ParentInvoicesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly toast = inject(ToastService);

  private readonly destroy$ = new Subject<void>();
  private pollSub: Subscription | null = null;

  detail: ParentInvoiceDetail | null = null;
  isLoading = false;
  errorMessage: string | null = null;
  isPaying = false;

  returnKind: PaymentReturnKind = 'none';
  returnDisplayState: PaymentReturnDisplayState = null;
  isPolling = false;

  readonly formatGbp = formatGbp;
  readonly formatBillingMonthLabel = formatBillingMonthLabel;
  readonly formatInstant = formatInstant;
  readonly formatMinutes = formatMinutes;
  readonly lineKindLabel = lineKindLabel;
  readonly balanceDueMinor = balanceDueMinor;
  readonly fundingModelLabel = fundingModelLabel;

  ngOnInit(): void {
    const invoiceId = this.route.snapshot.paramMap.get('invoiceId');
    if (!invoiceId) {
      this.errorMessage = 'Invoice ID is missing.';
      return;
    }

    this.parseReturnKind();

    if (this.returnKind !== 'none') {
      this.clearReturnParams();
    }

    this.loadDetail(invoiceId);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.pollSub?.unsubscribe();
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

  get returnAlertMessage(): string {
    switch (this.returnDisplayState) {
      case 'paid':
        return 'Payment received. This invoice is now marked as paid.';
      case 'failed':
        return 'Payment was not completed. You can try again when you are ready.';
      case 'cancelled':
        return 'Payment canceled. No payment was taken, and you can try again.';
      case 'processing':
        return 'Payment is still processing. We are checking for an update; you can refresh this invoice if it does not update shortly.';
      default:
        return '';
    }
  }

  get returnAlertVariant(): 'success' | 'error' | 'warning' | 'info' {
    switch (this.returnDisplayState) {
      case 'paid':
        return 'success';
      case 'failed':
        return 'error';
      case 'cancelled':
        return 'warning';
      case 'processing':
        return 'info';
      default:
        return 'info';
    }
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
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'payment.parentCheckout'));
        this.isPaying = false;
      },
    });
  }

  redirectTo(url: string): void {
    window.location.href = url;
  }

  downloadPdf(): void {
    if (!this.detail) return;

    this.apiService.downloadPdf(this.detail.invoiceId).subscribe({
      next: (blob) => {
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = this.detail!.invoiceNumber ? `INV-${this.detail!.invoiceNumber}.pdf` : `INV-${this.detail!.invoiceId}.pdf`;
        a.click();
        URL.revokeObjectURL(url);
      },
      error: () => {
        this.toast.error('Failed to download PDF. Please try again.');
      },
    });
  }

  private parseReturnKind(): void {
    const checkout = this.route.snapshot.queryParamMap.get('checkout');
    if (checkout === 'success') {
      this.returnKind = 'success';
    } else if (checkout === 'cancelled' || checkout === 'canceled') {
      this.returnKind = 'cancelled';
    }
  }

  private clearReturnParams(): void {
    const queryParams = { ...this.route.snapshot.queryParams };
    delete queryParams['checkout'];
    delete queryParams['session_id'];

    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: Object.keys(queryParams).length > 0 ? queryParams : {},
      replaceUrl: true,
    });
  }

  private loadDetail(invoiceId: string): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.apiService.getInvoice(invoiceId).subscribe({
      next: (detail) => {
        this.detail = detail;
        this.isLoading = false;
        this.deriveReturnDisplayState();

        if (this.returnKind === 'success' && !isTerminalStatus(detail.status)) {
          this.startPolling(invoiceId);
        }
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'payment.parentDetail'));
        this.isLoading = false;
      },
    });
  }

  private deriveReturnDisplayState(): void {
    if (this.returnKind === 'none') {
      this.returnDisplayState = null;
      return;
    }

    const status = this.detail?.status;
    if (status === 'paid') {
      this.returnDisplayState = 'paid';
    } else if (status === 'payment_failed') {
      this.returnDisplayState = 'failed';
    } else if (this.returnKind === 'cancelled') {
      this.returnDisplayState = 'cancelled';
    } else {
      this.returnDisplayState = 'processing';
    }
  }

  private startPolling(invoiceId: string): void {
    this.pollSub?.unsubscribe();
    this.isPolling = true;

    const maxTicks = POLL_MAX_DURATION_MS / POLL_INTERVAL_MS;
    let tick = 0;

    this.pollSub = timer(POLL_INTERVAL_MS, POLL_INTERVAL_MS)
      .pipe(
        takeUntil(this.destroy$),
        switchMap(() => this.apiService.getInvoice(invoiceId)),
        takeWhile(() => {
          tick++;
          if (tick >= maxTicks) return false;
          return !isTerminalStatus(this.detail?.status);
        }),
      )
      .subscribe({
        next: (detail) => {
          this.detail = detail;
          this.deriveReturnDisplayState();
          if (isTerminalStatus(detail.status)) {
            this.stopPolling();
          }
        },
        error: (err) => {
          const mapped = this.errorMapper.mapAndHandle(err);
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'payment.parentDetail'));
          this.stopPolling();
        },
        complete: () => {
          this.isPolling = false;
        },
      });
  }

  private stopPolling(): void {
    this.pollSub?.unsubscribe();
    this.pollSub = null;
    this.isPolling = false;
  }
}

function isTerminalStatus(status: ParentInvoiceStatus | undefined): boolean {
  return status === 'paid' || status === 'payment_failed';
}

function fundingModelLabel(model: string | null): string {
  if (model === 'term_time_only') return 'Term-time funding';
  if (model === 'stretched') return 'Stretched funding';
  return 'Funded hours';
}
