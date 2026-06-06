import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

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
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(dashboard).toBeTruthy();
    expect(managerChildren).toBeTruthy();
    expect(managerGuardians).toBeTruthy();
    expect(managerInvites).toBeTruthy();
    expect(practitionerAttendance).toBeTruthy();
    expect(parentInvoices).toBeFalsy();
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

  it('shows only attendance for practitioner role', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const managerInvites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(dashboard).toBeFalsy();
    expect(managerChildren).toBeFalsy();
    expect(managerInvites).toBeFalsy();
    expect(practitionerAttendance).toBeTruthy();
    expect(parentInvoices).toBeFalsy();
  });

  it('practitioner attendance link points to /staff/practitioner/attendance', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const attendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    expect(attendance.getAttribute('href')).toContain('/staff/practitioner/attendance');
  });

  it('shows only invoices for parent role', () => {
    authStub.role = ROLES.parent;
    fixture.detectChanges();

    const dashboard = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-dashboard"]');
    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const managerGuardians = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-guardians"]');
    const managerInvites = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-invites"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(dashboard).toBeFalsy();
    expect(managerChildren).toBeFalsy();
    expect(managerGuardians).toBeFalsy();
    expect(managerInvites).toBeFalsy();
    expect(practitionerAttendance).toBeFalsy();
    expect(parentInvoices).toBeTruthy();
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

  it('shows no links for null role', () => {
    authStub.role = null;
    fixture.detectChanges();

    const links = fixture.nativeElement.querySelectorAll('a[data-testid]');
    expect(links.length).toBe(0);
  });
});
