import { inject } from '@angular/core';
import { CanActivateFn, Router, UrlTree } from '@angular/router';

import { defaultRouteForRole } from '../constants/roles';
import { AuthService } from '../services/auth.service';

export const roleDefaultRedirectGuard: CanActivateFn = (): boolean | UrlTree => {
  const authService = inject(AuthService);
  const router = inject(Router);

  const role = authService.currentRole();
  if (!role) {
    return router.createUrlTree(['/signin']);
  }

  return router.createUrlTree([defaultRouteForRole(role)]);
};
