export type ManagerInvoiceStatus = 'draft' | 'issued' | 'payment_failed' | 'paid' | 'overdue';

export type ManagerInvoiceStatusFilter = 'all' | ManagerInvoiceStatus;

export type ManagerInvoiceDueStatus = 'not_due' | 'due' | 'overdue' | 'paid';

export interface ManagerInvoicePeriod {
  startDate: string;
  endDate: string;
}

export interface ManagerInvoiceListItem {
  invoiceId: string;
  invoiceKind: string;
  invoiceNumber: string | null;
  invoiceNumberDisplay: string;
  childId: string;
  childName: string;
  billingMonth: string;
  period: ManagerInvoicePeriod | null;
  status: ManagerInvoiceStatus;
  dueStatus: ManagerInvoiceDueStatus;
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
  generatedRunId: string | null;
  generatedRunStatus: string | null;
  generatedRunStartedAt: string | null;
  generatedRunCompletedAt: string | null;
  generatedRunExceptionCount: number | null;
  createdAt: string;
  updatedAt: string;
}

export interface ManagerInvoiceListResult {
  items: ManagerInvoiceListItem[];
  total: number;
  page: number;
  page_size: number;
}

export interface ManagerInvoiceGeneratedRunException {
  childId: string;
  childName: string;
  blockerCodes: string[];
}

export interface ManagerInvoiceCalculation {
  siteCoreHourlyRateMinor: number | null;
  rawAttendedMinutes: number | null;
  roundedAttendedMinutes: number | null;
  fundedAllowanceMinutes: number | null;
  fundedDeductionMinutes: number | null;
  coreBillableMinutes: number | null;
  includedSessionCount: number | null;
  coreSubtotalMinor: number | null;
  extrasTotalMinor: number | null;
}

export interface ManagerInvoiceLine {
  lineId: string;
  lineKind: string;
  description: string;
  sortOrder: number;
  quantityMinutes: number | null;
  unitAmountMinor: number | null;
  lineAmountMinor: number;
  rawAttendedMinutes: number | null;
  roundedAttendedMinutes: number | null;
  fundedAllowanceMinutes: number | null;
  fundedDeductionMinutes: number | null;
  coreBillableMinutes: number | null;
  sessionCount: number | null;
  fundingModel: string | null;
}

export interface ManagerInvoiceSiteProfile {
  nursery_name: string;
  phone: string;
  email: string;
  website: string;
  address_street: string;
  address_city: string;
  address_postcode: string;
}

export interface ManagerInvoiceParentContact {
  fullName: string;
  addressLine1: string;
  addressLine2: string;
  addressCity: string;
  addressPostcode: string;
  email: string;
  telephone: string;
}

export interface ManagerInvoiceDetail {
  invoiceId: string;
  invoiceKind: string;
  invoiceNumber: string | null;
  invoiceNumberDisplay: string;
  childId: string;
  childName: string;
  billingMonth: string;
  period: ManagerInvoicePeriod | null;
  status: ManagerInvoiceStatus;
  dueStatus: ManagerInvoiceDueStatus;
  currencyCode: string;
  subtotalMinor: number;
  fundedDeductionMinor: number;
  totalDueMinor: number;
  amountPaidMinor: number;
  issuedAt: string | null;
  lockedAt: string | null;
  dueAt: string | null;
  paidAt: string | null;
  paymentFailedAt: string | null;
  paymentStatusUpdatedAt: string | null;
  adjustsInvoiceId: string | null;
  adjustmentReasonCode: string | null;
  adjustmentReasonNote: string | null;
  generatedRunId: string | null;
  generatedRunStatus: string | null;
  generatedRunStartedAt: string | null;
  generatedRunCompletedAt: string | null;
  generatedRunExceptionCount: number | null;
  generatedRunExceptions: ManagerInvoiceGeneratedRunException[];
  site_profile: ManagerInvoiceSiteProfile | null;
  roomName: string | null;
  parentContact: ManagerInvoiceParentContact | null;
  calculation: ManagerInvoiceCalculation | null;
  lines: ManagerInvoiceLine[];
  createdAt: string;
  updatedAt: string;
}

export interface PaymentAttempt {
  paymentAttemptId: string;
  status: string;
  amountMinor: number;
  currencyCode: string;
  stripeCheckoutSessionId: string | null;
  stripePaymentIntentId: string | null;
  stripeExpiresAt: string | null;
  failureReason: string | null;
  providerErrorCode: string | null;
  providerErrorMessage: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface PaymentEvent {
  paymentEventId: string;
  paymentAttemptId: string;
  stripeEventId: string | null;
  stripeEventType: string | null;
  stripeCheckoutSessionId: string | null;
  stripePaymentIntentId: string | null;
  outcome: string;
  reasonCode: string;
  previousInvoiceStatus: string | null;
  newInvoiceStatus: string | null;
  attemptPreviousStatus: string | null;
  attemptNewStatus: string | null;
  amountMinor: number | null;
  currencyCode: string | null;
  webhookProcessingStatus: string;
  webhookProcessingReason: string | null;
  webhookReceivedAt: string | null;
  webhookProcessedAt: string | null;
  createdAt: string;
}

export interface ManagerPaymentStatus {
  invoiceId: string;
  status: string;
  dueStatus: string;
  currencyCode: string;
  totalDueMinor: number;
  amountPaidMinor: number;
  paidAt: string | null;
  paymentFailedAt: string | null;
  paymentStatusUpdatedAt: string | null;
  checkoutRetryAvailable: boolean;
  checkoutRetryReasonCode: string;
  latestPaymentAttempt: PaymentAttempt | null;
  latestPaymentEvent: PaymentEvent | null;
}

export interface PaginatedPaymentEvents {
  items: PaymentEvent[];
  limit: number;
  offset: number;
}
