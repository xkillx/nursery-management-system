import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  ManagerInvoiceStatus,
  ManagerInvoiceStatusFilter,
  ManagerInvoiceListItem,
  ManagerInvoiceListResult,
  ManagerInvoiceDetail,
  ManagerInvoiceLine,
  ManagerInvoiceCalculation,
  ManagerInvoiceGeneratedRunException,
  ManagerInvoicePeriod,
  ManagerInvoiceParentContact,
  PaymentAttempt,
  PaymentEvent,
  ManagerPaymentStatus,
  PaginatedPaymentEvents,
} from '../models/manager-invoices.models';
import { formatChildName } from '../utils/manager-list-formatters';

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
  issued_at: string | null;
  paid_at?: string | null;
  payment_failed_at?: string | null;
  payment_status_updated_at?: string | null;
  generated_run_id?: string | null;
  generated_run_status?: string | null;
  generated_run_started_at?: string | null;
  generated_run_completed_at?: string | null;
  generated_run_exception_count?: number | null;
  created_at: string;
  updated_at: string;
}

interface InvoiceListResponseApi {
  items: InvoiceListItemApi[];
  limit: number;
  offset: number;
}

interface InvoiceLineApi {
  line_id: string;
  line_kind: string;
  description: string;
  sort_order: number;
  quantity_minutes: number | null;
  unit_amount_minor: number | null;
  line_amount_minor: number;
  raw_attended_minutes?: number | null;
  rounded_attended_minutes?: number | null;
  funded_allowance_minutes?: number | null;
  funded_deduction_minutes?: number | null;
  core_billable_minutes?: number | null;
  session_count?: number | null;
}

interface InvoiceCalculationApi {
  core_hourly_rate_minor?: number | null;
  raw_attended_minutes?: number | null;
  rounded_attended_minutes?: number | null;
  funded_allowance_minutes?: number | null;
  funded_deduction_minutes?: number | null;
  core_billable_minutes?: number | null;
  included_session_count?: number | null;
  core_subtotal_minor?: number | null;
  extras_total_minor?: number | null;
}

interface GeneratedRunExceptionApi extends ChildNameApi {
  child_id: string;
  blocker_codes: string[];
}

interface ParentContactApi {
  full_name: string;
  address_line1: string;
  address_line2: string;
  address_city: string;
  address_postcode: string;
  email: string;
  telephone: string;
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
  locked_at?: string | null;
  due_at?: string | null;
  paid_at?: string | null;
  payment_failed_at?: string | null;
  payment_status_updated_at?: string | null;
  adjusts_invoice_id?: string | null;
  adjustment_reason_code?: string | null;
  adjustment_reason_note?: string | null;
  generated_run_id?: string | null;
  generated_run_status?: string | null;
  generated_run_started_at?: string | null;
  generated_run_completed_at?: string | null;
  generated_run_exception_count?: number | null;
  generated_run_exceptions?: GeneratedRunExceptionApi[] | null;
  site_profile?: { nursery_name: string; phone: string; email: string; website: string; address_street: string; address_city: string; address_postcode: string } | null;
  room_name?: string | null;
  parent_contact?: ParentContactApi | null;
  calculation?: InvoiceCalculationApi | null;
  lines?: InvoiceLineApi[] | null;
  created_at?: string;
  updated_at?: string;
}

interface PaymentAttemptApi {
  payment_attempt_id: string;
  status: string;
  amount_minor: number;
  currency_code: string;
  stripe_checkout_session_id: string | null;
  stripe_payment_intent_id: string | null;
  stripe_expires_at: string | null;
  failure_reason: string | null;
  provider_error_code: string | null;
  provider_error_message: string | null;
  created_at: string;
  updated_at: string;
}

interface PaymentEventApi {
  payment_event_id: string;
  payment_attempt_id: string;
  stripe_event_id: string | null;
  stripe_event_type: string | null;
  stripe_checkout_session_id: string | null;
  stripe_payment_intent_id: string | null;
  outcome: string;
  reason_code: string;
  previous_invoice_status: string | null;
  new_invoice_status: string | null;
  attempt_previous_status: string | null;
  attempt_new_status: string | null;
  amount_minor: number | null;
  currency_code: string | null;
  webhook_processing_status: string;
  webhook_processing_reason: string | null;
  webhook_received_at: string | null;
  webhook_processed_at: string | null;
  created_at: string;
}

interface ManagerPaymentStatusApi {
  invoice_id: string;
  status: string;
  due_status: string;
  currency_code: string;
  total_due_minor: number;
  amount_paid_minor: number;
  paid_at: string | null;
  payment_failed_at: string | null;
  payment_status_updated_at: string | null;
  checkout_retry_available: boolean;
  checkout_retry_reason_code: string;
  latest_payment_attempt: PaymentAttemptApi | null;
  latest_payment_event: PaymentEventApi | null;
}

interface PaginatedPaymentEventsApi {
  items: PaymentEventApi[];
  limit: number;
  offset: number;
}

@Injectable({ providedIn: 'root' })
export class ManagerInvoicesApiService {
  private readonly http = inject(HttpClient);

  listInvoices(params: {
    billingMonth?: string;
    billingMonthFrom?: string;
    billingMonthTo?: string;
    status: ManagerInvoiceStatusFilter;
    childId?: string;
    limit: number;
    offset: number;
  }): Observable<ManagerInvoiceListResult> {
    const queryObj: Record<string, string> = {
      limit: String(params.limit),
      offset: String(params.offset),
    };

    if (params.billingMonth) {
      queryObj['billing_month'] = params.billingMonth;
    }
    if (params.billingMonthFrom) {
      queryObj['billing_month_from'] = params.billingMonthFrom;
    }
    if (params.billingMonthTo) {
      queryObj['billing_month_to'] = params.billingMonthTo;
    }

    if (params.status !== 'all') {
      queryObj['status'] = params.status;
    }

    if (params.childId) {
      queryObj['child_id'] = params.childId;
    }

    return this.http
      .get<InvoiceListResponseApi>(apiUrl('/invoices'), {
        params: new HttpParams({ fromObject: queryObj }),
      })
      .pipe(map((res) => this.toListResult(res)));
  }

  getInvoice(invoiceId: string): Observable<ManagerInvoiceDetail> {
    return this.http
      .get<InvoiceDetailApi>(apiUrl(`/invoices/${invoiceId}`))
      .pipe(map((res) => this.toDetail(res)));
  }

  getPaymentStatus(invoiceId: string): Observable<ManagerPaymentStatus> {
    return this.http
      .get<ManagerPaymentStatusApi>(apiUrl(`/invoices/${invoiceId}/payment-status`))
      .pipe(map((res) => this.toPaymentStatus(res)));
  }

  listPaymentEvents(invoiceId: string, params: { limit: number; offset: number }): Observable<PaginatedPaymentEvents> {
    return this.http
      .get<PaginatedPaymentEventsApi>(apiUrl(`/invoices/${invoiceId}/payment-events`), {
        params: new HttpParams({ fromObject: { limit: String(params.limit), offset: String(params.offset) } }),
      })
      .pipe(map((res) => this.toPaginatedEvents(res)));
  }

  private toListResult(res: InvoiceListResponseApi): ManagerInvoiceListResult {
    return {
      items: res.items.map((i) => this.toListItem(i)),
      limit: res.limit,
      offset: res.offset,
    };
  }

  private toListItem(i: InvoiceListItemApi): ManagerInvoiceListItem {
    return {
      invoiceId: i.invoice_id,
      invoiceKind: i.invoice_kind,
      invoiceNumber: i.invoice_number,
      invoiceNumberDisplay: i.invoice_number_display ?? '',
      childId: i.child_id,
      childName: this.childName(i),
      billingMonth: i.billing_month,
      period: i.period ? this.toPeriod(i.period) : null,
      status: i.status as ManagerInvoiceStatus,
      dueStatus: (i.due_status ?? 'not_due') as ManagerInvoiceListItem['dueStatus'],
      currencyCode: i.currency_code ?? 'gbp',
      subtotalMinor: i.subtotal_minor,
      fundedDeductionMinor: i.funded_deduction_minor,
      totalDueMinor: i.total_due_minor,
      amountPaidMinor: i.amount_paid_minor ?? 0,
      dueAt: i.due_at ?? null,
      issuedAt: i.issued_at,
      paidAt: i.paid_at ?? null,
      paymentFailedAt: i.payment_failed_at ?? null,
      paymentStatusUpdatedAt: i.payment_status_updated_at ?? null,
      generatedRunId: i.generated_run_id ?? null,
      generatedRunStatus: i.generated_run_status ?? null,
      generatedRunStartedAt: i.generated_run_started_at ?? null,
      generatedRunCompletedAt: i.generated_run_completed_at ?? null,
      generatedRunExceptionCount: i.generated_run_exception_count ?? null,
      createdAt: i.created_at,
      updatedAt: i.updated_at,
    };
  }

  private toDetail(d: InvoiceDetailApi): ManagerInvoiceDetail {
    return {
      invoiceId: d.invoice_id,
      invoiceKind: d.invoice_kind ?? 'monthly',
      invoiceNumber: d.invoice_number,
      invoiceNumberDisplay: d.invoice_number_display ?? '',
      childId: d.child_id,
      childName: this.childName(d),
      billingMonth: d.billing_month,
      period: d.period ? this.toPeriod(d.period) : null,
      status: d.status as ManagerInvoiceStatus,
      dueStatus: (d.due_status ?? 'not_due') as ManagerInvoiceDetail['dueStatus'],
      currencyCode: d.currency_code ?? 'gbp',
      subtotalMinor: d.subtotal_minor,
      fundedDeductionMinor: d.funded_deduction_minor,
      totalDueMinor: d.total_due_minor,
      amountPaidMinor: d.amount_paid_minor ?? 0,
      issuedAt: d.issued_at ?? null,
      lockedAt: d.locked_at ?? null,
      dueAt: d.due_at ?? null,
      paidAt: d.paid_at ?? null,
      paymentFailedAt: d.payment_failed_at ?? null,
      paymentStatusUpdatedAt: d.payment_status_updated_at ?? null,
      adjustsInvoiceId: d.adjusts_invoice_id ?? null,
      adjustmentReasonCode: d.adjustment_reason_code ?? null,
      adjustmentReasonNote: d.adjustment_reason_note ?? null,
      generatedRunId: d.generated_run_id ?? null,
      generatedRunStatus: d.generated_run_status ?? null,
      generatedRunStartedAt: d.generated_run_started_at ?? null,
      generatedRunCompletedAt: d.generated_run_completed_at ?? null,
      generatedRunExceptionCount: d.generated_run_exception_count ?? null,
      generatedRunExceptions: (d.generated_run_exceptions ?? []).map((e) => this.toException(e)),
      site_profile: d.site_profile ?? null,
      roomName: d.room_name ?? null,
      parentContact: d.parent_contact ? this.toParentContact(d.parent_contact) : null,
      calculation: d.calculation ? this.toCalculation(d.calculation) : null,
      lines: (d.lines ?? [])
        .sort((a, b) => a.sort_order - b.sort_order)
        .map((l) => this.toLine(l)),
      createdAt: d.created_at ?? '',
      updatedAt: d.updated_at ?? '',
    };
  }

  private toPeriod(p: { start_date: string; end_date: string }): ManagerInvoicePeriod {
    return { startDate: p.start_date, endDate: p.end_date };
  }

  private toException(e: GeneratedRunExceptionApi): ManagerInvoiceGeneratedRunException {
    return {
      childId: e.child_id,
      childName: this.childName(e),
      blockerCodes: e.blocker_codes,
    };
  }

  private childName(child: ChildNameApi): string {
    return formatChildName({
      firstName: child.child_first_name,
      middleName: child.child_middle_name,
      lastName: child.child_last_name,
    });
  }

  private toParentContact(pc: ParentContactApi): ManagerInvoiceParentContact {
    return {
      fullName: pc.full_name,
      addressLine1: pc.address_line1,
      addressLine2: pc.address_line2,
      addressCity: pc.address_city,
      addressPostcode: pc.address_postcode,
      email: pc.email,
      telephone: pc.telephone,
    };
  }

  private toCalculation(c: InvoiceCalculationApi): ManagerInvoiceCalculation {
    return {
      siteCoreHourlyRateMinor: c.core_hourly_rate_minor ?? null,
      rawAttendedMinutes: c.raw_attended_minutes ?? null,
      roundedAttendedMinutes: c.rounded_attended_minutes ?? null,
      fundedAllowanceMinutes: c.funded_allowance_minutes ?? null,
      fundedDeductionMinutes: c.funded_deduction_minutes ?? null,
      coreBillableMinutes: c.core_billable_minutes ?? null,
      includedSessionCount: c.included_session_count ?? null,
      coreSubtotalMinor: c.core_subtotal_minor ?? null,
      extrasTotalMinor: c.extras_total_minor ?? null,
    };
  }

  private toLine(l: InvoiceLineApi): ManagerInvoiceLine {
    return {
      lineId: l.line_id,
      lineKind: l.line_kind,
      description: l.description,
      sortOrder: l.sort_order,
      quantityMinutes: l.quantity_minutes,
      unitAmountMinor: l.unit_amount_minor,
      lineAmountMinor: l.line_amount_minor,
      rawAttendedMinutes: l.raw_attended_minutes ?? null,
      roundedAttendedMinutes: l.rounded_attended_minutes ?? null,
      fundedAllowanceMinutes: l.funded_allowance_minutes ?? null,
      fundedDeductionMinutes: l.funded_deduction_minutes ?? null,
      coreBillableMinutes: l.core_billable_minutes ?? null,
      sessionCount: l.session_count ?? null,
    };
  }

  private toPaymentAttempt(a: PaymentAttemptApi): PaymentAttempt {
    return {
      paymentAttemptId: a.payment_attempt_id,
      status: a.status,
      amountMinor: a.amount_minor,
      currencyCode: a.currency_code,
      stripeCheckoutSessionId: a.stripe_checkout_session_id,
      stripePaymentIntentId: a.stripe_payment_intent_id,
      stripeExpiresAt: a.stripe_expires_at,
      failureReason: a.failure_reason,
      providerErrorCode: a.provider_error_code,
      providerErrorMessage: a.provider_error_message,
      createdAt: a.created_at,
      updatedAt: a.updated_at,
    };
  }

  private toPaymentEvent(e: PaymentEventApi): PaymentEvent {
    return {
      paymentEventId: e.payment_event_id,
      paymentAttemptId: e.payment_attempt_id,
      stripeEventId: e.stripe_event_id,
      stripeEventType: e.stripe_event_type,
      stripeCheckoutSessionId: e.stripe_checkout_session_id,
      stripePaymentIntentId: e.stripe_payment_intent_id,
      outcome: e.outcome,
      reasonCode: e.reason_code,
      previousInvoiceStatus: e.previous_invoice_status,
      newInvoiceStatus: e.new_invoice_status,
      attemptPreviousStatus: e.attempt_previous_status,
      attemptNewStatus: e.attempt_new_status,
      amountMinor: e.amount_minor,
      currencyCode: e.currency_code,
      webhookProcessingStatus: e.webhook_processing_status,
      webhookProcessingReason: e.webhook_processing_reason,
      webhookReceivedAt: e.webhook_received_at,
      webhookProcessedAt: e.webhook_processed_at,
      createdAt: e.created_at,
    };
  }

  private toPaymentStatus(s: ManagerPaymentStatusApi): ManagerPaymentStatus {
    return {
      invoiceId: s.invoice_id,
      status: s.status,
      dueStatus: s.due_status,
      currencyCode: s.currency_code,
      totalDueMinor: s.total_due_minor,
      amountPaidMinor: s.amount_paid_minor,
      paidAt: s.paid_at,
      paymentFailedAt: s.payment_failed_at,
      paymentStatusUpdatedAt: s.payment_status_updated_at,
      checkoutRetryAvailable: s.checkout_retry_available,
      checkoutRetryReasonCode: s.checkout_retry_reason_code,
      latestPaymentAttempt: s.latest_payment_attempt ? this.toPaymentAttempt(s.latest_payment_attempt) : null,
      latestPaymentEvent: s.latest_payment_event ? this.toPaymentEvent(s.latest_payment_event) : null,
    };
  }

  private toPaginatedEvents(res: PaginatedPaymentEventsApi): PaginatedPaymentEvents {
    return {
      items: res.items.map((e) => this.toPaymentEvent(e)),
      limit: res.limit,
      offset: res.offset,
    };
  }
}
