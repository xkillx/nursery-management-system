export type InvoiceRunStep = 'preflight' | 'drafts' | 'review' | 'issue' | 'result';

export type InvoiceRunBlockerCode = string;

export const InvoiceRunBlockerCodes = {
  // Preflight blockers
  MISSING_CHILD_NAME: 'missing_child_name',
  MISSING_CHILD_DATE_OF_BIRTH: 'missing_child_date_of_birth',
  MISSING_CHILD_START_DATE: 'missing_child_start_date',
  MISSING_PARENT_CARER_CONTACT: 'missing_parent_carer_contact',
  MISSING_BILLING_RATE: 'missing_billing_rate',
  MISSING_FUNDING_PROFILE: 'missing_funding_profile',
  INCOMPLETE_ATTENDANCE: 'incomplete_attendance',
  INVOICE_ALREADY_ISSUED: 'invoice_already_issued',
  // Generation blockers
  CHILD_NOT_FOUND: 'child_not_found',
  CHILD_NOT_IN_BILLING_MONTH: 'child_not_in_billing_month',
  // Issue blockers
  INVOICE_NOT_FOUND: 'invoice_not_found',
  INVOICE_NOT_IN_BILLING_MONTH: 'invoice_not_in_billing_month',
  INVOICE_NOT_DRAFT: 'invoice_not_draft',
  INVOICE_NOT_MONTHLY: 'invoice_not_monthly',
} as const;

export type InvoiceDraftLineKind = string;

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
  generationAction: 'created' | 'updated' | null;
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
