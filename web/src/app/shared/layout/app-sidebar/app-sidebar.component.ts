import { CommonModule } from '@angular/common';
import { Component, ChangeDetectorRef } from '@angular/core';
import { SidebarService } from '../../services/sidebar.service';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import { Subscription } from 'rxjs';

import { ROLES, ROLE_ROUTES } from '../../../core/constants/roles';
import { AuthService } from '../../../core/services/auth.service';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroSquares2x2,
  heroUserGroup,
  heroUsers,
  heroEnvelope,
  heroClipboardDocumentCheck,
  heroClipboardDocumentList,
  heroBuildingOffice2,
  heroClock,
  heroCurrencyPound,
  heroDocumentText,
} from '@ng-icons/heroicons/outline';

export type SidebarIcon =
  | 'dashboard'
  | 'children'
  | 'invites'
  | 'attendance'
  | 'attendance-corrections'
  | 'invoices'
  | 'rooms'
  | 'session-types'
  | 'billing-setup';

export type SidebarNavItem = {
  label: string;
  path: string;
  testId: string;
  icon: SidebarIcon;
  matchPaths?: string[];
};

export type SidebarNavGroup = {
  label: string;
  items: SidebarNavItem[];
};

@Component({
  selector: 'app-sidebar',
  imports: [
    CommonModule,
    RouterModule,
    NgIcon,
  ],
  providers: [
    provideIcons({
      heroSquares2x2,
      heroUserGroup,
      heroUsers,
      heroEnvelope,
      heroClipboardDocumentCheck,
      heroClipboardDocumentList,
      heroBuildingOffice2,
      heroClock,
      heroCurrencyPound,
      heroDocumentText,
    }),
  ],
  templateUrl: './app-sidebar.component.html',
})
export class AppSidebarComponent {

  readonly isExpanded$;
  readonly isMobileOpen$;
  readonly isHovered$;

  iconMap: Record<SidebarIcon, string> = {
    dashboard: 'heroSquares2x2',
    children: 'heroUserGroup',
    invites: 'heroEnvelope',
    attendance: 'heroClipboardDocumentCheck',
    'attendance-corrections': 'heroClipboardDocumentList',
    invoices: 'heroDocumentText',
    rooms: 'heroBuildingOffice2',
    'session-types': 'heroClock',
    'billing-setup': 'heroCurrencyPound',
  };

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

  isActive(item: SidebarNavItem): boolean {
    const url = this.router.url.split('?')[0];
    if (url === item.path) return true;
    if (item.matchPaths) {
      return item.matchPaths.some(mp => url.startsWith(mp));
    }
    return false;
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

  get navGroups(): SidebarNavGroup[] {
    const role = this.authService.currentRole();

    if (role === ROLES.manager) {
      return [
        {
          label: 'Overview',
          items: [
            { label: 'Dashboard', path: ROLE_ROUTES.managerDashboard, testId: 'staff-link-manager-dashboard', icon: 'dashboard' },
          ],
        },
        {
          label: 'People',
          items: [
            { label: 'Children', path: ROLE_ROUTES.managerChildren, testId: 'staff-link-manager-children', icon: 'children', matchPaths: ['/manager/children/'] },
          ],
        },
        {
          label: 'Attendance',
          items: [
            { label: 'Attendance', path: ROLE_ROUTES.managerAttendance, testId: 'staff-link-manager-attendance', icon: 'attendance' },
          ],
        },
        {
          label: 'Billing',
          items: [
            { label: 'Invoices', path: ROLE_ROUTES.managerInvoices, testId: 'staff-link-manager-invoices', icon: 'invoices', matchPaths: ['/manager/invoices/'] },
          ],
        },
        {
          label: 'Setup',
          items: [
            { label: 'Rooms', path: ROLE_ROUTES.managerRooms, testId: 'staff-link-manager-rooms', icon: 'rooms', matchPaths: ['/manager/rooms/'] },
            { label: 'Session types', path: ROLE_ROUTES.managerSessionTypes, testId: 'staff-link-manager-session-types', icon: 'session-types', matchPaths: ['/manager/session-types/'] },
            { label: 'Billing setup', path: ROLE_ROUTES.managerBillingSetup, testId: 'staff-link-manager-billing-setup', icon: 'billing-setup' },
          ],
        },
      ];
    }

    if (role === ROLES.practitioner) {
      return [
        {
          label: 'Workday',
          items: [
            { label: 'Attendance', path: ROLE_ROUTES.practitionerAttendance, testId: 'staff-link-practitioner-attendance', icon: 'attendance' },
          ],
        },
      ];
    }

    if (role === ROLES.owner) {
      return [
        {
          label: 'Overview',
          items: [
            { label: 'Overview', path: ROLE_ROUTES.ownerHome, testId: 'owner-link-overview', icon: 'dashboard' },
          ],
        },
        {
          label: 'Access',
          items: [
            { label: 'Manager access', path: ROLE_ROUTES.ownerManagerAccess, testId: 'owner-link-manager-access', icon: 'children' },
          ],
        },
        {
          label: 'Setup',
          items: [
            { label: 'Rooms', path: ROLE_ROUTES.ownerRooms, testId: 'owner-link-rooms', icon: 'rooms', matchPaths: ['/owner/rooms/'] },
            { label: 'Session types', path: ROLE_ROUTES.ownerSessionTypes, testId: 'owner-link-session-types', icon: 'session-types', matchPaths: ['/owner/session-types/'] },
          ],
        },
      ];
    }

    if (role === ROLES.parent) {
      return [
        {
          label: 'Billing',
          items: [
            { label: 'Invoices', path: ROLE_ROUTES.parentInvoices, testId: 'parent-link-invoices', icon: 'invoices', matchPaths: ['/parent/invoices/'] },
          ],
        },
      ];
    }

    return [];
  }
}
