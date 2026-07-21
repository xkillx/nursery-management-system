import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, map } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  ManagerInvoicePrefill,
  ManagerInvoicePrefillLine,
  DraftInvoiceResult,
  DraftLineResult,
} from '../models/manager-invoice-create.models';

interface PrefillLineApi {
  line_kind: string;
  description: string;
  sort_order: number;
  quantity_minutes: number;
  unit_amount_minor: number;
  line_amount_minor: number;
  funded_allowance_minutes: number;
  funded_deduction_minutes: number;
  core_billable_minutes: number;
  session_count: number;
}

interface PrefillEntitlementApi {
  funding_profile_id: string | null;
  funded_allowance_minutes: number;
  status_label: string;
}

interface PrefillResponseApi {
  child_id: string;
  child_first_name: string;
  child_middle_name: string | null;
  child_last_name: string | null;
  billing_month: string;
  entitlement_status: PrefillEntitlementApi;
  lines: PrefillLineApi[];
  subtotal_minor: number;
  funded_deduction_minor: number;
  total_due_minor: number;
  warnings: string[];
}

interface DraftLineRequestApi {
  line_kind: string;
  description: string;
  sort_order: number;
  quantity_minutes: number;
  unit_amount_minor: number;
  line_amount_minor: number;
}

interface CreateDraftRequestApi {
  child_id: string;
  billing_month: string;
  lines: DraftLineRequestApi[];
  payment_terms: string;
  internal_notes: string;
  parent_note: string;
}

interface DraftLineResponseApi {
  line_id: string;
  line_kind: string;
  description: string;
  sort_order: number;
  quantity_minutes: number;
  unit_amount_minor: number;
  line_amount_minor: number;
}

interface DraftInvoiceResponseApi {
  invoice_id: string;
  child_id: string;
  billing_month: string;
  status: string;
  lines: DraftLineResponseApi[];
  subtotal_minor: number;
  total_due_minor: number;
  payment_terms: string;
  internal_notes: string;
  parent_note: string;
  created_at: string;
  updated_at: string;
}

interface IssueInvoiceResponseApi {
  invoice_id: string;
  invoice_number: string;
  status: string;
  issued_at: string;
  locked_at: string;
  due_at: string;
  issued_run_id: string;
  total_due_minor: number;
}

@Injectable({ providedIn: 'root' })
export class ManagerInvoiceCreateApiService {
  private readonly http = inject(HttpClient);

  getPrefill(childId: string, billingMonth: string): Observable<ManagerInvoicePrefill> {
    return this.http
      .get<PrefillResponseApi>(apiUrl('/invoices/prefill'), {
        params: new HttpParams({ fromObject: { child_id: childId, billing_month: billingMonth } }),
      })
      .pipe(map((res) => this.toPrefill(res)));
  }

  createDraft(input: {
    childId: string;
    billingMonth: string;
    lines: { lineKind: string; description: string; sortOrder: number; quantityMinutes: number; unitAmountMinor: number; lineAmountMinor: number }[];
    paymentTerms: string;
    internalNotes: string;
    parentNote: string;
  }): Observable<DraftInvoiceResult> {
    const body: CreateDraftRequestApi = {
      child_id: input.childId,
      billing_month: input.billingMonth,
      lines: input.lines.map((l) => ({
        line_kind: l.lineKind,
        description: l.description,
        sort_order: l.sortOrder,
        quantity_minutes: l.quantityMinutes,
        unit_amount_minor: l.unitAmountMinor,
        line_amount_minor: l.lineAmountMinor,
      })),
      payment_terms: input.paymentTerms,
      internal_notes: input.internalNotes,
      parent_note: input.parentNote,
    };
    return this.http
      .post<DraftInvoiceResponseApi>(apiUrl('/invoices/drafts'), body)
      .pipe(map((res) => this.toDraftResult(res)));
  }

  createAndIssue(input: {
    childId: string;
    billingMonth: string;
    lines: { lineKind: string; description: string; sortOrder: number; quantityMinutes: number; unitAmountMinor: number; lineAmountMinor: number }[];
    paymentTerms: string;
    internalNotes: string;
    parentNote: string;
  }): Observable<{ invoiceId: string; invoiceNumber: string; status: string; totalDueMinor: number }> {
    const body: CreateDraftRequestApi = {
      child_id: input.childId,
      billing_month: input.billingMonth,
      lines: input.lines.map((l) => ({
        line_kind: l.lineKind,
        description: l.description,
        sort_order: l.sortOrder,
        quantity_minutes: l.quantityMinutes,
        unit_amount_minor: l.unitAmountMinor,
        line_amount_minor: l.lineAmountMinor,
      })),
      payment_terms: input.paymentTerms,
      internal_notes: input.internalNotes,
      parent_note: input.parentNote,
    };
    return this.http
      .post<IssueInvoiceResponseApi>(apiUrl('/invoices/drafts/issue'), body)
      .pipe(map((res) => ({
        invoiceId: res.invoice_id,
        invoiceNumber: res.invoice_number,
        status: res.status,
        totalDueMinor: res.total_due_minor,
      })));
  }

  private toPrefill(res: PrefillResponseApi): ManagerInvoicePrefill {
    return {
      childId: res.child_id,
      childFirstName: res.child_first_name,
      childMiddleName: res.child_middle_name,
      childLastName: res.child_last_name,
      billingMonth: res.billing_month,
      entitlementStatus: {
        fundingProfileId: res.entitlement_status.funding_profile_id,
        fundedAllowanceMinutes: res.entitlement_status.funded_allowance_minutes,
        statusLabel: res.entitlement_status.status_label,
      },
      lines: res.lines.map((l) => this.toPrefillLine(l)),
      subtotalMinor: res.subtotal_minor,
      fundedDeductionMinor: res.funded_deduction_minor,
      totalDueMinor: res.total_due_minor,
      warnings: res.warnings,
    };
  }

  private toPrefillLine(l: PrefillLineApi): ManagerInvoicePrefillLine {
    return {
      lineKind: l.line_kind,
      description: l.description,
      sortOrder: l.sort_order,
      quantityMinutes: l.quantity_minutes,
      unitAmountMinor: l.unit_amount_minor,
      lineAmountMinor: l.line_amount_minor,
      fundedAllowanceMinutes: l.funded_allowance_minutes,
      fundedDeductionMinutes: l.funded_deduction_minutes,
      coreBillableMinutes: l.core_billable_minutes,
      sessionCount: l.session_count,
    };
  }

  private toDraftResult(res: DraftInvoiceResponseApi): DraftInvoiceResult {
    return {
      invoiceId: res.invoice_id,
      childId: res.child_id,
      billingMonth: res.billing_month,
      status: res.status,
      lines: res.lines.map((l) => this.toDraftLine(l)),
      subtotalMinor: res.subtotal_minor,
      totalDueMinor: res.total_due_minor,
      paymentTerms: res.payment_terms,
      internalNotes: res.internal_notes,
      parentNote: res.parent_note,
      createdAt: res.created_at,
      updatedAt: res.updated_at,
    };
  }

  private toDraftLine(l: DraftLineResponseApi): DraftLineResult {
    return {
      lineId: l.line_id,
      lineKind: l.line_kind,
      description: l.description,
      sortOrder: l.sort_order,
      quantityMinutes: l.quantity_minutes,
      unitAmountMinor: l.unit_amount_minor,
      lineAmountMinor: l.line_amount_minor,
    };
  }
}
