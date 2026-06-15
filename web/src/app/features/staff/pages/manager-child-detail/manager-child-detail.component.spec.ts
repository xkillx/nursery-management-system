import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';

import { StaffApiService } from '../../data/staff-api.service';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerChildDetailComponent } from './manager-child-detail.component';

describe('ManagerChildDetailComponent', () => {
  let fixture: ComponentFixture<ManagerChildDetailComponent>;
  let component: ManagerChildDetailComponent;
  let staffApiMock: jasmine.SpyObj<StaffApiService>;

  const mockChild = {
    id: 'child-1',
    fullName: 'Emma Thompson',
    dateOfBirth: '2022-03-15',
    startDate: '2023-01-10',
    endDate: null,
    coreHourlyRateMinor: null,
    siteCoreHourlyRateMinor: 850,
    notes: null,
    isActive: true,
    leftAt: null,
    leftReasonCode: null,
    leftReasonNote: null,
    enrollmentComplete: false,
    missingRequirements: ['guardian_link'],
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  };

  const mockLinks = [
    {
      id: 'link-1',
      guardianId: 'guardian-1',
      childId: 'child-1',
      guardian: {
        id: 'guardian-1',
        fullName: 'Sarah Thompson',
        email: 'sarah@example.com',
        phone: '+44 7700 900001',
        isActive: true,
      },
      createdAt: '2026-06-07T10:00:00Z',
      updatedAt: '2026-06-07T10:00:00Z',
    },
  ];

  const mockGuardians = [
    {
      id: 'guardian-1',
      fullName: 'Sarah Thompson',
      email: 'sarah@example.com',
      phone: '+44 7700 900001',
      notes: null,
      isActive: true,
      deactivatedAt: null,
      deactivationReasonCode: null,
      deactivationReasonNote: null,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
    {
      id: 'guardian-2',
      fullName: 'John Smith',
      email: null,
      phone: null,
      notes: null,
      isActive: true,
      deactivatedAt: null,
      deactivationReasonCode: null,
      deactivationReasonNote: null,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
  ];

  const mockFundingProfile = {
    id: 'fp-1',
    childId: 'child-1',
    billingMonth: '2026-06',
    fundedAllowanceMinutes: 570,
    createdAt: '2026-06-01T10:00:00Z',
    updatedAt: '2026-06-08T12:00:00Z',
  };

  const mockRegistrationProfile = {
    child: { id: 'child-1', fullName: 'Emma Thompson', dateOfBirth: '2022-03-15' },
    profileExists: true,
    profile: { id: 'rp-1', createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
    demographicsHome: {
      sex: 'female',
      religion: null,
      ethnicOrigin: null,
      firstLanguage: 'English',
      otherLanguages: [],
      homeAddress: { line1: '10 High Street', town: 'Leeds', postcode: 'LS1 1AA' },
      homePostcode: 'LS1 1AA',
      homeTelephone: '0113 123 4567',
      disabilityStatus: null,
      disabilityNotes: null,
      accessRequirements: null,
      demographicsHomeReviewed: true,
    },
    medicalDietary: {
      medicalConditionsStatus: 'yes',
      medicalConditionsNotes: 'Asthma',
      prescribedMedicationStatus: 'yes',
      medicationNotes: 'Blue inhaler',
      immunisationStatus: 'up_to_date',
      immunisationCountry: 'United Kingdom',
      illnessDiagnosisHistory: null,
      dietaryRequirementsStatus: 'yes',
      dietaryRequirementsNotes: 'No gelatine',
      dietarySideEffects: null,
      medicalDietaryReviewed: true,
    },
    healthContacts: null,
    socialDevelopment: null,
    parentCarers: [{
      fullName: 'Sarah Thompson',
      relationshipToChild: 'Mother',
      address: null,
      telephone: '+44 7700 900001',
      email: 'sarah@example.com',
      workAddress: null,
      hasParentalResponsibility: true,
    }],
    emergencyContacts: [{
      fullName: 'Nina Patel',
      relationshipToChild: 'Aunt',
      address: null,
      telephone: '+44 7700 900002',
      email: null,
      workAddress: null,
      hasParentalResponsibility: null,
    }],
    authorisedCollectors: [{
      fullName: 'Omar Khan',
      relationshipToChild: 'Family friend',
      address: null,
      telephone: '+44 7700 900003',
      email: null,
      workAddress: null,
      hasParentalResponsibility: null,
    }],
    collection: {
      isSet: true,
      lastUpdatedAt: '2026-01-01T00:00:00Z',
      lastUpdatedByUserId: 'user-1',
      lastUpdatedByMembershipId: 'membership-1',
      over18CollectionAcknowledged: true,
      emergencyCollectionReviewed: true,
    },
    fundingSupport: null,
    routineCare: null,
    gdprDeclaration: null,
    completeness: { isComplete: false, missingSections: ['child_demographics_home'], sections: [] },
  };

  const mockConsents = {
    child: { id: 'child-1', full_name: 'Emma Thompson', date_of_birth: '2022-03-15' },
    current: {
      id: 'consent-1',
      child_id: 'child-1',
      version: 1,
      source: 'staff',
      paper_form_on_file: true,
      urgent_medical_treatment: true,
      urgent_medical_treatment_exceptions: null,
      plasters: true,
      safeguarding_reporting_acknowledgement: true,
      information_truthfulness_declaration: true,
      area_senco_liaison: false,
      health_visitor_liaison: true,
      transition_documents: true,
      local_outings: false,
      face_painting: true,
      parent_supplied_sun_cream: true,
      parent_supplied_nappy_cream: true,
      development_profile_photos: true,
      nursery_display_boards: true,
      promotional_literature: false,
      nursery_website: false,
      staff_student_coursework: true,
      social_media: false,
      social_media_channel_notes: null,
      notes_exceptions: null,
      entered_by_user_id: 'user-1',
      entered_by_membership_id: 'membership-1',
      created_at: '2026-01-01T00:00:00Z',
    },
    history: [],
    completeness: { is_complete: true, missing_decisions: [] },
  };

  function fundingNotFound404(): HttpErrorResponse {
    return new HttpErrorResponse({
      status: 404,
      error: { code: 'funding_profile_not_found', message: 'not found' },
    });
  }

  beforeEach(async () => {
    staffApiMock = jasmine.createSpyObj('StaffApiService', [
      'getChild', 'listChildGuardianLinks', 'listGuardians', 'updateChild',
      'createGuardianChildLink', 'getFundingProfile', 'upsertFundingProfile',
      'getRegistrationProfile',
      'getRegistrationWorkflowStatus',
      'getRegistrationConsents',
    ]);

    staffApiMock.getChild.and.returnValue(of(mockChild));
    staffApiMock.listChildGuardianLinks.and.returnValue(of(mockLinks));
    staffApiMock.listGuardians.and.returnValue(of(mockGuardians));
    staffApiMock.updateChild.and.returnValue(of(mockChild));
    staffApiMock.createGuardianChildLink.and.returnValue(of({} as any));
    staffApiMock.getFundingProfile.and.returnValue(of(mockFundingProfile));
    staffApiMock.upsertFundingProfile.and.returnValue(of(mockFundingProfile));
    staffApiMock.getRegistrationProfile.and.returnValue(of(mockRegistrationProfile));
    staffApiMock.getRegistrationConsents.and.returnValue(of(mockConsents));
    staffApiMock.getRegistrationWorkflowStatus.and.returnValue(of({
      child: { id: 'child-1', full_name: 'Ada', date_of_birth: '2022-01-15' },
      profile_completeness: { is_complete: false, missing_sections: [] },
      consent_completeness: { is_complete: false, missing_decisions: [] },
      can_mark_complete: false,
      is_reviewed_complete: false,
      needs_review: false,
      missing_groups: [],
    }));

    await TestBed.configureTestingModule({
      imports: [ManagerChildDetailComponent],
      providers: [
        { provide: StaffApiService, useValue: staffApiMock },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: (key: string) => key === 'childId' ? 'child-1' : null }, queryParamMap: { get: () => null } } } },
        ApiErrorMapper,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerChildDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('creates', () => {
    expect(component).toBeTruthy();
  });

  it('loads child detail on init', () => {
    expect(staffApiMock.getChild).toHaveBeenCalledWith('child-1');
    expect(component.child?.fullName).toBe('Emma Thompson');
  });

  it('loads registration profile and consents on init', () => {
    expect(staffApiMock.getRegistrationProfile).toHaveBeenCalledWith('child-1');
    expect(staffApiMock.getRegistrationConsents).toHaveBeenCalledWith('child-1');
    expect(component.registrationProfile?.demographicsHome?.firstLanguage).toBe('English');
    expect(component.currentConsent?.urgent_medical_treatment).toBeTrue();
  });

  it('loads linked guardians', () => {
    expect(staffApiMock.listChildGuardianLinks).toHaveBeenCalledWith('child-1');
    expect(component.linkedGuardians.length).toBe(1);
    expect(component.linkedGuardians[0].guardian.fullName).toBe('Sarah Thompson');
  });

  it('loads all active guardians for selector', () => {
    expect(staffApiMock.listGuardians).toHaveBeenCalledWith({ status: 'active', limit: 200, offset: 0 });
  });

  it('availableGuardians excludes already-linked guardians', () => {
    expect(component.availableGuardians.length).toBe(1);
    expect(component.availableGuardians[0].id).toBe('guardian-2');
  });

  it('shows missing enrollment requirements', () => {
    expect(component.child?.missingRequirements).toContain('guardian_link');
  });

  it('renders real medical and dietary alerts when present', () => {
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent as string;

    expect(text).toContain('Medical: Asthma');
    expect(text).toContain('Medication: Blue inhaler');
    expect(text).toContain('Dietary: No gelatine');
  });

  it('falls back to mock alerts and documents when real data is missing', () => {
    component.registrationProfile = { ...mockRegistrationProfile, medicalDietary: null };
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent as string;

    expect(component.medicalDietaryAlerts).toContain('Peanut allergy - severe');
    expect(text).toContain('Registration Form.pdf');
    expect(text).toContain('Unavailable');
  });

  it('shows collection password status but never the password value', () => {
    fixture.detectChanges();
    const text = fixture.nativeElement.textContent as string;

    expect(text).toContain('Collection password');
    expect(text).toContain('Set');
    expect(text).not.toContain('mysecret');
  });

  it('keeps child page usable when profile and consent loads fail', () => {
    staffApiMock.getRegistrationProfile.and.returnValue(throwError(() => new Error('profile failed')));
    staffApiMock.getRegistrationConsents.and.returnValue(throwError(() => new Error('consent failed')));

    component.ngOnInit();

    expect(component.child?.fullName).toBe('Emma Thompson');
    expect(component.profileLoadError).toContain('Could not load');
    expect(component.consentsLoadError).toContain('Could not load');
  });

  it('shows site core hourly rate formatted as GBP per hour', () => {
    const rate = component.formatSiteRate(component.child!.siteCoreHourlyRateMinor);
    expect(rate).toBe('£8.50/hr');
  });

  it('opens edit form', () => {
    component.onEditChild();
    expect(component.showEditForm).toBeTrue();
  });

  it('closes edit form', () => {
    component.onEditChild();
    component.closeEditForm();
    expect(component.showEditForm).toBeFalse();
  });

  it('saves child and reloads', () => {
    component.childId = 'child-1';
    component.onEditChild();
    component.saveChild({
      full_name: 'Emma Thompson',
      date_of_birth: '2022-03-15',
      start_date: '2023-01-10',
    });

    expect(staffApiMock.updateChild).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      full_name: 'Emma Thompson',
    }));
  });

  it('links guardian and reloads', () => {
    component.childId = 'child-1';
    component.selectedGuardianId = 'guardian-2';
    component.linkGuardian();

    expect(staffApiMock.createGuardianChildLink).toHaveBeenCalledWith({
      guardian_id: 'guardian-2',
      child_id: 'child-1',
    });
  });

  it('handles child load error', () => {
    staffApiMock.getChild.and.returnValue(throwError(() => new Error('not found')));
    component.childId = '';
    component.ngOnInit();

    expect(component.errorMessage).toBeTruthy();
  });

  // --- Funding editor tests ---

  it('loads funding profile for current month on init', () => {
    expect(staffApiMock.getFundingProfile).toHaveBeenCalledWith('child-1', component.selectedBillingMonth);
  });

  it('populates hours and minutes from funding profile', () => {
    expect(component.fundedHoursInput).toBe('9');
    expect(component.fundedMinutesInput).toBe('30');
  });

  it('shows funding profile data after load', () => {
    expect(component.fundingProfile).toBeTruthy();
    expect(component.fundingProfile?.fundedAllowanceMinutes).toBe(570);
  });

  it('selectedBillingMonth matches YYYY-MM format', () => {
    expect(component.selectedBillingMonth).toMatch(/^\d{4}-\d{2}$/);
  });

  it('handles missing funding profile (404) as not-set state', () => {
    staffApiMock.getFundingProfile.and.returnValue(throwError(() => fundingNotFound404()));
    component.childId = 'child-1';
    (component as any).loadFundingProfile();

    expect(component.fundingProfile).toBeNull();
    expect(component.fundedHoursInput).toBe('');
    expect(component.fundedMinutesInput).toBe('');
    expect(component.fundingNotSet).toBeTrue();
  });

  it('fundingNotSet is true when profile is null and not loading', () => {
    component.isLoadingFunding = false;
    component.fundingProfile = null;
    component.fundingErrorMessage = null;
    expect(component.fundingNotSet).toBeTrue();
  });

  it('fundingNotSet is false when profile exists', () => {
    component.fundingProfile = mockFundingProfile;
    expect(component.fundingNotSet).toBeFalse();
  });

  it('reloads funding profile on billing month change', () => {
    const previousCalls = staffApiMock.getFundingProfile.calls.count();
    component.selectedBillingMonth = '2026-05';
    component.onBillingMonthChange();

    expect(staffApiMock.getFundingProfile.calls.count()).toBe(previousCalls + 1);
    expect(staffApiMock.getFundingProfile).toHaveBeenCalledWith('child-1', '2026-05');
  });

  it('clears funding status on billing month change', () => {
    component.fundingStatusMessage = 'Saved';
    component.fundingErrorMessage = 'Error';
    component.onBillingMonthChange();

    expect(component.fundingStatusMessage).toBeNull();
    expect(component.fundingErrorMessage).toBeNull();
  });

  // --- Save and validation tests ---

  it('converts 9 hours 30 minutes to 570 total minutes on save', () => {
    component.fundedHoursInput = '9';
    component.fundedMinutesInput = '30';
    component.selectedBillingMonth = '2026-06';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).toHaveBeenCalledWith('child-1', {
      billing_month: '2026-06',
      funded_allowance_minutes: 570,
    });
  });

  it('saves explicit zero allowance', () => {
    component.fundedHoursInput = '0';
    component.fundedMinutesInput = '0';
    component.selectedBillingMonth = '2026-06';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).toHaveBeenCalledWith('child-1', {
      billing_month: '2026-06',
      funded_allowance_minutes: 0,
    });
  });

  it('treats blank hours as zero when minutes provided', () => {
    component.fundedHoursInput = '';
    component.fundedMinutesInput = '30';
    component.selectedBillingMonth = '2026-06';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).toHaveBeenCalledWith('child-1', {
      billing_month: '2026-06',
      funded_allowance_minutes: 30,
    });
  });

  it('treats blank minutes as zero when hours provided', () => {
    component.fundedHoursInput = '5';
    component.fundedMinutesInput = '';
    component.selectedBillingMonth = '2026-06';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).toHaveBeenCalledWith('child-1', {
      billing_month: '2026-06',
      funded_allowance_minutes: 300,
    });
  });

  it('rejects blank hours and minutes without calling API', () => {
    component.fundedHoursInput = '';
    component.fundedMinutesInput = '';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).not.toHaveBeenCalled();
    expect(component.fundingErrorMessage).toContain('Enter an allowance');
  });

  it('rejects non-integer hours', () => {
    component.fundedHoursInput = '1.5';
    component.fundedMinutesInput = '0';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).not.toHaveBeenCalled();
    expect(component.fundingErrorMessage).toBeTruthy();
  });

  it('rejects negative hours', () => {
    component.fundedHoursInput = '-1';
    component.fundedMinutesInput = '0';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).not.toHaveBeenCalled();
    expect(component.fundingErrorMessage).toContain('non-negative');
  });

  it('rejects minutes above 59', () => {
    component.fundedHoursInput = '0';
    component.fundedMinutesInput = '60';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).not.toHaveBeenCalled();
    expect(component.fundingErrorMessage).toContain('0 and 59');
  });

  it('rejects total exceeding 44640 minutes', () => {
    component.fundedHoursInput = '800';
    component.fundedMinutesInput = '0';

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).not.toHaveBeenCalled();
    expect(component.fundingErrorMessage).toContain('44640');
  });

  it('sets isSavingFunding during save and clears on success', () => {
    const savedProfile = { ...mockFundingProfile, fundedAllowanceMinutes: 300 };
    staffApiMock.upsertFundingProfile.and.returnValue(of(savedProfile));

    component.fundedHoursInput = '5';
    component.fundedMinutesInput = '0';
    component.isSavingFunding = false;

    component.saveFundingAllowance();

    expect(staffApiMock.upsertFundingProfile).toHaveBeenCalled();
    expect(component.isSavingFunding).toBeFalse();
    expect(component.fundingProfile?.fundedAllowanceMinutes).toBe(300);
  });

  it('updates profile and status after successful save', () => {
    const savedProfile = { ...mockFundingProfile, fundedAllowanceMinutes: 300, updatedAt: '2026-06-08T15:00:00Z' };
    staffApiMock.upsertFundingProfile.and.returnValue(of(savedProfile));

    component.fundedHoursInput = '5';
    component.fundedMinutesInput = '0';
    component.saveFundingAllowance();

    expect(component.fundingProfile?.fundedAllowanceMinutes).toBe(300);
    expect(component.fundingStatusMessage).toBe('Saved');
    expect(component.isSavingFunding).toBeFalse();
  });

  it('preserves entered values on validation error', () => {
    component.fundedHoursInput = 'abc';
    component.fundedMinutesInput = '0';

    component.saveFundingAllowance();

    expect(component.fundedHoursInput).toBe('abc');
    expect(component.fundedMinutesInput).toBe('0');
  });

  it('maps enrollment window conflict to actionable message', () => {
    staffApiMock.upsertFundingProfile.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        status: 409,
        error: { code: 'funding_month_outside_enrollment_window', message: 'outside window' },
      })
    ));

    component.fundedHoursInput = '5';
    component.fundedMinutesInput = '0';
    component.saveFundingAllowance();

    expect(component.fundingErrorMessage).toContain('does not overlap');
  });

  it('maps validation error with field errors', () => {
    staffApiMock.upsertFundingProfile.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        status: 422,
        error: { code: 'validation_error', message: 'too high', details: { field: 'funded_allowance_minutes' } },
      })
    ));

    component.fundedHoursInput = '5';
    component.fundedMinutesInput = '0';
    component.saveFundingAllowance();

    expect(component.fundingErrorMessage).toBeTruthy();
  });

  it('maps unknown errors with request id', () => {
    staffApiMock.upsertFundingProfile.and.returnValue(throwError(() =>
      new HttpErrorResponse({
        status: 500,
        error: { code: 'internal_error', message: 'Something went wrong.', request_id: 'req-123' },
      })
    ));

    component.fundedHoursInput = '5';
    component.fundedMinutesInput = '0';
    component.saveFundingAllowance();

    expect(component.fundingErrorMessage).toContain('req-123');
  });

  it('disables save while loading or saving funding', () => {
    component.isLoadingFunding = true;
    expect(component.isLoadingFunding || component.isSavingFunding).toBeTrue();

    component.isLoadingFunding = false;
    component.isSavingFunding = true;
    expect(component.isLoadingFunding || component.isSavingFunding).toBeTrue();
  });
});
