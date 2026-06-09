import { Injectable } from '@angular/core';
import { Observable, of } from 'rxjs';

import {
  InvoiceRunBlocker,
  InvoiceRunBlockerCode,
  InvoiceRunEligibleChild,
  InvoiceRunException,
  InvoiceRunPreflight,
  InvoiceRunPreflightSummary,
  InvoiceDraftLine,
  InvoiceDraftReviewItem,
  DraftGenerationResult,
  IssueResultSummary,
  IssuedInvoiceResult,
  IssueException,
  InvoiceRunStatus,
} from '../models/invoice-run.models';

// FE-21 replaces this service with real /api/v1 calls.

export interface MockMonthState {
  preflight: InvoiceRunPreflight;
  drafts: InvoiceDraftReviewItem[];
  issuedResults: IssuedInvoiceResult[];
}

const BLOCKERS: Record<string, InvoiceRunException[]> = {
  '2026-05': [
    {
      childId: 'child-block-1',
      childName: 'Yusuf Ali',
      blockers: [
        { code: 'incomplete_attendance', detail: '3 sessions missing check-out in May 2026' },
      ],
    },
    {
      childId: 'child-block-2',
      childName: 'Isla Murphy',
      blockers: [
        { code: 'missing_funding_profile', detail: 'No funded-hours allowance set for May 2026' },
        { code: 'missing_core_hourly_rate', detail: 'Core hourly rate not set on enrollment' },
      ],
    },
  ],
  '2026-03': [],
  '2026-01': [
    {
      childId: 'child-block-3',
      childName: 'Grace O\'Connor',
      blockers: [
        { code: 'existing_issued_invoice', detail: 'Invoice INV-202601-001 already issued' },
      ],
    },
  ],
};

const ELIGIBLE: Record<string, InvoiceRunEligibleChild[]> = {
  '2026-05': [
    { childId: 'child-1', childName: 'Amira Hassan', attendedMinutes: 2760, fundedDeductionMinor: 12000, totalDueMinor: 33600 },
    { childId: 'child-2', childName: 'Arjun Patel', attendedMinutes: 2640, fundedDeductionMinor: 9000, totalDueMinor: 36000 },
    { childId: 'child-3', childName: 'Emma Chen', attendedMinutes: 2520, fundedDeductionMinor: 0, totalDueMinor: 42000 },
    { childId: 'child-4', childName: 'Noah Williams', attendedMinutes: 0, fundedDeductionMinor: 0, totalDueMinor: 0 },
  ],
  '2026-03': [
    { childId: 'child-1', childName: 'Amira Hassan', attendedMinutes: 2880, fundedDeductionMinor: 12000, totalDueMinor: 34800 },
    { childId: 'child-2', childName: 'Arjun Patel', attendedMinutes: 2700, fundedDeductionMinor: 9000, totalDueMinor: 37500 },
  ],
  '2026-01': [],
};

const SUMMARIES: Record<string, InvoiceRunPreflightSummary> = {
  '2026-05': {
    totalChildren: 6,
    eligibleChildren: 4,
    blockedChildren: 2,
    sessionsIncluded: 44,
    attendedMinutes: 7920,
    fundedDeductionMinor: 21000,
    totalDueMinor: 111600,
  },
  '2026-03': {
    totalChildren: 2,
    eligibleChildren: 2,
    blockedChildren: 0,
    sessionsIncluded: 22,
    attendedMinutes: 5580,
    fundedDeductionMinor: 21000,
    totalDueMinor: 72300,
  },
  '2026-01': {
    totalChildren: 1,
    eligibleChildren: 0,
    blockedChildren: 1,
    sessionsIncluded: 0,
    attendedMinutes: 0,
    fundedDeductionMinor: 0,
    totalDueMinor: 0,
  },
};

function buildLines(child: InvoiceRunEligibleChild): InvoiceDraftLine[] {
  const ratePerMinute = 25; // £0.25/min = £15/hr
  const childcareAmount = child.attendedMinutes * ratePerMinute;

  return [
    {
      kind: 'core_childcare',
      description: 'Core childcare',
      quantityMinutes: child.attendedMinutes,
      unitAmountMinor: ratePerMinute,
      lineAmountMinor: childcareAmount,
    },
    {
      kind: 'funded_deduction',
      description: 'Funded-hours deduction',
      quantityMinutes: Math.round(child.fundedDeductionMinor / ratePerMinute),
      unitAmountMinor: ratePerMinute,
      lineAmountMinor: -child.fundedDeductionMinor,
    },
    {
      kind: 'extra',
      description: 'Extras',
      quantityMinutes: 0,
      unitAmountMinor: null,
      lineAmountMinor: 0,
    },
  ];
}

function buildDraft(child: InvoiceRunEligibleChild, billingMonth: string, action: 'created' | 'updated'): InvoiceDraftReviewItem {
  const lines = buildLines(child);
  const subtotal = lines.reduce((sum, l) => sum + l.lineAmountMinor, 0);

  return {
    invoiceId: `inv-${child.childId}-${billingMonth}`,
    childId: child.childId,
    childName: child.childName,
    billingMonth,
    status: 'draft' as InvoiceRunStatus,
    attendedMinutes: child.attendedMinutes,
    fundedDeductionMinor: child.fundedDeductionMinor,
    extrasMinor: 0,
    subtotalMinor: subtotal,
    netDueMinor: child.totalDueMinor,
    lines,
    invoiceNumber: null,
    issuedAt: null,
    generationAction: action,
  };
}

function buildPreflight(month: string): InvoiceRunPreflight {
  return {
    billingMonth: month,
    summary: SUMMARIES[month] ?? {
      totalChildren: 0, eligibleChildren: 0, blockedChildren: 0,
      sessionsIncluded: 0, attendedMinutes: 0, fundedDeductionMinor: 0, totalDueMinor: 0,
    },
    eligibleChildren: ELIGIBLE[month] ?? [],
    blockedChildren: BLOCKERS[month] ?? [],
  };
}

@Injectable({ providedIn: 'root' })
export class InvoiceRunMockService {
  private monthStates = new Map<string, MockMonthState>();

  loadPreflight(billingMonth: string): Observable<InvoiceRunPreflight> {
    return of(buildPreflight(billingMonth));
  }

  generateDrafts(billingMonth: string): Observable<DraftGenerationResult> {
    const eligible = ELIGIBLE[billingMonth] ?? [];
    const blocked = BLOCKERS[billingMonth] ?? [];
    const state = this.getOrCreateState(billingMonth);

    const existingIds = new Set(state.drafts.map(d => d.invoiceId));
    const existingIssuedIds = new Set(state.issuedResults.map(i => i.invoiceId));

    let createdCount = 0;
    let updatedCount = 0;

    for (const child of eligible) {
      const id = `inv-${child.childId}-${billingMonth}`;
      if (existingIssuedIds.has(id)) continue;

      if (existingIds.has(id)) {
        const idx = state.drafts.findIndex(d => d.invoiceId === id);
        state.drafts[idx] = buildDraft(child, billingMonth, 'updated');
        updatedCount++;
      } else {
        state.drafts.push(buildDraft(child, billingMonth, 'created'));
        createdCount++;
      }
    }

    state.drafts.sort((a, b) => a.childName.localeCompare(b.childName));

    return of({
      billingMonth,
      generatedCount: createdCount,
      updatedCount,
      blockedCount: blocked.length,
      blockedChildren: blocked,
    });
  }

  listDrafts(billingMonth: string): Observable<InvoiceDraftReviewItem[]> {
    const state = this.getOrCreateState(billingMonth);
    return of([...state.drafts].sort((a, b) => a.childName.localeCompare(b.childName)));
  }

  bulkIssue(billingMonth: string, invoiceIds: string[]): Observable<IssueResultSummary> {
    const state = this.getOrCreateState(billingMonth);
    const issued: IssuedInvoiceResult[] = [];
    const skipped: IssueException[] = [];
    const now = new Date().toISOString();
    let seq = 1;

    const readyDrafts = state.drafts
      .filter(d => d.status === 'draft' && invoiceIds.includes(d.invoiceId))
      .sort((a, b) => a.childName.localeCompare(b.childName));

    for (const draft of readyDrafts) {
      const num = `INV-${billingMonth.replace('-', '')}-${String(seq).padStart(4, '0')}`;
      draft.status = 'issued';
      draft.invoiceNumber = num;
      draft.issuedAt = now;

      const result: IssuedInvoiceResult = {
        invoiceId: draft.invoiceId,
        childId: draft.childId,
        childName: draft.childName,
        invoiceNumber: num,
        issuedAt: now,
        totalMinor: draft.netDueMinor,
      };

      issued.push(result);
      state.issuedResults.push(result);
      seq++;
    }

    const issuedIds = new Set(issued.map(i => i.invoiceId));
    for (const id of invoiceIds) {
      if (!issuedIds.has(id)) {
        const draft = state.drafts.find(d => d.invoiceId === id);
        skipped.push({
          invoiceId: id,
          childName: draft?.childName ?? 'Unknown',
          reason: draft ? 'Already issued' : 'Not found',
        });
      }
    }

    return of({
      billingMonth,
      issuedCount: issued.length,
      totalIssuedMinor: issued.reduce((s, i) => s + i.totalMinor, 0),
      issued,
      skipped,
    });
  }

  issueOne(invoiceId: string): Observable<IssueResultSummary> {
    for (const [, state] of this.monthStates) {
      const draft = state.drafts.find(d => d.invoiceId === invoiceId && d.status === 'draft');
      if (draft) {
        return this.bulkIssue(draft.billingMonth, [invoiceId]);
      }
    }
    return of({
      billingMonth: '',
      issuedCount: 0,
      totalIssuedMinor: 0,
      issued: [],
      skipped: [{ invoiceId, childName: 'Unknown', reason: 'Not found or not in draft status' }],
    });
  }

  resetMonth(billingMonth: string): void {
    this.monthStates.delete(billingMonth);
  }

  private getOrCreateState(billingMonth: string): MockMonthState {
    let state = this.monthStates.get(billingMonth);
    if (!state) {
      state = { preflight: buildPreflight(billingMonth), drafts: [], issuedResults: [] };
      this.monthStates.set(billingMonth, state);
    }
    return state;
  }
}
