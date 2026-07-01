export type ParentInvoiceStatus = 'issued' | 'payment_failed' | 'paid' | 'overdue';

export type ParentInvoiceDueStatus = 'due' | 'overdue' | 'paid';

export interface ParentInvoicePeriod {
  startDate: string;
  endDate: string;
}

export interface ParentInvoiceListItem {
  invoiceId: string;
  invoiceKind: string;
  invoiceNumber: string | null;
  invoiceNumberDisplay: string;
  childId: string;
  childName: string;
  billingMonth: string;
  period: ParentInvoicePeriod | null;
  status: ParentInvoiceStatus;
  dueStatus: ParentInvoiceDueStatus;
  currencyCode: string;
  subtotalMinor: number;
  fundedDeductionMinor: number;
  totalDueMinor: number;
  amountPaidMinor: number;
  dueAt: string | null;
  issuedAt: string | null;
  paidAt: string | null;
  paymentFailedAt: string | null;
  paymentStatusUpdatedAt: string | null;
}

export interface ParentInvoiceListResult {
  items: ParentInvoiceListItem[];
  limit: number;
  offset: number;
}

export interface ParentInvoiceCalculation {
  rawAttendedMinutes: number | null;
  roundedAttendedMinutes: number | null;
  fundedAllowanceMinutes: number | null;
  fundedDeductionMinutes: number | null;
  coreBillableMinutes: number | null;
  includedSessionCount: number | null;
  siteCoreHourlyRateMinor: number | null;
  coreSubtotalMinor: number | null;
  extrasTotalMinor: number | null;
}

export interface ParentInvoiceLine {
  lineKind: string;
  description: string;
  sortOrder: number;
  quantityMinutes: number | null;
  unitAmountMinor: number | null;
  lineAmountMinor: number;
}

export interface ParentInvoiceSiteProfile {
  nursery_name: string;
  phone: string;
  email: string;
  website: string;
  address_street: string;
  address_city: string;
  address_postcode: string;
}

export interface ParentInvoiceDetail {
  invoiceId: string;
  invoiceKind: string;
  invoiceNumber: string | null;
  invoiceNumberDisplay: string;
  childId: string;
  childName: string;
  billingMonth: string;
  period: ParentInvoicePeriod | null;
  status: ParentInvoiceStatus;
  dueStatus: ParentInvoiceDueStatus;
  currencyCode: string;
  subtotalMinor: number;
  fundedDeductionMinor: number;
  totalDueMinor: number;
  amountPaidMinor: number;
  issuedAt: string | null;
  dueAt: string | null;
  paidAt: string | null;
  paymentFailedAt: string | null;
  paymentStatusUpdatedAt: string | null;
  site_profile: ParentInvoiceSiteProfile | null;
  calculation: ParentInvoiceCalculation | null;
  lines: ParentInvoiceLine[];
}

export interface CheckoutSessionResult {
  checkoutSessionId: string;
  checkoutUrl: string;
  paymentAttemptId: string;
}

export interface ChildInvoiceGroup {
  childId: string;
  childName: string;
  invoices: ParentInvoiceListItem[];
}
