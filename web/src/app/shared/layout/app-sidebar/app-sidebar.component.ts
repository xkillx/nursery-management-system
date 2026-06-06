import { CommonModule } from '@angular/common';
import { Component, ChangeDetectorRef } from '@angular/core';
import { SidebarService } from '../../services/sidebar.service';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import { Subscription } from 'rxjs';

import { ROLES, ROLE_ROUTES } from '../../../core/constants/roles';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-sidebar',
  imports: [
    CommonModule,
    RouterModule,
  ],
  templateUrl: './app-sidebar.component.html',
})
export class AppSidebarComponent {

  readonly isExpanded$;
  readonly isMobileOpen$;
  readonly isHovered$;

  private subscription: Subscription = new Subscription();

  constructor(
    public sidebarService: SidebarService,
    private router: Router,
    private cdr: ChangeDetectorRef,
    private authService: AuthService
  ) {
    this.isExpanded$ = this.sidebarService.isExpanded$;
    this.isMobileOpen$ = this.sidebarService.isMobileOpen$;
    this.isHovered$ = this.sidebarService.isHovered$;
  }

  ngOnInit() {
    this.subscription.add(
      this.router.events.subscribe(event => {
        if (event instanceof NavigationEnd) {
          this.cdr.detectChanges();
        }
      })
    );

    this.subscription.add(
      this.isMobileOpen$.subscribe(isMobile => {
        if (isMobile === false) {
          this.cdr.detectChanges();
        }
      })
    );
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  isActive(path: string): boolean {
    return this.router.url === path;
  }

  onSidebarMouseEnter() {
    this.isExpanded$.subscribe(expanded => {
      if (!expanded) {
        this.sidebarService.setHovered(true);
      }
    }).unsubscribe();
  }

  closeSidebar() {
    this.sidebarService.setMobileOpen(false);
  }

  get navLinks(): { label: string; path: string; testId: string }[] {
    const role = this.authService.currentRole();

    if (role === ROLES.manager) {
      return [
        { label: 'Dashboard', path: ROLE_ROUTES.managerDashboard, testId: 'staff-link-manager-dashboard' },
        { label: 'Children', path: ROLE_ROUTES.managerChildren, testId: 'staff-link-manager-children' },
        { label: 'Guardians', path: ROLE_ROUTES.managerGuardians, testId: 'staff-link-manager-guardians' },
        { label: 'Invites', path: ROLE_ROUTES.managerInvites, testId: 'staff-link-manager-invites' },
        { label: 'Attendance', path: ROLE_ROUTES.practitionerAttendance, testId: 'staff-link-practitioner-attendance' },
      ];
    }

    if (role === ROLES.practitioner) {
      return [
        { label: 'Attendance', path: ROLE_ROUTES.practitionerAttendance, testId: 'staff-link-practitioner-attendance' },
      ];
    }

    if (role === ROLES.parent) {
      return [
        { label: 'Invoices', path: ROLE_ROUTES.parentInvoices, testId: 'parent-link-invoices' },
      ];
    }

    return [];
  }

  get sectionHeading(): string {
    const role = this.authService.currentRole();
    if (role === ROLES.parent) {
      return 'Parent Portal';
    }
    return 'Staff';
  }
}
