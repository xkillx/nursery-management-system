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

function findRouteChain(routes: Routes, path: string): any[] {
  const matches: any[] = [];
  collectByPath(routes, '', matches);
  return matches.filter((r) => path === '' || path.startsWith(r.fullPath));
}

function findBreadcrumbInChain(routes: Routes, path: string): { label: string; resolve?: unknown } | undefined {
  const chain = findRouteChain(routes, path);
  for (const r of chain) {
    const crumb = r.data?.['breadcrumb'];
    if (crumb && typeof crumb === 'object' && typeof crumb.label === 'string') {
      return crumb;
    }
  }
  return undefined;
}

function collectByPath(routes: Routes, parentPath: string, out: any[]): void {
  for (const r of routes) {
    if (r.path === undefined) continue;
    const fullPath = r.path === '' ? parentPath : parentPath ? `${parentPath}/${r.path}` : r.path;
    out.push({ ...r, fullPath });
    if (r.children) {
      collectByPath(r.children, fullPath, out);
    }
  }
}

function allDescendantRoutes(routes: Routes): any[] {
  const out: any[] = [];
  const walk = (rs: Routes) => {
    for (const r of rs) {
      out.push(r);
      if (r.children) walk(r.children);
    }
  };
  walk(routes);
  return out;
}

function topLevelRouteChildren(routes: Routes): any[] {
  return routes
    .flatMap(r => r.children ?? [])
    .filter(r => r.path !== undefined);
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
      const allPaths: any[] = [];
      collectByPath(routes, '', allPaths);
      const found = allPaths.some(r => r.fullPath === dynamic);
      expect(found).toBeTrue();
    });
  }

  it('child detail route is a child of the children list and inherits its role guard', () => {
    const listParent = allDescendantRoutes(routes)
      .find(r => r.path === 'staff/manager/children');
    const detailLeaf = allDescendantRoutes(routes)
      .find(r => r.path === ':childId' && r.data?.['breadcrumb']?.resolve);

    expect(listParent).toBeDefined();
    expect(detailLeaf).toBeDefined();
    expect(listParent!.data?.['roles']).toEqual(['manager']);
  });

  it('new registration route is a child of the children list and inherits its role guard', () => {
    const listParent = allDescendantRoutes(routes)
      .find(r => r.path === 'staff/manager/children');
    const newLeaf = allDescendantRoutes(routes)
      .find(r => r.path === 'new' && r.data?.['breadcrumb']?.label === 'New registration');

    expect(listParent).toBeDefined();
    expect(newLeaf).toBeDefined();
    expect(listParent!.data?.['roles']).toEqual(['manager']);
  });

  it('children list child routes declare "new" before ":childId" so the static segment wins', () => {
    const listParent = allDescendantRoutes(routes)
      .find(r => r.path === 'staff/manager/children');

    expect(listParent).toBeDefined();
    const childPaths = (listParent!.children ?? []).map((c: any) => c.path);
    const newIndex = childPaths.indexOf('new');
    const childIdIndex = childPaths.indexOf(':childId');
    expect(newIndex).toBeGreaterThanOrEqual(0);
    expect(childIdIndex).toBeGreaterThanOrEqual(0);
    expect(newIndex).toBeLessThan(childIdIndex);
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
    const invoicesRoute = allDescendantRoutes(routes)
      .find(r => r.path === 'staff/manager/invoices');

    expect(invoicesRoute).toBeDefined();
    expect(invoicesRoute!.data?.['roles']).toEqual(['manager']);
  });

  it('invoice detail route is a child of the invoices list and inherits its role guard', () => {
    const listParent = allDescendantRoutes(routes)
      .find(r => r.path === 'staff/manager/invoices');
    const detailLeaf = allDescendantRoutes(routes)
      .find(r => r.path === ':invoiceId' && r.data?.['breadcrumb']?.resolve);

    expect(listParent).toBeDefined();
    expect(detailLeaf).toBeDefined();
    expect(listParent!.data?.['roles']).toEqual(['manager']);
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
    const ownerRoomParent = allDescendantRoutes(routes)
      .find(r => r.path === 'owner/rooms');

    expect(ownerRoomParent).toBeDefined();
    expect(ownerRoomParent!.data?.['roles']).toEqual(['owner']);
  });

  it('manager room routes require manager role only', () => {
    const managerRoomParent = allDescendantRoutes(routes)
      .find(r => r.path === 'staff/manager/rooms');

    expect(managerRoomParent).toBeDefined();
    expect(managerRoomParent!.data?.['roles']).toEqual(['manager']);
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

  it('canonical parent invoice detail route is a child of the parent invoices list and inherits its role guard', () => {
    const listParent = allDescendantRoutes(routes)
      .find(r => r.path === 'app/invoices');
    const detailLeaf = allDescendantRoutes(routes)
      .find(r => r.path === ':invoiceId' && r.data?.['breadcrumb']?.resolve);

    expect(listParent).toBeDefined();
    expect(detailLeaf).toBeDefined();
    expect(listParent!.data?.['roles']).toEqual(['parent']);
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
    'staff/manager/children/new',
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
      const crumb = findBreadcrumbInChain(routes, path);
      expect(crumb).toBeDefined();
      expect(typeof crumb!.label).toBe('string');
      expect(crumb!.label.length).toBeGreaterThan(0);
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
