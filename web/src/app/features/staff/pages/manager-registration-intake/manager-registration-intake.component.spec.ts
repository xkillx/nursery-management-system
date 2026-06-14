import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { ManagerRegistrationIntakeComponent } from './manager-registration-intake.component';
import { StaffApiService } from '../../data/staff-api.service';
import { RegistrationDraftStorage } from '../../data/registration-draft.storage';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ConsentWritePayload, RegistrationContactEntry } from '../../models/registration-profile.models';

describe('ManagerRegistrationIntakeComponent', () => {
  let fixture: ComponentFixture<ManagerRegistrationIntakeComponent>;
  let component: ManagerRegistrationIntakeComponent;

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
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ManagerRegistrationIntakeComponent);
    component = fixture.componentInstance;
    component.isNewRegistration = true;
    fixture.detectChanges();
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
    component.step1.surname = 'Smith';
    component.step1.date_of_birth = '2022-01-01';
    component.step1.start_date = '2026-09-01';
    component.step1.home_address = '123 High Street';
    component.step1.first_language = 'English';
    component.step1.disability_status = 'no';
    component.step1.disability_notes = '';
    component.step1.access_requirements = '';

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

    component.step4.signer_name = 'Sarah Johnson';
    component.step4.signed_date = '2026-06-01';
    component.step4.paper_form_on_file = true;
    component.step4.safeguarding_reporting_acknowledgement = true;
    component.step4.information_sharing_consent = true;
    component.step4.gdpr_data_processing_consent = true;
    markAllConsentsReviewed();

    component.officeEvidence = {
      applicationDateStatus: 'complete',
      applicationDate: '2026-05-01',
      birthCertificatePassportStatus: 'unknown',
      proofOfAddressStatus: 'unknown',
      redBookStatus: 'unknown',
      handbookStatus: 'unknown',
      contractStatus: 'unknown',
      notes: '',
      depositStatus: 'unknown',
      depositPaidDate: '',
      sessionsDaysRequestedStatus: 'unknown',
      sessionsDaysRequested: '',
      termTimeOnlySpaceStatus: 'unknown',
      contractDate: '',
      handbookDate: '',
      redBookCheckedDate: '',
      birthCertificatePassportCheckedDate: '',
      proofOfAddressCheckedDate: '',
    };
  }

  describe('canSubmitLocally — happy path', () => {
    it('passes when all required fields are filled for a new registration', () => {
      fillRequiredForCompletion();
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — child profile', () => {
    it('blocks when first name missing', () => {
      fillRequiredForCompletion();
      component.step1.first_name = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when surname missing', () => {
      fillRequiredForCompletion();
      component.step1.surname = '';
      expect(component.canSubmitLocally()).toBe(false);
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
    it('blocks when signer name missing', () => {
      fillRequiredForCompletion();
      component.step4.signer_name = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when signed date missing', () => {
      fillRequiredForCompletion();
      component.step4.signed_date = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when paper form on file not confirmed', () => {
      fillRequiredForCompletion();
      component.step4.paper_form_on_file = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

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

  describe('canSubmitLocally — office evidence', () => {
    it('blocks when deposit status is blank', () => {
      fillRequiredForCompletion();
      component.officeEvidence.depositStatus = undefined;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when sessions/days status is blank', () => {
      fillRequiredForCompletion();
      component.officeEvidence.sessionsDaysRequestedStatus = undefined;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when term-time-only status is blank', () => {
      fillRequiredForCompletion();
      component.officeEvidence.termTimeOnlySpaceStatus = undefined;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('unknown office statuses are allowed for follow-up', () => {
      fillRequiredForCompletion();
      component.officeEvidence.handbookStatus = 'unknown';
      component.officeEvidence.contractStatus = 'unknown';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('submitRegistration', () => {
    it('routes to first failing step and sets error message when issues exist', () => {
      fillRequiredForCompletion();
      component.step2.allergy_status = 'unknown';
      component.currentStep = 'consents-evidence';

      component.submitRegistration();

      expect(component.currentStep).toBe('medical-health');
      expect(component.errorMessage).toContain('allergies');
      expect(component.finalCompletionIssues.length).toBeGreaterThan(0);
    });
  });
});
