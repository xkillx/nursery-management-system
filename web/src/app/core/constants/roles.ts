export const ROLES = {
  manager: 'manager',
  practitioner: 'practitioner',
  parent: 'parent',
} as const;

export type AppRole = (typeof ROLES)[keyof typeof ROLES];

export const ROLE_ROUTES = {
  managerDashboard: '/staff/manager/dashboard',
  managerChildren: '/staff/manager/children',
  managerGuardians: '/staff/manager/guardians',
  managerInvites: '/staff/manager/invites',
  managerFunding: '/staff/manager/funding',
  managerAttendanceCorrections: '/staff/manager/attendance-corrections',
  managerInvoiceRun: '/staff/manager/invoice-run',
  practitionerAttendance: '/staff/practitioner/attendance',
  practitionerAttendanceLegacy: '/staff/practitioner/attendance-children',
  parentInvoices: '/parent/invoices',
  signIn: '/signin',
} as const;

export function defaultRouteForRole(role: AppRole | null): string {
  switch (role) {
    case ROLES.manager:
      return ROLE_ROUTES.managerDashboard;
    case ROLES.practitioner:
      return ROLE_ROUTES.practitionerAttendance;
    case ROLES.parent:
      return ROLE_ROUTES.parentInvoices;
    default:
      return ROLE_ROUTES.signIn;
  }
}
