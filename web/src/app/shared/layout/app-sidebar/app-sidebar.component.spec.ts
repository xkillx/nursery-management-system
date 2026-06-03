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

  it('shows manager links for manager role', () => {
    authStub.role = ROLES.manager;
    fixture.detectChanges();

    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const managerGuardians = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-guardians"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(managerChildren).toBeTruthy();
    expect(managerGuardians).toBeTruthy();
    expect(practitionerAttendance).toBeTruthy();
    expect(parentInvoices).toBeFalsy();
  });

  it('hides manager links for practitioner role', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(managerChildren).toBeFalsy();
    expect(practitionerAttendance).toBeTruthy();
    expect(parentInvoices).toBeFalsy();
  });

  it('shows only invoices for parent role', () => {
    authStub.role = ROLES.parent;
    fixture.detectChanges();

    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const managerGuardians = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-guardians"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');
    const parentInvoices = fixture.nativeElement.querySelector('[data-testid="parent-link-invoices"]');

    expect(managerChildren).toBeFalsy();
    expect(managerGuardians).toBeFalsy();
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
