import { Routes } from '@angular/router';
import { routes } from './app.routes';

function flattenPaths(routes: Routes, parentPath = ''): string[] {
  const paths: string[] = [];
  for (const route of routes) {
    const fullPath = parentPath ? `${parentPath}/${route.path}` : (route.path ?? '');
    if (route.path !== undefined && route.path !== '') {
      paths.push(fullPath);
    }
    if (route.children) {
      paths.push(...flattenPaths(route.children, fullPath));
    }
  }
  return paths;
}

describe('app.routes', () => {
  const paths = flattenPaths(routes);

  const removedDemoPaths = [
    'calendar', 'profile', 'form-elements', 'basic-tables',
    'blank', 'invoice', 'line-chart', 'bar-chart', 'alerts', 'avatars',
    'badge', 'buttons', 'images', 'videos',
  ];

  for (const demo of removedDemoPaths) {
    it(`does not register demo route /${demo}`, () => {
      expect(paths).not.toContain(demo);
    });
  }

  const mvpPaths = [
    'staff/manager/dashboard',
    'staff/manager/children',
    'staff/manager/guardians',
    'staff/manager/invites',
    'staff/manager/funding',
    'staff/manager/invoice-run',
    'staff/manager/invoices',
    'staff/practitioner/attendance',
    'staff/practitioner/attendance-children',
    'parent/invoices',
    'signin',
    'signup',
    'forgot-password',
    'reset-password',
    'invite-accept',
  ];

  const dynamicPaths = [
    'staff/manager/children/:childId',
    'staff/manager/invoices/:invoiceId',
  ];

  for (const mvp of mvpPaths) {
    it(`registers MVP route /${mvp}`, () => {
      expect(paths).toContain(mvp);
    });
  }

  for (const dynamic of dynamicPaths) {
    it(`registers dynamic MVP route /${dynamic}`, () => {
      const allPaths = routes.flatMap(r => r.children ?? []);
      const found = allPaths.some(r => r.path === dynamic);
      expect(found).toBeTrue();
    });
  }

  it('child detail route requires manager role only', () => {
    const detailRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/manager/children/:childId');

    expect(detailRoute).toBeDefined();
    expect(detailRoute!.data?.['roles']).toEqual(['manager']);
  });

  it('legacy attendance-children route is a redirect, not a component route', () => {
    const legacyRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/practitioner/attendance-children');

    expect(legacyRoute).toBeDefined();
    expect(legacyRoute!.redirectTo).toBe('staff/practitioner/attendance');
    expect(legacyRoute!.component).toBeUndefined();
  });

  it('manager invites route requires manager role only', () => {
    const invitesRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/manager/invites');

    expect(invitesRoute).toBeDefined();
    expect(invitesRoute!.data?.['roles']).toEqual(['manager']);
  });

  it('funding overview route requires manager role only', () => {
    const fundingRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/manager/funding');

    expect(fundingRoute).toBeDefined();
    expect(fundingRoute!.data?.['roles']).toEqual(['manager']);
  });

  it('invoice run route requires manager role only', () => {
    const invoiceRunRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/manager/invoice-run');

    expect(invoiceRunRoute).toBeDefined();
    expect(invoiceRunRoute!.data?.['roles']).toEqual(['manager']);
  });

  it('invoices list route requires manager role only', () => {
    const invoicesRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/manager/invoices');

    expect(invoicesRoute).toBeDefined();
    expect(invoicesRoute!.data?.['roles']).toEqual(['manager']);
  });

  it('invoice detail route requires manager role only', () => {
    const detailRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/manager/invoices/:invoiceId');

    expect(detailRoute).toBeDefined();
    expect(detailRoute!.data?.['roles']).toEqual(['manager']);
  });
});
