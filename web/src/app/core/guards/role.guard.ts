import { inject } from '@angular/core';
import { CanActivateFn, Router, UrlTree } from '@angular/router';

import { defaultRouteForRole } from '../constants/roles';
import { AuthService } from '../services/auth.service';

export const roleGuard: CanActivateFn = (route): boolean | UrlTree => {
  const authService = inject(AuthService);
  const router = inject(Router);

  const allowedRoles = (route.data?.['roles'] as string[] | undefined) ?? [];
  const role = authService.currentRole();

  if (!role) {
    return router.createUrlTree(['/signin']);
  }

  if (allowedRoles.includes(role)) {
    return true;
  }

  return router.createUrlTree([defaultRouteForRole(role)]);
};
