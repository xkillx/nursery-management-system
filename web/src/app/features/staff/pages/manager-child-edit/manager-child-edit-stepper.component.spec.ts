import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { of } from 'rxjs';

import { ManagerChildEditStepperComponent } from './manager-child-edit-stepper.component';
import { StaffApiService } from '../../data/staff-api.service';
import { RegistrationDraftStorage } from '../../data/registration-draft.storage';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ConsentRecord, ConsentWritePayload, RegistrationContactEntry } from '../../models/child-legacy-compat.models';
import { ToastService } from '../../../../shared/services/toast.service';
import { ChildRecord } from '../../models/children.models';

describe('ManagerChildEditStepperComponent', () => {
  let fixture: ComponentFixture<ManagerChildEditStepperComponent>;
  let component: ManagerChildEditStepperComponent;
  let toastErrorSpy: jasmine.Spy;

  beforeEach(async () => {
    localStorage.clear();
    await TestBed.configureTestingModule({
      imports: [ManagerChildEditStepperComponent],
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

    fixture = TestBed.createComponent(ManagerChildEditStepperComponent);
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
      email: 'sarah.johnson@example.com',
      workAddress: null,
      hasParentalResponsibility: null,
    }];
    component.selectedParentId = 'parent-1';
    component.step3.parent1_has_responsibility = true;
    component.step3.parent1_address_street = '123 High Street';
    component.step3.parent1_address_city = 'London';
    component.step3.parent1_address_postcode = 'SW1A 1AA';

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
    ];
    component.emergencyAuthorisedFlags = [false];
    component.emergencyContactAddresses = [''];

    component.step3.collection_password = '';

    component.step4.safeguarding_reporting_acknowledgement = true;
    component.step4.information_sharing_consent = true;
    component.step4.urgent_medical_treatment = true;
    component.step4.plasters = true;
    component.step4.information_truthfulness_declaration = true;
    component.step4.gdpr_data_processing_consent = true;
    component.step4.signer_name = 'Sarah Johnson';
    component.step4.signed_date = component.todayIso;
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
    it('requires parent selection', () => {
      fillRequiredForCompletion();
      component.selectedParentId = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('satisfies completion when parental responsibility is no', () => {
      fillRequiredForCompletion();
      component.step3.parent1_has_responsibility = false;
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('requires at least one emergency contact with name, relationship, phone', () => {
      fillRequiredForCompletion();
      component.emergencyContactsDraft = [emptyContact()];
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when authorised collector present without password', () => {
      fillRequiredForCompletion();
      component.emergencyContactsDraft[0].fullName = 'Mary Hill';
      component.emergencyContactsDraft[0].relationshipToChild = 'Grandparent';
      component.emergencyContactsDraft[0].telephone = '020 9999 0000';
      component.emergencyAuthorisedFlags = [true];
      component.step3.collection_password = '';
      expect(component.canSubmitLocally()).toBe(false);
      component.step3.collection_password = 'secret123';
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('does not require password when no non-parent authorised collector', () => {
      fillRequiredForCompletion();
      component.emergencyAuthorisedFlags = [false];
      component.step3.collection_password = '';
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('canSubmitLocally — consents', () => {
    it('blocks when truthfulness declaration is not granted (required)', () => {
      fillRequiredForCompletion();
      component.step4.information_truthfulness_declaration = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when GDPR consent is not granted (required)', () => {
      fillRequiredForCompletion();
      component.step4.gdpr_data_processing_consent = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when required-acknowledged item is untouched', () => {
      fillRequiredForCompletion();
      component.consentsReviewed['safeguarding_reporting_acknowledgement'] = false;
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('allows required-acknowledged No answer (advisory only, not blocking)', () => {
      fillRequiredForCompletion();
      component.step4.safeguarding_reporting_acknowledgement = false;
      component.consentsReviewed['safeguarding_reporting_acknowledgement'] = true;
      expect(component.canSubmitLocally()).toBe(true);
      expect(component.consentAdvisories.length).toBeGreaterThan(0);
    });

    it('allows optional consents to remain false without blocking', () => {
      fillRequiredForCompletion();
      component.setConsentValue('social_media', false);
      component.setConsentValue('face_painting', false);
      expect(component.canSubmitLocally()).toBe(true);
    });

    it('blocks when signer_name is empty', () => {
      fillRequiredForCompletion();
      component.step4.signer_name = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('blocks when signed_date is empty', () => {
      fillRequiredForCompletion();
      component.step4.signed_date = '';
      expect(component.canSubmitLocally()).toBe(false);
    });

    it('passes when all required and required-acknowledged items are answered and audit trail is filled', () => {
      fillRequiredForCompletion();
      expect(component.canSubmitLocally()).toBe(true);
    });
  });

  describe('tier classification', () => {
    it('classifies GDPR and truthfulness as required', () => {
      expect(component.consentTier('gdpr_data_processing_consent')).toBe('required');
      expect(component.consentTier('information_truthfulness_declaration')).toBe('required');
    });

    it('classifies safeguarding, information sharing, urgent medical, and plasters as required-acknowledged', () => {
      expect(component.consentTier('safeguarding_reporting_acknowledgement')).toBe('required_acknowledged');
      expect(component.consentTier('information_sharing_consent')).toBe('required_acknowledged');
      expect(component.consentTier('urgent_medical_treatment')).toBe('required_acknowledged');
      expect(component.consentTier('plasters')).toBe('required_acknowledged');
    });

    it('classifies professional liaison, activities, and photos as optional', () => {
      expect(component.consentTier('area_senco_liaison')).toBe('optional');
      expect(component.consentTier('local_outings')).toBe('optional');
      expect(component.consentTier('social_media')).toBe('optional');
    });


  });

  describe('consent advisories', () => {
    it('records a "No" on a required-acknowledged item as an advisory', () => {
      fillRequiredForCompletion();
      component.step4.urgent_medical_treatment = false;
      component.consentsReviewed['urgent_medical_treatment'] = true;
      const advisories = component.collectConsentAdvisories();
      expect(advisories.length).toBe(1);
      expect(advisories[0].field).toBe('urgent_medical_treatment');
    });

    it('does not record a "Yes" on a required-acknowledged item as an advisory', () => {
      fillRequiredForCompletion();
      expect(component.collectConsentAdvisories().length).toBe(0);
    });
  });

  describe('consent values changed since snapshot', () => {
    it('returns false when no original snapshot exists', () => {
      fillRequiredForCompletion();
      component.originalStep4Snapshot = null;
      expect(component.consentValuesChangedSince(null)).toBe(false);
    });

    it('returns true when a boolean consent differs from the snapshot', () => {
      fillRequiredForCompletion();
      component.originalStep4Snapshot = { ...component.step4 };
      component.step4.social_media = !component.originalStep4Snapshot.social_media;
      expect(component.consentValuesChangedSince(component.originalStep4Snapshot)).toBe(true);
    });

    it('returns false when values match the snapshot', () => {
      fillRequiredForCompletion();
      component.originalStep4Snapshot = { ...component.step4 };
      expect(component.consentValuesChangedSince(component.originalStep4Snapshot)).toBe(false);
    });
  });

  describe('Mark Reviewed/Complete button removed in edit mode', () => {
    it('does not render the no-op Mark Reviewed/Complete button in edit mode', () => {
      fillRequiredForCompletion();
      component.isNewRegistration = false;
      component.currentStep = 'consents-evidence';
      fixture.detectChanges();
      const buttons: HTMLButtonElement[] = Array.from(fixture.nativeElement.querySelectorAll('button'));
      const labels = buttons.map(b => (b.textContent || '').trim());
      expect(labels.some(l => l.includes('Mark Reviewed/Complete'))).toBe(false);
      expect(labels.some(l => l.includes('Save changes'))).toBe(true);
    });
  });

  describe('Reason for change in edit-mode save', () => {
    it('includes consent_change_reason in the API payload when a value differs and a reason is set', () => {
      fillRequiredForCompletion();
      component.isNewRegistration = false;
      component.childId = 'child-1';
      component.currentStep = 'consents-evidence';
      component.originalStep4Snapshot = { ...component.step4 };
      component.loadedSections.add('consent');
      component.step4.social_media = !component.originalStep4Snapshot.social_media;
      component.step4.consent_change_reason = 'Parent called to withdraw social media consent';

      const staffApi = TestBed.inject(StaffApiService);
      const updateSpy = spyOn(staffApi, 'updateChildConsent').and.returnValue(of({ ...component.step4 } as unknown as ConsentRecord));

      component.saveConsentsEvidence();

      expect(updateSpy).toHaveBeenCalled();
      const payload = updateSpy.calls.mostRecent().args[1] as unknown as ConsentWritePayload;
      expect(payload.consent_change_reason).toBe('Parent called to withdraw social media consent');
      expect(payload.signer_name).toBe('Sarah Johnson');
      expect(payload.signed_date).toBe(component.todayIso);
    });

    it('omits consent_change_reason when values match the snapshot', () => {
      fillRequiredForCompletion();
      component.isNewRegistration = false;
      component.childId = 'child-1';
      component.currentStep = 'consents-evidence';
      component.originalStep4Snapshot = { ...component.step4 };
      component.loadedSections.add('consent');
      component.step4.consent_change_reason = 'Should not be sent';

      const staffApi = TestBed.inject(StaffApiService);
      const updateSpy = spyOn(staffApi, 'updateChildConsent').and.returnValue(of({ ...component.step4 } as unknown as ConsentRecord));

      component.saveConsentsEvidence();

      const payload = updateSpy.calls.mostRecent().args[1] as unknown as ConsentWritePayload;
      expect(payload.consent_change_reason).toBeNull();
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

    it('saveContactsCollection blocks on missing parent selection', () => {
      fillRequiredForCompletion();
      component.selectedParentId = '';
      component.currentStep = 'contacts-collection';

      component.saveContactsCollection();

      expect(component.currentStep).toBe('contacts-collection');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('saveConsentsEvidence blocks on untouched required-acknowledged consent', () => {
      fillRequiredForCompletion();
      component.consentsReviewed['safeguarding_reporting_acknowledgement'] = false;
      component.currentStep = 'consents-evidence';

      component.saveConsentsEvidence();

      expect(component.currentStep).toBe('consents-evidence');
      expect(toastErrorSpy).toHaveBeenCalled();
    });

    it('saveConsentsEvidence allows No on a required-acknowledged item', () => {
      fillRequiredForCompletion();
      component.step4.safeguarding_reporting_acknowledgement = false;
      component.consentsReviewed['safeguarding_reporting_acknowledgement'] = true;
      component.isNewRegistration = false;
      component.childId = 'child-1';
      component.currentStep = 'consents-evidence';
      component.loadedSections.add('consent');

      spyOn(component['staffApi'], 'updateChildConsent').and.returnValue(of({ ...component.step4 } as unknown as ConsentRecord));
      component.saveConsentsEvidence();

      expect(toastErrorSpy).not.toHaveBeenCalled();
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

  describe('wizard step list', () => {
    it('new-registration step list has five entries with the expected keys', () => {
      component.isNewRegistration = true;
      const keys = component.steps.map(s => s.key);
      expect(component.steps.length).toBe(5);
      expect(keys).toEqual([
        'child-basics',
        'medical-health',
        'contacts-collection',
        'consents-evidence',
        'funding',
      ]);
    });

    it('edit-registration step list has five entries with the expected keys', () => {
      component.isNewRegistration = false;
      const keys = component.steps.map(s => s.key);
      expect(component.steps.length).toBe(5);
      expect(keys).toEqual([
        'child-basics',
        'medical-health',
        'contacts-collection',
        'consents-evidence',
        'funding',
      ]);
    });
  });







  describe('page heading — edit vs new registration', () => {
    it('renders "Child registration" heading in edit mode', () => {
      component.isNewRegistration = false;
      component.currentStep = 'child-basics';
      fixture.detectChanges();

      const headings: HTMLElement[] = Array.from(fixture.nativeElement.querySelectorAll('h2'));
      const text = headings.map(h => (h.textContent || '').trim()).join('|');
      expect(text).toContain('Child Profile');
      expect(text).not.toContain('Edit child');
    });

    it('renders "Child registration" heading in new-registration mode', () => {
      component.isNewRegistration = true;
      component.currentStep = 'child-basics';
      fixture.detectChanges();

      const headings: HTMLElement[] = Array.from(fixture.nativeElement.querySelectorAll('h2'));
      const text = headings.map(h => (h.textContent || '').trim()).join('|');
      expect(text).toContain('Child Profile');
      expect(text).not.toContain('Edit child');
    });
  });



  describe('photo picker', () => {
    it('onPhotoSelected stores file and creates preview URL for valid JPEG under 5 MB', () => {
      const file = new File(['x'.repeat(100)], 'photo.jpg', { type: 'image/jpeg' });
      const input = document.createElement('input');
      input.type = 'file';
      const dt = new DataTransfer();
      dt.items.add(file);
      input.files = dt.files;
      const event = new Event('change');
      Object.defineProperty(event, 'target', { value: input });

      component.onPhotoSelected(event);

      expect(component.selectedPhotoFile).toBe(file);
      expect(component.photoPreviewUrl).toBeTruthy();
      expect(component.photoErrorMessage).toBeNull();
    });

    it('onPhotoSelected stores file and creates preview URL for valid PNG under 5 MB', () => {
      const file = new File(['x'.repeat(100)], 'photo.png', { type: 'image/png' });
      const input = document.createElement('input');
      input.type = 'file';
      const dt = new DataTransfer();
      dt.items.add(file);
      input.files = dt.files;
      const event = new Event('change');
      Object.defineProperty(event, 'target', { value: input });

      component.onPhotoSelected(event);

      expect(component.selectedPhotoFile).toBe(file);
      expect(component.photoPreviewUrl).toBeTruthy();
      expect(component.photoErrorMessage).toBeNull();
    });

    it('onPhotoSelected rejects file over 5 MB', () => {
      const largeFile = new File(['x'.repeat(6 * 1024 * 1024)], 'large.jpg', { type: 'image/jpeg' });
      const input = document.createElement('input');
      input.type = 'file';
      const dt = new DataTransfer();
      dt.items.add(largeFile);
      input.files = dt.files;
      const event = new Event('change');
      Object.defineProperty(event, 'target', { value: input });

      component.onPhotoSelected(event);

      expect(component.selectedPhotoFile).toBeNull();
      expect(component.photoErrorMessage).toContain('5 MB');
    });

    it('onPhotoSelected rejects non-JPEG/PNG file', () => {
      const file = new File(['data'], 'doc.pdf', { type: 'application/pdf' });
      const input = document.createElement('input');
      input.type = 'file';
      const dt = new DataTransfer();
      dt.items.add(file);
      input.files = dt.files;
      const event = new Event('change');
      Object.defineProperty(event, 'target', { value: input });

      component.onPhotoSelected(event);

      expect(component.selectedPhotoFile).toBeNull();
      expect(component.photoErrorMessage).toContain('JPEG and PNG');
    });

    it('onPhotoSelected replaces old file and revokes old preview URL', () => {
      const file1 = new File(['a'], 'photo1.jpg', { type: 'image/jpeg' });
      const file2 = new File(['b'], 'photo2.jpg', { type: 'image/jpeg' });

      const input1 = document.createElement('input');
      input1.type = 'file';
      const dt1 = new DataTransfer();
      dt1.items.add(file1);
      input1.files = dt1.files;
      const event1 = new Event('change');
      Object.defineProperty(event1, 'target', { value: input1 });
      component.onPhotoSelected(event1);

      const firstUrl = component.photoPreviewUrl;
      expect(firstUrl).toBeTruthy();

      const input2 = document.createElement('input');
      input2.type = 'file';
      const dt2 = new DataTransfer();
      dt2.items.add(file2);
      input2.files = dt2.files;
      const event2 = new Event('change');
      Object.defineProperty(event2, 'target', { value: input2 });
      component.onPhotoSelected(event2);

      expect(component.selectedPhotoFile).toBe(file2);
      expect(component.photoPreviewUrl).toBeTruthy();
      expect(component.photoPreviewUrl).not.toBe(firstUrl);
    });

    it('removePhoto clears stored file and preview URL in registration mode', () => {
      const file = new File(['x'], 'photo.jpg', { type: 'image/jpeg' });
      component.selectedPhotoFile = file;
      component.photoPreviewUrl = 'blob:test';
      component.isNewRegistration = true;

      component.removePhoto();

      expect(component.selectedPhotoFile).toBeNull();
      expect(component.photoPreviewUrl).toBeNull();
    });

    it('removePhoto calls API in edit mode with existing photo', () => {
      component.isNewRegistration = false;
      component.childId = 'child-1';
      component.child = { id: 'child-1', photoUrl: 'https://example.com/photo.jpg' } as unknown as ChildRecord;

      const staffApi = TestBed.inject(StaffApiService);
      const removeSpy = spyOn(staffApi, 'removePhoto').and.returnValue(of({ photo_url: null }));

      component.removePhoto();

      expect(removeSpy).toHaveBeenCalledWith('child-1');
      expect(component.child!.photoUrl).toBeNull();
    });


  });

  describe('developmental concerns gate', () => {
    it('developmentalConcernsGate returns no by default', () => {
      expect(component.developmentalConcernsGate).toBe('no');
    });

    it('setDevelopmentalConcernsGate(yes) sets gate to yes', () => {
      (component as unknown as { setDevelopmentalConcernsGate: (v: 'yes' | 'no') => void }).setDevelopmentalConcernsGate('yes');
      expect(component.developmentalConcernsGate).toBe('yes');
    });

    it('setDevelopmentalConcernsGate(no) clears all concern booleans', () => {
      component.step2.concern_walking = true;
      component.step2.concern_speech_language = true;
      component.step2.concern_hearing = true;
      component.step2.concern_sight = true;
      component.step2.concern_emotional_wellbeing = true;
      component.step2.concern_behaviour = true;

      (component as unknown as { setDevelopmentalConcernsGate: (v: 'yes' | 'no') => void }).setDevelopmentalConcernsGate('no');

      expect(component.step2.concern_walking).toBeFalse();
      expect(component.step2.concern_speech_language).toBeFalse();
      expect(component.step2.concern_hearing).toBeFalse();
      expect(component.step2.concern_sight).toBeFalse();
      expect(component.step2.concern_emotional_wellbeing).toBeFalse();
      expect(component.step2.concern_behaviour).toBeFalse();
    });

    it('setDevelopmentalConcernsGate(yes) does not modify existing concern values', () => {
      component.step2.concern_walking = true;
      component.step2.concern_speech_language = false;

      (component as unknown as { setDevelopmentalConcernsGate: (v: 'yes' | 'no') => void }).setDevelopmentalConcernsGate('yes');

      expect(component.step2.concern_walking).toBeTrue();
      expect(component.step2.concern_speech_language).toBeFalse();
    });

    it('step2FieldError returns error when gate is yes and no concerns checked', () => {
      (component as unknown as { setDevelopmentalConcernsGate: (v: 'yes' | 'no') => void }).setDevelopmentalConcernsGate('yes');
      expect(component.developmentalConcernsGate).toBe('yes');

      const error = (component as unknown as { step2FieldError: (f: string) => string | null }).step2FieldError('developmental_concerns_gate');
      expect(error).toContain('Select at least one developmental concern');
    });

    it('step2FieldError returns null when gate is no', () => {
      const error = (component as unknown as { step2FieldError: (f: string) => string | null }).step2FieldError('developmental_concerns_gate');
      expect(error).toBeNull();
    });

    it('step2FieldError returns null when gate is yes and at least one concern checked', () => {
      (component as unknown as { setDevelopmentalConcernsGate: (v: 'yes' | 'no') => void }).setDevelopmentalConcernsGate('yes');
      component.step2.concern_speech_language = true;

      const error = (component as unknown as { step2FieldError: (f: string) => string | null }).step2FieldError('developmental_concerns_gate');
      expect(error).toBeNull();
    });

    it('collectMedicalSafetyIssues returns no gate issue when gate is no', () => {
      fillRequiredForCompletion();

      const issues = (component as unknown as { collectFinalCompletionIssues: () => { field: string; message: string }[] }).collectFinalCompletionIssues();
      const gateIssue = issues.find((i) => i.field === 'developmental_concerns_gate');
      expect(gateIssue).toBeUndefined();
    });

    it('collectMedicalSafetyIssues returns no gate issue when gate is yes and at least one concern checked', () => {
      fillRequiredForCompletion();
      component.step2.concern_speech_language = true;

      const issues = (component as unknown as { collectFinalCompletionIssues: () => { field: string; message: string }[] }).collectFinalCompletionIssues();
      const gateIssue = issues.find((i) => i.field === 'developmental_concerns_gate');
      expect(gateIssue).toBeUndefined();
    });
  });
});

