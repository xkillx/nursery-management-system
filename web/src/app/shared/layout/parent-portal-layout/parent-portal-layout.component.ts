import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NavigationEnd, Router, RouterModule } from '@angular/router';

import { ROLE_ROUTES } from '../../../core/constants/roles';
import { ThemeToggleButtonComponent } from '../../components/common/theme-toggle/theme-toggle-button.component';
import { UserDropdownComponent } from '../../components/header/user-dropdown/user-dropdown.component';
import { ToastContainerComponent } from '../../components/ui/toast/toast-container.component';

@Component({
  selector: 'app-parent-portal-layout',
  imports: [
    CommonModule,
    RouterModule,
    ThemeToggleButtonComponent,
    UserDropdownComponent,
    ToastContainerComponent,
  ],
  templateUrl: './parent-portal-layout.component.html',
})
export class ParentPortalLayoutComponent {
  readonly navLinks = [
    { label: 'Invoices', path: ROLE_ROUTES.parentInvoices, testId: 'parent-link-invoices' },
  ];

  constructor(private router: Router) {}

  isActive(path: string): boolean {
    return this.router.url === path;
  }
}
