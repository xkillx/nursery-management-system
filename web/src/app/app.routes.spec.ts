import { Routes } from '@angular/router';
import { routes } from './app.routes';
import { authGuard } from './core/guards/auth.guard';
import { roleDefaultRedirectGuard } from './core/guards/role-default-redirect.guard';
import { roleGuard } from './core/guards/role.guard';

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

function findLeafRoute(routes: Routes, path: string) {
  const matches: any[] = [];
  collectByPath(routes, '', matches);
  return matches.find((r) => r.fullPath === path);
}

function collectByPath(routes: Routes, parentPath: string, out: any[]): void {
  for (const r of routes) {
    if (r.path === undefined) continue;
    const fullPath = r.path === '' ? parentPath : parentPath ? `${parentPath}/${r.path}` : r.path;
    if (r.component) {
      out.push({ ...r, fullPath });
    }
    if (r.children) {
      collectByPath(r.children, fullPath, out);
    }
  }
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
    'staff/manager/rooms',
    'staff/manager/rooms/new',
    'staff/practitioner/attendance',
    'staff/practitioner/attendance-children',
    'owner',
    'owner/manager-access',
    'owner/rooms',
    'owner/rooms/new',
    'app/invoices',
    'signin',
    'signup',
    'forgot-password',
    'reset-password',
    'invite-accept',
  ];

  const dynamicPaths = [
    'staff/manager/children/:childId',
    'staff/manager/invoices/:invoiceId',
    'staff/manager/rooms/:roomId/edit',
    'owner/rooms/:roomId/edit',
    'app/invoices/:invoiceId',
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

  it('owner overview route requires owner role only', () => {
    const ownerRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'owner');

    expect(ownerRoute).toBeDefined();
    expect(ownerRoute!.data?.['roles']).toEqual(['owner']);
  });

  it('owner manager-access route requires owner role only', () => {
    const accessRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'owner/manager-access');

    expect(accessRoute).toBeDefined();
    expect(accessRoute!.data?.['roles']).toEqual(['owner']);
  });

  it('owner room routes require owner role only', () => {
    const ownerRoomRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('owner/rooms'));

    expect(ownerRoomRoutes.length).toBe(3);
    for (const route of ownerRoomRoutes) {
      expect(route.data?.['roles']).toEqual(['owner']);
    }
  });

  it('manager room routes require manager role only', () => {
    const managerRoomRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('staff/manager/rooms'));

    expect(managerRoomRoutes.length).toBe(3);
    for (const route of managerRoomRoutes) {
      expect(route.data?.['roles']).toEqual(['manager']);
    }
  });

  it('does not register practitioner room routes', () => {
    expect(paths).not.toContain('staff/practitioner/rooms');
    expect(paths).not.toContain('staff/practitioner/rooms/new');
  });

  it('canonical parent invoices list route requires parent role only', () => {
    const parentRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'app/invoices');

    expect(parentRoute).toBeDefined();
    expect(parentRoute!.data?.['roles']).toEqual(['parent']);
  });

  it('canonical parent invoice detail route requires parent role only', () => {
    const detailRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'app/invoices/:invoiceId');

    expect(detailRoute).toBeDefined();
    expect(detailRoute!.data?.['roles']).toEqual(['parent']);
  });

  it('legacy /parent/invoices redirects to /app/invoices', () => {
    const redirect = routes.find(r => r.path === 'parent/invoices');
    expect(redirect).toBeDefined();
    expect(redirect!.redirectTo).toBe('app/invoices');
    expect(redirect!.component).toBeUndefined();
  });

  it('legacy /parent/invoices/:invoiceId redirects to /app/invoices/:invoiceId', () => {
    const redirect = routes.find(r => r.path === 'parent/invoices/:invoiceId');
    expect(redirect).toBeDefined();
    expect(redirect!.redirectTo).toBe('app/invoices/:invoiceId');
    expect(redirect!.component).toBeUndefined();
  });

  it('root path includes authGuard and roleDefaultRedirectGuard', () => {
    const rootRoute = routes.find(r => r.path === '' && r.pathMatch === 'full');

    expect(rootRoute).toBeDefined();
    expect(rootRoute!.canActivate).toContain(authGuard);
    expect(rootRoute!.canActivate).toContain(roleDefaultRedirectGuard);
  });

  it('protected manager routes include both authGuard and roleGuard', () => {
    const managerRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('staff/manager'));

    for (const route of managerRoutes) {
      if (route.redirectTo) continue;
      expect(route.canActivate).toContain(authGuard);
      expect(route.canActivate).toContain(roleGuard);
    }
  });

  it('protected owner routes include both authGuard and roleGuard', () => {
    const ownerRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('owner'));

    for (const route of ownerRoutes) {
      expect(route.canActivate).toContain(authGuard);
      expect(route.canActivate).toContain(roleGuard);
    }
  });

  it('protected parent routes include both authGuard and roleGuard', () => {
    const parentRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('app/invoices'));

    for (const route of parentRoutes) {
      expect(route.canActivate).toContain(authGuard);
      expect(route.canActivate).toContain(roleGuard);
    }
  });

  it('practitioner attendance route allows both manager and practitioner', () => {
    const attendanceRoute = routes
      .flatMap(r => r.children ?? [])
      .find(r => r.path === 'staff/practitioner/attendance');

    expect(attendanceRoute).toBeDefined();
    expect(attendanceRoute!.data?.['roles']).toEqual(['manager', 'practitioner']);
  });

  it('staff routes do not include parent or owner roles', () => {
    const staffRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('staff') && r.data?.['roles']);

    for (const route of staffRoutes) {
      expect((route.data?.['roles'] as string[])).not.toContain('parent');
      expect((route.data?.['roles'] as string[])).not.toContain('owner');
    }
  });

  it('owner routes do not include manager, practitioner, or parent roles', () => {
    const ownerRoutes = routes
      .flatMap(r => r.children ?? [])
      .filter(r => r.path?.startsWith('owner') && r.data?.['roles']);

    for (const route of ownerRoutes) {
      expect((route.data?.['roles'] as string[])).not.toContain('manager');
      expect((route.data?.['roles'] as string[])).not.toContain('practitioner');
      expect((route.data?.['roles'] as string[])).not.toContain('parent');
    }
  });
});

describe('app.routes breadcrumb wiring', () => {
  const breadcrumbPaths = [
    'staff/manager/dashboard',
    'staff/manager/children',
    'staff/manager/children/:childId',
    'staff/manager/children/:childId/registration',
    'staff/manager/registrations/new',
    'staff/manager/registrations/:childId/intake',
    'staff/manager/guardians',
    'staff/manager/invites',
    'staff/manager/attendance-corrections',
    'staff/manager/rooms',
    'staff/manager/rooms/new',
    'staff/manager/rooms/:roomId/edit',
    'staff/manager/funding',
    'staff/manager/invoice-run',
    'staff/manager/invoices',
    'staff/manager/invoices/:invoiceId',
    'staff/practitioner/attendance',
    'owner',
    'owner/manager-access',
    'owner/rooms',
    'owner/rooms/new',
    'owner/rooms/:roomId/edit',
    'app/invoices',
    'app/invoices/:invoiceId',
  ];

  for (const path of breadcrumbPaths) {
    it(`declares a breadcrumb Crumb on /${path}`, () => {
      const route = findLeafRoute(routes, path);
      expect(route).toBeDefined();
      const crumb = route!.data?.['breadcrumb'];
      expect(crumb).toBeDefined();
      expect(typeof crumb.label).toBe('string');
      expect(crumb.label.length).toBeGreaterThan(0);
    });
  }

  it('does not declare a root "Settings" Crumb on the shared layout (Home icon is prepended by the component)', () => {
    const layout = routes.find((r) => r.path === '' && (r.children?.length ?? 0) > 0);
    expect(layout).toBeDefined();
    expect(layout!.data?.['breadcrumb']).toBeUndefined();
  });

  it('does not declare a breadcrumb on auth or 404 routes', () => {
    const authPaths = ['signin', 'signup', 'forgot-password', 'reset-password', 'invite-accept', '**'];
    for (const p of authPaths) {
      const route = routes.find((r) => r.path === p);
      expect(route?.data?.['breadcrumb']).toBeUndefined();
    }
  });

  it('uses a resolve function for dynamic child-name and invoice-number segments', () => {
    const childDetail = findLeafRoute(routes, 'staff/manager/children/:childId');
    expect(typeof childDetail!.data.breadcrumb.resolve).toBe('function');

    const managerInvoice = findLeafRoute(routes, 'staff/manager/invoices/:invoiceId');
    expect(typeof managerInvoice!.data.breadcrumb.resolve).toBe('function');

    const parentInvoice = findLeafRoute(routes, 'app/invoices/:invoiceId');
    expect(typeof parentInvoice!.data.breadcrumb.resolve).toBe('function');

    const ownerRoomEdit = findLeafRoute(routes, 'owner/rooms/:roomId/edit');
    expect(typeof ownerRoomEdit!.data.breadcrumb.resolve).toBe('function');
  });
});
