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
  managerDashboard: '/manager/dashboard',
  managerChildren: '/manager/children',
  managerInvites: '/manager/invites',
  managerRooms: '/manager/rooms',
  managerSessionTypes: '/manager/session-types',
  managerSessionTemplates: '/manager/session-templates',
  managerFunding: '/manager/funding',
  managerBillingSetup: '/manager/billing-setup',
  managerAttendanceCorrections: '/manager/attendance-corrections',
  managerInvoiceRun: '/manager/invoice-run',
  managerInvoices: '/manager/invoices',
  managerAttendance: '/manager/attendance',
  practitionerAttendance: '/practitioner/attendance',
  practitionerAttendanceLegacy: '/practitioner/attendance-children',
  parentInvoices: '/parent/invoices',
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
