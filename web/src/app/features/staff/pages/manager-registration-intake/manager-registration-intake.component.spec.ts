import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ManagerRegistrationIntakeComponent } from './manager-registration-intake.component';
import { StaffApiService } from '../../data/staff-api.service';
import { RegistrationDraftStorage } from '../../data/registration-draft.storage';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ConsentWritePayload, RegistrationContactEntry } from '../../models/registration-profile.models';
import { ToastService } from '../../../../shared/services/toast.service';

describe('ManagerRegistrationIntakeComponent', () => {
  let fixture: ComponentFixture<ManagerRegistrationIntakeComponent>;
  let component: ManagerRegistrationIntakeComponent;
  let toastErrorSpy: jasmine.Spy;

  beforeEach(async () => {
    localStorage.clear();
    await TestBed.configureTestingModule({
      imports: [ManagerRegistrationIntakeComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        StaffApiService,
        RegistrationDraftStorage,
        ApiErrorMapper,
        ToastService,
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerRegistrationIntakeComponent);
    component = fixture.componentInstance;
    component.isNewRegistration = true;
    fixture.detectChanges();
    const toast = TestBed.inject(ToastService);
    toastErrorSpy = spyOn(toast, 'error').and.callThrough();
  });

  function emptyContact(): RegistrationContactEntry {
    return {
      fullName: '',
      relationshipToChild: '',
      address: null,
      telephone: '',
      email: null,
      workAddress: null,
      hasParentalResponsibility: null,
    };
  }

  function markAllConsentsReviewed(): void {
    const keys = Object.keys(component.step4) as (keyof ConsentWritePayload)[];
    for (const key of keys) {
      if (typeof component.step4[key] === 'boolean') {
        component.consentsReviewed[key] = true;
      }
    }
  }

  function fillRequiredForCompletion(): void {
    component.step1.first_name = 'James';
    component.step1.last_name = 'Smith';
    component.step1.date_of_birth = '2022-01-01';
    component.step1.start_date = '2026-09-01';
    component.step1.home_address = '123 High Street';
    component.step1.first_language = 'English';
    component.step1.disability_status = 'no';
    component.step1.disability_notes = '';
    component.step1.access_requirements = '';
    component.step1.primary_room_id = 'room-1';
    component.step1.registration_date = component.todayIso;

    component.step2.allergy_status = 'no';
    component.step2.allergy_details = '';
    component.step2.medication_status = 'no';
    component.step2.medication_name = '';
    component.step2.medication_dosage = '';
    component.step2.dietary_status = 'none';
    component.step2.special_dietary_requirements = '';
    component.step2.medical_history_status = 'none';
    component.step2.illness_diagnosis_history = '';
    component.step2.social_services_status = 'no';
    component.step2.social_services_details = '';

    component.parentCarersDraft = [{
      fullName: 'Sarah Johnson',
      relationshipToChild: 'Mother',
      address: null,
      telephone: '020 1234 5678',
      email: null,
      workAddress: null,
      hasParentalResponsibility: null,
    }];
    component.step3.parent1_has_responsibility = true;

    component.emergencyContactsDraft = [
      {
        fullName: 'Mary Hill',
        relationshipToChild: 'Grandparent',
        address: null,
        telephone: '020 9999 0000',
        email: null,
        workAddress: null,
        hasParentalResponsibility: null,
      },
      emptyContact(),
    ];
    component.emergencyAuthorisedFlags = [false, false];
    component.emergencyContactAddresses = ['', ''];

    component.step3.funding_support_answer = 'no';
    component.step3.working_tax_credit = false;
    component.step3.college_uni_paid_to_parent = false;
    component.step3.funding_3yo_term_time = false;
    component.step3.funding_2yo_term_time = false;
    component.step3.other_funding_selected = false;
    component.step3.other_benefits = '';
    component.step3.collection_password = '';

    component.step4.safeguarding_reporting_acknowledgement = true;
    component.step4.information_sharing_consent = true;
    component.step4.information_truthfulness_declaration = true;
    component.step4.gdpr_data_processing_consent = true;
    markAllConsentsReviewed();
  }

  describe('canSubmitLocally — happy path', () => {
    it('passes when all required fields are filled for a new registration', () => {
      fillRequiredForCompletion();
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('Step 1 — primary room and registration date', () => {
    it('blocks Step 1 completion when primary_room_id is empty', () => {
      fillRequiredForCompletion();
      component.step1.primary_room_id = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks Step 1 completion when registration_date is empty', () => {
      fillRequiredForCompletion();
      component.step1.registration_date = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks Step 1 completion when registration_date is in the future', () => {
      fillRequiredForCompletion();
      const future = new Date(Date.now() + 1000 * 60 * 60 * 24).toISOString().slice(0, 10);
      component.step1.registration_date = future;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('defaults registration_date to today for a new registration', () => {
      expect(component.step1.registration_date).toBe(component.todayIso);
    });

    it('hydrates primary_room_id and registration_date from a restored draft', () => {
      const draft = {
        currentStep: 'child-basics',
        step1: {
          first_name: 'James',
          last_name: 'Smith',
          date_of_birth: '2022-01-01',
          start_date: '2026-09-01',
          home_address: '123 High Street',
          first_language: 'English',
          primary_room_id: 'room-99',
          registration_date: '2026-01-15',
        },
      };
      localStorage.setItem('nursery.registration_intake.draft', JSON.stringify(draft));
      component.restoreDraftIfPresentPublic();
      expect(component.step1.primary_room_id).toBe('room-99');
      expect(component.step1.registration_date).toBe('2026-01-15');
    });
  });

  describe('canSubmitLocally — child profile', () => {
    it('blocks when first name missing', () => {
      fillRequiredForCompletion();
      component.step1.first_name = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('does not block when last name is missing — optional field', () => {
      fillRequiredForCompletion();
      component.step1.last_name = '';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('blocks when date of birth missing', () => {
      fillRequiredForCompletion();
      component.step1.date_of_birth = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when start date missing', () => {
      fillRequiredForCompletion();
      component.step1.start_date = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when home address missing', () => {
      fillRequiredForCompletion();
      component.step1.home_address = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when first language missing', () => {
      fillRequiredForCompletion();
      component.step1.first_language = '';
      expect(component.canSubmitLocally()).toBe(false);
    });
  });

  describe('canSubmitLocally — disability / SEND / access', () => {
    it('blocks when disability status is blank', () => {
      fillRequiredForCompletion();
      component.step1.disability_status = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when disability status is unknown', () => {
      fillRequiredForCompletion();
      component.step1.disability_status = 'unknown';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires notes when disability status is yes', () => {
      fillRequiredForCompletion();
      component.step1.disability_status = 'yes';
      component.step1.disability_notes = '';
      component.step1.access_requirements = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step1.disability_notes = 'Asthma plan';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('satisfies completion when disability status is no', () => {
      fillRequiredForCompletion();
      component.step1.disability_status = 'no';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — allergies', () => {
    it('blocks when allergies blank', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when allergies unknown', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = 'unknown';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('satisfies completion when allergies no', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = 'no';
      component.step2.allergy_details = '';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('requires details when allergies yes', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = 'yes';
      component.step2.allergy_details = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step2.allergy_details = 'Peanuts — severe';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — medication', () => {
    it('blocks when medication yes without name or dosage', () => {
      fillRequiredForCompletion();
      component.step2.medication_status = 'yes';
      component.step2.medication_name = '';
      component.step2.medication_dosage = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('satisfies completion when medication yes with name and dosage', () => {
      fillRequiredForCompletion();
      component.step2.medication_status = 'yes';
      component.step2.medication_name = 'Inhaler';
      component.step2.medication_dosage = '2 puffs';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('satisfies completion when medication no', () => {
      fillRequiredForCompletion();
      component.step2.medication_status = 'no';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('blocks when medication unknown', () => {
      fillRequiredForCompletion();
      component.step2.medication_status = 'unknown';
      expect(component.canSubmitLocally()).toBe(false);
    });
  });

  describe('canSubmitLocally — dietary and medical history', () => {
    it('blocks when dietary blank', () => {
      fillRequiredForCompletion();
      component.step2.dietary_status = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when dietary unknown', () => {
      fillRequiredForCompletion();
      component.step2.dietary_status = 'unknown';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires notes when dietary details selected', () => {
      fillRequiredForCompletion();
      component.step2.dietary_status = 'details';
      component.step2.special_dietary_requirements = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step2.special_dietary_requirements = 'Halal';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('blocks when medical history unknown', () => {
      fillRequiredForCompletion();
      component.step2.medical_history_status = 'unknown';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires notes when medical history details selected', () => {
      fillRequiredForCompletion();
      component.step2.medical_history_status = 'details';
      component.step2.illness_diagnosis_history = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step2.illness_diagnosis_history = 'Asthma diagnosis 2023';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('none satisfies dietary and medical history completion', () => {
      fillRequiredForCompletion();
      component.step2.dietary_status = 'none';
      component.step2.special_dietary_requirements = '';
      component.step2.medical_history_status = 'none';
      component.step2.illness_diagnosis_history = '';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — social services', () => {
    it('blocks when social services unknown', () => {
      fillRequiredForCompletion();
      component.step2.social_services_status = 'unknown';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires details when social services yes', () => {
      fillRequiredForCompletion();
      component.step2.social_services_status = 'yes';
      component.step2.social_services_details = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step2.social_services_details = 'CIN plan in place';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('satisfies completion when social services no', () => {
      fillRequiredForCompletion();
      component.step2.social_services_status = 'no';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — contacts and collection', () => {
    it('requires primary parent full name', () => {
      fillRequiredForCompletion();
      component.parentCarersDraft[0].fullName = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires primary parent relationship', () => {
      fillRequiredForCompletion();
      component.parentCarersDraft[0].relationshipToChild = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires primary parent phone', () => {
      fillRequiredForCompletion();
      component.parentCarersDraft[0].telephone = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('requires parental responsibility answer', () => {
      fillRequiredForCompletion();
      component.step3.parent1_has_responsibility = null;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('satisfies completion when parental responsibility is no', () => {
      fillRequiredForCompletion();
      component.step3.parent1_has_responsibility = false;
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('requires at least one emergency contact with name, relationship, phone', () => {
      fillRequiredForCompletion();
      component.emergencyContactsDraft = [emptyContact(), emptyContact()];
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when authorised collector present without password', () => {
      fillRequiredForCompletion();
      component.emergencyContactsDraft[0].fullName = 'Mary Hill';
      component.emergencyContactsDraft[0].relationshipToChild = 'Grandparent';
      component.emergencyContactsDraft[0].telephone = '020 9999 0000';
      component.emergencyAuthorisedFlags = [true, false];
      component.step3.collection_password = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step3.collection_password = 'secret123';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('does not require password when no non-parent authorised collector', () => {
      fillRequiredForCompletion();
      component.emergencyAuthorisedFlags = [false, false];
      component.step3.collection_password = '';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — funding', () => {
    it('blocks when funding answer missing', () => {
      fillRequiredForCompletion();
      component.step3.funding_support_answer = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when funding yes without any option selected', () => {
      fillRequiredForCompletion();
      component.step3.funding_support_answer = 'yes';
      component.step3.working_tax_credit = false;
      component.step3.college_uni_paid_to_parent = false;
      component.step3.funding_3yo_term_time = false;
      component.step3.funding_2yo_term_time = false;
      component.step3.other_funding_selected = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('satisfies when funding yes with working tax credit', () => {
      fillRequiredForCompletion();
      component.step3.funding_support_answer = 'yes';
      component.step3.working_tax_credit = true;
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('blocks when funding yes with other but no details', () => {
      fillRequiredForCompletion();
      component.step3.funding_support_answer = 'yes';
      component.step3.other_funding_selected = true;
      component.step3.other_benefits = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step3.other_benefits = 'Council scheme';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — consents', () => {
    it('blocks when safeguarding acknowledgement off', () => {
      fillRequiredForCompletion();
      component.step4.safeguarding_reporting_acknowledgement = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when information sharing consent off', () => {
      fillRequiredForCompletion();
      component.step4.information_sharing_consent = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when GDPR consent off', () => {
      fillRequiredForCompletion();
      component.step4.gdpr_data_processing_consent = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('untouched default consent values block new registration completion', () => {
      fillRequiredForCompletion();
      component.consentsReviewed = {};
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('explicit No on optional consent satisfies reviewed gate', () => {
      fillRequiredForCompletion();
      component.setConsentValue('social_media', false);
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('existing registration skips optional-consent reviewed gate', () => {
      fillRequiredForCompletion();
      component.consentsReviewed = {};
      component.isNewRegistration = false;
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('submitRegistration', () => {
    it('routes to first failing step and shows toast when issues exist', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = 'unknown';
      component.currentStep = 'consents-evidence';

      component.submitRegistration();

      expect(component.currentStep).toBe('medical-health');
      expect(toastErrorSpy).toHaveBeenCalledWith(jasmine.any(String), { title: 'Check required details' });
      expect(component.finalCompletionIssues.length).toBeGreaterThan(0);
    });
  });

  describe('step continuation — toast and focus', () => {
    it('saveChildBasics blocks on missing first name and toasts', () => {
      component.currentStep = 'child-basics';
      component.step1.first_name = '';

      component.saveChildBasics();

      expect(component.currentStep).toBe('child-basics');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('saveMedicalHealth blocks on blank allergy and toasts', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = '';
      component.currentStep = 'medical-health';

      component.saveMedicalHealth();

      expect(component.currentStep).toBe('medical-health');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('saveMedicalHealth blocks on legacy unknown medication', () => {
      fillRequiredForCompletion();
      component.step2.medication_status = 'unknown';
      component.currentStep = 'medical-health';

      component.saveMedicalHealth();

      expect(component.currentStep).toBe('medical-health');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('saveContactsCollection blocks on missing primary phone', () => {
      fillRequiredForCompletion();
      component.parentCarersDraft[0].telephone = '';
      component.currentStep = 'contacts-collection';

      component.saveContactsCollection();

      expect(component.currentStep).toBe('contacts-collection');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('saveConsentsEvidence blocks on missing safeguarding consent', () => {
      fillRequiredForCompletion();
      component.step4.safeguarding_reporting_acknowledgement = false;
      component.currentStep = 'consents-evidence';

      component.saveConsentsEvidence();

      expect(component.currentStep).toBe('consents-evidence');
      expect(toastErrorSpy).toHaveBeenCalled();
    });
  });

  describe('step navigation lock', () => {
    it('canOpenStep blocks forward jump when prior step incomplete', () => {
      component.step1.first_name = '';

      expect(component.canOpenStep('medical-health')).toBe(false);
      expect(component.canOpenStep('contacts-collection')).toBe(false);
    });

    it('canOpenStep allows back navigation regardless of completion', () => {
      component.currentStep = 'consents-evidence';

      expect(component.canOpenStep('child-basics')).toBe(true);
      expect(component.canOpenStep('medical-health')).toBe(true);
    });

    it('goToStep toasts and routes to first blocking prior issue', () => {
      component.currentStep = 'child-basics';
      component.step1.first_name = '';

      component.goToStep('medical-health');

      expect(component.currentStep).toBe('child-basics');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('stepIsComplete reflects validation state, not index alone', () => {
      component.currentStep = 'medical-health';
      component.step1.first_name = '';

      expect(component.stepIsComplete('child-basics')).toBe(false);
    });
  });

  describe('option arrays — no manager-facing Unknown', () => {
    it('yesNoUnknownOptions drops unknown for required controls', () => {
      const values = component.yesNoUnknownOptions.map(o => o.value);
      expect(values).not.toContain('unknown');
      expect(values).toContain('yes');
      expect(values).toContain('no');
    });

    it('disabilityStatusOptions drops unknown', () => {
      const values = component.disabilityStatusOptions.map(o => o.value);
      expect(values).not.toContain('unknown');
    });

    it('noneDetailsUnknownOptions drops unknown', () => {
      const values = component.noneDetailsUnknownOptions.map(o => o.value);
      expect(values).not.toContain('unknown');
    });

    it('waitingListOptions drops unknown', () => {
      const values = component.waitingListOptions.map(o => o.value);
      expect(values).not.toContain('unknown');
    });

    it('addReferralEntry default waiting list status is not unknown', () => {
      component.referralsDraft = [];
      component.addReferralEntry();
      expect(component.referralsDraft[0].waitingListStatus).toBe('not_applicable');
    });
  });
});
