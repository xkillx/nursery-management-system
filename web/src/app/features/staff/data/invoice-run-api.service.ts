import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, forkJoin, map, of, switchMap } from 'rxjs';

import { apiUrl } from '../../../core/config/api.config';
import {
  InvoiceDraftLine,
  InvoiceDraftReviewItem,
  InvoiceRunBlocker,
  InvoiceRunEligibleChild,
  InvoiceRunException,
  InvoiceRunPreflight,
  InvoiceRunPreflightSummary,
  DraftGenerationResult,
  IssueException,
  IssueResultSummary,
  IssuedInvoiceResult,
} from '../models/invoice-run.models';
import { formatChildName } from '../utils/manager-list-formatters';

interface ChildNameApi {
  child_first_name: string;
  child_middle_name?: string | null;
  child_last_name?: string | null;
}

interface PreflightSummaryApi {
  total_children_count: number;
  eligible_children_count: number;
  blocked_children_count: number;
  included_session_count: number;
  rounded_attended_minutes: number;
  funded_deduction_minor: number;
  total_due_minor: number;
}

interface PreflightEligibleChildApi extends ChildNameApi {
  child_id: string;
  rounded_attended_minutes: number;
  funded_deduction_minor: number;
  total_due_minor: number;
}

interface PreflightBlockerApi {
  code: string;
  message: string;
}

interface PreflightBlockedChildApi extends ChildNameApi {
  child_id: string;
  blockers: PreflightBlockerApi[];
}

interface PreflightResponseApi {
  billing_month: string;
  summary: PreflightSummaryApi;
  eligible_children: PreflightEligibleChildApi[];
  blocked_children: PreflightBlockedChildApi[];
}

interface DraftGenerationSummaryApi {
  eligible_count: number;
  success_count: number;
  blocked_count: number;
  total_due_minor: number;
}

interface DraftGenerationItemApi extends ChildNameApi {
  child_id: string;
  action: 'created' | 'updated';
  invoice_id: string;
}

interface DraftGenerationBlockedApi extends ChildNameApi {
  child_id: string;
  blockers: PreflightBlockerApi[];
}

interface DraftGenerationResponseApi {
  run_id: string;
  billing_month: string;
  status: 'completed' | 'completed_with_exceptions';
  summary: DraftGenerationSummaryApi;
  generated: DraftGenerationItemApi[];
  blocked: DraftGenerationBlockedApi[];
}

interface InvoiceListItemApi extends ChildNameApi {
  invoice_id: string;
  child_id: string;
  billing_month: string;
  status: string;
  subtotal_minor: number;
  funded_deduction_minor: number;
  total_due_minor: number;
  invoice_number: string | null;
  issued_at: string | null;
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
}

interface InvoiceCalculationApi {
  rounded_attended_minutes: number;
  funded_deduction_minutes: number;
  extras_total_minor: number;
}

interface InvoiceDetailApi extends ChildNameApi {
  invoice_id: string;
  child_id: string;
  billing_month: string;
  status: string;
  subtotal_minor: number;
  funded_deduction_minor: number;
  total_due_minor: number;
  invoice_number: string | null;
  issued_at: string | null;
  calculation: InvoiceCalculationApi;
  lines: InvoiceLineApi[];
}

interface SingleIssueResponseApi {
  invoice_id: string;
  invoice_number: string;
  issued_at: string;
  total_due_minor: number;
}

interface BulkIssueIssuedApi extends ChildNameApi {
  invoice_id: string;
  invoice_number: string;
  child_id: string;
  issued_at: string;
  total_due_minor: number;
}

interface BulkIssueBlockedApi extends ChildNameApi {
  invoice_id: string;
  child_id: string;
  blockers: PreflightBlockerApi[];
}

interface BulkIssueResponseApi {
  run_id: string;
  billing_month: string;
  status: 'completed' | 'completed_with_exceptions';
  summary: {
    issued_count: number;
    blocked_count: number;
    total_due_minor: number;
  };
  issued: BulkIssueIssuedApi[];
  blocked: BulkIssueBlockedApi[];
}

@Injectable({ providedIn: 'root' })
export class InvoiceRunApiService {
  private readonly http = inject(HttpClient);

  loadPreflight(billingMonth: string): Observable<InvoiceRunPreflight> {
    return this.http
      .get<PreflightResponseApi>(apiUrl('/invoices/drafts/preflight'), {
        params: new HttpParams({ fromObject: { billing_month: billingMonth } }),
      })
      .pipe(map((res) => this.toPreflight(res)));
  }

  generateDrafts(billingMonth: string, childIds?: string[]): Observable<DraftGenerationResult> {
    const body: Record<string, unknown> = { billing_month: billingMonth };
    if (childIds && childIds.length > 0) {
      body['child_ids'] = childIds;
    }
    return this.http
      .post<DraftGenerationResponseApi>(apiUrl('/invoice-runs/drafts'), body)
      .pipe(map((res) => this.toDraftGenerationResult(res)));
  }

  listDrafts(billingMonth: string): Observable<InvoiceDraftReviewItem[]> {
    return this.http
      .get<InvoiceListResponseApi>(apiUrl('/invoices'), {
        params: new HttpParams({
          fromObject: {
            billing_month: billingMonth,
            status: 'draft',
            limit: '200',
            offset: '0',
          },
        }),
      })
      .pipe(
        switchMap((listRes) => {
          if (listRes.items.length === 0) return of([]);

          const detailCalls = listRes.items.map((item) =>
            this.http
              .get<InvoiceDetailApi>(apiUrl(`/invoices/${item.invoice_id}`))
              .pipe(map((detail) => this.toDraftReviewItem(detail, item))),
          );
          return forkJoin(detailCalls);
        }),
      );
  }

  bulkIssue(billingMonth: string, invoiceIds: string[]): Observable<IssueResultSummary> {
    return this.http
      .post<BulkIssueResponseApi>(apiUrl('/invoices/bulk-issue'), {
        billing_month: billingMonth,
        invoice_ids: invoiceIds,
        confirm: true,
      })
      .pipe(map((res) => this.toIssueResult(res)));
  }

  issueOne(invoiceId: string): Observable<IssueResultSummary> {
    return this.http
      .post<SingleIssueResponseApi>(apiUrl(`/invoices/${invoiceId}/issue`), { confirm: true })
      .pipe(
        switchMap((issueRes) =>
          this.http.get<InvoiceDetailApi>(apiUrl(`/invoices/${invoiceId}`)).pipe(
            map((detail) => this.toSingleIssueResult(issueRes, detail)),
          ),
        ),
      );
  }

  private toPreflight(res: PreflightResponseApi): InvoiceRunPreflight {
    return {
      billingMonth: res.billing_month,
      summary: this.toPreflightSummary(res.summary),
      eligibleChildren: res.eligible_children.map((c) => this.toEligibleChild(c)),
      blockedChildren: res.blocked_children.map((c) => this.toException(c)),
    };
  }

  private toPreflightSummary(s: PreflightSummaryApi): InvoiceRunPreflightSummary {
    return {
      totalChildren: s.total_children_count,
      eligibleChildren: s.eligible_children_count,
      blockedChildren: s.blocked_children_count,
      sessionsIncluded: s.included_session_count,
      attendedMinutes: s.rounded_attended_minutes,
      fundedDeductionMinor: s.funded_deduction_minor,
      totalDueMinor: s.total_due_minor,
    };
  }

  private toEligibleChild(c: PreflightEligibleChildApi): InvoiceRunEligibleChild {
    return {
      childId: c.child_id,
      childName: this.childName(c),
      attendedMinutes: c.rounded_attended_minutes,
      fundedDeductionMinor: c.funded_deduction_minor,
      totalDueMinor: c.total_due_minor,
    };
  }

  private toException(c: PreflightBlockedChildApi): InvoiceRunException {
    return {
      childId: c.child_id,
      childName: this.childName(c),
      blockers: c.blockers.map((b) => this.toBlocker(b)),
    };
  }

  private toBlocker(b: PreflightBlockerApi): InvoiceRunBlocker {
    return {
      code: b.code,
      detail: b.message,
    };
  }

  private toDraftGenerationResult(res: DraftGenerationResponseApi): DraftGenerationResult {
    let generatedCount = 0;
    let updatedCount = 0;
    for (const item of res.generated) {
      if (item.action === 'created') generatedCount++;
      else updatedCount++;
    }
    return {
      billingMonth: res.billing_month,
      generatedCount,
      updatedCount,
      blockedCount: res.summary.blocked_count,
      blockedChildren: res.blocked.map((b) => this.toException(b)),
    };
  }

  private toDraftReviewItem(detail: InvoiceDetailApi, listItem: InvoiceListItemApi): InvoiceDraftReviewItem {
    const generatedAction = listItem.issued_at ? null : 'created';
    return {
      invoiceId: detail.invoice_id,
      childId: detail.child_id,
      childName: this.childName(detail),
      billingMonth: detail.billing_month,
      status: detail.status as InvoiceDraftReviewItem['status'],
      attendedMinutes: detail.calculation?.rounded_attended_minutes ?? 0,
      fundedDeductionMinor: detail.funded_deduction_minor,
      extrasMinor: detail.calculation?.extras_total_minor ?? 0,
      subtotalMinor: detail.subtotal_minor,
      netDueMinor: detail.total_due_minor,
      lines: (detail.lines ?? [])
        .sort((a, b) => a.sort_order - b.sort_order)
        .map((l) => this.toDraftLine(l)),
      invoiceNumber: detail.invoice_number,
      issuedAt: detail.issued_at,
      generationAction: generatedAction,
    };
  }

  private toDraftLine(l: InvoiceLineApi): InvoiceDraftLine {
    return {
      kind: l.line_kind,
      description: l.description,
      quantityMinutes: l.quantity_minutes ?? 0,
      unitAmountMinor: l.unit_amount_minor,
      lineAmountMinor: l.line_amount_minor,
    };
  }

  private toIssueResult(res: BulkIssueResponseApi): IssueResultSummary {
    return {
      billingMonth: res.billing_month,
      issuedCount: res.summary.issued_count,
      totalIssuedMinor: res.summary.total_due_minor,
      issued: res.issued.map((i) => this.toIssuedResult(i)),
      skipped: res.blocked.map((b) => this.toIssueException(b)),
    };
  }

  private toIssuedResult(i: BulkIssueIssuedApi): IssuedInvoiceResult {
    return {
      invoiceId: i.invoice_id,
      childId: i.child_id,
      childName: this.childName(i),
      invoiceNumber: i.invoice_number,
      issuedAt: i.issued_at,
      totalMinor: i.total_due_minor,
    };
  }

  private toIssueException(b: BulkIssueBlockedApi): IssueException {
    return {
      invoiceId: b.invoice_id,
      childName: this.childName(b),
      reason: b.blockers.map((bl) => bl.message).join('; ') || 'Blocked',
    };
  }

  private toSingleIssueResult(issueRes: SingleIssueResponseApi, detail: InvoiceDetailApi): IssueResultSummary {
    return {
      billingMonth: detail.billing_month,
      issuedCount: 1,
      totalIssuedMinor: issueRes.total_due_minor,
      issued: [
        {
          invoiceId: detail.invoice_id,
          childId: detail.child_id,
          childName: this.childName(detail),
          invoiceNumber: issueRes.invoice_number,
          issuedAt: issueRes.issued_at,
          totalMinor: issueRes.total_due_minor,
        },
      ],
      skipped: [],
    };
  }

  private childName(child: ChildNameApi): string {
    return formatChildName({
      firstName: child.child_first_name,
      middleName: child.child_middle_name,
      lastName: child.child_last_name,
    });
  }
}
