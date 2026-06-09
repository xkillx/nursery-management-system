export interface AttendanceSummary {
  checkedInToday: number;
  notInYet: number;
  enrollmentIncomplete: number;
  incompleteAttendance: number;
}

export interface IncompleteAttendanceItem {
  id: string;
  childName: string;
  localDateLabel: string;
  issue: string;
  actionHint: string;
  isToday: boolean;
  childId?: string;
  localDate?: string;
  sessionId?: string;
}

export interface InvoiceRunStatus {
  billingMonthLabel: string;
  eligibleChildren: number;
  blockedChildren: number;
  draftInvoices: number;
  issuedInvoices: number;
  lastRunLabel: string;
  nextStep: string;
}

export type PaymentFollowUpStatus = 'overdue' | 'payment_failed' | 'issued';

export interface PaymentFollowUpInvoice {
  id: string;
  invoiceNumber: string;
  childName: string;
  payerName: string;
  status: PaymentFollowUpStatus;
  dueDateLabel: string;
  outstandingMinor: number;
}

export interface QuickAction {
  label: string;
  description: string;
  route?: string;
  disabled?: boolean;
}

export interface ManagerDashboardSnapshot {
  attendanceSummary: AttendanceSummary;
  incompleteAttendance: IncompleteAttendanceItem[];
  invoiceRunStatus: InvoiceRunStatus;
  paymentFollowUp: PaymentFollowUpInvoice[];
  quickActions: QuickAction[];
}

const PAYMENT_URGENCY: Record<PaymentFollowUpStatus, number> = {
  overdue: 0,
  payment_failed: 1,
  issued: 2,
};

export function sortPaymentFollowUp(items: PaymentFollowUpInvoice[]): PaymentFollowUpInvoice[] {
  return [...items].sort(
    (a, b) => PAYMENT_URGENCY[a.status] - PAYMENT_URGENCY[b.status],
  );
}

export function formatGbp(minorUnits: number): string {
  const pounds = minorUnits / 100;
  return `£${pounds.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

export const MANAGER_DASHBOARD_MOCK: ManagerDashboardSnapshot = {
  attendanceSummary: {
    checkedInToday: 14,
    notInYet: 3,
    enrollmentIncomplete: 1,
    incompleteAttendance: 4,
  },

  incompleteAttendance: [
    {
      id: 'inc-1',
      childName: 'Liam Okafor',
      localDateLabel: 'Today',
      issue: 'Missing check-out',
      actionHint: 'Needs manager correction',
      isToday: true,
    },
    {
      id: 'inc-2',
      childName: 'Sofia Ahmed',
      localDateLabel: 'Today',
      issue: 'Missing check-out',
      actionHint: 'Needs manager correction',
      isToday: true,
    },
    {
      id: 'inc-3',
      childName: 'Emma Chen',
      localDateLabel: '3 Jun 2026',
      issue: 'Missing check-out',
      actionHint: 'Needs manager correction',
      isToday: false,
    },
    {
      id: 'inc-4',
      childName: 'Noah Williams',
      localDateLabel: '1 Jun 2026',
      issue: 'No check-in or check-out recorded',
      actionHint: 'Needs manager correction',
      isToday: false,
    },
  ],

  invoiceRunStatus: {
    billingMonthLabel: 'June 2026',
    eligibleChildren: 22,
    blockedChildren: 2,
    draftInvoices: 18,
    issuedInvoices: 2,
    lastRunLabel: '4 Jun 2026 at 09:15',
    nextStep: 'Review blocked children before issuing remaining drafts',
  },

  paymentFollowUp: [
    {
      id: 'pay-1',
      invoiceNumber: 'INV-2026-0047',
      childName: 'Mia Thompson',
      payerName: 'Sarah Thompson',
      status: 'overdue',
      dueDateLabel: '28 May 2026',
      outstandingMinor: 45000,
    },
    {
      id: 'pay-2',
      invoiceNumber: 'INV-2026-0051',
      childName: 'Arjun Patel',
      payerName: 'Deepa Patel',
      status: 'overdue',
      dueDateLabel: '1 Jun 2026',
      outstandingMinor: 45000,
    },
    {
      id: 'pay-3',
      invoiceNumber: 'INV-2026-0039',
      childName: 'Oliver Brown',
      payerName: 'James Brown',
      status: 'payment_failed',
      dueDateLabel: '4 Jun 2026',
      outstandingMinor: 22500,
    },
    {
      id: 'pay-4',
      invoiceNumber: 'INV-2026-0055',
      childName: 'Amira Hassan',
      payerName: 'Fatima Hassan',
      status: 'payment_failed',
      dueDateLabel: '4 Jun 2026',
      outstandingMinor: 45000,
    },
    {
      id: 'pay-5',
      invoiceNumber: 'INV-2026-0058',
      childName: 'Lucas Davies',
      payerName: 'Helen Davies',
      status: 'issued',
      dueDateLabel: '12 Jun 2026',
      outstandingMinor: 45000,
    },
  ],

  quickActions: [
    {
      label: 'Open attendance',
      description: 'Check-in and check-out for today',
      route: '/staff/practitioner/attendance',
    },
    {
      label: 'Attendance corrections',
      description: 'Review and correct attendance records',
      route: '/staff/manager/attendance-corrections',
    },
    {
      label: 'Manage children',
      description: 'Add, edit, and manage child records',
      route: '/staff/manager/children',
    },
    {
      label: 'Manage guardians',
      description: 'Add, edit, and manage guardian records',
      route: '/staff/manager/guardians',
    },
    {
      label: 'Start invoice run',
      description: 'Generate invoices for the current billing month',
      route: '/staff/manager/invoice-run',
    },
    {
      label: 'Review payment follow-up',
      description: 'Chase overdue and failed payments',
      disabled: true,
    },
  ],
};
