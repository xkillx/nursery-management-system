export type InvoiceRunStep = 'preflight' | 'drafts' | 'review' | 'issue' | 'result';

export type InvoiceRunBlockerCode =
  | 'incomplete_attendance'
  | 'missing_funding_profile'
  | 'missing_core_hourly_rate'
  | 'missing_guardian_link'
  | 'existing_issued_invoice';

export type InvoiceDraftLineKind = 'core_childcare' | 'funded_deduction' | 'extra';

export type InvoiceRunStatus = 'draft' | 'issued';

export interface InvoiceRunBlocker {
  code: InvoiceRunBlockerCode;
  detail: string;
}

export interface InvoiceRunException {
  childId: string;
  childName: string;
  blockers: InvoiceRunBlocker[];
}

export interface InvoiceRunPreflightSummary {
  totalChildren: number;
  eligibleChildren: number;
  blockedChildren: number;
  sessionsIncluded: number;
  attendedMinutes: number;
  fundedDeductionMinor: number;
  totalDueMinor: number;
}

export interface InvoiceRunEligibleChild {
  childId: string;
  childName: string;
  attendedMinutes: number;
  fundedDeductionMinor: number;
  totalDueMinor: number;
}

export interface InvoiceRunPreflight {
  billingMonth: string;
  summary: InvoiceRunPreflightSummary;
  eligibleChildren: InvoiceRunEligibleChild[];
  blockedChildren: InvoiceRunException[];
}

export interface InvoiceDraftLine {
  kind: InvoiceDraftLineKind;
  description: string;
  quantityMinutes: number;
  unitAmountMinor: number | null;
  lineAmountMinor: number;
}

export interface InvoiceDraftReviewItem {
  invoiceId: string;
  childId: string;
  childName: string;
  billingMonth: string;
  status: InvoiceRunStatus;
  attendedMinutes: number;
  fundedDeductionMinor: number;
  extrasMinor: number;
  subtotalMinor: number;
  netDueMinor: number;
  lines: InvoiceDraftLine[];
  invoiceNumber: string | null;
  issuedAt: string | null;
  generationAction: 'created' | 'updated';
}

export interface DraftGenerationResult {
  billingMonth: string;
  generatedCount: number;
  updatedCount: number;
  blockedCount: number;
  blockedChildren: InvoiceRunException[];
}

export interface IssueResultSummary {
  billingMonth: string;
  issuedCount: number;
  totalIssuedMinor: number;
  issued: IssuedInvoiceResult[];
  skipped: IssueException[];
}

export interface IssuedInvoiceResult {
  invoiceId: string;
  childId: string;
  childName: string;
  invoiceNumber: string;
  issuedAt: string;
  totalMinor: number;
}

export interface IssueException {
  invoiceId: string;
  childName: string;
  reason: string;
}
