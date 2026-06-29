import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter, Router } from '@angular/router';

import { ROLES } from '../../../core/constants/roles';
import { AuthService } from '../../../core/services/auth.service';
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
    fixture = TestBed.createComponent(AppSidebarComponent);
  });

  describe('manager role', () => {
    beforeEach(() => {
      authStub.role = ROLES.manager;
    });

    it('shows dashboard, children, invites, attendance, billing, and site settings', () => {
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
      const managerInvites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
      const managerInvoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
      const managerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-attendance"]');
      const siteSettings = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-site-settings"]');
      const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

      expect(dashboard).toBeTruthy();
      expect(managerChildren).toBeTruthy();
      expect(managerInvites).toBeTruthy();
      expect(managerInvoices).toBeTruthy();
      expect(managerAttendance).toBeTruthy();
      expect(siteSettings).toBeTruthy();
      expect(parentInvoices).toBeFalsy();
    });

    it('does not show owner links', () => {
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]')).toBeFalsy();
    });

    it('sees five grouped headings', () => {
      fixture.detectChanges();

      const headings = fixture.nativeElement.querySelectorAll('h2');
      const labels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

      expect(labels).toContain('Overview');
      expect(labels).toContain('People');
      expect(labels).toContain('Attendance');
      expect(labels).toContain('Billing');
      expect(labels).toContain('Setup');
      expect(labels.length).toBe(5);
    });

    it('dashboard link points to /manager/dashboard', () => {
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      expect(dashboard.getAttribute('href')).toContain('/manager/dashboard');
    });

    it('invites link points to /manager/invites', () => {
      fixture.detectChanges();

      const invites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
      expect(invites).toBeTruthy();
      expect(invites.getAttribute('href')).toContain('/manager/invites');
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

    it('highlights Attendance corrections with query params', () => {
      const router = TestBed.inject(Router);
      spyOnProperty(router, 'url', 'get').and.returnValue('/manager/attendance-corrections?childId=abc');
      fixture.detectChanges();

      const corrections = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-attendance-corrections"]');
      expect(corrections.getAttribute('aria-current')).toBe('page');
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

  describe('practitioner role', () => {
    beforeEach(() => {
      authStub.role = ROLES.practitioner;
    });

    it('shows only attendance', () => {
      fixture.detectChanges();

      const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
      const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
      const managerInvites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
      const managerInvoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
      const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
      const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

      expect(dashboard).toBeFalsy();
      expect(managerChildren).toBeFalsy();
      expect(managerInvites).toBeFalsy();
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
        'staff-link-manager-invites',
        'staff-link-practitioner-attendance',
        'staff-link-manager-invoices',
        'staff-link-manager-attendance-corrections',
        'staff-link-manager-rooms',
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
        'staff-link-manager-invites',
        'staff-link-practitioner-attendance',
        'staff-link-manager-invoices',
        'staff-link-manager-attendance-corrections',
        'staff-link-manager-rooms',
      ];

      for (const id of staffIds) {
        expect(fixture.nativeElement.querySelector(`[data-testid="${id}"]`)).toBeFalsy();
      }

      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-overview"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-manager-access"]')).toBeFalsy();
      expect(fixture.nativeElement.querySelector('[data-testid="owner-link-rooms"]')).toBeFalsy();
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
