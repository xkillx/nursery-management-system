import { CommonModule } from '@angular/common';
import { Component, ChangeDetectorRef, OnInit, OnDestroy, inject } from '@angular/core';
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
  heroCog6Tooth,
  heroDocumentText,
  heroCalendarDays,
  heroBuildingOffice2,
  heroRectangleStack,
  heroReceiptPercent,
  heroChevronDown,
} from '@ng-icons/heroicons/outline';

export type SidebarIcon =
  | 'dashboard'
  | 'children'
  | 'parents'
  | 'invites'
  | 'attendance'
  | 'attendance-corrections'
  | 'bookings'
  | 'invoices'
  | 'site-settings'
  | 'site-profile'
  | 'rooms'
  | 'session-types'
  | 'session-templates'
  | 'fees-billing'
  | 'term-calendar'
  | 'closure-days';

export interface SidebarNavItem {
  label: string;
  path: string;
  testId: string;
  icon: SidebarIcon;
  matchPaths?: string[];
}

export interface SidebarNavGroup {
  label: string;
  items: SidebarNavItem[];
  isAccordion?: boolean;
  accordionKey?: string;
}

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
      heroCog6Tooth,
      heroDocumentText,
      heroCalendarDays,
      heroBuildingOffice2,
      heroRectangleStack,
      heroReceiptPercent,
      heroChevronDown,
    }),
  ],
  templateUrl: './app-sidebar.component.html',
})
export class AppSidebarComponent implements OnInit, OnDestroy {

  iconMap: Record<SidebarIcon, string> = {
    dashboard: 'heroSquares2x2',
    children: 'heroUserGroup',
    parents: 'heroUsers',
    invites: 'heroEnvelope',
    attendance: 'heroClipboardDocumentCheck',
    'attendance-corrections': 'heroClipboardDocumentList',
    bookings: 'heroCalendarDays',
    invoices: 'heroDocumentText',
    'site-settings': 'heroCog6Tooth',
    'site-profile': 'heroBuildingOffice2',
    rooms: 'heroCog6Tooth',
    'session-types': 'heroRectangleStack',
    'session-templates': 'heroClipboardDocumentList',
    'fees-billing': 'heroReceiptPercent',
    'term-calendar': 'heroCalendarDays',
    'closure-days': 'heroCalendarDays',
  };

  private subscription: Subscription = new Subscription();

  readonly sidebarService = inject(SidebarService);
  private readonly router = inject(Router);
  private readonly cdr = inject(ChangeDetectorRef);
  private readonly authService = inject(AuthService);

  readonly isExpanded$ = this.sidebarService.isExpanded$;
  readonly isMobileOpen$ = this.sidebarService.isMobileOpen$;
  readonly isHovered$ = this.sidebarService.isHovered$;
  readonly accordionExpanded$ = this.sidebarService.accordionExpanded$;

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

  onAccordionHeaderClick() {
    this.sidebarService.toggleAccordion();
    this.router.navigate(['/manager/site-settings']);
    this.closeSidebar();
  }

  isAccordionExpanded(): boolean {
    let expanded = true;
    this.accordionExpanded$.subscribe(val => expanded = val).unsubscribe();
    return expanded;
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
            { label: 'Parents', path: ROLE_ROUTES.managerParents, testId: 'staff-link-manager-parents', icon: 'parents', matchPaths: ['/manager/parents/'] },
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
            { label: 'Bookings', path: ROLE_ROUTES.managerBookings, testId: 'staff-link-manager-bookings', icon: 'bookings', matchPaths: ['/manager/bookings'] },
            { label: 'Invoices', path: ROLE_ROUTES.managerInvoices, testId: 'staff-link-manager-invoices', icon: 'invoices', matchPaths: ['/manager/invoices/'] },
          ],
        },
          {
            label: 'Settings',
            isAccordion: true,
            accordionKey: 'settings',
            items: [
              { label: 'Site profile', path: ROLE_ROUTES.managerSiteProfile, testId: 'staff-link-manager-site-profile', icon: 'site-profile', matchPaths: ['/manager/site-settings/profile'] },
              { label: 'Rooms & capacity', path: ROLE_ROUTES.managerRooms, testId: 'staff-link-manager-rooms', icon: 'rooms', matchPaths: ['/manager/site-settings/rooms'] },
              { label: 'Session types', path: ROLE_ROUTES.managerSessionTypes, testId: 'staff-link-manager-session-types', icon: 'session-types', matchPaths: ['/manager/site-settings/session-types'] },
              { label: 'Session templates', path: ROLE_ROUTES.managerSessionTemplatesSetup, testId: 'staff-link-manager-session-templates', icon: 'session-templates', matchPaths: ['/manager/site-settings/session-templates'] },
              { label: 'Fees & billing', path: ROLE_ROUTES.managerBillingSetup, testId: 'staff-link-manager-billing-setup', icon: 'fees-billing', matchPaths: ['/manager/site-settings/billing-setup'] },
              { label: 'Term calendar', path: ROLE_ROUTES.managerTermCalendar, testId: 'staff-link-manager-term-calendar', icon: 'term-calendar', matchPaths: ['/manager/site-settings/term-calendar'] },
              { label: 'Closure days', path: ROLE_ROUTES.managerClosureDays, testId: 'staff-link-manager-closure-days', icon: 'closure-days', matchPaths: ['/manager/site-settings/closure-days'] },
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
