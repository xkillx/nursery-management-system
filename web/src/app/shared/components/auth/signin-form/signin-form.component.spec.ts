import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { BehaviorSubject, of, throwError } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute, Router } from '@angular/router';

import { SigninFormComponent } from './signin-form.component';
import { AuthService } from '../../../../core/services/auth.service';
import { MembershipModel } from '../../../../core/models/auth.models';

describe('SigninFormComponent', () => {
  let component: SigninFormComponent;
  let fixture: ComponentFixture<SigninFormComponent>;
  let authServiceMock: jasmine.SpyObj<AuthService>;
  let routerMock: jasmine.SpyObj<Router>;

  const mockMembership: MembershipModel = {
    membership_id: '00000000-0000-0000-0000-000000000010',
    tenant_id: '00000000-0000-0000-0000-000000000020',
    tenant_name: 'Little Sprouts Nursery',
    branch_id: '00000000-0000-0000-0000-000000000030',
    branch_name: 'Main Branch',
    role: 'manager',
  };

  const mockMembership2: MembershipModel = {
    membership_id: '00000000-0000-0000-0000-000000000011',
    tenant_id: '00000000-0000-0000-0000-000000000020',
    tenant_name: 'Little Sprouts Nursery',
    branch_id: '00000000-0000-0000-0000-000000000031',
    branch_name: 'Second Branch',
    role: 'practitioner',
  };

  beforeEach(async () => {
    authServiceMock = jasmine.createSpyObj('AuthService', ['login', 'currentRole'], {
      authState$: new BehaviorSubject(null),
    });
    routerMock = jasmine.createSpyObj('Router', ['navigateByUrl']);

    await TestBed.configureTestingModule({
      imports: [SigninFormComponent],
      providers: [
        { provide: ActivatedRoute, useValue: { snapshot: { queryParamMap: { get: () => null } } } },
        { provide: Router, useValue: routerMock },
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authServiceMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(SigninFormComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('direct one-scope success calls login(email, password) and navigates', fakeAsync(() => {
    authServiceMock.login.and.returnValue(of({} as any));
    authServiceMock.currentRole.and.returnValue('manager');

    component.email = 'user@test.com';
    component.password = 'password1';
    component.onSignIn();
    tick();

    expect(authServiceMock.login).toHaveBeenCalledWith('user@test.com', 'password1');
    expect(routerMock.navigateByUrl).toHaveBeenCalledWith('/staff/manager/dashboard');
  }));

  it('challenge response renders choices by labels and not IDs', fakeAsync(() => {
    const errorBody = {
      code: 'membership_selection_required',
      message: 'Choose a nursery to continue.',
      request_id: 'req-1',
      available_memberships: [mockMembership, mockMembership2],
    };
    authServiceMock.login.and.returnValue(throwError(() =>
      new HttpErrorResponse({ error: errorBody, status: 400 })
    ));

    component.email = 'user@test.com';
    component.password = 'password1';
    component.onSignIn();
    tick();

    expect(component.isMembershipChallenge).toBeTrue();
    expect(component.membershipChoices.length).toBe(2);
    expect(component.membershipChoices[0].tenant_name).toBe('Little Sprouts Nursery');
    expect(component.membershipChoices[0].branch_name).toBe('Main Branch');
    expect(component.membershipChallengeMessage).toBe('Choose a nursery to continue.');

    fixture.detectChanges();
    const compiled = fixture.nativeElement as HTMLElement;
    const text = compiled.textContent!;
    expect(text).toContain('Little Sprouts Nursery');
    expect(text).toContain('Main Branch');
    expect(text).toContain('Second Branch');
    expect(text).toContain('practitioner');
  }));

  it('selecting a choice calls login(email, password, membership_id) and navigates on success', fakeAsync(() => {
    const errorBody = {
      code: 'membership_selection_required',
      message: 'Choose a nursery to continue.',
      available_memberships: [mockMembership, mockMembership2],
    };
    authServiceMock.login.and.returnValues(
      throwError(() => new HttpErrorResponse({ error: errorBody, status: 400 })),
      of({} as any),
    );
    authServiceMock.currentRole.and.returnValue('manager');

    component.email = 'user@test.com';
    component.password = 'password1';
    component.onSignIn();
    tick();

    expect(component.isMembershipChallenge).toBeTrue();

    component.selectMembership(mockMembership.membership_id);
    component.onSignIn();
    tick();

    expect(authServiceMock.login).toHaveBeenCalledWith(
      'user@test.com', 'password1', mockMembership.membership_id,
    );
    expect(routerMock.navigateByUrl).toHaveBeenCalledWith('/staff/manager/dashboard');
  }));

  it('editing email clears the picker', fakeAsync(() => {
    const errorBody = {
      code: 'membership_selection_required',
      message: 'Choose a nursery to continue.',
      available_memberships: [mockMembership],
    };
    authServiceMock.login.and.returnValue(throwError(() =>
      new HttpErrorResponse({ error: errorBody, status: 400 })
    ));

    component.email = 'user@test.com';
    component.password = 'password1';
    component.onSignIn();
    tick();

    expect(component.isMembershipChallenge).toBeTrue();

    component.onEmailChange('new@test.com');
    expect(component.isMembershipChallenge).toBeFalse();
    expect(component.membershipChoices.length).toBe(0);
  }));

  it('editing password clears the picker', fakeAsync(() => {
    const errorBody = {
      code: 'membership_selection_required',
      message: 'Choose a nursery to continue.',
      available_memberships: [mockMembership],
    };
    authServiceMock.login.and.returnValue(throwError(() =>
      new HttpErrorResponse({ error: errorBody, status: 400 })
    ));

    component.email = 'user@test.com';
    component.password = 'password1';
    component.onSignIn();
    tick();

    expect(component.isMembershipChallenge).toBeTrue();

    component.onPasswordChange('newpassword');
    expect(component.isMembershipChallenge).toBeFalse();
  }));

  it('stale retry challenge replaces choices and displays the stale message', fakeAsync(() => {
    const initialChallenge = {
      code: 'membership_selection_required',
      message: 'Choose a nursery to continue.',
      available_memberships: [mockMembership, mockMembership2],
    };
    const staleChallenge = {
      code: 'membership_selection_required',
      message: 'That access is no longer available. Choose another nursery or contact your manager.',
      available_memberships: [mockMembership],
    };

    authServiceMock.login.and.returnValues(
      throwError(() => new HttpErrorResponse({ error: initialChallenge, status: 400 })),
      throwError(() => new HttpErrorResponse({ error: staleChallenge, status: 400 })),
    );

    component.email = 'user@test.com';
    component.password = 'password1';
    component.onSignIn();
    tick();

    expect(component.membershipChoices.length).toBe(2);

    component.selectMembership('stale-id');
    component.onSignIn();
    tick();

    expect(component.isMembershipChallenge).toBeTrue();
    expect(component.membershipChoices.length).toBe(1);
    expect(component.membershipChallengeMessage).toContain('no longer available');
  }));
});
