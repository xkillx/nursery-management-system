export const ROLES = {
  manager: 'manager',
  practitioner: 'practitioner',
  parent: 'parent',
} as const;

export type AppRole = (typeof ROLES)[keyof typeof ROLES];

export function defaultRouteForRole(role: AppRole | null): string {
  switch (role) {
    case ROLES.manager:
      return '/staff/manager/children';
    case ROLES.practitioner:
      return '/staff/practitioner/attendance-children';
    case ROLES.parent:
      return '/parent/invoices';
    default:
      return '/signin';
  }
}
