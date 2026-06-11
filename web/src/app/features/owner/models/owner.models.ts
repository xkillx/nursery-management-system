export interface OwnerAttendanceSummary {
  checkedInTodayCount: number;
  incompleteAttendanceCount: number;
}

export interface OwnerFundingReadiness {
  includedChildCount: number;
  flaggedChildCount: number;
  missingProfileCount: number;
  explicitZeroCount: number;
  underOneHourCount: number;
  above160HoursCount: number;
}

export interface OwnerInvoicePaymentHealth {
  draftCount: number;
  issuedCount: number;
  overdueCount: number;
  paymentFailedCount: number;
  paidCount: number;
  totalIssuedMinor: number;
  totalPaidMinor: number;
  outstandingMinor: number;
  overdueOutstandingMinor: number;
  failedPaymentCount: number;
}

export interface OwnerSiteSummary {
  siteId: string;
  siteName: string;
  setupStatus: string;
  activeManagerCount: number;
  pendingManagerInviteCount: number;
  activeChildrenCount: number;
  attendance: OwnerAttendanceSummary;
  fundingReadiness: OwnerFundingReadiness;
  invoicePaymentHealth: OwnerInvoicePaymentHealth;
}

export interface OwnerSiteSummaryTotals {
  activeManagerCount: number;
  pendingManagerInviteCount: number;
  activeChildrenCount: number;
  checkedInTodayCount: number;
  incompleteAttendanceCount: number;
  draftCount: number;
  issuedCount: number;
  overdueCount: number;
  paymentFailedCount: number;
  paidCount: number;
  totalIssuedMinor: number;
  totalPaidMinor: number;
  outstandingMinor: number;
  overdueOutstandingMinor: number;
}

export interface OwnerSiteSummariesResponse {
  billingMonth: string;
  attendanceLocalDate: string;
  currencyCode: string;
  totals: OwnerSiteSummaryTotals;
  sites: OwnerSiteSummary[];
}

export interface OwnerManagerAccessRecord {
  membershipId: string;
  userId: string;
  email: string;
  isActive: boolean;
}

export type OwnerGrantOutcome =
  | 'manager_membership_granted'
  | 'manager_membership_reactivated'
  | 'manager_membership_already_active'
  | 'manager_invite_pending'
  | 'granted'
  | 'reactivated'
  | 'already_active'
  | 'invite_pending';

export interface OwnerGrantInviteDetails {
  email: string;
  expiresAt: string;
}

export interface OwnerGrantManagerAccessResult {
  outcome: OwnerGrantOutcome;
  membershipId: string | null;
  invite: OwnerGrantInviteDetails | null;
}
