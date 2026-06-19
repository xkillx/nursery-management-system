export const ROLES = {
  owner: 'owner',
  manager: 'manager',
  practitioner: 'practitioner',
  parent: 'parent',
} as const;

export type AppRole = (typeof ROLES)[keyof typeof ROLES];

export const ROLE_ROUTES = {
  ownerHome: '/owner',
  ownerManagerAccess: '/owner/manager-access',
  ownerRooms: '/owner/rooms',
  ownerSessionTypes: '/owner/session-types',
  managerDashboard: '/staff/manager/dashboard',
  managerChildren: '/staff/manager/children',
  managerGuardians: '/staff/manager/guardians',
  managerInvites: '/staff/manager/invites',
  managerRooms: '/staff/manager/rooms',
  managerSessionTypes: '/staff/manager/session-types',
  managerFunding: '/staff/manager/funding',
  managerAttendanceCorrections: '/staff/manager/attendance-corrections',
  managerInvoiceRun: '/staff/manager/invoice-run',
  managerInvoices: '/staff/manager/invoices',
  practitionerAttendance: '/staff/practitioner/attendance',
  practitionerAttendanceLegacy: '/staff/practitioner/attendance-children',
  parentInvoices: '/app/invoices',
  signIn: '/signin',
} as const;

export function defaultRouteForRole(role: AppRole | null): string {
  switch (role) {
    case ROLES.owner:
      return ROLE_ROUTES.ownerHome;
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
