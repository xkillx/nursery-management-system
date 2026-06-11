export function formatGbp(minorUnits: number): string {
  const pounds = minorUnits / 100;
  return `£${pounds.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

const knownStatusLabels: Record<string, string> = {
  complete: 'Complete',
  incomplete_attendance: 'Incomplete attendance',
  incomplete_funding: 'Incomplete funding',
  incomplete_setup: 'Incomplete setup',
};

export function formatSetupStatus(status: string): string {
  if (knownStatusLabels[status]) {
    return knownStatusLabels[status];
  }
  return status.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export function formatGrantOutcome(outcome: string): string {
  const labels: Record<string, string> = {
    manager_membership_granted: 'Manager access granted.',
    granted: 'Manager access granted.',
    manager_membership_reactivated: 'Manager access reactivated.',
    reactivated: 'Manager access reactivated.',
    manager_membership_already_active: 'Manager already has active access.',
    already_active: 'Manager already has active access.',
    manager_invite_pending: 'Manager invite sent. The user will receive an email to set up their account.',
    invite_pending: 'Manager invite sent. The user will receive an email to set up their account.',
  };
  return labels[outcome] ?? outcome.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export function isExceptionSite(site: {
  activeManagerCount: number;
  invoicePaymentHealth: { overdueOutstandingMinor: number; paymentFailedCount: number; outstandingMinor: number };
  attendance: { incompleteAttendanceCount: number };
  fundingReadiness: { flaggedChildCount: number };
}): boolean {
  if (site.activeManagerCount === 0) return true;
  if (site.invoicePaymentHealth.overdueOutstandingMinor > 0) return true;
  if (site.invoicePaymentHealth.paymentFailedCount > 0) return true;
  if (site.invoicePaymentHealth.outstandingMinor > 0) return true;
  if (site.attendance.incompleteAttendanceCount > 0) return true;
  if (site.fundingReadiness.flaggedChildCount > 0) return true;
  return false;
}
