import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  ParentInvoiceStatus,
  ParentInvoiceListItem,
  ParentInvoiceListResult,
  ParentInvoiceDetail,
  ParentInvoiceLine,
  ParentInvoiceCalculation,
  ParentInvoicePeriod,
  CheckoutSessionResult,
} from '../models/parent-invoices.models';
import { formatChildName } from '../../staff/utils/manager-list-formatters';

interface ChildNameApi {
  child_first_name: string;
  child_middle_name?: string | null;
  child_last_name?: string | null;
}

interface InvoiceListItemApi extends ChildNameApi {
  invoice_id: string;
  invoice_kind: string;
  invoice_number: string | null;
  invoice_number_display?: string;
  child_id: string;
  billing_month: string;
  period?: { start_date: string; end_date: string } | null;
  status: string;
  due_status?: string;
  currency_code?: string;
  subtotal_minor: number;
  funded_deduction_minor: number;
  total_due_minor: number;
  amount_paid_minor?: number;
  due_at?: string | null;
  issued_at?: string | null;
  paid_at?: string | null;
  payment_failed_at?: string | null;
  payment_status_updated_at?: string | null;
}

interface InvoiceListResponseApi {
  items: InvoiceListItemApi[];
  total: number;
  page: number;
  page_size: number;
}

interface InvoiceCalculationApi {
  raw_attended_minutes?: number | null;
  rounded_attended_minutes?: number | null;
  funded_allowance_minutes?: number | null;
  funded_deduction_minutes?: number | null;
  core_billable_minutes?: number | null;
  included_session_count?: number | null;
  core_hourly_rate_minor?: number | null;
  core_subtotal_minor?: number | null;
  extras_total_minor?: number | null;
}

interface InvoiceLineApi {
  line_kind: string;
  description: string;
  sort_order: number;
  quantity_minutes?: number | null;
  unit_amount_minor?: number | null;
  line_amount_minor: number;
  funding_model?: string | null;
}

interface InvoiceDetailApi extends ChildNameApi {
  invoice_id: string;
  invoice_kind?: string;
  invoice_number: string | null;
  invoice_number_display?: string;
  child_id: string;
  billing_month: string;
  period?: { start_date: string; end_date: string } | null;
  status: string;
  due_status?: string;
  currency_code?: string;
  subtotal_minor: number;
  funded_deduction_minor: number;
  total_due_minor: number;
  amount_paid_minor?: number;
  issued_at?: string | null;
  due_at?: string | null;
  paid_at?: string | null;
  payment_failed_at?: string | null;
  payment_status_updated_at?: string | null;
  site_profile?: { nursery_name: string; phone: string; email: string; website: string; address_street: string; address_city: string; address_postcode: string } | null;
  calculation?: InvoiceCalculationApi | null;
  lines?: InvoiceLineApi[] | null;
}

interface CheckoutSessionApi {
  checkout_session_id: string;
  checkout_url: string;
  payment_attempt_id: string;
}

@Injectable({ providedIn: 'root' })
export class ParentInvoicesApiService {
  private readonly http = inject(HttpClient);

  listInvoices(params: {
    page: number;
    pageSize: number;
    billingMonth?: string;
    billingMonthFrom?: string;
    billingMonthTo?: string;
    status?: ParentInvoiceStatus;
    childId?: string;
  }): Observable<ParentInvoiceListResult> {
    const queryObj: Record<string, string> = {
      page: String(params.page),
      page_size: String(params.pageSize),
    };

    if (params.billingMonth) queryObj['billing_month'] = params.billingMonth;
    if (params.billingMonthFrom) queryObj['billing_month_from'] = params.billingMonthFrom;
    if (params.billingMonthTo) queryObj['billing_month_to'] = params.billingMonthTo;
    if (params.status) queryObj['status'] = params.status;
    if (params.childId) queryObj['child_id'] = params.childId;

    return this.http
      .get<InvoiceListResponseApi>(apiUrl('/parent/invoices'), {
        params: new HttpParams({ fromObject: queryObj }),
      })
      .pipe(map((res) => this.toListResult(res)));
  }

  getInvoice(invoiceId: string): Observable<ParentInvoiceDetail> {
    return this.http
      .get<InvoiceDetailApi>(apiUrl(`/parent/invoices/${invoiceId}`))
      .pipe(map((res) => this.toDetail(res)));
  }

  createCheckoutSession(invoiceId: string): Observable<CheckoutSessionResult> {
    return this.http
      .post<CheckoutSessionApi>(apiUrl(`/parent/invoices/${invoiceId}/checkout-sessions`), null)
      .pipe(map((res) => this.toCheckoutSession(res)));
  }

  private toListResult(res: InvoiceListResponseApi): ParentInvoiceListResult {
    return {
      items: res.items.map((i) => this.toListItem(i)),
      total: res.total,
      page: res.page,
      pageSize: res.page_size,
    };
  }

  private toListItem(i: InvoiceListItemApi): ParentInvoiceListItem {
    return {
      invoiceId: i.invoice_id,
      invoiceKind: i.invoice_kind,
      invoiceNumber: i.invoice_number,
      invoiceNumberDisplay: i.invoice_number_display ?? '',
      childId: i.child_id,
      childName: this.childName(i),
      billingMonth: i.billing_month,
      period: i.period ? this.toPeriod(i.period) : null,
      status: i.status as ParentInvoiceStatus,
      dueStatus: (i.due_status ?? 'due') as ParentInvoiceListItem['dueStatus'],
      currencyCode: i.currency_code ?? 'gbp',
      subtotalMinor: i.subtotal_minor,
      fundedDeductionMinor: i.funded_deduction_minor,
      totalDueMinor: i.total_due_minor,
      amountPaidMinor: i.amount_paid_minor ?? 0,
      dueAt: i.due_at ?? null,
      issuedAt: i.issued_at ?? null,
      paidAt: i.paid_at ?? null,
      paymentFailedAt: i.payment_failed_at ?? null,
      paymentStatusUpdatedAt: i.payment_status_updated_at ?? null,
    };
  }

  private toDetail(d: InvoiceDetailApi): ParentInvoiceDetail {
    return {
      invoiceId: d.invoice_id,
      invoiceKind: d.invoice_kind ?? 'monthly',
      invoiceNumber: d.invoice_number,
      invoiceNumberDisplay: d.invoice_number_display ?? '',
      childId: d.child_id,
      childName: this.childName(d),
      billingMonth: d.billing_month,
      period: d.period ? this.toPeriod(d.period) : null,
      status: d.status as ParentInvoiceStatus,
      dueStatus: (d.due_status ?? 'due') as ParentInvoiceDetail['dueStatus'],
      currencyCode: d.currency_code ?? 'gbp',
      subtotalMinor: d.subtotal_minor,
      fundedDeductionMinor: d.funded_deduction_minor,
      totalDueMinor: d.total_due_minor,
      amountPaidMinor: d.amount_paid_minor ?? 0,
      issuedAt: d.issued_at ?? null,
      dueAt: d.due_at ?? null,
      paidAt: d.paid_at ?? null,
      paymentFailedAt: d.payment_failed_at ?? null,
      paymentStatusUpdatedAt: d.payment_status_updated_at ?? null,
      site_profile: d.site_profile ?? null,
      calculation: d.calculation ? this.toCalculation(d.calculation) : null,
      lines: (d.lines ?? [])
        .sort((a, b) => a.sort_order - b.sort_order)
        .map((l) => this.toLine(l)),
    };
  }

  private childName(child: ChildNameApi): string {
    return formatChildName({
      firstName: child.child_first_name,
      middleName: child.child_middle_name,
      lastName: child.child_last_name,
    });
  }

  private toPeriod(p: { start_date: string; end_date: string }): ParentInvoicePeriod {
    return { startDate: p.start_date, endDate: p.end_date };
  }

  private toCalculation(c: InvoiceCalculationApi): ParentInvoiceCalculation {
    return {
      rawAttendedMinutes: c.raw_attended_minutes ?? null,
      roundedAttendedMinutes: c.rounded_attended_minutes ?? null,
      fundedAllowanceMinutes: c.funded_allowance_minutes ?? null,
      fundedDeductionMinutes: c.funded_deduction_minutes ?? null,
      coreBillableMinutes: c.core_billable_minutes ?? null,
      includedSessionCount: c.included_session_count ?? null,
      siteCoreHourlyRateMinor: c.core_hourly_rate_minor ?? null,
      coreSubtotalMinor: c.core_subtotal_minor ?? null,
      extrasTotalMinor: c.extras_total_minor ?? null,
    };
  }

  private toLine(l: InvoiceLineApi): ParentInvoiceLine {
    return {
      lineKind: l.line_kind,
      description: l.description,
      sortOrder: l.sort_order,
      quantityMinutes: l.quantity_minutes ?? null,
      unitAmountMinor: l.unit_amount_minor ?? null,
      lineAmountMinor: l.line_amount_minor,
      fundingModel: l.funding_model ?? null,
    };
  }

  private toCheckoutSession(c: CheckoutSessionApi): CheckoutSessionResult {
    return {
      checkoutSessionId: c.checkout_session_id,
      checkoutUrl: c.checkout_url,
      paymentAttemptId: c.payment_attempt_id,
    };
  }
}
