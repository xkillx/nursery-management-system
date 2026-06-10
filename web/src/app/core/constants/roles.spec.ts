import { ROLES, ROLE_ROUTES, defaultRouteForRole } from './roles';

describe('roles constants', () => {
  it('maps manager to managerDashboard route', () => {
    expect(defaultRouteForRole(ROLES.manager)).toBe(ROLE_ROUTES.managerDashboard);
    expect(defaultRouteForRole(ROLES.manager)).toBe('/staff/manager/dashboard');
  });

  it('maps practitioner to practitionerAttendance route', () => {
    expect(defaultRouteForRole(ROLES.practitioner)).toBe(ROLE_ROUTES.practitionerAttendance);
    expect(defaultRouteForRole(ROLES.practitioner)).toBe('/staff/practitioner/attendance');
  });

  it('maps parent to parentInvoices route', () => {
    expect(defaultRouteForRole(ROLES.parent)).toBe(ROLE_ROUTES.parentInvoices);
    expect(defaultRouteForRole(ROLES.parent)).toBe('/parent/invoices');
  });

  it('maps null role to signIn route', () => {
    expect(defaultRouteForRole(null)).toBe(ROLE_ROUTES.signIn);
    expect(defaultRouteForRole(null)).toBe('/signin');
  });
});
