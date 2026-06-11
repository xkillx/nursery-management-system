import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NavigationEnd, Router, RouterModule } from '@angular/router';

import { ROLE_ROUTES } from '../../../core/constants/roles';
import { ThemeToggleButtonComponent } from '../../components/common/theme-toggle/theme-toggle-button.component';
import { UserDropdownComponent } from '../../components/header/user-dropdown/user-dropdown.component';
import { ToastContainerComponent } from '../../components/ui/toast/toast-container.component';

@Component({
  selector: 'app-owner-layout',
  imports: [
    CommonModule,
    RouterModule,
    ThemeToggleButtonComponent,
    UserDropdownComponent,
    ToastContainerComponent,
  ],
  templateUrl: './owner-layout.component.html',
})
export class OwnerLayoutComponent {
  readonly navLinks = [
    { label: 'Overview', path: ROLE_ROUTES.ownerHome, testId: 'owner-link-overview' },
    { label: 'Manager access', path: '/owner/manager-access', testId: 'owner-link-manager-access' },
  ];

  constructor(private router: Router) {}

  isActive(path: string): boolean {
    const url = this.router.url.split('?')[0];
    if (path === ROLE_ROUTES.ownerHome) {
      return url === path;
    }
    return url.startsWith(path);
  }
}
