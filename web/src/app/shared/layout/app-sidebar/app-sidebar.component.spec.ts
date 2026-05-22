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

    expect(managerChildren).toBeTruthy();
    expect(managerGuardians).toBeTruthy();
  });

  it('hides manager links for practitioner role', () => {
    authStub.role = ROLES.practitioner;
    fixture.detectChanges();

    const managerChildren = fixture.nativeElement.querySelector('[data-testid="staff-link-manager-children"]');
    const practitionerAttendance = fixture.nativeElement.querySelector('[data-testid="staff-link-practitioner-attendance"]');

    expect(managerChildren).toBeFalsy();
    expect(practitionerAttendance).toBeTruthy();
  });
});
