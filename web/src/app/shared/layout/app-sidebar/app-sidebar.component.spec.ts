import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';

import { ROLES } from '../../../core/constants/roles';
import { AuthService } from '../../../core/services/auth.service';
import { SidebarService } from '../../services/sidebar.service';
import { AppSidebarComponent } from './app-sidebar.component';

class AuthServiceStub {
  role: string | null = null;

  currentRole(): string | null {
    return this.role;
  }
}

describe('AppSidebarComponent', () => {
  let fixture: ComponentFixture<AppSidebarComponent>;
  let authStub: AuthServiceStub;
  let sidebarService: SidebarService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppSidebarComponent],
      providers: [
        provideRouter([]),
        {
          provide: AuthService,
          useClass: AuthServiceStub,
        },
      ],
    }).compileComponents();

    authStub = TestBed.inject(AuthService) as unknown as AuthServiceStub;
    sidebarService = TestBed.inject(SidebarService);
    fixture = TestBed.createComponent(AppSidebarComponent);
  });

  describe('manager role', () => {
    beforeEach(() => {
      authStub.role = ROLES.manager;
    });

    it('shows dashboard, children, attendance, billing, and settings sub-items', () => {
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
      const managerInvoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
      const managerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-attendance"]');
      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      const rooms = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-rooms"]');
      const sessionTypes = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-session-types"]');
      const sessionTemplates = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-session-templates"]');
      const billingSetup = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-billing-setup"]');
      const termCalendar = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-term-calendar"]');
      const closureDays = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-closure-days"]');
      const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

      expect(dashboard).toBeTruthy();
      expect(managerChildren).toBeTruthy();
      expect(managerInvoices).toBeTruthy();
      expect(managerAttendance).toBeTruthy();
      expect(siteProfile).toBeTruthy();
      expect(rooms).toBeTruthy();
      expect(sessionTypes).toBeTruthy();
      expect(sessionTemplates).toBeTruthy();
      expect(billingSetup).toBeTruthy();
      expect(termCalendar).toBeTruthy();
      expect(closureDays).toBeTruthy();
      expect(parentInvoices).toBeFalsy();
    });

    it('does not show owner links', () => {
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]')).toBeFalsy();
    });

    it('sees five grouped headings including Settings accordion', () => {
      fixture.detectChanges();

      const headings = fixture.nativeElement.querySelectorAll('h2');
      const headingLabels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

      const accordionButtons = fixture.nativeElement.querySelectorAll('button[data-testid^="sidebar-accordion-"]');
      const accordionLabels = Array.from(accordionButtons).map(b => (b as HTMLElement).textContent?.trim());

      expect(headingLabels).toContain('Overview');
      expect(headingLabels).toContain('People');
      expect(headingLabels).toContain('Attendance');
      expect(headingLabels).toContain('Billing');
      expect(accordionLabels).toContain('Settings');
      expect(headingLabels.length).toBe(4);
      expect(accordionLabels.length).toBe(1);
    });

    it('dashboard link points to /manager/dashboard', () => {
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      expect(dashboard.getAttribute('href')).toContain('/manager/dashboard');
    });

    it('invoices link points to /manager/invoices', () => {
      fixture.detectChanges();

      const invoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
      expect(invoices).toBeTruthy();
      expect(invoices.getAttribute('href')).toContain('/manager/invoices');
    });

    it('attendance link points to /manager/attendance', () => {
      fixture.detectChanges();

      const attendance = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-attendance"]');
      expect(attendance).toBeTruthy();
      expect(attendance.getAttribute('href')).toContain('/manager/attendance');
    });

    it('sets aria-current="page" on the active link', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/manager/dashboard');
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      expect(dashboard.getAttribute('aria-current')).toBe('page');

      const children = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
      expect(children.getAttribute('aria-current')).toBeNull();
    });

    it('highlights Children when on a child detail route', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/manager/children/abc-123');
      fixture.detectChanges();

      const children = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
      expect(children.getAttribute('aria-current')).toBe('page');

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      expect(dashboard.getAttribute('aria-current')).toBeNull();
    });

    it('highlights Invoices when on an invoice detail route', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/manager/invoices/inv-456');
      fixture.detectChanges();

      const invoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
      expect(invoices.getAttribute('aria-current')).toBe('page');
    });

    it('does not contain TailAdmin demo labels', () => {
      fixture.detectChanges();

      const text = fixture.nativeElement.textContent;
      const demoLabels = ['Ecommerce', 'Charts', 'Forms', 'UI Elements', 'Calendar', 'Authentication', 'Sign Up'];

      for (const label of demoLabels) {
        expect(text).not.toContain(label);
      }
    });
  });

  describe('Settings accordion', () => {
    beforeEach(() => {
      authStub.role = ROLES.manager;
    });

    it('defaults to expanded state', () => {
      fixture.detectChanges();

      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfile).toBeTruthy();
    });

    it('contains exactly 7 settings items when expanded', () => {
      fixture.detectChanges();

      const settingsItems = [
        'staff-link-manager-site-profile',
        'staff-link-manager-rooms',
        'staff-link-manager-session-types',
        'staff-link-manager-session-templates',
        'staff-link-manager-billing-setup',
        'staff-link-manager-term-calendar',
        'staff-link-manager-closure-days',
      ];

      for (const testId of settingsItems) {
        expect(fixture.nativeElement.querySelector(`[data-testid="${testId}"]`)).toBeTruthy();
      }
    });

    it('collapses items when accordion is toggled', () => {
      fixture.detectChanges();

      sidebarService.toggleAccordion();
      fixture.detectChanges();

      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfile).toBeFalsy();
    });

    it('expands items when accordion is toggled back', () => {
      fixture.detectChanges();

      sidebarService.toggleAccordion();
      fixture.detectChanges();

      sidebarService.toggleAccordion();
      fixture.detectChanges();

      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfile).toBeTruthy();
    });

    it('shows chevron arrow on accordion header', () => {
      fixture.detectChanges();

      const accordionBtn = fixture.nativeElement.querySelector('button[data-testid="sidebar-accordion-Settings"]');
      expect(accordionBtn).toBeTruthy();

      const chevron = accordionBtn.querySelector('ng-icon');
      expect(chevron).toBeTruthy();
    });

    it('highlights active settings sub-item', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/manager/site-settings/rooms');
      fixture.detectChanges();

      const rooms = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-rooms"]');
      expect(rooms.getAttribute('aria-current')).toBe('page');

      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfile.getAttribute('aria-current')).toBeNull();
    });

    it('site profile link points to /manager/site-settings/profile', () => {
      fixture.detectChanges();

      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfile.getAttribute('href')).toContain('/manager/site-settings/profile');
    });

    it('rooms link points to /manager/site-settings/rooms', () => {
      fixture.detectChanges();

      const rooms = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-rooms"]');
      expect(rooms.getAttribute('href')).toContain('/manager/site-settings/rooms');
    });

    it('session types link points to /manager/site-settings/session-types', () => {
      fixture.detectChanges();

      const sessionTypes = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-session-types"]');
      expect(sessionTypes.getAttribute('href')).toContain('/manager/site-settings/session-types');
    });

    it('session templates link points to /manager/site-settings/session-templates', () => {
      fixture.detectChanges();

      const sessionTemplates = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-session-templates"]');
      expect(sessionTemplates.getAttribute('href')).toContain('/manager/site-settings/session-templates');
    });

    it('fees & billing link points to /manager/site-settings/billing-setup', () => {
      fixture.detectChanges();

      const billing = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-billing-setup"]');
      expect(billing.getAttribute('href')).toContain('/manager/site-settings/billing-setup');
    });

    it('term calendar link points to /manager/site-settings/term-calendar', () => {
      fixture.detectChanges();

      const termCalendar = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-term-calendar"]');
      expect(termCalendar.getAttribute('href')).toContain('/manager/site-settings/term-calendar');
    });

    it('closure days link points to /manager/site-settings/closure-days', () => {
      fixture.detectChanges();

      const closureDays = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-closure-days"]');
      expect(closureDays.getAttribute('href')).toContain('/manager/site-settings/closure-days');
    });

    it('accordion state persists during navigation', () => {
      fixture.detectChanges();

      sidebarService.setAccordionExpanded(false);
      fixture.detectChanges();

      const siteProfile = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfile).toBeFalsy();

      fixture.detectChanges();

      const siteProfileAfter = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-profile"]');
      expect(siteProfileAfter).toBeFalsy();
    });

    it('sets aria-expanded on accordion button', () => {
      fixture.detectChanges();

      const accordionBtn = fixture.nativeElement.querySelector('button[data-testid="sidebar-accordion-Settings"]');
      expect(accordionBtn.getAttribute('aria-expanded')).toBe('true');

      sidebarService.toggleAccordion();
      fixture.detectChanges();

      expect(accordionBtn.getAttribute('aria-expanded')).toBe('false');
    });
  });

  describe('practitioner role', () => {
    beforeEach(() => {
      authStub.role = ROLES.practitioner;
    });

    it('shows only attendance', () => {
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
      const managerInvoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
      const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
      const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

      expect(dashboard).toBeFalsy();
      expect(managerChildren).toBeFalsy();
      expect(managerInvoices).toBeFalsy();
      expect(practitionerAttendance).toBeTruthy();
      expect(parentInvoices).toBeFalsy();
    });

    it('does not show owner or parent links', () => {
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]')).toBeFalsy();
    });

    it('sees Workday group heading', () => {
      fixture.detectChanges();

      const headings = fixture.nativeElement.querySelectorAll('h2');
      const labels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

      expect(labels).toContain('Workday');
      expect(labels.length).toBe(1);
    });

    it('attendance link points to /practitioner/attendance', () => {
      fixture.detectChanges();

      const attendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
      expect(attendance.getAttribute('href')).toContain('/practitioner/attendance');
    });
  });

  describe('owner role', () => {
    beforeEach(() => {
      authStub.role = ROLES.owner;
    });

    it('renders owner overview and manager access links', () => {
      fixture.detectChanges();

      const overview = fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]');
      const managerAccess = fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]');

      expect(overview).toBeTruthy();
      expect(managerAccess).toBeTruthy();
    });

    it('does not render staff or parent links', () => {
      fixture.detectChanges();

      const staffIds = [
        'staff-link-manager-dashboard',
        'staff-link-manager-children',
        'staff-link-manager-invoices',
        'staff-link-practitioner-attendance',
        'staff-link-manager-attendance-corrections',
        'staff-link-manager-rooms',
        'staff-link-manager-site-profile',
      ];

      for (const id of staffIds) {
        expect(fixture.nativeElement.querySelector(`[data-testid="${id}"]`)).toBeFalsy();
      }

      expect(fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]')).toBeFalsy();
    });

    it('sees Overview and Access group headings', () => {
      fixture.detectChanges();

      const headings = fixture.nativeElement.querySelectorAll('h2');
      const labels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

      expect(labels).toContain('Overview');
      expect(labels).toContain('Access');
      expect(labels.length).toBe(2);
    });

    it('overview link points to /owner', () => {
      fixture.detectChanges();

      const overview = fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]');
      expect(overview.getAttribute('href')).toContain('/owner');
    });

    it('manager access link points to /owner/manager-access', () => {
      fixture.detectChanges();

      const managerAccess = fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]');
      expect(managerAccess.getAttribute('href')).toContain('/owner/manager-access');
    });

    it('highlights overview when on /owner', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/owner');
      fixture.detectChanges();

      const overview = fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]');
      expect(overview.getAttribute('aria-current')).toBe('page');

      const managerAccess = fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]');
      expect(managerAccess.getAttribute('aria-current')).toBeNull();
    });

    it('highlights manager access when on /owner/manager-access', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/owner/manager-access');
      fixture.detectChanges();

      const managerAccess = fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]');
      expect(managerAccess.getAttribute('aria-current')).toBe('page');

      const overview = fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]');
      expect(overview.getAttribute('aria-current')).toBeNull();
    });

  });

  describe('parent role', () => {
    beforeEach(() => {
      authStub.role = ROLES.parent;
    });

    it('renders invoices link', () => {
      fixture.detectChanges();

      const invoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');
      expect(invoices).toBeTruthy();
    });

    it('does not render staff or owner links', () => {
      fixture.detectChanges();

      const staffIds = [
        'staff-link-manager-dashboard',
        'staff-link-manager-children',
        'staff-link-manager-invoices',
        'staff-link-practitioner-attendance',
        'staff-link-manager-attendance-corrections',
        'staff-link-manager-rooms',
        'staff-link-manager-site-profile',
      ];

      for (const id of staffIds) {
        expect(fixture.nativeElement.querySelector(`[data-testid="${id}"]`)).toBeFalsy();
      }

      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]')).toBeFalsy();
    });

    it('sees Billing group heading', () => {
      fixture.detectChanges();

      const headings = fixture.nativeElement.querySelectorAll('h2');
      const labels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

      expect(labels).toContain('Billing');
      expect(labels.length).toBe(1);
    });

    it('invoices link points to /parent/invoices', () => {
      fixture.detectChanges();

      const invoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');
      expect(invoices.getAttribute('href')).toContain('/parent/invoices');
    });

    it('highlights invoices when on /parent/invoices', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/parent/invoices');
      fixture.detectChanges();

      const invoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');
      expect(invoices.getAttribute('aria-current')).toBe('page');
    });

    it('highlights invoices when on invoice detail route', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/parent/invoices/inv-789');
      fixture.detectChanges();

      const invoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');
      expect(invoices.getAttribute('aria-current')).toBe('page');
    });
  });

  describe('unknown role', () => {
    it('shows no links for null role', () => {
      authStub.role = null;
      fixture.detectChanges();

      const links = fixture.nativeElement.querySelectorAll('a[data-testid]');
      expect(links.length).toBe(0);
    });
  });
});
