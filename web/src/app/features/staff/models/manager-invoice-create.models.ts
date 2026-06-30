export interface ManagerInvoicePrefillLine {
  lineKind: string;
  description: string;
  sortOrder: number;
  quantityMinutes: number;
  unitAmountMinor: number;
  lineAmountMinor: number;
  fundedAllowanceMinutes: number;
  fundedDeductionMinutes: number;
  coreBillableMinutes: number;
  sessionCount: number;
}

export interface ManagerInvoicePrefillEntitlement {
  fundingProfileId: string | null;
  fundedAllowanceMinutes: number;
  statusLabel: string;
}

export interface ManagerInvoicePrefill {
  childId: string;
  childFirstName: string;
  childMiddleName: string | null;
  childLastName: string | null;
  billingMonth: string;
  entitlementStatus: ManagerInvoicePrefillEntitlement;
  lines: ManagerInvoicePrefillLine[];
  subtotalMinor: number;
  fundedDeductionMinor: number;
  totalDueMinor: number;
  warnings: string[];
}

export interface FormInvoiceLine {
  id: string;
  lineKind: string;
  description: string;
  sortOrder: number;
  quantityMinutes: number;
  unitAmountMinor: number;
  lineAmountMinor: number;
  fundedAllowanceMinutes: number;
  fundedDeductionMinutes: number;
  coreBillableMinutes: number;
  sessionCount: number;
  isFundingOffset: boolean;
}

export interface DraftInvoiceResult {
  invoiceId: string;
  childId: string;
  billingMonth: string;
  status: string;
  lines: DraftLineResult[];
  subtotalMinor: number;
  totalDueMinor: number;
  paymentTerms: string;
  internalNotes: string;
  createdAt: string;
  updatedAt: string;
}

export interface DraftLineResult {
  lineId: string;
  lineKind: string;
  description: string;
  sortOrder: number;
  quantityMinutes: number;
  unitAmountMinor: number;
  lineAmountMinor: number;
}

export interface ChildSearchResult {
  childId: string;
  childFirstName: string;
  childMiddleName: string | null;
  childLastName: string | null;
  parentName: string;
  roomName: string;
  ageGroup: string;
}
