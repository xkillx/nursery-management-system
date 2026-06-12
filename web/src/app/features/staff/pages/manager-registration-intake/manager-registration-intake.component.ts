import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { combineLatest, Observable } from 'rxjs';
import {
  heroAcademicCap,
  heroArrowLeft,
  heroArrowRight,
  heroCamera,
  heroCheck,
  heroClipboardDocumentCheck,
  heroDocumentCheck,
  heroExclamationTriangle,
  heroHeart,
  heroIdentification,
  heroInformationCircle,
  heroLockClosed,
  heroPaperAirplane,
  heroPlusCircle,
  heroShieldCheck,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { CheckboxComponent } from '../../../../shared/components/form/input/checkbox.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { RadioComponent } from '../../../../shared/components/form/input/radio.component';
import { SelectComponent, type Option } from '../../../../shared/components/form/select/select.component';
import { TextAreaComponent } from '../../../../shared/components/form/input/text-area.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord, ChildWritePayload } from '../../models/children.models';
import {
  ConsentWritePayload,
  OfficeUseChecklist,
  RegistrationContactEntry,
  RegistrationProfileResponse,
  RegistrationWorkflowStatus,
} from '../../models/registration-profile.models';

type StepperStep =
  | 'child-basics'
  | 'medical-health'
  | 'contacts-collection'
  | 'consents-evidence'
  | 'review-complete';

type IntakeStep = {
  key: StepperStep;
  label: string;
  shortLabel: string;
  description: string;
};

type Step1Field =
  | 'first_name'
  | 'surname'
  | 'date_of_birth'
  | 'start_date'
  | 'sex'
  | 'first_language'
  | 'home_address'
  | 'home_postcode'
  | 'home_telephone'
  | 'notes'
  | 'religion'
  | 'ethnic_origin'
  | 'other_languages'
  | 'disability_status'
  | 'disability_notes'
  | 'access_requirements';

type Step1RequiredField = Extract<Step1Field, 'first_name' | 'surname' | 'date_of_birth' | 'start_date'>;

type ConsentItem = {
  key: keyof ConsentWritePayload;
  label: string;
  detail: string;
};

@Component({
  selector: 'app-manager-registration-intake',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    AlertComponent,
    CheckboxComponent,
    FormFieldComponent,
    InputFieldComponent,
    LoadingStateComponent,
    RadioComponent,
    SelectComponent,
    TextAreaComponent,
    DatePickerComponent,
  ],
  providers: [
    provideIcons({
      heroAcademicCap,
      heroArrowLeft,
      heroArrowRight,
      heroCamera,
      heroCheck,
      heroClipboardDocumentCheck,
      heroDocumentCheck,
      heroExclamationTriangle,
      heroHeart,
      heroIdentification,
      heroInformationCircle,
      heroLockClosed,
      heroPaperAirplane,
      heroPlusCircle,
      heroShieldCheck,
      heroUserGroup,
    }),
  ],
  templateUrl: './manager-registration-intake.component.html',
})
export class ManagerRegistrationIntakeComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  readonly steps: IntakeStep[] = [
    {
      key: 'child-basics',
      label: 'Child Details',
      shortLabel: 'Details',
      description: 'Basic information',
    },
    {
      key: 'medical-health',
      label: 'Medical & Health',
      shortLabel: 'Medical',
      description: 'Health and dietary',
    },
    {
      key: 'contacts-collection',
      label: 'Contacts & Security',
      shortLabel: 'Contacts',
      description: 'Carers and collection',
    },
    {
      key: 'consents-evidence',
      label: 'Permissions & Consents',
      shortLabel: 'Consents',
      description: 'Terms and decisions',
    },
    {
      key: 'review-complete',
      label: 'Review',
      shortLabel: 'Review',
      description: 'Complete registration',
    },
  ];

  readonly languageOptions = ['English', 'Polish', 'Punjabi', 'Arabic', 'Urdu', 'Spanish', 'Other'];
  readonly relationshipOptions = ['Mother', 'Father', 'Parent', 'Carer', 'Grandparent', 'Aunt', 'Uncle', 'Other'];
  readonly sexOptions: Option[] = [
    { value: 'male', label: 'Male' },
    { value: 'female', label: 'Female' },
    { value: 'other', label: 'Other' },
  ];
  readonly languageSelectOptions: Option[] = this.languageOptions.map((language) => ({
    value: language,
    label: language,
  }));
  readonly relationshipSelectOptions: Option[] = this.relationshipOptions.map((relationship) => ({
    value: relationship,
    label: relationship,
  }));
  readonly religionOptions: Option[] = [
    { value: 'christian', label: 'Christian' },
    { value: 'muslim', label: 'Muslim' },
    { value: 'hindu', label: 'Hindu' },
    { value: 'sikh', label: 'Sikh' },
    { value: 'jewish', label: 'Jewish' },
    { value: 'buddhist', label: 'Buddhist' },
    { value: 'none', label: 'No religion' },
    { value: 'other', label: 'Other' },
  ];
  readonly ethnicOriginOptions: Option[] = [
    { value: 'white_british', label: 'White British' },
    { value: 'white_irish', label: 'White Irish' },
    { value: 'white_other', label: 'White Other' },
    { value: 'mixed_white_caribbean', label: 'Mixed White & Caribbean' },
    { value: 'mixed_white_african', label: 'Mixed White & African' },
    { value: 'mixed_white_asian', label: 'Mixed White & Asian' },
    { value: 'mixed_other', label: 'Mixed Other' },
    { value: 'asian_indian', label: 'Asian Indian' },
    { value: 'asian_pakistani', label: 'Asian Pakistani' },
    { value: 'asian_bangladeshi', label: 'Asian Bangladeshi' },
    { value: 'asian_chinese', label: 'Asian Chinese' },
    { value: 'asian_other', label: 'Asian Other' },
    { value: 'black_caribbean', label: 'Black Caribbean' },
    { value: 'black_african', label: 'Black African' },
    { value: 'black_other', label: 'Black Other' },
    { value: 'other', label: 'Other Ethnic Group' },
  ];
  readonly yesNoOptions: Option[] = [
    { value: 'yes', label: 'Yes' },
    { value: 'no', label: 'No' },
  ];
  readonly evidenceStatusOptions: Option[] = [
    { value: 'unknown', label: 'Unknown' },
    { value: 'complete', label: 'Complete' },
    { value: 'missing', label: 'Still needed' },
    { value: 'not_applicable', label: 'Not applicable' },
  ];
  readonly todayIso = new Date().toISOString().slice(0, 10);
  readonly step1RequiredFields: Step1RequiredField[] = [
    'first_name',
    'surname',
    'date_of_birth',
    'start_date',
  ];
  readonly concernItems: { key: string; label: string }[] = [
    { key: 'concern_walking', label: 'Walking / Motor Skills' },
    { key: 'concern_speech_language', label: 'Speech and Language' },
    { key: 'concern_hearing', label: 'Hearing' },
    { key: 'concern_sight', label: 'Sight / Vision' },
    { key: 'concern_emotional_wellbeing', label: 'Emotional Wellbeing' },
    { key: 'concern_behaviour', label: 'Behaviour' },
  ];
  readonly immunisationOptions = [
    { value: 'up_to_date', label: 'Fully Up-to-Date' },
    { value: 'partial', label: 'Partially (Delayed)' },
    { value: 'refused', label: 'Not Immunised' },
  ];

  readonly professionalConsentItems: ConsentItem[] = [
    {
      key: 'area_senco_liaison',
      label: 'Area SENCO Liaison',
      detail: 'Staff may discuss this child with the Special Educational Needs Co-ordinator.',
    },
    {
      key: 'health_visitor_liaison',
      label: 'Health Visitor Liaison',
      detail: "Staff may discuss the child's development with their Health Visitor.",
    },
    {
      key: 'transition_documents',
      label: 'Transition Documents',
      detail: "The nursery may send transition documents to the child's next setting.",
    },
  ];

  readonly activityConsentItems: ConsentItem[] = [
    {
      key: 'local_outings',
      label: 'Regular outings in the local area for community learning',
      detail: '',
    },
    {
      key: 'face_painting',
      label: 'Face painting during themed days and celebrations',
      detail: '',
    },
    {
      key: 'parent_supplied_sun_cream',
      label: 'Permission for staff to apply parent-supplied sun cream',
      detail: '',
    },
    {
      key: 'parent_supplied_nappy_cream',
      label: 'Permission for staff to apply parent-supplied nappy cream',
      detail: '',
    },
  ];

  readonly photoConsentItems: ConsentItem[] = [
    { key: 'development_profile_photos', label: 'Development Records', detail: '' },
    { key: 'nursery_display_boards', label: 'Nursery Display Boards', detail: '' },
    { key: 'promotional_literature', label: 'Promotional Literature', detail: '' },
    { key: 'nursery_website', label: 'Nursery Website', detail: '' },
    { key: 'staff_student_coursework', label: 'Staff/Student Qualification Coursework', detail: '' },
  ];

  currentStep: StepperStep = 'child-basics';
  childId: string | null = null;
  child: ChildRecord | null = null;
  workflowStatus: RegistrationWorkflowStatus | null = null;
  isNewRegistration = true;

  isLoading = false;
  isSaving = false;
  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};
  successMessage: string | null = null;
  step1Submitted = false;
  step1Touched: Partial<Record<Step1Field, boolean>> = {};

  step1 = {
    first_name: '',
    surname: '',
    date_of_birth: '',
    start_date: '',
    sex: '',
    first_language: '',
    home_address: '',
    home_postcode: '',
    home_telephone: '',
    notes: '',
    religion: '',
    ethnic_origin: '',
    other_languages: '',
    disability_status: '',
    disability_notes: '',
    access_requirements: '',
  };

  step2 = {
    has_allergies: false,
    allergy_details: '',
    on_medication: false,
    medication_name: '',
    medication_dosage: '',
    medication_storage: '',
    immunisation_status: '',
    immunisation_country: '',
    illness_diagnosis_history: '',
    dietary_side_effects: '',
    doctor_practice: '',
    doctor_name: '',
    doctor_phone: '',
    health_visitor_name: '',
    health_visitor_clinic: '',
    health_visitor_phone: '',
    social_services_involvement: false,
    social_services_details: '',
    social_worker_contact: '',
    concern_walking: false,
    concern_speech_language: false,
    concern_hearing: false,
    concern_sight: false,
    concern_emotional_wellbeing: false,
    concern_behaviour: false,
    professional_referrals: '',
    routine_care_notes: '',
  };

  step3 = {
    collection_password: '',
    collection_password_hint: '',
    national_insurance_number: '',
    applying_for_funding: false,
    early_years_pupil_premium: false,
    working_tax_credit: false,
    college_uni_paid_to_parent: false,
    college_uni_paid_to_nursery: false,
    funding_3yo_term_time: false,
    funding_2yo_term_time: false,
    parent1_address: '',
    parent1_has_responsibility: true,
    show_second_parent: false,
    second_parent_name: '',
    second_parent_relationship: '',
    second_parent_telephone: '',
    second_parent_email: '',
    second_parent_address: '',
    second_parent_has_responsibility: true,
  };

  step4: ConsentWritePayload = {
    signer_name: '',
    signed_date: '',
    paper_form_on_file: true,
    urgent_medical_treatment: true,
    plasters: true,
    safeguarding_reporting_acknowledgement: true,
    area_senco_liaison: true,
    health_visitor_liaison: true,
    transition_documents: true,
    local_outings: true,
    face_painting: true,
    parent_supplied_sun_cream: true,
    parent_supplied_nappy_cream: true,
    development_profile_photos: true,
    nursery_display_boards: true,
    promotional_literature: true,
    nursery_website: true,
    staff_student_coursework: true,
    social_media: true,
    social_media_channel_notes: 'TikTok, Instagram, Facebook',
    urgent_medical_treatment_exceptions: null,
    notes_exceptions: null,
  };

  step4_gdpr = {
    gdpr_declared_by_name: '',
    gdpr_declaration_date: '',
  };

  officeEvidence: Partial<OfficeUseChecklist> = {
    applicationDateStatus: 'complete',
    applicationDate: '',
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

  parentCarersDraft: RegistrationContactEntry[] = [this.emptyContact('Mother')];
  emergencyContactsDraft: RegistrationContactEntry[] = [this.emptyContact('Grandparent'), this.emptyContact('Aunt')];
  emergencyAuthorisedFlags = [true, false];

  ngOnInit(): void {
    const childIdParam = this.route.snapshot.paramMap.get('childId');
    if (childIdParam) {
      this.isNewRegistration = false;
      this.childId = childIdParam;
      this.loadChildAndStatus();
    }
  }

  get stepIndex(): number {
    return this.steps.findIndex(s => s.key === this.currentStep);
  }

  get isFirstStep(): boolean {
    return this.currentStep === 'child-basics';
  }

  get isLastStep(): boolean {
    return this.currentStep === 'review-complete';
  }

  get canMarkComplete(): boolean {
    return this.workflowStatus?.can_mark_complete ?? false;
  }

  get isReviewedComplete(): boolean {
    return this.workflowStatus?.is_reviewed_complete ?? false;
  }

  get needsReview(): boolean {
    return this.workflowStatus?.needs_review ?? false;
  }

  get childFullNameDraft(): string {
    return [this.step1.first_name.trim(), this.step1.surname.trim()].filter(Boolean).join(' ');
  }

  get currentStepNumber(): number {
    return this.stepIndex + 1;
  }

  get step1MissingRequiredFields(): Step1RequiredField[] {
    return this.step1RequiredFields.filter(field => !!this.step1FieldError(field));
  }



  get allergyStatusLabel(): string {
    return this.step2.has_allergies ? 'Recorded' : 'No known allergies recorded';
  }

  nextStep(): void {
    const idx = this.stepIndex;
    if (idx < this.steps.length - 1) {
      this.currentStep = this.steps[idx + 1].key;
      this.errorMessage = null;
    }
  }

  prevStep(): void {
    const idx = this.stepIndex;
    if (idx > 0) {
      this.currentStep = this.steps[idx - 1].key;
      this.errorMessage = null;
    }
  }

  goToStep(step: StepperStep): void {
    if (!this.canOpenStep(step)) {
      this.errorMessage = 'Create the child record before continuing to later registration steps.';
      return;
    }
    this.currentStep = step;
    this.errorMessage = null;
  }

  stepIsActive(step: StepperStep): boolean {
    return step === this.currentStep;
  }

  stepIsComplete(step: StepperStep): boolean {
    return this.steps.findIndex(s => s.key === step) < this.stepIndex;
  }

  canOpenStep(step: StepperStep): boolean {
    return step === 'child-basics' || !!this.childId;
  }

  step1FieldError(field: Step1Field): string | null {
    if (field === 'first_name' && !this.step1.first_name.trim()) {
      return 'Enter the child\'s first name.';
    }
    if (field === 'surname' && !this.step1.surname.trim()) {
      return 'Enter the child\'s surname.';
    }
    if (field === 'date_of_birth') {
      if (!this.step1.date_of_birth) {
        return 'Enter the child\'s date of birth.';
      }
      if (this.step1.date_of_birth > this.todayIso) {
        return 'Date of birth cannot be in the future.';
      }
    }
    if (field === 'start_date' && !this.step1.start_date) {
      return 'Enter the proposed start date.';
    }
    return null;
  }

  shouldShowStep1Error(field: Step1Field): boolean {
    return (this.step1Submitted || !!this.step1Touched[field]) && !!this.step1FieldError(field);
  }

  step1VisibleError(field: Step1Field): string {
    return this.shouldShowStep1Error(field) ? this.step1FieldError(field) ?? '' : '';
  }

  markStep1Touched(field: Step1Field): void {
    this.step1Touched[field] = true;
  }

  saveChildBasics(advance = true): void {
    this.step1Submitted = true;

    if (this.step1MissingRequiredFields.length > 0) {
      this.errorMessage = advance
        ? 'Complete the required child details before continuing.'
        : 'Complete the required child details before saving a draft.';
      this.focusFirstStep1Error();
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;
    this.fieldErrors = {};
    this.successMessage = null;

    const payload: ChildWritePayload = {
      full_name: this.childFullNameDraft,
      date_of_birth: this.step1.date_of_birth,
      start_date: this.step1.start_date,
      notes: this.step1.notes.trim() || undefined,
    };

    if (this.childId) {
      this.saveStep1Profile(this.childId, advance);
      return;
    }

    this.staffApi.createChild(payload).subscribe({
      next: (child) => {
        this.child = child;
        this.childId = child.id;
        this.saveStep1Profile(child.id, advance);
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  saveMedicalHealth(): void {
    if (!this.childId) {
      this.errorMessage = 'Create the child record before saving medical information.';
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;

    const medicationNotes = [
      this.step2.medication_name && `Medication: ${this.step2.medication_name}`,
      this.step2.medication_dosage && `Dosage: ${this.step2.medication_dosage}`,
      this.step2.medication_storage && `Storage: ${this.step2.medication_storage}`,
    ].filter(Boolean).join('\n');

    this.staffApi.patchRegistrationProfile(this.childId, {
      medical_dietary: {
        medical_conditions_status: this.step2.has_allergies ? 'yes' : 'no',
        medical_conditions_notes: this.step2.allergy_details.trim() || null,
        prescribed_medication_status: this.step2.on_medication ? 'yes' : 'no',
        medication_notes: medicationNotes || null,
        dietary_requirements_status: this.step2.has_allergies ? 'yes' : 'no',
        dietary_requirements_notes: this.step2.allergy_details.trim() || null,
        dietary_side_effects: this.step2.dietary_side_effects.trim() || null,
        immunisation_status: this.step2.immunisation_status || null,
        immunisation_country: this.step2.immunisation_country.trim() || null,
        illness_diagnosis_history: this.step2.illness_diagnosis_history.trim() || null,
        medical_dietary_reviewed: true,
      },
      health_contacts: {
        doctor_name: this.step2.doctor_name.trim() || null,
        doctor_address: this.step2.doctor_practice.trim() || null,
        doctor_phone: this.step2.doctor_phone.trim() || null,
        health_visitor_name: this.step2.health_visitor_name.trim() || null,
        health_visitor_address: this.step2.health_visitor_clinic.trim() || null,
        health_visitor_phone: this.step2.health_visitor_phone.trim() || null,
        health_contacts_reviewed: true,
      },
      social_development: {
        social_services_status: this.step2.social_services_involvement ? 'yes' : 'no',
        social_services_notes: this.step2.social_services_details.trim() || null,
        social_worker_contact_details: this.step2.social_worker_contact.trim() || null,
        concern_walking: this.step2.concern_walking ? 'yes' : 'no',
        concern_speech_language: this.step2.concern_speech_language ? 'yes' : 'no',
        concern_hearing: this.step2.concern_hearing ? 'yes' : 'no',
        concern_sight: this.step2.concern_sight ? 'yes' : 'no',
        concern_emotional_wellbeing: this.step2.concern_emotional_wellbeing ? 'yes' : 'no',
        concern_behaviour: this.step2.concern_behaviour ? 'yes' : 'no',
        social_development_reviewed: true,
      },
      routine_care: {
        routine_care_notes: this.step2.routine_care_notes.trim() || null,
        routine_care_reviewed: true,
      },
    }).subscribe({
      next: () => {
        this.isSaving = false;
        this.nextStep();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  saveContactsCollection(): void {
    if (!this.childId) {
      this.errorMessage = 'Create the child record before saving contacts.';
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;

    const parent1 = this.parentCarersDraft[0]
      ? {
          ...this.parentCarersDraft[0],
          hasParentalResponsibility: this.step3.parent1_has_responsibility || null,
          address: this.step3.parent1_address
            ? { text: this.step3.parent1_address.trim() }
            : this.parentCarersDraft[0].address,
        }
      : null;

    const parentCarers: RegistrationContactEntry[] = [];
    if (parent1) parentCarers.push(parent1);

    if (this.step3.show_second_parent && this.step3.second_parent_name.trim()) {
      parentCarers.push({
        fullName: this.step3.second_parent_name.trim(),
        relationshipToChild: this.step3.second_parent_relationship.trim() || null,
        address: this.step3.second_parent_address.trim()
          ? { text: this.step3.second_parent_address.trim() } as unknown as Record<string, unknown>
          : null,
        telephone: this.step3.second_parent_telephone.trim() || null,
        email: this.step3.second_parent_email.trim() || null,
        workAddress: null,
        hasParentalResponsibility: this.step3.second_parent_has_responsibility || null,
      });
    }

    const emergencyContacts = this.filterContacts(this.emergencyContactsDraft);
    const authorisedCollectors = this.emergencyContactsDraft
      .filter((contact, index) => this.emergencyAuthorisedFlags[index] && this.contactHasValue(contact))
      .map((contact) => ({ ...contact, hasParentalResponsibility: null }));

    this.staffApi.patchRegistrationProfile(this.childId, {
      parent_carers: parentCarers,
      emergency_contacts: emergencyContacts,
      authorised_collectors: authorisedCollectors,
      funding_support: {
        benefits_contribute_to_fees: this.step3.applying_for_funding ? 'yes' : 'unknown',
        working_tax_credit: this.step3.working_tax_credit ? 'yes' : 'unknown',
        college_uni_paid_to_parent: this.step3.college_uni_paid_to_parent ? 'yes' : 'unknown',
        college_uni_paid_to_nursery: this.step3.college_uni_paid_to_nursery ? 'yes' : 'unknown',
        funding_3yo_term_time: this.step3.funding_3yo_term_time ? 'yes' : 'unknown',
        funding_2yo_term_time: this.step3.funding_2yo_term_time ? 'yes' : 'unknown',
        funding_support_notes: this.step3.national_insurance_number
          ? `National Insurance Number captured for funding verification: ${this.step3.national_insurance_number}`
          : null,
        funding_support_reviewed: true,
      },
      collection: {
        over18_collection_acknowledged: true,
        emergency_collection_reviewed: true,
      },
    }).subscribe({
      next: () => {
        if (this.step3.collection_password) {
          this.staffApi.setRegistrationCollectionPassword(this.childId!, this.step3.collection_password).subscribe({
            next: () => {
              this.step3.collection_password = '';
              this.isSaving = false;
              this.nextStep();
            },
            error: (err) => {
              this.isSaving = false;
              const mapped = this.errorMapper.mapAndHandle(err);
              this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
            },
          });
        } else {
          this.isSaving = false;
          this.nextStep();
        }
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  saveConsentsEvidence(): void {
    if (!this.childId) {
      this.errorMessage = 'Create the child record before saving consents.';
      return;
    }
    if (!this.step4.signer_name.trim() || !this.step4.signed_date) {
      this.errorMessage = 'Parent/guardian name and signed date are required.';
      return;
    }
    if (!this.step4.paper_form_on_file || !this.step4.safeguarding_reporting_acknowledgement) {
      this.errorMessage = 'Paper form evidence and safeguarding acknowledgement must be confirmed.';
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;

    const consentPayload: ConsentWritePayload = {
      ...this.step4,
      signer_name: this.step4.signer_name.trim(),
      social_media_channel_notes: this.step4.social_media_channel_notes?.trim() || null,
      urgent_medical_treatment_exceptions: this.step4.urgent_medical_treatment_exceptions?.trim() || null,
      notes_exceptions: this.step4.notes_exceptions?.trim() || null,
    };

    const gdprPayload = this.step4_gdpr.gdpr_declared_by_name.trim()
      ? {
          gdpr_declaration: {
            gdpr_declared_by_name: this.step4_gdpr.gdpr_declared_by_name.trim(),
            gdpr_declaration_date: this.step4_gdpr.gdpr_declaration_date || null,
          },
        }
      : {};

    this.staffApi.createRegistrationConsent(this.childId, consentPayload).subscribe({
      next: () => {
        const officePatch: Record<string, unknown> = {
          ...this.officeEvidence,
          applicationDate: this.officeEvidence.applicationDate || null,
        };
        const gdprObservable = this.step4_gdpr.gdpr_declared_by_name.trim()
          ? this.staffApi.patchRegistrationProfile(this.childId!, gdprPayload)
          : undefined;

        const requests: Observable<unknown>[] = [
          this.staffApi.patchRegistrationOfficeUseChecklist(this.childId!, officePatch),
        ];
        if (gdprObservable) requests.push(gdprObservable);

        (requests.length > 1
          ? combineLatest(requests)
          : requests[0]
        ).subscribe({
          next: () => {
            this.isSaving = false;
            this.loadStatus();
            this.nextStep();
          },
          error: () => {
            this.isSaving = false;
            this.loadStatus();
            this.nextStep();
          },
        });
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  markComplete(): void {
    if (!this.childId) return;
    this.isSaving = true;
    this.errorMessage = null;

    this.staffApi.createRegistrationCompletionAttestation(this.childId).subscribe({
      next: () => {
        this.isSaving = false;
        this.successMessage = 'Registration marked as reviewed and complete.';
        this.loadStatus();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  navigateToChildDetail(): void {
    if (this.childId) {
      this.router.navigate(['/staff/manager/children', this.childId]);
    }
  }

  addEmergencyContact(): void {
    this.emergencyContactsDraft.push(this.emptyContact(''));
    this.emergencyAuthorisedFlags.push(false);
  }

  removeEmergencyContact(index: number): void {
    this.emergencyContactsDraft.splice(index, 1);
    this.emergencyAuthorisedFlags.splice(index, 1);
    if (this.emergencyContactsDraft.length === 0) {
      this.addEmergencyContact();
    }
  }

  protected getConcernValue(key: string): boolean {
    return (this.step2 as Record<string, boolean | string>)[key] === true;
  }

  protected setConcernValue(key: string, value: boolean): void {
    (this.step2 as Record<string, boolean | string>)[key] = value;
  }

  protected trackByIndex(index: number): number {
    return index;
  }

  protected consentValue(key: keyof ConsentWritePayload): boolean {
    return this.step4[key] === true;
  }

  protected setConsentValue(key: keyof ConsentWritePayload, checked: boolean): void {
    (this.step4[key] as boolean) = checked;
  }

  protected profileCompleteLabel(): string {
    return this.workflowStatus?.profile_completeness?.is_complete ? 'Complete' : 'Incomplete';
  }

  protected consentCompleteLabel(): string {
    return this.workflowStatus?.consent_completeness?.is_complete ? 'Complete' : 'Incomplete';
  }

  protected officeCompleteLabel(): string {
    return this.workflowStatus?.office_completeness?.is_complete ? 'Complete' : 'Incomplete';
  }

  private saveStep1Profile(childId: string, advance: boolean): void {
    this.staffApi.patchRegistrationProfile(childId, {
      demographics_home: {
        sex: this.step1.sex || null,
        first_language: this.step1.first_language || null,
        home_address: this.stringToAddress(this.step1.home_address),
        home_postcode: this.step1.home_postcode.trim() || null,
        home_telephone: this.step1.home_telephone.trim() || null,
        religion: this.step1.religion.trim() || null,
        ethnic_origin: this.step1.ethnic_origin.trim() || null,
        other_languages: this.parseOtherLanguages(this.step1.other_languages),
        disability_status: this.parseYesNoUnknown(this.step1.disability_status),
        disability_notes: this.step1.disability_notes.trim() || null,
        access_requirements: this.step1.access_requirements.trim() || null,
        demographics_home_reviewed: true,
      },
    }).subscribe({
      next: () => {
        this.isSaving = false;
        if (advance) {
          this.nextStep();
        } else {
          this.successMessage = 'Child details saved.';
        }
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  private loadChildAndStatus(): void {
    if (!this.childId) return;
    this.isLoading = true;

    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.populateStep1FromChild(child);
        this.loadRegistrationProfile();
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  private loadRegistrationProfile(): void {
    if (!this.childId) return;

    this.staffApi.getRegistrationProfile(this.childId).subscribe({
      next: (profile) => {
        this.populateDraftsFromProfile(profile);
        this.loadStatus();
      },
      error: () => {
        this.loadStatus();
      },
    });
  }

  private loadStatus(): void {
    if (!this.childId) return;
    this.staffApi.getRegistrationWorkflowStatus(this.childId).subscribe({
      next: (status) => {
        this.workflowStatus = status;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
        this.workflowStatus = null;
      },
    });
  }

  private populateStep1FromChild(child: ChildRecord): void {
    const parts = child.fullName.trim().split(/\s+/);
    this.step1.first_name = parts.slice(0, -1).join(' ') || child.fullName;
    this.step1.surname = parts.length > 1 ? parts[parts.length - 1] : '';
    this.step1.date_of_birth = child.dateOfBirth;
    this.step1.start_date = child.startDate;
    this.step1.notes = child.notes ?? '';
  }

  private populateDraftsFromProfile(profile: RegistrationProfileResponse): void {
    if (profile.demographicsHome) {
      this.step1.sex = profile.demographicsHome.sex ?? '';
      this.step1.first_language = profile.demographicsHome.firstLanguage ?? '';
      this.step1.home_address = this.addressToString(profile.demographicsHome.homeAddress);
      this.step1.home_postcode = profile.demographicsHome.homePostcode ?? '';
      this.step1.home_telephone = profile.demographicsHome.homeTelephone ?? '';
      this.step1.religion = profile.demographicsHome.religion ?? '';
      this.step1.ethnic_origin = profile.demographicsHome.ethnicOrigin ?? '';
      this.step1.other_languages = (profile.demographicsHome.otherLanguages ?? []).join(', ');
      this.step1.disability_status = profile.demographicsHome.disabilityStatus ?? '';
      this.step1.disability_notes = profile.demographicsHome.disabilityNotes ?? '';
      this.step1.access_requirements = profile.demographicsHome.accessRequirements ?? '';
    }

    if (profile.medicalDietary) {
      this.step2.has_allergies =
        profile.medicalDietary.dietaryRequirementsStatus === 'yes'
        || profile.medicalDietary.medicalConditionsStatus === 'yes';
      this.step2.allergy_details =
        profile.medicalDietary.dietaryRequirementsNotes
        ?? profile.medicalDietary.medicalConditionsNotes
        ?? '';
      this.step2.on_medication = profile.medicalDietary.prescribedMedicationStatus === 'yes';
      this.step2.medication_name = profile.medicalDietary.medicationNotes ?? '';
      this.step2.immunisation_status = profile.medicalDietary.immunisationStatus ?? '';
      this.step2.immunisation_country = profile.medicalDietary.immunisationCountry ?? '';
      this.step2.illness_diagnosis_history = profile.medicalDietary.illnessDiagnosisHistory ?? '';
      this.step2.dietary_side_effects = profile.medicalDietary.dietarySideEffects ?? '';
    }

    if (profile.healthContacts) {
      this.step2.doctor_name = profile.healthContacts.doctorName ?? '';
      this.step2.doctor_practice = profile.healthContacts.doctorAddress ?? '';
      this.step2.doctor_phone = profile.healthContacts.doctorPhone ?? '';
      this.step2.health_visitor_name = profile.healthContacts.healthVisitorName ?? '';
      this.step2.health_visitor_clinic = profile.healthContacts.healthVisitorAddress ?? '';
      this.step2.health_visitor_phone = profile.healthContacts.healthVisitorPhone ?? '';
    }

    if (profile.socialDevelopment) {
      this.step2.social_services_involvement = profile.socialDevelopment.socialServicesStatus === 'yes';
      this.step2.social_services_details = profile.socialDevelopment.socialServicesNotes ?? '';
      this.step2.social_worker_contact = profile.socialDevelopment.socialWorkerContactDetails ?? '';
      this.step2.concern_walking = profile.socialDevelopment.concernWalking === 'yes';
      this.step2.concern_speech_language = profile.socialDevelopment.concernSpeechLanguage === 'yes';
      this.step2.concern_hearing = profile.socialDevelopment.concernHearing === 'yes';
      this.step2.concern_sight = profile.socialDevelopment.concernSight === 'yes';
      this.step2.concern_emotional_wellbeing = profile.socialDevelopment.concernEmotionalWellbeing === 'yes';
      this.step2.concern_behaviour = profile.socialDevelopment.concernBehaviour === 'yes';
      this.step2.professional_referrals = profile.socialDevelopment.professionalReferrals
        ? profile.socialDevelopment.professionalReferrals.map(r =>
            `${r.type}${r.referredDate ? ` (${r.referredDate})` : ''}${r.notes ? `: ${r.notes}` : ''}`
          ).join('; ')
        : '';
    }

    if (profile.routineCare) {
      this.step2.routine_care_notes = profile.routineCare.routineCareNotes ?? '';
    }

    this.parentCarersDraft = profile.parentCarers.length
      ? profile.parentCarers.map(contact => ({ ...contact }))
      : [this.emptyContact('Mother')];
    if (profile.parentCarers.length > 1) {
      this.step3.show_second_parent = true;
      this.step3.second_parent_name = profile.parentCarers[1].fullName ?? '';
      this.step3.second_parent_relationship = profile.parentCarers[1].relationshipToChild ?? '';
      this.step3.second_parent_telephone = profile.parentCarers[1].telephone ?? '';
      this.step3.second_parent_email = profile.parentCarers[1].email ?? '';
      this.step3.second_parent_address = this.addressToString(profile.parentCarers[1].address);
      this.step3.second_parent_has_responsibility = profile.parentCarers[1].hasParentalResponsibility ?? true;
    }
    this.step3.parent1_has_responsibility = profile.parentCarers[0]?.hasParentalResponsibility ?? true;

    this.emergencyContactsDraft = profile.emergencyContacts.length
      ? profile.emergencyContacts.map(contact => ({ ...contact }))
      : [this.emptyContact('Grandparent'), this.emptyContact('Aunt')];
    this.emergencyAuthorisedFlags = this.emergencyContactsDraft.map((contact) =>
      profile.authorisedCollectors.some((collector) => collector.fullName === contact.fullName && !!contact.fullName),
    );
    if (!this.emergencyAuthorisedFlags.some(Boolean) && this.emergencyAuthorisedFlags.length > 0) {
      this.emergencyAuthorisedFlags[0] = true;
    }

    if (profile.fundingSupport) {
      this.step3.applying_for_funding = profile.fundingSupport.benefitsContributeToFees === 'yes';
      this.step3.working_tax_credit = profile.fundingSupport.workingTaxCredit === 'yes';
      this.step3.college_uni_paid_to_parent = profile.fundingSupport.collegeUniPaidToParent === 'yes';
      this.step3.college_uni_paid_to_nursery = profile.fundingSupport.collegeUniPaidToNursery === 'yes';
      this.step3.funding_3yo_term_time = profile.fundingSupport.funding3yoTermTime === 'yes';
      this.step3.funding_2yo_term_time = profile.fundingSupport.funding2yoTermTime === 'yes';
    }
  }

  private emptyContact(relationshipToChild: string): RegistrationContactEntry {
    return {
      fullName: '',
      relationshipToChild,
      address: null,
      telephone: null,
      email: null,
      workAddress: null,
      hasParentalResponsibility: null,
    };
  }

  private filterContacts(entries: RegistrationContactEntry[]): RegistrationContactEntry[] {
    return entries
      .filter(contact => this.contactHasValue(contact))
      .map(contact => ({
        ...contact,
        fullName: contact.fullName.trim(),
        relationshipToChild: contact.relationshipToChild?.trim() || null,
        telephone: contact.telephone?.trim() || null,
        email: contact.email?.trim() || null,
      }));
  }

  private contactHasValue(contact: RegistrationContactEntry): boolean {
    return !!(
      contact.fullName.trim()
      || contact.relationshipToChild?.trim()
      || contact.telephone?.trim()
      || contact.email?.trim()
    );
  }

  private addressToString(addr: Record<string, unknown> | null | undefined): string {
    if (!addr) return '';
    if (typeof addr['text'] === 'string') return addr['text'];
    return Object.values(addr).filter((v): v is string => typeof v === 'string' && !!v).join(', ');
  }

  private stringToAddress(value: string): Record<string, unknown> | null {
    const trimmed = value.trim();
    return trimmed ? { text: trimmed } : null;
  }

  private parseOtherLanguages(value: string): string[] | null {
    const trimmed = value.trim();
    if (!trimmed) return null;
    return trimmed.split(',').map(s => s.trim()).filter(Boolean);
  }

  private parseYesNoUnknown(value: string): string | null {
    const trimmed = value.trim().toLowerCase();
    if (trimmed === 'yes') return 'yes';
    if (trimmed === 'no') return 'no';
    if (trimmed === 'unknown' || !trimmed) return null;
    return null;
  }

  private focusFirstStep1Error(): void {
    const fieldIds: Record<Step1RequiredField, string> = {
      first_name: 'child-first-name',
      surname: 'child-surname',
      date_of_birth: 'child-date-of-birth',
      start_date: 'child-start-date',
    };
    const firstInvalidField = this.step1MissingRequiredFields[0];
    if (!firstInvalidField) return;

    setTimeout(() => {
      const element = globalThis.document?.getElementById(fieldIds[firstInvalidField]);
      element?.focus();
    });
  }
}
