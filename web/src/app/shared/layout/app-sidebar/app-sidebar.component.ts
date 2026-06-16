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
  heroBanknotes,
  heroDocumentPlus,
  heroDocumentText,
  heroHomeModern,
} from '@ng-icons/heroicons/outline';

export type SidebarIcon =
  | 'dashboard'
  | 'children'
  | 'guardians'
  | 'invites'
  | 'attendance'
  | 'attendance-corrections'
  | 'funding'
  | 'invoice-run'
  | 'invoices'
  | 'rooms';

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
      heroBanknotes,
      heroDocumentPlus,
      heroDocumentText,
      heroHomeModern,
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
    guardians: 'heroUsers',
    invites: 'heroEnvelope',
    attendance: 'heroClipboardDocumentCheck',
    'attendance-corrections': 'heroClipboardDocumentList',
    funding: 'heroBanknotes',
    'invoice-run': 'heroDocumentPlus',
    invoices: 'heroDocumentText',
    rooms: 'heroHomeModern',
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
            { label: 'Children', path: ROLE_ROUTES.managerChildren, testId: 'staff-link-manager-children', icon: 'children', matchPaths: ['/staff/manager/children/'] },
            { label: 'Guardians', path: ROLE_ROUTES.managerGuardians, testId: 'staff-link-manager-guardians', icon: 'guardians' },
            { label: 'Invites', path: ROLE_ROUTES.managerInvites, testId: 'staff-link-manager-invites', icon: 'invites' },
          ],
        },
        {
          label: 'Attendance',
          items: [
            { label: 'Attendance', path: ROLE_ROUTES.practitionerAttendance, testId: 'staff-link-practitioner-attendance', icon: 'attendance' },
            { label: 'Attendance corrections', path: ROLE_ROUTES.managerAttendanceCorrections, testId: 'staff-link-manager-attendance-corrections', icon: 'attendance-corrections' },
          ],
        },
        {
          label: 'Billing',
          items: [
            { label: 'Funding', path: ROLE_ROUTES.managerFunding, testId: 'staff-link-manager-funding', icon: 'funding' },
            { label: 'Invoice run', path: ROLE_ROUTES.managerInvoiceRun, testId: 'staff-link-manager-invoice-run', icon: 'invoice-run' },
            { label: 'Invoices', path: ROLE_ROUTES.managerInvoices, testId: 'staff-link-manager-invoices', icon: 'invoices', matchPaths: ['/staff/manager/invoices/'] },
          ],
        },
        {
          label: 'Setup',
          items: [
            { label: 'Rooms', path: ROLE_ROUTES.managerRooms, testId: 'staff-link-manager-rooms', icon: 'rooms', matchPaths: ['/staff/manager/rooms/'] },
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
            { label: 'Manager access', path: ROLE_ROUTES.ownerManagerAccess, testId: 'owner-link-manager-access', icon: 'guardians' },
          ],
        },
        {
          label: 'Setup',
          items: [
            { label: 'Rooms', path: ROLE_ROUTES.ownerRooms, testId: 'owner-link-rooms', icon: 'rooms', matchPaths: ['/owner/rooms/'] },
          ],
        },
      ];
    }

    if (role === ROLES.parent) {
      return [
        {
          label: 'Billing',
          items: [
            { label: 'Invoices', path: ROLE_ROUTES.parentInvoices, testId: 'parent-link-invoices', icon: 'invoices', matchPaths: ['/app/invoices/'] },
          ],
        },
      ];
    }

    return [];
  }
}
