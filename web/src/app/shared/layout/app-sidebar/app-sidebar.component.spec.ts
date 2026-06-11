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

  it('shows dashboard, children, guardians, invites, and attendance for manager role', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const managerGuardians = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-guardians"]');
    const managerInvites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
    const managerFunding = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-funding"]');
    const managerInvoiceRun = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoice-run"]');
    const managerInvoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(dashboard).toBeTruthy();
    expect(managerChildren).toBeTruthy();
    expect(managerGuardians).toBeTruthy();
    expect(managerInvites).toBeTruthy();
    expect(managerFunding).toBeTruthy();
    expect(managerInvoiceRun).toBeTruthy();
    expect(managerInvoices).toBeTruthy();
    expect(practitionerAttendance).toBeTruthy();
    expect(parentInvoices).toBeFalsy();
  });

  it('manager sees four grouped headings', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const headings = fixture.nativeElement.querySelectorAll('h2');
    const labels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

    expect(labels).toContain('Overview');
    expect(labels).toContain('People');
    expect(labels).toContain('Attendance');
    expect(labels).toContain('Billing');
  });

  it('manager dashboard link points to /staff/manager/dashboard', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    expect(dashboard.getAttribute('href')).toContain('/staff/manager/dashboard');
  });

  it('invites link points to /staff/manager/invites', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const invites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
    expect(invites).toBeTruthy();
    expect(invites.getAttribute('href')).toContain('/staff/manager/invites');
  });

  it('invoices link points to /staff/manager/invoices', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const invoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
    expect(invoices).toBeTruthy();
    expect(invoices.getAttribute('href')).toContain('/staff/manager/invoices');
  });

  it('shows only attendance for practitioner role', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const managerInvites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
    const managerInvoiceRun = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoice-run"]');
    const managerInvoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(dashboard).toBeFalsy();
    expect(managerChildren).toBeFalsy();
    expect(managerInvites).toBeFalsy();
    expect(managerInvoiceRun).toBeFalsy();
    expect(managerInvoices).toBeFalsy();
    expect(practitionerAttendance).toBeTruthy();
    expect(parentInvoices).toBeFalsy();
  });

  it('practitioner sees Workday group heading', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const headings = fixture.nativeElement.querySelectorAll('h2');
    const labels = Array.from(headings).map(h => (h as HTMLElement).textContent?.trim());

    expect(labels).toContain('Workday');
    expect(labels.length).toBe(1);
  });

  it('practitioner attendance link points to /staff/practitioner/attendance', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const attendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    expect(attendance.getAttribute('href')).toContain('/staff/practitioner/attendance');
  });

  it('shows no links for parent role', () => {
    authStub.role = ROLES.parent;
    fixture.detectChanges();

    const links = fixture.nativeElement.querySelectorAll('a[data-testid]');
    expect(links.length).toBe(0);
  });

  it('shows no links for owner role', () => {
    authStub.role = ROLES.owner;
    fixture.detectChanges();

    const links = fixture.nativeElement.querySelectorAll('a[data-testid]');
    expect(links.length).toBe(0);
  });

  it('shows no links for null role', () => {
    authStub.role = null;
    fixture.detectChanges();

    const links = fixture.nativeElement.querySelectorAll('a[data-testid]');
    expect(links.length).toBe(0);
  });

  it('does not contain TailAdmin demo labels', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    const demoLabels = ['Ecommerce', 'Charts', 'Forms', 'UI Elements', 'Calendar', 'Authentication', 'Sign Up'];

    for (const label of demoLabels) {
      expect(text).not.toContain(label);
    }
  });

  it('sets aria-current="page" on the active link', () => {
    authStub.role = ROLES.manager;
    const router = TestBed.inject(Router);
    spyOnProperty(router, 'url', 'get').and.returnValue('/staff/manager/dashboard');
    fixture.detectChanges();

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    expect(dashboard.getAttribute('aria-current')).toBe('page');

    const children = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    expect(children.getAttribute('aria-current')).toBeNull();
  });

  it('highlights Children when on a child detail route', () => {
    authStub.role = ROLES.manager;
    const router = TestBed.inject(Router);
    spyOnProperty(router, 'url', 'get').and.returnValue('/staff/manager/children/abc-123');
    fixture.detectChanges();

    const children = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    expect(children.getAttribute('aria-current')).toBe('page');

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    expect(dashboard.getAttribute('aria-current')).toBeNull();
  });

  it('highlights Invoices when on an invoice detail route', () => {
    authStub.role = ROLES.manager;
    const router = TestBed.inject(Router);
    spyOnProperty(router, 'url', 'get').and.returnValue('/staff/manager/invoices/inv-456');
    fixture.detectChanges();

    const invoices = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invoices"]');
    expect(invoices.getAttribute('aria-current')).toBe('page');
  });

  it('highlights Attendance corrections with query params', () => {
    authStub.role = ROLES.manager;
    const router = TestBed.inject(Router);
    spyOnProperty(router, 'url', 'get').and.returnValue('/staff/manager/attendance-corrections?childId=abc');
    fixture.detectChanges();

    const corrections = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-attendance-corrections"]');
    expect(corrections.getAttribute('aria-current')).toBe('page');
  });
});
