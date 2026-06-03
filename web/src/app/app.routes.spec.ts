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
    'signup', 'calendar', 'profile', 'form-elements', 'basic-tables',
    'blank', 'invoice', 'line-chart', 'bar-chart', 'alerts', 'avatars',
    'badge', 'buttons', 'images', 'videos',
  ];

  for (const demo of removedDemoPaths) {
    it(`does not register demo route /${demo}`, () => {
      expect(paths).not.toContain(demo);
    });
  }

  const mvpPaths = [
    'staff/manager/children',
    'staff/manager/guardians',
    'staff/practitioner/attendance-children',
    'parent/invoices',
    'signin',
  ];

  for (const mvp of mvpPaths) {
    it(`registers MVP route /${mvp}`, () => {
      expect(paths).toContain(mvp);
    });
  }
});
