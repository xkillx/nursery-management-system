import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpErrorResponse } from '@angular/common/http';
import { ActivatedRoute } from '@angular/router';
import { of, throwError } from 'rxjs';

import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ManagerChildRegistrationComponent } from './manager-child-registration.component';
import { RegistrationProfileResponse } from '../../models/registration-profile.models';
import { ChildRecord } from '../../models/children.models';

describe('ManagerChildRegistrationComponent', () => {
  let fixture: ComponentFixture<ManagerChildRegistrationComponent>;
  let component: ManagerChildRegistrationComponent;
  let staffApiMock: jasmine.SpyObj<StaffApiService>;
  let roomsApiMock: jasmine.SpyObj<StaffRoomsApiService>;
  let authStub: Partial<AuthService>;

  const mockChild: ChildRecord = {
    id: 'child-1',
    firstName: 'Emma',
    middleName: null,
    lastName: 'Thompson',
    fullName: 'Emma Thompson',
    dateOfBirth: '2022-03-15',
    startDate: '2024-01-01',
    endDate: null,
    coreHourlyRateMinor: null,
    siteCoreHourlyRateMinor: null,
    notes: null,
    isActive: true,
    leftAt: null,
    leftReasonCode: null,
    leftReasonNote: null,
    primaryRoomId: 'room-1',
    enrollmentComplete: false,
    missingRequirements: [],
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  };

  const mockProfile: RegistrationProfileResponse = {
    child: { id: 'child-1', fullName: 'Emma Thompson', dateOfBirth: '2022-03-15' },
    profileExists: true,
    profile: { id: 'rp-1', createdAt: '2026-01-01T00:00:00Z', updatedAt: '2026-01-01T00:00:00Z' },
    demographicsHome: {
      sex: 'female', religion: null, ethnicOrigin: null, firstLanguage: 'English',
      otherLanguages: null, homeAddress: null, homePostcode: null, homeTelephone: null,
      disabilityStatus: 'no', disabilityNotes: null, accessRequirements: null,
      demographicsHomeReviewed: false,
    },
    medicalDietary: {
      medicalConditionsStatus: 'unknown', medicalConditionsNotes: null,
      prescribedMedicationStatus: 'unknown', medicationNotes: null,
      immunisationStatus: 'unknown', immunisationCountry: null,
      illnessDiagnosisHistory: null, dietaryRequirementsStatus: 'unknown',
      dietaryRequirementsNotes: null, dietarySideEffects: null,
      medicalDietaryReviewed: false,
    },
    healthContacts: {
      doctorName: null, doctorAddress: null, doctorPhone: null,
      healthVisitorName: null, healthVisitorAddress: null, healthVisitorPhone: null,
      healthContactsReviewed: false,
    },
    socialDevelopment: {
      socialServicesStatus: 'unknown', socialServicesNotes: null,
      socialWorkerName: null, socialWorkerPhone: null, socialWorkerEmail: null, concernWalking: 'unknown',
      concernSpeechLanguage: 'unknown', concernHearing: 'unknown',
      concernSight: 'unknown', concernEmotionalWellbeing: 'unknown',
      concernBehaviour: 'unknown', professionalReferrals: [],
      socialDevelopmentReviewed: false,
    },
    parentCarers: [],
    emergencyContacts: [],
    authorisedCollectors: [],
    collection: { isSet: false, lastUpdatedAt: null, lastUpdatedByUserId: null, lastUpdatedByMembershipId: null, over18CollectionAcknowledged: false, emergencyCollectionReviewed: false },
    fundingSupport: {
      benefitsContributeToFees: 'unknown', workingTaxCredit: 'unknown',
      collegeUniPaidToParent: 'unknown', collegeUniPaidToNursery: 'unknown',
      funding3yoTermTime: 'unknown', funding2yoTermTime: 'unknown',
      fundingSupportNotes: null, fundingSupportReviewed: false,
    },
    routineCare: { routineCareNotes: null, routineCareReviewed: false },
    gdprDeclaration: { gdprDeclaredByName: null, gdprDeclaredAt: null, gdprDeclarationDate: null },
    completeness: { isComplete: false, missingSections: ['child_demographics_home'], sections: [{ code: 'child_demographics_home', status: 'incomplete', missingFields: ['review_required'] }] },
  };

  beforeEach(async () => {
    staffApiMock = jasmine.createSpyObj('StaffApiService', [
      'getRegistrationProfile', 'patchRegistrationProfile',
      'setRegistrationCollectionPassword',
      'getChild', 'updateChild',
    ]);

    staffApiMock.getRegistrationProfile.and.returnValue(of(mockProfile));
    staffApiMock.patchRegistrationProfile.and.returnValue(of(mockProfile));
    staffApiMock.setRegistrationCollectionPassword.and.returnValue(of(mockProfile));
    staffApiMock.getChild.and.returnValue(of(mockChild));
    staffApiMock.updateChild.and.returnValue(of(mockChild));

    roomsApiMock = jasmine.createSpyObj('StaffRoomsApiService', ['listRooms']);
    roomsApiMock.listRooms.and.returnValue(of([]));

    authStub = {
      activeMembership: (() => ({ branch_id: 'branch-1' })) as AuthService['activeMembership'],
    };

    await TestBed.configureTestingModule({
      imports: [ManagerChildRegistrationComponent],
      providers: [
        { provide: StaffApiService, useValue: staffApiMock },
        { provide: StaffRoomsApiService, useValue: roomsApiMock },
        { provide: AuthService, useValue: authStub },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: (key: string) => key === 'childId' ? 'child-1' : null } } } },
        ApiErrorMapper,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerChildRegistrationComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('creates', () => {
    expect(component).toBeTruthy();
  });

  it('loads registration profile on init', () => {
    expect(staffApiMock.getRegistrationProfile).toHaveBeenCalledWith('child-1');
    expect(component.childName).toBe('Emma Thompson');
  });

  it('initializes section drafts from profile', () => {
    expect(component.demoHomeDraft).toBeTruthy();
    expect(component.demoHomeDraft?.sex).toBe('female');
    expect(component.demoHomeDraft?.disabilityStatus).toBe('no');
    expect(component.medicalDietaryDraft).toBeTruthy();
    expect(component.healthContactsDraft).toBeTruthy();
    expect(component.socialDevDraft).toBeTruthy();
    expect(component.fundingSupportDraft).toBeTruthy();
    expect(component.routineCareDraft).toBeTruthy();
  });

  it('initializes contact drafts as empty arrays', () => {
    expect(component.parentCarersDraft).toEqual([]);
    expect(component.emergencyContactsDraft).toEqual([]);
    expect(component.authorisedCollectorsDraft).toEqual([]);
  });

  it('renders profile completion badge', () => {
    const rendered = fixture.nativeElement as HTMLElement;
    expect(rendered.textContent).toContain('Registration profile');
  });

  it('shows missing registration sections when incomplete', () => {
    const rendered = fixture.nativeElement as HTMLElement;
    expect(rendered.textContent).toContain('Missing registration sections');
  });

  it('does not contain placeholder text about fields rendered later', () => {
    const rendered = fixture.nativeElement as HTMLElement;
    expect(rendered.textContent).not.toContain('would be rendered here');
  });

  it('does not contain placeholder text about checklist items rendered later', () => {
    const rendered = fixture.nativeElement as HTMLElement;
    expect(rendered.textContent).not.toContain('would be rendered here');
  });

  /* eslint-disable @typescript-eslint/no-explicit-any */
  const c = () => component as any;

  it('saves demographics home section', () => {
    c().saveDemographicsHome();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      demographics_home: jasmine.objectContaining({ sex: 'female' }),
    }));
  });

  it('saves medical dietary section', () => {
    c().saveMedicalDietary();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      medical_dietary: jasmine.objectContaining({ medical_conditions_status: 'unknown' }),
    }));
  });

  it('saves health contacts section', () => {
    c().saveHealthContacts();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      health_contacts: jasmine.any(Object),
    }));
  });

  it('saves funding support section', () => {
    c().saveFundingSupport();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      funding_support: jasmine.any(Object),
    }));
  });

  it('saves routine care section', () => {
    c().saveRoutineCare();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      routine_care: jasmine.any(Object),
    }));
  });

  it('saves GDPR declaration', () => {
    c().saveGdprDeclaration();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      gdpr_declaration: jasmine.any(Object),
    }));
  });

  it('saves parent/carer contacts array', () => {
    component.parentCarersDraft = [{ fullName: 'Sarah Thompson', relationshipToChild: 'Mother', address: null, telephone: '+44 7700 900001', email: null, workAddress: null, hasParentalResponsibility: true }];
    c().saveContacts('parent_carers');
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', {
      parent_carers: [{ fullName: 'Sarah Thompson', relationshipToChild: 'Mother', address: null, telephone: '+44 7700 900001', email: null, workAddress: null, hasParentalResponsibility: true }],
    });
  });

  it('saves emergency contacts array', () => {
    component.emergencyContactsDraft = [{ fullName: 'Jane Doe', relationshipToChild: 'Aunt', address: null, telephone: null, email: null, workAddress: null, hasParentalResponsibility: null }];
    c().saveContacts('emergency_contacts');
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', {
      emergency_contacts: [{ fullName: 'Jane Doe', relationshipToChild: 'Aunt', address: null, telephone: null, email: null, workAddress: null, hasParentalResponsibility: null }],
    });
  });

  it('saves authorised collectors array', () => {
    component.authorisedCollectorsDraft = [{ fullName: 'Bob Smith', relationshipToChild: 'Grandfather', address: null, telephone: null, email: null, workAddress: null, hasParentalResponsibility: null }];
    c().saveContacts('authorised_collectors');
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', {
      authorised_collectors: [{ fullName: 'Bob Smith', relationshipToChild: 'Grandfather', address: null, telephone: null, email: null, workAddress: null, hasParentalResponsibility: null }],
    });
  });

  it('saves collection password via dedicated endpoint', () => {
    component.collectionPassword = 'mysecret';
    c().setCollectionPassword();
    expect(staffApiMock.setRegistrationCollectionPassword).toHaveBeenCalledWith('child-1', 'mysecret');
  });

  it('clears collection password input after successful save', () => {
    component.collectionPassword = 'mysecret';
    c().setCollectionPassword();
    expect(component.collectionPassword).toBe('');
  });

  it('does not display collection password value', () => {
    component.collectionPassword = 'mysecret';
    c().setCollectionPassword();
    const rendered = fixture.nativeElement as HTMLElement;
    expect(rendered.textContent).not.toContain('mysecret');
  });

  it('saves collection review flags', () => {
    component.collectionOver18 = true;
    component.collectionEmergencyReviewed = true;
    c().saveCollectionFlags();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalledWith('child-1', jasmine.objectContaining({
      collection: jasmine.objectContaining({ over18_collection_acknowledged: true, emergency_collection_reviewed: true }),
    }));
  });

  it('adds a contact row to parent carers', () => {
    c().addContactRow(component.parentCarersDraft);
    expect(component.parentCarersDraft.length).toBe(1);
    expect(component.parentCarersDraft[0].fullName).toBe('');
  });

  it('removes a contact row', () => {
    component.parentCarersDraft = [{ fullName: 'Test', relationshipToChild: null, address: null, telephone: null, email: null, workAddress: null, hasParentalResponsibility: null }];
    c().removeContactRow(component.parentCarersDraft, 0);
    expect(component.parentCarersDraft.length).toBe(0);
  });

  it('shows section error on API failure without losing draft values', () => {
    staffApiMock.patchRegistrationProfile.and.returnValue(throwError(() =>
      new HttpErrorResponse({ status: 422, error: { code: 'validation_error', message: 'Invalid value' } })
    ));
    component.demoHomeDraft!.sex = 'female';
    c().saveDemographicsHome();
    expect(staffApiMock.patchRegistrationProfile).toHaveBeenCalled();
    expect(component.sectionErrors['demographics_home']).toBeTruthy();
    expect(component.demoHomeDraft?.sex).toBe('female');
  });

  it('reloads profile drafts after successful save', () => {
    staffApiMock.patchRegistrationProfile.and.returnValue(of(mockProfile));
    c().saveDemographicsHome();
    expect(component.sectionMessages['demographics_home']).toBe('Section saved.');
    expect(component.demoHomeDraft?.sex).toBe('female');
  });

  it('adds emergency contact row', () => {
    c().addContactRow(component.emergencyContactsDraft);
    expect(component.emergencyContactsDraft.length).toBe(1);
  });

  it('adds authorised collector row', () => {
    c().addContactRow(component.authorisedCollectorsDraft);
    expect(component.authorisedCollectorsDraft.length).toBe(1);
  });

  it('loads child record to expose primaryRoomId', () => {
    expect(staffApiMock.getChild).toHaveBeenCalledWith('child-1');
    expect(component.primaryRoomId).toBe('room-1');
  });

  it('shows paper-form completion date from the loaded profile', () => {
    component.profile = { ...mockProfile, paperFormCompletedDate: '2026-05-01' };
    expect(component.paperFormCompletedDateDisplay).toBe('2026-05-01');
  });

  it('shows em-dash when paper-form completion date is missing', () => {
    component.profile = { ...mockProfile, paperFormCompletedDate: null };
    expect(component.paperFormCompletedDateDisplay).toBe('—');
  });

  it('saves primary room via updateChild', () => {
    component.primaryRoomId = 'room-2';
    c().savePrimaryRoom();
    expect(staffApiMock.updateChild).toHaveBeenCalledWith('child-1', jasmine.objectContaining({ primary_room_id: 'room-2' }));
    expect(component.roomSaveMessage).toBe('Primary room saved.');
  });

  it('shows error when primary room save fails', () => {
    staffApiMock.updateChild.and.returnValue(throwError(() => new HttpErrorResponse({ status: 500 })));
    component.primaryRoomId = 'room-2';
    c().savePrimaryRoom();
    expect(component.roomSaveError).toBeTruthy();
  });
});
