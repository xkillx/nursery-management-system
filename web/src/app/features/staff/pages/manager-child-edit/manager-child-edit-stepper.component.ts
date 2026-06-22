import { CommonModule } from '@angular/common';
import { Component, ElementRef, HostListener, inject, OnDestroy, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { Subject, debounceTime, takeUntil } from 'rxjs';
import {
  heroAcademicCap,
  heroArrowLeft,
  heroArrowRight,
  heroCalendarDays,
  heroCamera,
  heroChartBarSquare,
  heroCheck,
  heroClipboardDocument,
  heroClipboardDocumentCheck,
  heroClipboardDocumentList,
  heroCloudArrowUp,
  heroDocumentCheck,
  heroDocumentText,
  heroExclamationTriangle,
  heroEye,
  heroEyeSlash,
  heroHeart,
  heroHome,
  heroIdentification,
  heroInformationCircle,
  heroLanguage,
  heroLockClosed,
  heroPaperAirplane,
  heroPlusCircle,
  heroShieldCheck,
  heroUser,
  heroUserGroup,
  heroXMark,
  heroCalendar,
} from '@ng-icons/heroicons/outline';

import { environment } from '../../../../../environments/environment';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { AuthService } from '../../../../core/services/auth.service';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { CheckboxComponent } from '../../../../shared/components/form/input/checkbox.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { RadioComponent } from '../../../../shared/components/form/input/radio.component';
import { SelectComponent, type Option } from '../../../../shared/components/form/select/select.component';
import { TextAreaComponent } from '../../../../shared/components/form/input/text-area.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { StaffSessionType, StaffSessionTypesApiService } from '../../data/session-types-api.service';
import { RegistrationDraftStorage } from '../../data/registration-draft.storage';
import { ToastService } from '../../../../shared/services/toast.service';
import { ChildRecord, ChildWritePayload } from '../../models/children.models';
import {
  ConsentRecord,
  ConsentWritePayload,
  RegistrationContactEntry,
  StepperProfileView,
  CompleteRegistrationPayload,
  StepperCompletionStatus,
} from '../../models/child-legacy-compat.models';
import {
  ChildProfile,
  ChildHealthProfile,
  ChildSafeguardingProfile,
  ChildCollectionSettings,
  ChildFundingRecord,
  ChildConsentInput,
  ChildContact,
  ChildProfileInput,
  ChildHealthProfileInput,
  ChildSafeguardingProfileInput,
  ChildCollectionSettingsInput,
  CreateChildPayload,
} from '../../models/child-profile.models';

type StepperStep =
  | 'child-basics'
  | 'medical-health'
  | 'contacts-collection'
  | 'consents-evidence'
  | 'session-pattern';

type YesNoUnknownStatus = '' | 'yes' | 'no' | 'unknown';
type NoneDetailsUnknownStatus = '' | 'none' | 'details' | 'unknown';

type FinalCompletionIssue = {
  stepKey: StepperStep;
  field: string;
  message: string;
};

type ConsentAdvisory = {
  stepKey: 'consents-evidence';
  field: string;
  message: string;
};

type ConsentTier = 'required' | 'required_acknowledged' | 'optional';

const CONSENT_TIERS: Record<keyof ConsentWritePayload, ConsentTier> = {
  gdpr_data_processing_consent: 'required',
  information_truthfulness_declaration: 'required',
  safeguarding_reporting_acknowledgement: 'required_acknowledged',
  information_sharing_consent: 'required_acknowledged',
  urgent_medical_treatment: 'required_acknowledged',
  plasters: 'required_acknowledged',
  area_senco_liaison: 'optional',
  health_visitor_liaison: 'optional',
  transition_documents: 'optional',
  local_outings: 'optional',
  face_painting: 'optional',
  parent_supplied_sun_cream: 'optional',
  parent_supplied_nappy_cream: 'optional',
  development_profile_photos: 'optional',
  nursery_display_boards: 'optional',
  promotional_literature: 'optional',
  nursery_website: 'optional',
  staff_student_coursework: 'optional',
  social_media: 'optional',
  urgent_medical_treatment_exceptions: 'optional',
  notes_exceptions: 'optional',
  social_media_channel_notes: 'optional',
  signer_name: 'optional',
  signed_date: 'optional',
  paper_form_on_file: 'optional',
  consent_change_reason: 'optional',
};

const REQUIRED_ACKNOWLEDGED_KEYS = (Object.keys(CONSENT_TIERS) as (keyof ConsentWritePayload)[]).filter(
  (k) => CONSENT_TIERS[k] === 'required_acknowledged',
);

const REQUIRED_KEYS = (Object.keys(CONSENT_TIERS) as (keyof ConsentWritePayload)[]).filter(
  (k) => CONSENT_TIERS[k] === 'required',
);

type IntakeStep = {
  key: StepperStep;
  label: string;
  shortLabel: string;
  description: string;
};

type Step1Field =
  | 'first_name'
  | 'middle_name'
  | 'last_name'
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
  | 'access_requirements'
  | 'primary_room_id'
  | 'registration_date';

type Step1RequiredField = Extract<
  Step1Field,
  'first_name' | 'date_of_birth' | 'start_date' | 'home_address' | 'first_language' | 'primary_room_id' | 'registration_date'
>;

type ReferralEntry = {
  type: string;
  referredDate: string;
  referredBy: string;
  waitingListStatus: string;
  notes: string;
};

type ConsentItem = {
  key: keyof ConsentWritePayload;
  label: string;
  detail: string;
};

type RegistrationDraft = {
  currentStep: StepperStep;
  step1: {
    first_name: string;
    middle_name: string;
    last_name: string;
    date_of_birth: string;
    start_date: string;
    sex: string;
    first_language: string;
    home_address: string;
    home_postcode: string;
    home_telephone: string;
    notes: string;
    religion: string;
    ethnic_origin: string;
    other_languages: string;
    disability_status: string;
    disability_notes: string;
    access_requirements: string;
    primary_room_id: string;
    registration_date: string;
  };
  step2: {
    allergy_status: YesNoUnknownStatus;
    allergy_details: string;
    medication_status: YesNoUnknownStatus;
    medication_name: string;
    medication_dosage: string;
    medication_storage: string;
    medication_side_effects: string;
    immunisation_status: string;
    immunisation_country: string;
    medical_history_status: NoneDetailsUnknownStatus;
    illness_diagnosis_history: string;
    dietary_status: NoneDetailsUnknownStatus;
    special_dietary_requirements: string;
    dietary_side_effects: string;
    doctor_address: string;
    doctor_name: string;
    doctor_phone: string;
    health_visitor_name: string;
    health_visitor_clinic: string;
    health_visitor_phone: string;
    social_services_status: YesNoUnknownStatus;
    social_services_details: string;
    social_worker_name: string;
    social_worker_phone: string;
    social_worker_email: string;
    concern_walking: boolean;
    concern_speech_language: boolean;
    concern_hearing: boolean;
    concern_sight: boolean;
    concern_emotional_wellbeing: boolean;
    concern_behaviour: boolean;
    routine_care_notes: string;
  };
  step3: {
    collection_password: string;
    collection_password_hint: string;
    national_insurance_number: string;
    funding_support_answer: string;
    applying_for_funding: boolean;
    early_years_pupil_premium: boolean;
    working_tax_credit: boolean;
    college_uni_paid_to_parent: boolean;
    college_uni_paid_to_nursery: boolean;
    funding_3yo_term_time: boolean;
    funding_2yo_term_time: boolean;
    parent1_address: string;
    parent1_work_address: string;
    parent1_has_responsibility: boolean | null;
    show_second_parent: boolean;
    second_parent_name: string;
    second_parent_relationship: string;
    second_parent_telephone: string;
    second_parent_email: string;
    second_parent_address: string;
    second_parent_work_address: string;
    second_parent_has_responsibility: boolean | null;
    other_benefits: string;
    other_funding_selected: boolean;
    has_funding_support: boolean;
  };
  step4: ConsentWritePayload;
  step5?: {
    patternEffectiveFrom: string;
    patternEffectiveTo?: string;
    patternEntries: { dayOfWeek: number; sessionTypeId: string }[];
  };
  consentsReviewed: Partial<Record<keyof ConsentWritePayload, boolean>>;
  parentCarersDraft: RegistrationContactEntry[];
  emergencyContactsDraft: RegistrationContactEntry[];
  emergencyAuthorisedFlags: boolean[];
  emergencyContactAddresses: string[];
  referralsDraft: ReferralEntry[];
};

@Component({
  selector: 'app-manager-child-edit-stepper',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    AlertComponent,
    ButtonComponent,
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
      heroCalendarDays,
      heroCamera,
      heroChartBarSquare,
      heroCheck,
      heroClipboardDocument,
      heroClipboardDocumentCheck,
      heroClipboardDocumentList,
      heroCloudArrowUp,
      heroDocumentCheck,
      heroDocumentText,
      heroExclamationTriangle,
      heroEye,
      heroEyeSlash,
      heroHeart,
      heroHome,
      heroIdentification,
      heroInformationCircle,
      heroLanguage,
      heroLockClosed,
      heroPaperAirplane,
      heroPlusCircle,
      heroShieldCheck,
      heroUser,
      heroUserGroup,
      heroXMark,
      heroCalendar,
    }),
  ],
  templateUrl: './manager-child-edit-stepper.component.html',
})
export class ManagerChildEditStepperComponent implements OnInit, OnDestroy {
  private readonly staffApi = inject(StaffApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly draftStorage = inject(RegistrationDraftStorage);
  private readonly toast = inject(ToastService);
  private readonly host = inject(ElementRef<HTMLElement>);
  private readonly destroy$ = new Subject<void>();
  private readonly draftChanges$ = new Subject<void>();
  private dismissTimeout: ReturnType<typeof setTimeout> | null = null;
  private hasRestoredDraft = false;

  readonly steps: readonly IntakeStep[] = [
    {
      key: 'child-basics',
      label: 'Child Details',
      shortLabel: 'Child Info',
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
      key: 'session-pattern',
      label: 'Session Pattern',
      shortLabel: 'Pattern',
      description: 'Planned weekly attendance',
    },
  ];

  readonly languageOptions = [
    'English',
    'Mandarin Chinese',
    'Hindi',
    'Spanish',
    'French',
    'Modern Standard Arabic',
    'Bengali',
    'Portuguese',
    'Russian',
    'Urdu',
  ];
  readonly relationshipOptions = ['Mother', 'Father', 'Parent', 'Carer', 'Grandparent', 'Aunt', 'Uncle', 'Other'];
  showCollectionPassword = false;

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
  readonly yesNoUnknownOptions: Option[] = [
    { value: 'yes', label: 'Yes' },
    { value: 'no', label: 'No' },
  ];
  readonly disabilityStatusOptions: Option[] = [
    { value: 'yes', label: 'Yes' },
    { value: 'no', label: 'No' },
  ];
  readonly noneDetailsUnknownOptions: Option[] = [
    { value: 'none', label: 'None' },
    { value: 'details', label: 'Details recorded' },
  ];
  readonly parentalResponsibilityOptions: Option[] = [
    { value: 'yes', label: 'Yes' },
    { value: 'no', label: 'No' },
  ];
  readonly todayIso = new Date().toISOString().slice(0, 10);
  readonly step1RequiredFields: Step1RequiredField[] = [
    'first_name',
    'date_of_birth',
    'start_date',
    'home_address',
    'first_language',
    'primary_room_id',
    'registration_date',
  ];
  roomOptions: Option[] = [];
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
    { key: 'social_media', label: 'Nursery Social Media Accounts', detail: '' },
  ];

  private _currentStep: StepperStep = 'child-basics';
  get currentStep(): StepperStep { return this._currentStep; }
  set currentStep(step: StepperStep) {
    this._currentStep = step;
    if (step === 'consents-evidence' && this.parentCarersDraft[0]?.fullName && !this.step4.signer_name) {
      this.step4.signer_name = this.parentCarersDraft[0].fullName;
    }
    if (step === 'session-pattern') {
      this.loadSessionPatternSupportData();
    }
  }
  childId: string | null = null;
  child: ChildRecord | null = null;
  workflowStatus: StepperCompletionStatus | null = null;
  isNewRegistration = true;

  isLoading = false;
  isSaving = false;
  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};
  successMessage: string | null = null;
  step1Submitted = false;
  step1Touched: Partial<Record<Step1Field, boolean>> = {};
  step2Submitted = false;
  step2Touched: Record<string, boolean> = {};
  step3Submitted = false;
  step3Touched: Record<string, boolean> = {};
  fundingSubmitted = false;
  hasStoredDraft = false;
  draftRestoredAt: string | null = null;
  draftSavedAt: string | null = null;
  isDraftRestoredBannerVisible = false;

  protected showDebugPanel = false;
  protected copiedDebug = false;
  protected readonly isProduction = environment.production;

  step1 = {
    first_name: '',
    middle_name: '',
    last_name: '',
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
    primary_room_id: '',
    registration_date: this.todayIso,
  };

  step2 = {
    allergy_status: '' as YesNoUnknownStatus,
    allergy_details: '',
    medication_status: '' as YesNoUnknownStatus,
    medication_name: '',
    medication_dosage: '',
    medication_storage: '',
    medication_side_effects: '',
    immunisation_status: '',
    immunisation_country: '',
    medical_history_status: '' as NoneDetailsUnknownStatus,
    illness_diagnosis_history: '',
    dietary_status: '' as NoneDetailsUnknownStatus,
    special_dietary_requirements: '',
    dietary_side_effects: '',
    doctor_address: '',
    doctor_name: '',
    doctor_phone: '',
    health_visitor_name: '',
    health_visitor_clinic: '',
    health_visitor_phone: '',
    social_services_status: '' as YesNoUnknownStatus,
    social_services_details: '',
    social_worker_name: '',
    social_worker_phone: '',
    social_worker_email: '',
    concern_walking: false,
    concern_speech_language: false,
    concern_hearing: false,
    concern_sight: false,
    concern_emotional_wellbeing: false,
    concern_behaviour: false,
    routine_care_notes: '',
  };

  step3 = {
    collection_password: '',
    collection_password_hint: '',
    national_insurance_number: '',
    funding_support_answer: '',
    applying_for_funding: false,
    early_years_pupil_premium: false,
    working_tax_credit: false,
    college_uni_paid_to_parent: false,
    college_uni_paid_to_nursery: false,
    funding_3yo_term_time: false,
    funding_2yo_term_time: false,
    parent1_address: '',
    parent1_work_address: '',
    parent1_has_responsibility: null as boolean | null,
    show_second_parent: false,
    second_parent_name: '',
    second_parent_relationship: '',
    second_parent_telephone: '',
    second_parent_email: '',
    second_parent_address: '',
    second_parent_work_address: '',
    second_parent_has_responsibility: null as boolean | null,
    other_benefits: '',
    other_funding_selected: false,
    has_funding_support: false,
  };

  step4: ConsentWritePayload = {
    urgent_medical_treatment: false,
    plasters: false,
    safeguarding_reporting_acknowledgement: false,
    information_truthfulness_declaration: false,
    information_sharing_consent: false,
    gdpr_data_processing_consent: false,
    area_senco_liaison: false,
    health_visitor_liaison: false,
    transition_documents: false,
    local_outings: false,
    face_painting: false,
    parent_supplied_sun_cream: false,
    parent_supplied_nappy_cream: false,
    development_profile_photos: false,
    nursery_display_boards: false,
    promotional_literature: false,
    nursery_website: false,
    staff_student_coursework: false,
    social_media: false,
    urgent_medical_treatment_exceptions: null,
    notes_exceptions: null,
    signer_name: '',
    signed_date: '',
    consent_change_reason: null,
  };

  step4NoReasons: Partial<Record<keyof ConsentWritePayload, string>> = {};

  patternEffectiveFrom = '';
  patternEffectiveTo = '';
  patternEntries: { dayOfWeek: number; sessionTypeId: string }[] = [];
  patternError: string | null = null;
  availableSessionTypes: StaffSessionType[] = [];
  sessionPatternLoading = false;
  sessionPatternLoadError: string | null = null;

  consentsReviewed: Partial<Record<keyof ConsentWritePayload, boolean>> = {};
  consentAdvisories: ConsentAdvisory[] = [];
  finalCompletionIssues: FinalCompletionIssue[] = [];
  originalStep4Snapshot: ConsentWritePayload | null = null;

  parentCarersDraft: RegistrationContactEntry[] = [this.emptyContact('Mother')];
  emergencyContactsDraft: RegistrationContactEntry[] = [this.emptyContact('Grandparent')];
  emergencyAuthorisedFlags = [true];
  emergencyContactAddresses: string[] = [''];
  referralsDraft: ReferralEntry[] = [];

  readonly referralTypeOptions: Option[] = [
    { value: 'community_paediatrician', label: 'Community Paediatrician' },
    { value: 'speech_language_therapist', label: 'Speech and Language Therapist' },
    { value: 'eyis', label: 'EYIS — Early Support Service' },
    { value: 'other', label: 'Other professional' },
  ];
  readonly waitingListOptions: Option[] = [
    { value: 'on_waiting_list', label: 'On waiting list' },
    { value: 'seen_completed', label: 'Seen / Completed' },
    { value: 'not_applicable', label: 'Not applicable' },
  ];

  ngOnInit(): void {
    const childIdParam = this.route.snapshot.paramMap.get('childId');
    if (childIdParam) {
      this.isNewRegistration = false;
      this.childId = childIdParam;
      this.markAllConsentsReviewed();
      this.loadChildAndStatus();
      this.loadRoomOptions();
      return;
    }

    this.loadRoomOptions();
    this.restoreDraftIfPresent();
    this.subscribeToDraftAutoSave();
  }

  private loadRoomOptions(): void {
    const branchId = this.auth.activeMembership()?.branch_id;
    if (!branchId) {
      this.roomOptions = [];
      return;
    }
    this.roomsApi.listRooms(branchId, { includeArchived: false }).subscribe({
      next: (rooms) => {
        this.roomOptions = rooms
          .filter((room) => room.isActive)
          .map((room) => ({ value: room.id, label: room.name }));
      },
      error: () => {
        this.roomOptions = [];
      },
    });
  }

  ngOnDestroy(): void {
    if (this.dismissTimeout) {
      clearTimeout(this.dismissTimeout);
      this.dismissTimeout = null;
    }
    this.destroy$.next();
    this.destroy$.complete();
  }

  @HostListener('input')
  @HostListener('change')
  protected onFormInput(): void {
    this.notifyDraftChanged();
  }

  get stepIndex(): number {
    return this.steps.findIndex(s => s.key === this.currentStep);
  }

  get isFirstStep(): boolean {
    return this.currentStep === 'child-basics';
  }

  get isLastStep(): boolean {
    return this.stepIndex === this.steps.length - 1;
  }

  get canMarkComplete(): boolean {
    return this.workflowStatus?.canMarkComplete ?? false;
  }

  get isReviewedComplete(): boolean {
    return this.workflowStatus?.isReviewedComplete ?? false;
  }

  get needsReview(): boolean {
    return this.workflowStatus?.needsReview ?? false;
  }

  get childFullNameDraft(): string {
    return [this.step1.first_name.trim(), this.step1.middle_name.trim(), this.step1.last_name.trim()].filter(Boolean).join(' ');
  }

  get currentStepNumber(): number {
    return this.stepIndex + 1;
  }

  get step1MissingRequiredFields(): Step1RequiredField[] {
    return this.step1RequiredFields.filter(field => !!this.step1FieldError(field));
  }



  get allergyStatusLabel(): string {
    if (this.step2.allergy_status === 'yes') return 'Recorded';
    if (this.step2.allergy_status === 'no') return 'No known allergies';
    if (this.step2.allergy_status === 'unknown') return 'Unknown — follow-up required';
    return 'No known allergies recorded';
  }

  nextStep(): void {
    const idx = this.stepIndex;
    if (idx < this.steps.length - 1) {
      this.currentStep = this.steps[idx + 1].key;
      this.errorMessage = null;
      setTimeout(() => this.focusStepHeading(), 50);
    }
  }

  prevStep(): void {
    const idx = this.stepIndex;
    if (idx > 0) {
      this.currentStep = this.steps[idx - 1].key;
      this.errorMessage = null;
      setTimeout(() => this.focusStepHeading(), 50);
    }
  }

  goToStep(step: StepperStep): void {
    if (!this.canOpenStep(step)) {
      const requestedIdx = this.steps.findIndex(s => s.key === step);
      for (let i = 0; i < requestedIdx; i++) {
        const priorIssue = this.firstBlockingIssueForStep(this.steps[i].key);
        if (priorIssue) {
          this.handleValidationFailure(priorIssue);
          return;
        }
      }
      this.toast.error('Complete the current step before continuing.', { title: 'Check required details' });
      return;
    }
    this.currentStep = step;
    this.errorMessage = null;
  }

  stepIsActive(step: StepperStep): boolean {
    return step === this.currentStep;
  }

  stepIsComplete(step: StepperStep): boolean {
    const idx = this.steps.findIndex(s => s.key === step);
    if (idx >= this.stepIndex) return false;
    return this.issuesForStep(step).length === 0;
  }

  canOpenStep(step: StepperStep): boolean {
    const requestedIdx = this.steps.findIndex(s => s.key === step);
    if (requestedIdx <= this.stepIndex) return true;
    if (!this.isNewRegistration && step !== 'child-basics' && !this.childId) return false;
    for (let i = 0; i < requestedIdx; i++) {
      if (this.issuesForStep(this.steps[i].key).length > 0) {
        return false;
      }
    }
    return true;
  }

  isSessionPatternStep(): boolean {
    return this.currentStep === 'session-pattern';
  }

  step1FieldError(field: Step1Field): string | null {
    if (field === 'first_name' && !this.step1.first_name.trim()) {
      return 'Enter the child\'s first name.';
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
    if (field === 'home_address' && !this.step1.home_address.trim()) {
      return 'Enter the child\'s home address.';
    }
    if (field === 'first_language' && !this.step1.first_language.trim()) {
      return 'Select the primary language spoken at home.';
    }
    if (field === 'disability_status') {
      if (this.step1.disability_status !== 'yes' && this.step1.disability_status !== 'no') {
        return 'Confirm whether the child has a disability or SEND/access needs.';
      }
    }
    if (field === 'disability_notes') {
      if (this.step1.disability_status === 'yes' && !this.step1.disability_notes.trim() && !this.step1.access_requirements.trim()) {
        return 'Record disability or access details, or set disability to No.';
      }
    }
    if (field === 'primary_room_id' && !this.step1.primary_room_id) {
      return 'Pick a primary room.';
    }
    if (field === 'registration_date') {
      if (!this.step1.registration_date) {
        return 'Enter the registration date.';
      }
      if (this.step1.registration_date > this.todayIso) {
        return 'The registration date cannot be in the future.';
      }
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

  protected markStep2Touched(field: string): void {
    this.step2Touched[field] = true;
  }

  protected markStep3Touched(field: string): void {
    this.step3Touched[field] = true;
  }

  protected step3FieldError(field: string): string | null {
    const primary = this.parentCarersDraft[0];
    if (field === 'primary_full_name' && !primary?.fullName?.trim()) {
      return 'Enter the primary parent/carer full name.';
    }
    if (field === 'primary_relationship' && !primary?.relationshipToChild?.trim()) {
      return 'Select the primary parent/carer relationship to the child.';
    }
    if (field === 'primary_telephone' && !primary?.telephone?.trim()) {
      return 'Enter the primary parent/carer contact number.';
    }
    if (field === 'parent1_has_responsibility' && this.step3.parent1_has_responsibility === null) {
      return 'Confirm whether the primary parent/carer holds parental responsibility.';
    }
    return null;
  }

  protected shouldShowStep3Error(field: string): boolean {
    return (this.step3Submitted || !!this.step3Touched[field]) && !!this.step3FieldError(field);
  }

  protected step3VisibleError(field: string): string {
    return this.shouldShowStep3Error(field) ? this.step3FieldError(field) ?? '' : '';
  }

  protected step2FieldError(field: string): string | null {
    if (field === 'allergy_status' && !this.step2.allergy_status) {
      return 'Confirm whether the child has any known allergies.';
    }
    if (field === 'medication_status' && !this.step2.medication_status) {
      return 'Confirm whether the child takes regular medication.';
    }
    if (field === 'medical_history_status' && !this.step2.medical_history_status) {
      return 'Confirm medical history.';
    }
    if (field === 'dietary_status' && !this.step2.dietary_status) {
      return 'Confirm dietary requirements.';
    }
    if (field === 'social_services_status' && !this.step2.social_services_status) {
      return 'Confirm social services involvement.';
    }
    return null;
  }

  protected shouldShowStep2Error(field: string): boolean {
    return (this.step2Submitted || !!this.step2Touched[field]) && !!this.step2FieldError(field);
  }

  protected step2VisibleError(field: string): string {
    return this.shouldShowStep2Error(field) ? this.step2FieldError(field) ?? '' : '';
  }

  saveChildBasics(advance = true): void {
    this.step1Submitted = true;

    const firstIssue = this.firstBlockingIssueForStep('child-basics');
    if (firstIssue) {
      this.handleValidationFailure(firstIssue);
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;
    this.fieldErrors = {};
    this.successMessage = null;

    const payload: ChildWritePayload = {
      first_name: this.step1.first_name.trim(),
      middle_name: this.step1.middle_name.trim() || null,
      last_name: this.step1.last_name.trim() || null,
      date_of_birth: this.step1.date_of_birth,
      start_date: this.step1.start_date,
      notes: this.step1.notes.trim() || undefined,
    };

    if (this.childId) {
      this.saveStep1Profile(this.childId, advance);
      return;
    }

    if (this.isNewRegistration) {
      this.errorMessage = null;
      this.isSaving = false;
      if (advance) {
        this.nextStep();
      }
      return;
    }

    // In the new model the child is created atomically at submit time, not
    // mid-step. If we reach this branch the stepper is in an inconsistent
    // state (no childId, not a new registration). Just clear the saving
    // flag; the user will need to restart as a new registration.
    this.isSaving = false;
    this.errorMessage = 'Please complete the wizard as a new registration.';
  }

  saveMedicalHealth(): void {
    this.step2Submitted = true;
    if (!this.childId && !this.isNewRegistration) {
      this.errorMessage = 'Create the child record before saving medical information.';
      return;
    }
    const firstIssue = this.firstBlockingIssueForStep('medical-health');
    if (firstIssue) {
      this.handleValidationFailure(firstIssue);
      return;
    }
    if (this.isNewRegistration) {
      this.nextStep();
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;

    const medicationNotes = [
      this.step2.medication_name && `Medication: ${this.step2.medication_name}`,
      this.step2.medication_dosage && `Dosage: ${this.step2.medication_dosage}`,
      this.step2.medication_storage && `Storage: ${this.step2.medication_storage}`,
    ].filter(Boolean).join('\n');

    const dietaryNotes = [
      this.step2.allergy_details.trim(),
      this.step2.special_dietary_requirements.trim(),
    ].filter(Boolean).join('; ');

    const referrals = this.referralsDraft
      .filter(r => r.type && r.referredBy)
      .map(r => ({
        type: r.type,
        referred_date: r.referredDate || null,
        referred_by: r.referredBy.trim(),
        waiting_list_status: r.waitingListStatus || 'unknown',
        notes: r.notes.trim() || null,
      }));

    this.staffApi.patchChildHealth(this.childId!, {
      medical_conditions_status: (this.step2.allergy_status || 'unknown') as any,
      medical_conditions_notes: this.step2.allergy_details.trim() || null,
      prescribed_medication_status: (this.step2.medication_status || 'unknown') as any,
      medication_notes: medicationNotes || null,
      dietary_requirements_status: this.dietaryApiStatus() as any,
      dietary_requirements_notes: dietaryNotes || null,
      dietary_side_effects: this.step2.dietary_side_effects.trim() || null,
      immunisation_status: (this.step2.immunisation_status || 'unknown') as any,
      immunisation_country: this.step2.immunisation_country.trim() || null,
      illness_diagnosis_history: this.step2.illness_diagnosis_history.trim() || null,
      doctor_name: this.step2.doctor_name.trim() || null,
      doctor_address: this.step2.doctor_address.trim() || null,
      doctor_phone: this.step2.doctor_phone.trim() || null,
      health_visitor_name: this.step2.health_visitor_name.trim() || null,
      health_visitor_address: this.step2.health_visitor_clinic.trim() || null,
      health_visitor_phone: this.step2.health_visitor_phone.trim() || null,
    }).subscribe({
      next: () => this.saveStep2Safeguarding(this.childId!, referrals),
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  private saveStep2Safeguarding(childId: string, referrals: any[]): void {
    this.staffApi.patchChildSafeguarding(childId, {
      social_services_status: (this.step2.social_services_status || 'unknown') as any,
      social_services_notes: this.step2.social_services_details.trim() || null,
      social_worker_name: this.step2.social_worker_name.trim() || null,
      social_worker_phone: this.step2.social_worker_phone.trim() || null,
      social_worker_email: this.step2.social_worker_email.trim() || null,
      concern_walking: (this.step2.concern_walking ? 'yes' : 'no') as any,
      concern_speech_language: (this.step2.concern_speech_language ? 'yes' : 'no') as any,
      concern_hearing: (this.step2.concern_hearing ? 'yes' : 'no') as any,
      concern_sight: (this.step2.concern_sight ? 'yes' : 'no') as any,
      concern_emotional_wellbeing: (this.step2.concern_emotional_wellbeing ? 'yes' : 'no') as any,
      concern_behaviour: (this.step2.concern_behaviour ? 'yes' : 'no') as any,
      professional_referrals: referrals.length > 0 ? referrals : [],
    }).subscribe({
      next: () => this.saveStep2RoutineCare(this.childId!),
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  private saveStep2RoutineCare(childId: string): void {
    this.staffApi.patchChildProfile(childId, {
      routine_care_notes: this.step2.routine_care_notes.trim() || null,
      routine_care_reviewed: true,
      demographics_home_reviewed: true,
      medical_dietary_reviewed: true,
      health_contacts_reviewed: true,
      social_development_reviewed: true,
      parent_responsibility_reviewed: true,
      emergency_collection_reviewed: true,
    } as any).subscribe({
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
    if (!this.childId && !this.isNewRegistration) {
      this.errorMessage = 'Create the child record before saving contacts.';
      return;
    }
    this.step3Submitted = true;
    this.fundingSubmitted = true;
    const firstIssue = this.firstBlockingIssueForStep('contacts-collection');
    if (firstIssue) {
      this.handleValidationFailure(firstIssue);
      return;
    }
    if (this.isNewRegistration) {
      this.nextStep();
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
          workAddress: this.step3.parent1_work_address
            ? { text: this.step3.parent1_work_address.trim() }
            : this.parentCarersDraft[0].workAddress,
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
        workAddress: this.step3.second_parent_work_address.trim()
          ? { text: this.step3.second_parent_work_address.trim() } as unknown as Record<string, unknown>
          : null,
        hasParentalResponsibility: this.step3.second_parent_has_responsibility || null,
      });
    }

    const emergencyContacts = this.filterContacts(this.emergencyContactsDraft)
      .map((contact, index) => ({
        ...contact,
        address: this.emergencyContactAddresses[index]?.trim()
          ? { text: this.emergencyContactAddresses[index].trim() } as unknown as Record<string, unknown>
          : contact.address,
      }));
    const authorisedCollectors = this.emergencyContactsDraft
      .filter((contact, index) => this.emergencyAuthorisedFlags[index] && this.contactHasValue(contact))
      .map((contact, index) => ({
        ...contact,
        hasParentalResponsibility: null,
        address: this.emergencyContactAddresses[index]?.trim()
          ? { text: this.emergencyContactAddresses[index].trim() } as unknown as Record<string, unknown>
          : contact.address,
      }));

    this.staffApi.putChildContacts(this.childId!, {
      parentCarers: parentCarers as any,
      emergencyContacts: emergencyContacts as any,
      authorisedCollectors: authorisedCollectors as any,
    }).subscribe({
      next: () => this.saveStep3Funding(this.childId!, this.buildFundingSupportPayload()),
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  private saveStep3Funding(childId: string, funding: any): void {
    this.staffApi.patchChildFunding(childId, funding).subscribe({
      next: () => this.saveStep3Collection(childId),
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  private saveStep3Collection(childId: string): void {
    const password = this.step3.collection_password;
    this.staffApi.putChildCollectionSettings(childId, { password: password || '' } as any).subscribe({
      next: () => {
        this.step3.collection_password = '';
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

  saveConsentsEvidence(): void {
    if (!this.childId && !this.isNewRegistration) {
      this.errorMessage = 'Create the child record before saving consents.';
      return;
    }
    const firstIssue = this.firstBlockingIssueForStep('consents-evidence');
    if (firstIssue) {
      this.handleValidationFailure(firstIssue);
      return;
    }
    if (this.isNewRegistration) {
      return;
    }

    this.isSaving = true;
    this.errorMessage = null;

    const { social_media_channel_notes: _ignored, ...step4Base } = this.step4;
    const signedDate = this.step4.signed_date || this.todayIso;
    const valuesChanged = this.consentValuesChangedSince(this.originalStep4Snapshot);
    const consentChangeReason = valuesChanged
      ? this.step4.consent_change_reason?.trim() || null
      : null;

    const consentPayload: ConsentWritePayload = {
      ...step4Base,
      signer_name: this.step4.signer_name.trim(),
      signed_date: signedDate,
      urgent_medical_treatment_exceptions: this.step4.urgent_medical_treatment_exceptions?.trim() || null,
      notes_exceptions: this.step4.notes_exceptions?.trim() || null,
      consent_change_reason: consentChangeReason,
    };

    this.staffApi.updateChildConsent(this.childId!, consentPayload).subscribe({
      next: (saved) => {
        this.isSaving = false;
        this.successMessage = 'Consents & evidence saved.';
        this.originalStep4Snapshot = {
          ...this.step4,
          consent_change_reason: null,
        };
        this.step4.consent_change_reason = null;
        if (saved) {
          this.workflowStatus = {
            ...(this.workflowStatus ?? {
              isReviewedComplete: false,
              canMarkComplete: false,
              needsReview: true,
              missingGroups: [],
              currentConsent: null,
            }),
            isReviewedComplete: !!saved.safeguarding_reporting_acknowledgement,
            canMarkComplete: !!saved.safeguarding_reporting_acknowledgement,
            currentConsent: saved,
          };
        }
        this.loadStatus();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake'));
      },
    });
  }

  submitRegistration(): void {
    const issues = this.collectFinalCompletionIssues();
    this.finalCompletionIssues = issues;
    this.consentAdvisories = this.collectConsentAdvisories();
    if (issues.length > 0) {
      this.handleValidationFailure(issues[0]);
      return;
    }

    if (this.isNewRegistration) {
      this.continueFromConsents();
      return;
    }

    // Edit-mode: no global "Mark Reviewed/Complete" gesture. Edit mode shows
    // only the per-section Save Changes buttons. The server derives the
    // consent completion state from the saved record.
  }

  continueFromConsents(): void {
    this.step1Submitted = true;
    this.step2Submitted = true;
    this.step3Submitted = true;
    this.fundingSubmitted = true;

    const issues = this.collectFinalCompletionIssues();
    this.finalCompletionIssues = issues;
    this.consentAdvisories = this.collectConsentAdvisories();

    // Filter to only consents-evidence issues
    const consentIssues = issues.filter(i => i.stepKey === 'consents-evidence' || i.stepKey === 'contacts-collection' || i.stepKey === 'child-basics' || i.stepKey === 'medical-health');
    if (consentIssues.length > 0) {
      this.handleValidationFailure(consentIssues[0]);
      return;
    }

    if (!this.validateFundingSection()) {
      return;
    }

    this.nextStep();
    this.initialisePatternDefaultEffectiveDate();
  }

  createChildFromSessionPatternStep(): void {
    this.patternError = null;

    if (this.patternEntries.length === 0) {
      this.patternError = 'Add at least one booked session.';
      return;
    }

    if (this.patternEffectiveFrom && !/^\d{4}-\d{2}-\d{2}$/.test(this.patternEffectiveFrom)) {
      this.patternError = 'Effective date must be in YYYY-MM-DD format.';
      return;
    }

    if (this.patternEffectiveTo && !/^\d{4}-\d{2}-\d{2}$/.test(this.patternEffectiveTo)) {
      this.patternError = 'Effective to date must be in YYYY-MM-DD format.';
      return;
    }

    if (this.patternEffectiveFrom && this.patternEffectiveTo && this.patternEffectiveTo < this.patternEffectiveFrom) {
      this.patternError = 'Effective to date must be on or after the effective from date.';
      return;
    }

    const payload = this.buildCompleteRegistrationPayload();
    if (this.patternEntries.length > 0) {
      const bp: any = {
        effective_from: this.patternEffectiveFrom || this.step1.start_date,
        entries: this.patternEntries.map(e => ({
          day_of_week: e.dayOfWeek,
          session_type_id: e.sessionTypeId,
        })),
      };
      if (this.patternEffectiveTo) {
        bp.effective_to = this.patternEffectiveTo;
      }
      (payload as any).booking_pattern = bp;
    }

    this.isSaving = true;
    this.errorMessage = null;

    this.staffApi.createChildWithFullProfile(payload).subscribe({
      next: (result) => {
        this.childId = result.id;
        this.draftStorage.clear();
        this.hasStoredDraft = false;
        this.draftSavedAt = null;
        this.draftRestoredAt = null;
        this.isDraftRestoredBannerVisible = false;
        this.isSaving = false;
        this.toast.success('Child created.');
        this.router.navigate(['/staff/manager/children', result.id]);
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'registration.intake')) || 'Registration could not be completed.';
      },
    });
  }

  loadSessionPatternSupportData(): void {
    const branchId = this.auth.activeMembership()?.branch_id;
    if (!branchId) return;
    if (this.sessionPatternLoading) return;
    this.sessionPatternLoading = true;
    this.sessionPatternLoadError = null;

    this.sessionTypesApi.listSessionTypes(branchId, { includeArchived: false }).subscribe({
      next: (types) => {
        this.availableSessionTypes = types;
        this.sessionPatternLoading = false;
      },
      error: () => {
        this.sessionPatternLoadError = 'Failed to load session types.';
        this.sessionPatternLoading = false;
      },
    });
  }

  initialisePatternDefaultEffectiveDate(): void {
    if (!this.patternEffectiveFrom && this.step1.start_date) {
      this.patternEffectiveFrom = this.step1.start_date;
    }
  }

  goToSessionTypes(): void {
    this.router.navigate(['/staff/manager/session-types']);
  }

  togglePatternEntry(day: number, sessionTypeId: string): void {
    const idx = this.patternEntries.findIndex(
      (e) => e.dayOfWeek === day && e.sessionTypeId === sessionTypeId,
    );
    if (idx >= 0) {
      this.patternEntries.splice(idx, 1);
    } else {
      this.patternEntries.push({ dayOfWeek: day, sessionTypeId });
    }
  }

  isPatternEntrySelected(day: number, sessionTypeId: string): boolean {
    return this.patternEntries.some(
      (e) => e.dayOfWeek === day && e.sessionTypeId === sessionTypeId,
    );
  }

  patternEntriesForDay(day: number): { sessionTypeId: string; sessionType?: StaffSessionType }[] {
    return this.patternEntries
      .filter((e) => e.dayOfWeek === day)
      .map((e) => ({
        sessionTypeId: e.sessionTypeId,
        sessionType: this.availableSessionTypes.find((s) => s.id === e.sessionTypeId),
      }));
  }

  protected toggleDebugPanel(): void {
    this.showDebugPanel = !this.showDebugPanel;
  }

  protected copyDebugModels(): void {
    navigator.clipboard.writeText(this.debugModels);
    this.copiedDebug = true;
    setTimeout(() => (this.copiedDebug = false), 2000);
  }

  protected get debugModels(): string {
    return JSON.stringify({
      step1: this.step1,
      step2: this.step2,
      step3: this.step3,
      step4: this.step4,
      step5: {
        patternEffectiveFrom: this.patternEffectiveFrom,
        patternEntries: this.patternEntries,
      },
      consentsReviewed: this.consentsReviewed,
      parentCarersDraft: this.parentCarersDraft,
      emergencyContactsDraft: this.emergencyContactsDraft,
      referralsDraft: this.referralsDraft,
      child: this.child,
      workflowStatus: this.workflowStatus,
    }, null, 2);
  }

  protected issuesForStep(step: StepperStep): FinalCompletionIssue[] {
    return this.collectFinalCompletionIssues().filter(issue => issue.stepKey === step);
  }

  protected firstBlockingIssueForStep(step: StepperStep): FinalCompletionIssue | null {
    const issues = this.issuesForStep(step);
    return issues.length > 0 ? issues[0] : null;
  }

  protected hasFieldIssue(field: string): boolean {
    return this.finalCompletionIssues.some(issue => issue.field === field);
  }

  protected fieldIssueMessage(field: string): string {
    const issue = this.finalCompletionIssues.find(i => i.field === field);
    return issue ? issue.message : '';
  }

  private handleValidationFailure(issue: FinalCompletionIssue): void {
    this.finalCompletionIssues = this.collectFinalCompletionIssues();
    if (this.currentStep !== issue.stepKey) {
      this.currentStep = issue.stepKey;
    }
    this.toast.error(issue.message, { title: 'Check required details' });
    this.focusIssueField(issue.field);
  }

  private fieldFocusTarget(field: string): string {
    const map: Record<string, string> = {
      first_name: 'child-first-name',
      middle_name: 'child-middle-name',
      last_name: 'child-last-name',
      date_of_birth: 'child-date-of-birth',
      start_date: 'child-start-date',
      home_address: 'child-home-address',
      first_language: 'child-first-language',
      disability_status: 'child-disability-status-group',
      disability_notes: 'child-disability-notes',
      access_requirements: 'child-access-requirements',
      allergy_status: 'allergy-status-group',
      allergy_details: 'allergy-details',
      medication_status: 'medication-status-group',
      medication_name: 'medication-name',
      medication_dosage: 'medication-dosage',
      dietary_status: 'dietary-status-group',
      special_dietary_requirements: 'special-dietary-requirements',
      medical_history_status: 'medical-history-status-group',
      illness_diagnosis_history: 'illness-diagnosis-history',
      social_services_status: 'social-services-status-group',
      social_services_details: 'social-services-details',
      primary_full_name: 'primary-full-name',
      primary_relationship: 'primary-relationship',
      primary_telephone: 'primary-telephone',
      parent1_has_responsibility: 'parent1-responsibility-group',
      emergency_contacts: 'emergency-contacts-group',
      collection_password: 'collection-password',
      primary_room_id: 'child-primary-room',
      registration_date: 'child-registration-date',
      funding_support_answer: 'funding-support-yes',
      funding_options: 'funding-working-tax-credit',
      other_benefits: 'otherFunding',
      safeguarding_reporting_acknowledgement: 'safeguarding-reporting-consent',
      information_sharing_consent: 'information-sharing-consent',
      gdpr_data_processing_consent: 'gdpr-consent',
      information_truthfulness_declaration: 'truthfulness-declaration',
    };
    return map[field] ?? field;
  }

  private focusIssueField(field: string): void {
    const targetId = this.fieldFocusTarget(field);
    setTimeout(() => {
      const root = this.host.nativeElement as HTMLElement;
      let el: HTMLElement | null = root.querySelector<HTMLElement>(`#${CSS.escape(targetId)}`);
      if (!el) {
        const wrapper = root.querySelector<HTMLElement>(`[data-focus-target="${targetId}"]`);
        if (wrapper) {
          el = wrapper.querySelector<HTMLElement>('input, select, textarea, button, [tabindex]:not([tabindex="-1"])');
        }
      }
      if (el) {
        el.focus();
        el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    }, 60);
  }

  navigateToChildDetail(): void {
    if (this.childId) {
      this.router.navigate(['/staff/manager/children', this.childId]);
    }
  }

  addEmergencyContact(): void {
    this.emergencyContactsDraft.push(this.emptyContact(''));
    this.emergencyAuthorisedFlags.push(false);
    this.emergencyContactAddresses.push('');
    this.notifyDraftChanged();
  }

  removeEmergencyContact(index: number): void {
    this.emergencyContactsDraft.splice(index, 1);
    this.emergencyAuthorisedFlags.splice(index, 1);
    this.emergencyContactAddresses.splice(index, 1);
    if (this.emergencyContactsDraft.length === 0) {
      this.addEmergencyContact();
    }
    this.notifyDraftChanged();
  }

  addReferralEntry(): void {
    this.referralsDraft.push({
      type: 'other',
      referredDate: '',
      referredBy: '',
      waitingListStatus: 'not_applicable',
      notes: '',
    });
    this.notifyDraftChanged();
  }

  removeReferralEntry(index: number): void {
    this.referralsDraft.splice(index, 1);
    this.notifyDraftChanged();
  }

  protected getConcernValue(key: string): boolean {
    return (this.step2 as Record<string, boolean | string>)[key] === true;
  }

  protected setConcernValue(key: string, value: boolean): void {
    (this.step2 as Record<string, boolean | string>)[key] = value;
    this.notifyDraftChanged();
  }

  protected trackByIndex(index: number): number {
    return index;
  }

  protected consentValue(key: keyof ConsentWritePayload): boolean {
    return this.step4[key] === true;
  }

  protected consentReviewed(key: keyof ConsentWritePayload): boolean {
    return this.consentsReviewed[key] === true;
  }

  setConsentValue(key: keyof ConsentWritePayload, checked: boolean): void {
    (this.step4[key] as boolean) = checked;
    this.consentsReviewed[key] = true;
    this.consentAdvisories = this.collectConsentAdvisories();
    this.notifyDraftChanged();
  }

  toggleConsent(key: keyof ConsentWritePayload): void {
    const next = !this.consentValue(key);
    (this.step4[key] as boolean) = next;
    this.consentsReviewed[key] = true;
    this.consentAdvisories = this.collectConsentAdvisories();
    this.notifyDraftChanged();
  }

  consentTier(key: keyof ConsentWritePayload): ConsentTier {
    return CONSENT_TIERS[key] ?? 'optional';
  }

  setNoReason(key: keyof ConsentWritePayload, value: string): void {
    this.step4NoReasons = { ...this.step4NoReasons, [key]: value };
  }

  advisoryMessage(field: string): string {
    const labels: Record<string, string> = {
      safeguarding_reporting_acknowledgement: 'Safeguarding reporting acknowledgement',
      information_sharing_consent: 'Information sharing consent',
      urgent_medical_treatment: 'Urgent medical treatment',
      plasters: 'First aid / plasters',
    };
    return labels[field] ?? field;
  }

  private markAllConsentsReviewed(): void {
    const keys = Object.keys(this.step4) as (keyof ConsentWritePayload)[];
    for (const key of keys) {
      if (typeof this.step4[key] === 'boolean') {
        this.consentsReviewed[key] = true;
      }
    }
  }

  protected setFundingSupportAnswer(answer: 'yes' | 'no'): void {
    this.step3.funding_support_answer = answer;
    this.step3.has_funding_support = answer === 'yes';
    this.fundingSubmitted = false;

    if (answer === 'no') {
      this.clearFundingOptions();
    }

    this.notifyDraftChanged();
  }

  protected setFundingOption(
    key: 'working_tax_credit' | 'college_uni_paid_to_parent' | 'funding_3yo_term_time' | 'funding_2yo_term_time',
    checked: boolean,
  ): void {
    this.step3[key] = checked;
    this.notifyDraftChanged();
  }

  protected setOtherFundingSelected(checked: boolean): void {
    this.step3.other_funding_selected = checked;
    if (!checked) {
      this.step3.other_benefits = '';
    }
    this.notifyDraftChanged();
  }

  protected get fundingAnswerError(): string {
    if (!this.fundingSubmitted || this.step3.funding_support_answer) {
      return '';
    }
    return 'Select Yes or No to continue.';
  }

  protected get fundingOptionsError(): string {
    if (!this.fundingSubmitted || this.step3.funding_support_answer !== 'yes' || this.hasSelectedFundingOption()) {
      return '';
    }
    return 'Select at least one funding or benefit option.';
  }

  protected get otherFundingError(): string {
    if (!this.fundingSubmitted || !this.step3.other_funding_selected || this.step3.other_benefits.trim()) {
      return '';
    }
    return 'Enter the funding or benefit details.';
  }

  protected profileCompleteLabel(): string {
    return this.workflowStatus?.isReviewedComplete ? 'Complete' : 'Incomplete';
  }

  protected consentCompleteLabel(): string {
    return this.workflowStatus?.currentConsent ? 'Complete' : 'Incomplete';
  }

  private validateFundingSection(): boolean {
    this.fundingSubmitted = true;

    if (this.fundingAnswerError) {
      this.errorMessage = this.fundingAnswerError;
      this.focusFundingControl('funding-support-yes');
      return false;
    }

    if (this.fundingOptionsError) {
      this.errorMessage = this.fundingOptionsError;
      this.focusFundingControl('funding-working-tax-credit');
      return false;
    }

    if (this.otherFundingError) {
      this.errorMessage = this.otherFundingError;
      this.focusFundingControl('otherFunding');
      return false;
    }

    this.errorMessage = null;
    return true;
  }

  private focusFundingControl(id: string): void {
    setTimeout(() => {
      (this.host.nativeElement as HTMLElement).querySelector<HTMLElement>(`#${id}`)?.focus();
    }, 0);
  }

  private hasSelectedFundingOption(): boolean {
    return this.step3.working_tax_credit
      || this.step3.college_uni_paid_to_parent
      || this.step3.funding_3yo_term_time
      || this.step3.funding_2yo_term_time
      || this.step3.other_funding_selected;
  }

  private clearFundingOptions(): void {
    this.step3.applying_for_funding = false;
    this.step3.early_years_pupil_premium = false;
    this.step3.working_tax_credit = false;
    this.step3.college_uni_paid_to_parent = false;
    this.step3.college_uni_paid_to_nursery = false;
    this.step3.funding_3yo_term_time = false;
    this.step3.funding_2yo_term_time = false;
    this.step3.other_funding_selected = false;
    this.step3.other_benefits = '';
    this.step3.national_insurance_number = '';
  }

  private buildFundingSupportPayload(): Record<string, unknown> {
    const answer = this.step3.funding_support_answer;
    const statusFor = (selected: boolean): string => {
      if (answer === 'yes') return selected ? 'yes' : 'no';
      if (answer === 'no') return 'no';
      return 'unknown';
    };

    return {
      benefits_contribute_to_fees: answer || 'unknown',
      working_tax_credit: statusFor(this.step3.working_tax_credit),
      college_uni_paid_to_parent: statusFor(this.step3.college_uni_paid_to_parent),
      college_uni_paid_to_nursery: statusFor(false),
      funding_3yo_term_time: statusFor(this.step3.funding_3yo_term_time),
      funding_2yo_term_time: statusFor(this.step3.funding_2yo_term_time),
      funding_support_notes: this.step3.other_funding_selected
        ? this.step3.other_benefits.trim() || null
        : null,
      funding_support_reviewed: true,
    };
  }

  private buildCompleteRegistrationPayload(): CompleteRegistrationPayload {
    const medicationNotes = [
      this.step2.medication_name && `Medication: ${this.step2.medication_name}`,
      this.step2.medication_dosage && `Dosage: ${this.step2.medication_dosage}`,
      this.step2.medication_storage && `Storage: ${this.step2.medication_storage}`,
    ].filter(Boolean).join('\n');

    const parentCarers: ChildContact[] = [];
    if (this.parentCarersDraft[0] && this.parentCarersDraft[0].fullName.trim()) {
      parentCarers.push({
        fullName: this.parentCarersDraft[0].fullName.trim(),
        relationshipToChild: this.parentCarersDraft[0].relationshipToChild || null,
        address: this.step3.parent1_address.trim() ? { text: this.step3.parent1_address.trim() } : null,
        telephone: this.parentCarersDraft[0].telephone || null,
        email: this.parentCarersDraft[0].email || null,
        hasParentalResponsibility: this.step3.parent1_has_responsibility || null,
      });
    }
    if (this.step3.show_second_parent && this.step3.second_parent_name.trim()) {
      parentCarers.push({
        fullName: this.step3.second_parent_name.trim(),
        relationshipToChild: this.step3.second_parent_relationship.trim() || null,
        address: this.step3.second_parent_address.trim() ? { text: this.step3.second_parent_address.trim() } : null,
        telephone: this.step3.second_parent_telephone.trim() || null,
        email: this.step3.second_parent_email.trim() || null,
        hasParentalResponsibility: this.step3.second_parent_has_responsibility || null,
      });
    }

    const emergencyContacts: ChildContact[] = this.emergencyContactsDraft
      .filter(c => c.fullName.trim())
      .map(c => ({
        fullName: c.fullName.trim(),
        relationshipToChild: c.relationshipToChild?.trim() || null,
        telephone: c.telephone?.trim() || null,
        email: c.email?.trim() || null,
      }));

    const authorisedCollectors: ChildContact[] = this.emergencyContactsDraft
      .filter((c, i) => this.emergencyAuthorisedFlags[i] && c.fullName.trim())
      .map(c => ({
        fullName: c.fullName.trim(),
        relationshipToChild: c.relationshipToChild?.trim() || null,
        telephone: c.telephone?.trim() || null,
        email: c.email?.trim() || null,
        hasParentalResponsibility: null,
      }));

    const referrals = this.referralsDraft
      .filter(r => r.type && r.referredBy)
      .map(r => ({
        type: r.type,
        referred_date: r.referredDate || null,
        referred_by: r.referredBy.trim(),
        waiting_list_status: r.waitingListStatus || 'unknown',
        notes: r.notes.trim() || null,
      }));

    const child: ChildWritePayload = {
      first_name: this.step1.first_name.trim(),
      middle_name: this.step1.middle_name.trim() || null,
      last_name: this.step1.last_name.trim() || null,
      date_of_birth: this.step1.date_of_birth,
      start_date: this.step1.start_date,
      notes: this.step1.notes.trim() || undefined,
    };

    const profile: ChildProfileInput = {
      sex: this.step1.sex || null,
      religion: this.step1.religion.trim() || null,
      ethnic_origin: this.step1.ethnic_origin.trim() || null,
      first_language: this.step1.first_language.trim() || null,
      other_languages: this.step1.other_languages || null,
      home_postcode: this.step1.home_postcode.trim() || null,
      home_telephone: this.step1.home_telephone.trim() || null,
      disability_status: this.parseYesNoUnknownFromStr(this.step1.disability_status) ?? 'unknown',
      disability_notes: this.step1.disability_notes.trim() || null,
      access_requirements: this.step1.access_requirements.trim() || null,
      routine_care_notes: this.step2.routine_care_notes.trim() || null,
      registration_date: this.step1.registration_date || null,
      demographics_home_reviewed: true,
      medical_dietary_reviewed: true,
      health_contacts_reviewed: true,
      social_development_reviewed: true,
      parent_responsibility_reviewed: true,
      emergency_collection_reviewed: true,
      routine_care_reviewed: true,
    };
    if (this.step1.home_address.trim()) {
      profile.home_address = { text: this.step1.home_address.trim() };
    }

    const health: ChildHealthProfileInput = {
      medical_conditions_status: this.step2.allergy_status || 'unknown',
      medical_conditions_notes: this.step2.allergy_details.trim() || null,
      prescribed_medication_status: this.step2.medication_status || 'unknown',
      medication_notes: medicationNotes || null,
      dietary_requirements_status: this.dietaryApiStatus(),
      dietary_requirements_notes: [this.step2.allergy_details.trim(), this.step2.special_dietary_requirements.trim()].filter(Boolean).join('; ') || null,
      dietary_side_effects: this.step2.dietary_side_effects.trim() || null,
      immunisation_status: this.step2.immunisation_status || 'unknown',
      immunisation_country: this.step2.immunisation_country.trim() || null,
      illness_diagnosis_history: this.step2.illness_diagnosis_history.trim() || null,
      doctor_name: this.step2.doctor_name.trim() || null,
      doctor_address: this.step2.doctor_address.trim() || null,
      doctor_phone: this.step2.doctor_phone.trim() || null,
      health_visitor_name: this.step2.health_visitor_name.trim() || null,
      health_visitor_address: this.step2.health_visitor_clinic.trim() || null,
      health_visitor_phone: this.step2.health_visitor_phone.trim() || null,
    };

    const safeguarding: ChildSafeguardingProfileInput = {
      social_services_status: this.step2.social_services_status || 'unknown',
      social_services_notes: this.step2.social_services_details.trim() || null,
      social_worker_name: this.step2.social_worker_name.trim() || null,
      social_worker_phone: this.step2.social_worker_phone.trim() || null,
      social_worker_email: this.step2.social_worker_email.trim() || null,
      concern_walking: this.step2.concern_walking ? 'yes' : 'no',
      concern_speech_language: this.step2.concern_speech_language ? 'yes' : 'no',
      concern_hearing: this.step2.concern_hearing ? 'yes' : 'no',
      concern_sight: this.step2.concern_sight ? 'yes' : 'no',
      concern_emotional_wellbeing: this.step2.concern_emotional_wellbeing ? 'yes' : 'no',
      concern_behaviour: this.step2.concern_behaviour ? 'yes' : 'no',
      professional_referrals: referrals,
    };

    const funding = this.buildFundingSupportPayload() as {
      benefits_contribute_to_fees: string;
      working_tax_credit: string;
      college_uni_paid_to_parent: string;
      college_uni_paid_to_nursery: string;
      funding_3yo_term_time: string;
      funding_2yo_term_time: string;
      funding_support_notes?: string | null;
      funding_support_reviewed: boolean;
    };

    const consent: ChildConsentInput = {
      urgent_medical_treatment: this.step4.urgent_medical_treatment,
      urgent_medical_treatment_exceptions: this.step4.urgent_medical_treatment_exceptions?.trim() || null,
      plasters: this.step4.plasters,
      safeguarding_reporting_acknowledgement: this.step4.safeguarding_reporting_acknowledgement,
      information_sharing_consent: this.step4.information_sharing_consent,
      information_truthfulness_declaration: this.step4.information_truthfulness_declaration,
      gdpr_data_processing_consent: this.step4.gdpr_data_processing_consent,
      area_senco_liaison: this.step4.area_senco_liaison,
      health_visitor_liaison: this.step4.health_visitor_liaison,
      transition_documents: this.step4.transition_documents,
      local_outings: this.step4.local_outings,
      face_painting: this.step4.face_painting,
      parent_supplied_sun_cream: this.step4.parent_supplied_sun_cream,
      parent_supplied_nappy_cream: this.step4.parent_supplied_nappy_cream,
      development_profile_photos: this.step4.development_profile_photos,
      nursery_display_boards: this.step4.nursery_display_boards,
      promotional_literature: this.step4.promotional_literature,
      nursery_website: this.step4.nursery_website,
      staff_student_coursework: this.step4.staff_student_coursework,
      social_media: this.step4.social_media,
      notes_exceptions: this.step4.notes_exceptions?.trim() || null,
      signer_name: this.step4.signer_name.trim(),
      signed_date: this.step4.signed_date || this.step1.registration_date || this.todayIso,
    };

    const collectionSettings: ChildCollectionSettingsInput = {
      over_18_collection_acknowledged: true,
    };
    if (this.step3.collection_password) {
      collectionSettings.password = this.step3.collection_password;
    }

    const payload: CreateChildPayload = {
      child,
      profile,
      health,
      safeguarding,
      contacts: {
        parent_carers: parentCarers.map(c => this.toContactWire(c)),
        emergency_contacts: emergencyContacts.map(c => this.toContactWire(c)),
        authorised_collectors: authorisedCollectors.map(c => this.toContactWire(c)),
      },
      consent,
      funding,
      collection_settings: collectionSettings,
      room: {
        room_id: this.step1.primary_room_id,
        start_date: this.step1.start_date,
      },
    };

    return payload;
  }

  private toContactWire(c: ChildContact): Record<string, unknown> {
    return {
      full_name: c.fullName.trim(),
      relationship_to_child: c.relationshipToChild?.trim() || null,
      address: c.address ?? null,
      telephone: c.telephone?.trim() || null,
      email: c.email?.trim() || null,
      has_parental_responsibility: c.hasParentalResponsibility ?? null,
    };
  }

  private deriveSignerName(): string {
    const parent = this.parentCarersDraft[0];
    if (parent && parent.fullName.trim()) return parent.fullName.trim();
    const fullName = [this.step1.first_name, this.step1.last_name]
      .map(s => s.trim())
      .filter(Boolean)
      .join(' ');
    return fullName || 'On file';
  }

  canSubmitLocally(): boolean {
    const issues = this.collectFinalCompletionIssues();
    this.consentAdvisories = this.collectConsentAdvisories();
    this.finalCompletionIssues = issues;
    return issues.length === 0;
  }

  collectFinalCompletionIssues(): FinalCompletionIssue[] {
    const issues: FinalCompletionIssue[] = [];

    if (!this.step1.first_name.trim()) {
      issues.push({ stepKey: 'child-basics', field: 'first_name', message: 'Enter the child\'s first name.' });
    }
    if (!this.step1.date_of_birth) {
      issues.push({ stepKey: 'child-basics', field: 'date_of_birth', message: 'Enter the child\'s date of birth.' });
    }
    if (!this.step1.start_date) {
      issues.push({ stepKey: 'child-basics', field: 'start_date', message: 'Enter the proposed start date.' });
    }
    if (!this.step1.home_address.trim()) {
      issues.push({ stepKey: 'child-basics', field: 'home_address', message: 'Enter the child\'s home address.' });
    }
    if (!this.step1.first_language.trim()) {
      issues.push({ stepKey: 'child-basics', field: 'first_language', message: 'Select the primary language spoken at home.' });
    }
    if (!this.step1.primary_room_id) {
      issues.push({ stepKey: 'child-basics', field: 'primary_room_id', message: 'Pick a primary room.' });
    }
    if (!this.step1.registration_date) {
      issues.push({ stepKey: 'child-basics', field: 'registration_date', message: 'Enter the registration date.' });
    } else if (this.step1.registration_date > this.todayIso) {
      issues.push({ stepKey: 'child-basics', field: 'registration_date', message: 'The registration date cannot be in the future.' });
    }

    if (this.step1.disability_status !== 'yes' && this.step1.disability_status !== 'no') {
      issues.push({ stepKey: 'child-basics', field: 'disability_status', message: 'Confirm whether the child has a disability or SEND/access needs.' });
    } else if (this.step1.disability_status === 'yes' && !this.step1.disability_notes.trim() && !this.step1.access_requirements.trim()) {
      issues.push({ stepKey: 'child-basics', field: 'disability_notes', message: 'Record disability or access details, or set disability to No.' });
    }

    this.collectMedicalSafetyIssues(issues);
    this.collectContactsIssues(issues);
    this.collectFundingIssues(issues);
    this.collectConsentsIssues(issues);
    this.collectSessionPatternIssues(issues);

    return issues;
  }

  private collectMedicalSafetyIssues(issues: FinalCompletionIssue[]): void {
    const pushSafetyIssue = (
      status: YesNoUnknownStatus,
      field: string,
      label: string,
      detailsValue: string,
      detailsMissingMsg: string,
    ): void => {
      if (status === '') {
        issues.push({ stepKey: 'medical-health', field, message: `Confirm ${label}.` });
      } else if (status === 'unknown') {
        issues.push({ stepKey: 'medical-health', field, message: `${label[0].toUpperCase()}${label.slice(1)} cannot be marked Unknown for final completion — follow up before completing.` });
      } else if (status === 'yes' && !detailsValue && detailsMissingMsg) {
        issues.push({ stepKey: 'medical-health', field, message: detailsMissingMsg });
      }
    };

    pushSafetyIssue(
      this.step2.allergy_status,
      'allergy_status',
      'whether the child has any known allergies',
      this.step2.allergy_details.trim(),
      'Record allergy details, or set allergies to No known allergies.',
    );

    pushSafetyIssue(
      this.step2.medication_status,
      'medication_status',
      'whether the child takes regular medication',
      '',
      '',
    );
    if (this.step2.medication_status === 'yes') {
      if (!this.step2.medication_name.trim()) {
        issues.push({ stepKey: 'medical-health', field: 'medication_name', message: 'Record medication name, or set medication to No regular medication.' });
      }
      if (!this.step2.medication_dosage.trim()) {
        issues.push({ stepKey: 'medical-health', field: 'medication_dosage', message: 'Record medication dosage and frequency, or set medication to No regular medication.' });
      }
    }

    this.collectNoneDetailsIssue(
      issues,
      this.step2.dietary_status,
      this.step2.special_dietary_requirements.trim(),
      'dietary_status',
      'dietary requirements',
      'special_dietary_requirements',
      'Record dietary details, or set dietary requirements to None.',
    );
    this.collectNoneDetailsIssue(
      issues,
      this.step2.medical_history_status,
      this.step2.illness_diagnosis_history.trim(),
      'medical_history_status',
      'medical history',
      'illness_diagnosis_history',
      'Record medical history details, or set medical history to None.',
    );

    pushSafetyIssue(
      this.step2.social_services_status,
      'social_services_status',
      'social services involvement',
      this.step2.social_services_details.trim(),
      'Record social services details, or set involvement to No.',
    );
  }

  private collectNoneDetailsIssue(
    issues: FinalCompletionIssue[],
    status: NoneDetailsUnknownStatus,
    detailsValue: string,
    statusField: string,
    label: string,
    detailsField: string,
    detailsMissingMsg: string,
  ): void {
    if (status === '') {
      issues.push({ stepKey: 'medical-health', field: statusField, message: `Confirm ${label}.` });
    } else if (status === 'unknown') {
      issues.push({ stepKey: 'medical-health', field: statusField, message: `${label[0].toUpperCase()}${label.slice(1)} cannot be marked Unknown for final completion — follow up before completing.` });
    } else if (status === 'details' && !detailsValue) {
      issues.push({ stepKey: 'medical-health', field: detailsField, message: detailsMissingMsg });
    }
  }

  private collectContactsIssues(issues: FinalCompletionIssue[]): void {
    const primary = this.parentCarersDraft[0];
    if (!primary?.fullName?.trim()) {
      issues.push({ stepKey: 'contacts-collection', field: 'primary_full_name', message: 'Record the primary parent/carer full name.' });
    }
    if (!primary?.relationshipToChild?.trim()) {
      issues.push({ stepKey: 'contacts-collection', field: 'primary_relationship', message: 'Record the primary parent/carer relationship to the child.' });
    }
    if (!primary?.telephone?.trim()) {
      issues.push({ stepKey: 'contacts-collection', field: 'primary_telephone', message: 'Record the primary parent/carer phone number.' });
    }
    if (this.step3.parent1_has_responsibility === null) {
      issues.push({ stepKey: 'contacts-collection', field: 'parent1_has_responsibility', message: 'Confirm whether the primary parent/carer holds parental responsibility.' });
    }

    const validEmergency = this.emergencyContactsDraft.filter(contact =>
      contact.fullName.trim() && contact.relationshipToChild?.trim() && contact.telephone?.trim(),
    );
    if (validEmergency.length === 0) {
      issues.push({ stepKey: 'contacts-collection', field: 'emergency_contacts', message: 'Add at least one emergency contact with name, relationship, and phone number.' });
    }

    const hasAuthorisedNonParent = this.emergencyContactsDraft.some((contact, index) =>
      this.emergencyAuthorisedFlags[index] && contact.fullName.trim(),
    );
    if (hasAuthorisedNonParent && !this.step3.collection_password.trim()) {
      issues.push({ stepKey: 'contacts-collection', field: 'collection_password', message: 'Set an authorised collection password before completing.' });
    }
  }

  private collectFundingIssues(issues: FinalCompletionIssue[]): void {
    if (!this.step3.funding_support_answer) {
      issues.push({ stepKey: 'contacts-collection', field: 'funding_support_answer', message: 'Select Yes or No for funding/benefits support.' });
      return;
    }
    if (this.step3.funding_support_answer === 'yes') {
      if (!this.hasSelectedFundingOption()) {
        issues.push({ stepKey: 'contacts-collection', field: 'funding_options', message: 'Select at least one funding or benefit option.' });
      }
      if (this.step3.other_funding_selected && !this.step3.other_benefits.trim()) {
        issues.push({ stepKey: 'contacts-collection', field: 'other_benefits', message: 'Enter the funding or benefit details.' });
      }
    }
  }

  private collectConsentsIssues(issues: FinalCompletionIssue[]): void {
    for (const key of REQUIRED_KEYS) {
      if (this.step4[key] !== true) {
        issues.push({
          stepKey: 'consents-evidence',
          field: key as string,
          message: this.requiredConsentMessage(key),
        });
      }
    }

    for (const key of REQUIRED_ACKNOWLEDGED_KEYS) {
      if (!this.consentsReviewed[key]) {
        issues.push({
          stepKey: 'consents-evidence',
          field: key as string,
          message: this.requiredAcknowledgedMessage(key),
        });
      }
    }

    if (!this.step4.signer_name.trim()) {
      issues.push({
        stepKey: 'consents-evidence',
        field: 'signer_name',
        message: 'Record the parent or carer full name who signed the consent.',
      });
    }
    if (!this.step4.signed_date) {
      issues.push({
        stepKey: 'consents-evidence',
        field: 'signed_date',
        message: 'Record the date the consent was signed.',
      });
    }
  }

  private collectSessionPatternIssues(issues: FinalCompletionIssue[]): void {
    if (this.isNewRegistration && this.currentStep === 'session-pattern' && this.patternEntries.length === 0) {
      issues.push({
        stepKey: 'session-pattern',
        field: 'pattern_entries',
        message: 'Add at least one booked session.',
      });
    }
  }

  private requiredConsentMessage(key: keyof ConsentWritePayload): string {
    if (key === 'gdpr_data_processing_consent') {
      return 'Confirm GDPR data processing consent.';
    }
    if (key === 'information_truthfulness_declaration') {
      return 'Confirm the truthfulness declaration.';
    }
    return 'This required consent must be granted.';
  }

  private requiredAcknowledgedMessage(key: keyof ConsentWritePayload): string {
    const labels: Record<string, string> = {
      safeguarding_reporting_acknowledgement: 'safeguarding reporting acknowledgement',
      information_sharing_consent: 'information sharing consent',
      urgent_medical_treatment: 'urgent medical treatment',
      plasters: 'first aid/plasters',
    };
    return `Record an explicit Yes or No decision for ${labels[key as string] ?? key}.`;
  }

  collectConsentAdvisories(): ConsentAdvisory[] {
    const advisories: ConsentAdvisory[] = [];
    for (const key of REQUIRED_ACKNOWLEDGED_KEYS) {
      if (this.step4[key] === false) {
        advisories.push({
          stepKey: 'consents-evidence',
          field: key as string,
          message: this.requiredAcknowledgedMessage(key),
        });
      }
    }
    return advisories;
  }

  consentValuesChangedSince(snapshot: ConsentWritePayload | null): boolean {
    if (!snapshot) return false;
    const booleanKeys = (Object.keys(this.step4) as (keyof ConsentWritePayload)[]).filter(
      (k) => typeof this.step4[k] === 'boolean',
    );
    for (const key of booleanKeys) {
      if (this.step4[key] !== snapshot[key]) {
        return true;
      }
    }
    return false;
  }

  private parseYesNoUnknownFromStr(value: string): string | null {
    const trimmed = value.trim().toLowerCase();
    if (trimmed === 'yes') return 'yes';
    if (trimmed === 'no') return 'no';
    return null;
  }

  private dietaryApiStatus(): string {
    if (this.step2.dietary_status === 'details') return 'yes';
    if (this.step2.dietary_status === 'none') return 'no';
    return 'unknown';
  }

  private coerceYesNoUnknown(value: string | null | undefined): YesNoUnknownStatus {
    if (value === 'yes' || value === 'no' || value === 'unknown') return value;
    return '';
  }

  private legacyBooleanToYesNoUnknown(value: boolean | undefined): YesNoUnknownStatus {
    return value === true ? 'yes' : '';
  }

  private deriveDietaryStatusFromProfile(value: string | null | undefined): NoneDetailsUnknownStatus {
    if (value === 'yes') return 'details';
    if (value === 'no') return 'none';
    if (value === 'unknown') return 'unknown';
    return '';
  }

  private deriveMedicalHistoryStatusFromProfile(notes: string, reviewed: boolean): NoneDetailsUnknownStatus {
    if (notes.trim()) return 'details';
    if (reviewed) return 'none';
    return '';
  }

  private saveStep1Profile(childId: string, advance: boolean): void {
    this.staffApi.patchChildProfile(childId, {
      sex: this.step1.sex || null,
      first_language: this.step1.first_language || null,
      home_address: this.stringToAddress(this.step1.home_address) as any,
      home_postcode: this.step1.home_postcode.trim() || null,
      home_telephone: this.step1.home_telephone.trim() || null,
      religion: this.step1.religion.trim() || null,
      ethnic_origin: this.step1.ethnic_origin.trim() || null,
      other_languages: this.step1.other_languages || null,
      disability_status: this.parseYesNoUnknown(this.step1.disability_status) as any,
      disability_notes: this.step1.disability_notes.trim() || null,
      access_requirements: this.step1.access_requirements.trim() || null,
      demographics_home_reviewed: true,
      registration_date: this.step1.registration_date || null,
      medical_dietary_reviewed: true,
      health_contacts_reviewed: true,
      social_development_reviewed: true,
      parent_responsibility_reviewed: true,
      emergency_collection_reviewed: true,
      routine_care_reviewed: true,
    } as any).subscribe({
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
        this.loadChildView();
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  private loadChildView(): void {
    if (!this.childId) return;

    this.staffApi.getStepperView(this.childId).subscribe({
      next: (view) => {
        this.populateDraftsFromView(view);
        this.loadRegistrationConsents();
        this.loadStatus();
      },
      error: () => {
        this.loadStatus();
      },
    });
  }

  private loadRegistrationConsents(): void {
    if (!this.childId) return;
    this.staffApi.getChildConsent(this.childId).subscribe({
      next: (consent) => {
        if (consent) {
          this.populateStep4FromConsent(consent);
        }
      },
      error: () => {
        // Consents not yet recorded for this child; defaults remain.
      },
    });
  }

  private populateStep4FromConsent(record: ConsentRecord): void {
    this.step4 = {
      ...this.step4,
      urgent_medical_treatment: record.urgent_medical_treatment,
      urgent_medical_treatment_exceptions: record.urgent_medical_treatment_exceptions,
      plasters: record.plasters,
      safeguarding_reporting_acknowledgement: record.safeguarding_reporting_acknowledgement,
      information_truthfulness_declaration: record.information_truthfulness_declaration,
      area_senco_liaison: record.area_senco_liaison,
      health_visitor_liaison: record.health_visitor_liaison,
      transition_documents: record.transition_documents,
      local_outings: record.local_outings,
      face_painting: record.face_painting,
      parent_supplied_sun_cream: record.parent_supplied_sun_cream,
      parent_supplied_nappy_cream: record.parent_supplied_nappy_cream,
      development_profile_photos: record.development_profile_photos,
      nursery_display_boards: record.nursery_display_boards,
      promotional_literature: record.promotional_literature,
      nursery_website: record.nursery_website,
      staff_student_coursework: record.staff_student_coursework,
      social_media: record.social_media,
      notes_exceptions: record.notes_exceptions,
      signer_name: record.signer_name ?? '',
      signed_date: record.signed_date ?? '',
      consent_change_reason: null,
    };
    this.originalStep4Snapshot = { ...this.step4 };
    this.consentAdvisories = this.collectConsentAdvisories();
    this.markAllConsentsReviewed();
  }

  private loadStatus(): void {
    if (!this.childId) {
      this.isLoading = false;
      return;
    }
    this.staffApi.getChildConsent(this.childId).subscribe({
      next: (consent) => {
        this.workflowStatus = {
          isReviewedComplete: !!consent?.safeguarding_reporting_acknowledgement,
          canMarkComplete: !!consent?.safeguarding_reporting_acknowledgement,
          needsReview: !consent,
          missingGroups: consent ? [] : ['safeguarding_reporting_acknowledgement'],
          currentConsent: consent,
        };
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
        this.workflowStatus = null;
      },
    });
  }

  private populateStep1FromChild(child: ChildRecord): void {
    this.step1.first_name = child.firstName ?? '';
    this.step1.middle_name = child.middleName ?? '';
    this.step1.last_name = child.lastName ?? '';
    this.step1.date_of_birth = child.dateOfBirth;
    this.step1.start_date = child.startDate;
    this.step1.notes = child.notes ?? '';
    this.step1.primary_room_id = child.primaryRoomId ?? '';
  }

  private populateDraftsFromView(view: {
    profile: ChildProfile | null;
    health: ChildHealthProfile | null;
    safeguarding: ChildSafeguardingProfile | null;
    contacts: {
      parentCarers: RegistrationContactEntry[];
      emergencyContacts: RegistrationContactEntry[];
      authorisedCollectors: RegistrationContactEntry[];
    };
    collection: ChildCollectionSettings | null;
    funding: ChildFundingRecord | null;
    consent: ConsentRecord | null;
  }): void {
    if (view.profile) {
      const p = view.profile;
      if (p.registration_date) {
        this.step1.registration_date = p.registration_date;
      }
      this.step1.sex = p.sex ?? '';
      this.step1.first_language = p.first_language ?? '';
      this.step1.home_address = this.addressToString(p.home_address as any);
      this.step1.home_postcode = p.home_postcode ?? '';
      this.step1.home_telephone = p.home_telephone ?? '';
      this.step1.religion = p.religion ?? '';
      this.step1.ethnic_origin = p.ethnic_origin ?? '';
      this.step1.other_languages = p.other_languages ?? '';
      this.step1.disability_status = p.disability_status ?? '';
      this.step1.disability_notes = p.disability_notes ?? '';
      this.step1.access_requirements = p.access_requirements ?? '';
      this.step2.routine_care_notes = p.routine_care_notes ?? '';
    }

    if (view.health) {
      const h = view.health;
      this.step2.allergy_status = this.coerceYesNoUnknown(h.medical_conditions_status);
      const notes = h.dietary_requirements_notes ?? h.medical_conditions_notes ?? '';
      const parts = notes.split('; ').filter(Boolean);
      this.step2.allergy_details = parts[0] ?? '';
      this.step2.special_dietary_requirements = parts[1] ?? '';
      this.step2.medication_status = this.coerceYesNoUnknown(h.prescribed_medication_status);
      this.step2.medication_name = h.medication_notes ?? '';
      this.step2.immunisation_status = h.immunisation_status ?? '';
      this.step2.immunisation_country = h.immunisation_country ?? '';
      this.step2.illness_diagnosis_history = h.illness_diagnosis_history ?? '';
      this.step2.dietary_side_effects = h.dietary_side_effects ?? '';
      this.step2.dietary_status = this.deriveDietaryStatusFromProfile(h.dietary_requirements_status);
      this.step2.medical_history_status = this.deriveMedicalHistoryStatusFromProfile(
        h.illness_diagnosis_history ?? '',
        !!h.illness_diagnosis_history,
      );

      this.step2.doctor_name = h.doctor_name ?? '';
      this.step2.doctor_address = h.doctor_address ?? '';
      this.step2.doctor_phone = h.doctor_phone ?? '';
      this.step2.health_visitor_name = h.health_visitor_name ?? '';
      this.step2.health_visitor_clinic = h.health_visitor_address ?? '';
      this.step2.health_visitor_phone = h.health_visitor_phone ?? '';
    }

    if (view.safeguarding) {
      const s = view.safeguarding;
      this.step2.social_services_status = this.coerceYesNoUnknown(s.social_services_status);
      this.step2.social_services_details = s.social_services_notes ?? '';
      this.step2.social_worker_name = s.social_worker_name ?? '';
      this.step2.social_worker_phone = s.social_worker_phone ?? '';
      this.step2.social_worker_email = s.social_worker_email ?? '';
      this.step2.concern_walking = s.concern_walking === 'yes';
      this.step2.concern_speech_language = s.concern_speech_language === 'yes';
      this.step2.concern_hearing = s.concern_hearing === 'yes';
      this.step2.concern_sight = s.concern_sight === 'yes';
      this.step2.concern_emotional_wellbeing = s.concern_emotional_wellbeing === 'yes';
      this.step2.concern_behaviour = s.concern_behaviour === 'yes';
      this.referralsDraft = s.professional_referrals?.length
        ? s.professional_referrals.map((r: any) => ({
            type: r.type,
            referredDate: r.referred_date ?? '',
            referredBy: r.referred_by ?? '',
            waitingListStatus: r.waiting_list_status ?? 'unknown',
            notes: r.notes ?? '',
          }))
        : [];
    }

    this.parentCarersDraft = view.contacts.parentCarers.length
      ? view.contacts.parentCarers.map(contact => ({ ...contact }))
      : [this.emptyContact('Mother')];
    if (view.contacts.parentCarers.length > 1) {
      this.step3.show_second_parent = true;
      this.step3.second_parent_name = view.contacts.parentCarers[1].fullName ?? '';
      this.step3.second_parent_relationship = view.contacts.parentCarers[1].relationship_to_child ?? '';
      this.step3.second_parent_telephone = view.contacts.parentCarers[1].telephone ?? '';
      this.step3.second_parent_email = view.contacts.parentCarers[1].email ?? '';
      this.step3.second_parent_address = this.addressToString(view.contacts.parentCarers[1].address as any);
      this.step3.second_parent_has_responsibility = view.contacts.parentCarers[1].has_parental_responsibility ?? null;
    }
    this.step3.parent1_has_responsibility = view.contacts.parentCarers[0]?.has_parental_responsibility ?? null;

    this.emergencyContactsDraft = view.contacts.emergencyContacts.length
      ? view.contacts.emergencyContacts.map(contact => ({ ...contact }))
      : [this.emptyContact('Grandparent')];
    this.emergencyAuthorisedFlags = this.emergencyContactsDraft.map((contact) =>
      view.contacts.authorisedCollectors.some(
        (collector) => collector.fullName === contact.fullName && !!contact.fullName,
      ),
    );
    if (!this.emergencyAuthorisedFlags.some(Boolean) && this.emergencyAuthorisedFlags.length > 0) {
      this.emergencyAuthorisedFlags[0] = true;
    }

    if (view.funding) {
      const f = view.funding;
      this.step3.funding_support_answer =
        f.benefits_contribute_to_fees === 'yes'
          ? 'yes'
          : f.benefits_contribute_to_fees === 'no'
            ? 'no'
            : '';
      this.step3.has_funding_support = this.step3.funding_support_answer === 'yes';
      this.step3.applying_for_funding = this.step3.has_funding_support;
      this.step3.working_tax_credit = f.working_tax_credit === 'yes';
      this.step3.college_uni_paid_to_parent = f.college_uni_paid_to_parent === 'yes';
      this.step3.college_uni_paid_to_nursery = f.college_uni_paid_to_nursery === 'yes';
      this.step3.funding_3yo_term_time = f.funding_3yo_term_time === 'yes';
      this.step3.funding_2yo_term_time = f.funding_2yo_term_time === 'yes';
      this.step3.other_benefits = f.funding_support_notes ?? '';
      this.step3.other_funding_selected = !!this.step3.other_benefits.trim();
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

  protected roomLabelFor(roomId: string | null | undefined): string {
    if (!roomId) return '';
    return this.roomOptions.find((r) => r.value === roomId)?.label ?? '';
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

  private parseYesNoUnknown(value: string): string | null {
    const trimmed = value.trim().toLowerCase();
    if (trimmed === 'yes') return 'yes';
    if (trimmed === 'no') return 'no';
    if (trimmed === 'unknown' || !trimmed) return null;
    return null;
  }

  private focusStepHeading(): void {
    const activeStep = this.host.nativeElement.querySelector('[aria-current="step"]') as HTMLElement | null;
    activeStep?.focus();
  }

  private subscribeToDraftAutoSave(): void {
    this.draftChanges$
      .pipe(debounceTime(500), takeUntil(this.destroy$))
      .subscribe(() => this.persistDraft());
  }

  protected notifyDraftChanged(): void {
    if (!this.isNewRegistration) {
      return;
    }
    this.draftChanges$.next();
  }

  protected discardDraft(): void {
    if (this.dismissTimeout) {
      clearTimeout(this.dismissTimeout);
      this.dismissTimeout = null;
    }
    this.draftStorage.clear();
    this.hasStoredDraft = false;
    this.draftRestoredAt = null;
    this.draftSavedAt = null;
    this.isDraftRestoredBannerVisible = false;
    this.resetDrafts();
    this.currentStep = 'child-basics';
    this.successMessage = 'Draft cleared. You can start a new registration.';
  }

  protected dismissDraftBanner(): void {
    if (this.dismissTimeout) {
      clearTimeout(this.dismissTimeout);
      this.dismissTimeout = null;
    }
    this.isDraftRestoredBannerVisible = false;
  }

  protected dismissErrorMessage(): void {
    this.errorMessage = null;
  }

  protected dismissSuccessMessage(): void {
    this.successMessage = null;
  }

  protected formatDraftTimestamp(value: string | null): string {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    return date.toLocaleString();
  }

  private persistDraft(): void {
    if (!this.isNewRegistration || this.isSaving) {
      return;
    }
    const payload: RegistrationDraft = {
      currentStep: this.currentStep,
      step1: { ...this.step1 },
      step2: { ...this.step2 },
      step3: { ...this.step3 },
      step4: { ...this.step4 },
      step5: {
        patternEffectiveFrom: this.patternEffectiveFrom,
        patternEffectiveTo: this.patternEffectiveTo,
        patternEntries: [...this.patternEntries],
      },
      consentsReviewed: { ...this.consentsReviewed },
      parentCarersDraft: this.parentCarersDraft.map(contact => ({ ...contact })),
      emergencyContactsDraft: this.emergencyContactsDraft.map(contact => ({ ...contact })),
      emergencyAuthorisedFlags: [...this.emergencyAuthorisedFlags],
      emergencyContactAddresses: [...this.emergencyContactAddresses],
      referralsDraft: this.referralsDraft.map(r => ({ ...r })),
    };
    this.draftStorage.save(payload, this.currentStep);
    this.draftSavedAt = new Date().toISOString();
    this.hasStoredDraft = true;
  }

  private restoreDraftIfPresent(): void {
    if (!this.isNewRegistration) return;
    const raw = this.draftStorage.load();
    if (!raw) return;

    try {
      const draft = JSON.parse(raw) as Partial<RegistrationDraft>;
      this.applyDraft(draft);
      this.draftRestoredAt = this.draftStorage.loadSavedAt();
      this.draftSavedAt = this.draftRestoredAt;
      this.hasStoredDraft = true;
      this.isDraftRestoredBannerVisible = true;
      this.dismissTimeout = setTimeout(() => {
        this.isDraftRestoredBannerVisible = false;
        this.dismissTimeout = null;
      }, 5000);
    } catch {
      this.draftStorage.clear();
    }
  }

  restoreDraftIfPresentPublic(): void {
    this.restoreDraftIfPresent();
  }

  private applyDraft(draft: Partial<RegistrationDraft>): void {
    if (draft.step1) this.step1 = { ...this.step1, ...draft.step1 };
    if (draft.step2) {
      const legacy = draft.step2 as Partial<typeof draft.step2> & {
        has_allergies?: boolean;
        on_medication?: boolean;
        social_services_involvement?: boolean;
      };
      this.step2 = { ...this.step2, ...draft.step2 };
      this.step2.allergy_status = this.step2.allergy_status || this.legacyBooleanToYesNoUnknown(legacy.has_allergies);
      this.step2.medication_status = this.step2.medication_status || this.legacyBooleanToYesNoUnknown(legacy.on_medication);
      this.step2.social_services_status = this.step2.social_services_status || this.legacyBooleanToYesNoUnknown(legacy.social_services_involvement);
      this.step2.social_worker_name = this.step2.social_worker_name ?? '';
      this.step2.social_worker_phone = this.step2.social_worker_phone ?? '';
      this.step2.social_worker_email = this.step2.social_worker_email ?? '';
    }
    if (draft.step3) {
      this.step3 = { ...this.step3, ...draft.step3 };
      if (!this.step3.funding_support_answer && this.step3.has_funding_support) {
        this.step3.funding_support_answer = 'yes';
      }
      this.step3.has_funding_support = this.step3.funding_support_answer === 'yes';
      if (this.step3.funding_support_answer === 'no') {
        this.clearFundingOptions();
      }
    }
    if (draft.step4) {
      this.step4 = { ...this.step4, ...draft.step4 };
      for (const key of Object.keys(this.step4) as (keyof ConsentWritePayload)[]) {
        if (typeof this.step4[key] === 'boolean' && this.step4[key] === true) {
          this.consentsReviewed[key] = true;
        }
      }
    }
    if (draft.step5) {
      this.patternEffectiveFrom = draft.step5.patternEffectiveFrom ?? '';
      this.patternEffectiveTo = draft.step5.patternEffectiveTo ?? '';
      this.patternEntries = draft.step5.patternEntries ? draft.step5.patternEntries.map(e => ({ ...e })) : [];
    }
    if (draft.consentsReviewed) this.consentsReviewed = { ...this.consentsReviewed, ...draft.consentsReviewed };
    if (draft.parentCarersDraft?.length) {
      this.parentCarersDraft = draft.parentCarersDraft.map(contact => ({ ...contact }));
    }
    if (draft.emergencyContactsDraft?.length) {
      this.emergencyContactsDraft = draft.emergencyContactsDraft.map(contact => ({ ...contact }));
    }
    if (draft.emergencyAuthorisedFlags?.length) {
      this.emergencyAuthorisedFlags = [...draft.emergencyAuthorisedFlags];
    }
    if (draft.emergencyContactAddresses?.length) {
      this.emergencyContactAddresses = [...draft.emergencyContactAddresses];
    }
    if (draft.referralsDraft?.length) {
      this.referralsDraft = draft.referralsDraft.map(r => ({ ...r }));
    }
    if (draft.currentStep) {
      const validKeys = this.steps.map(step => step.key);
      if (validKeys.includes(draft.currentStep as StepperStep)) {
        this.currentStep = draft.currentStep as StepperStep;
      } else {
        this.currentStep = 'consents-evidence';
      }
    }
    this.step1Submitted = !!draft.step1?.first_name?.trim();
    this.successMessage = 'Restored your in-progress registration draft.';
  }

  private isEmptyDraft(): boolean {
    const step1Empty = Object.values(this.step1).every(value => !String(value ?? '').trim());
    const step2Empty = Object.entries(this.step2).every(([key, value]) => {
      if (typeof value === 'boolean') return value === false;
      return !String(value ?? '').trim();
    }) && this.referralsDraft.length === 0;
    const step3Empty = Object.entries(this.step3).every(([key, value]) => {
      if (typeof value === 'boolean') return value === false;
      return !String(value ?? '').trim();
    });
    return step1Empty && step2Empty && step3Empty;
  }

  private resetDrafts(): void {
    this.step1 = {
      first_name: '',
      middle_name: '',
      last_name: '',
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
      primary_room_id: '',
      registration_date: this.todayIso,
    };
    this.step2 = {
      allergy_status: '',
      allergy_details: '',
      medication_status: '',
      medication_name: '',
      medication_dosage: '',
      medication_storage: '',
      medication_side_effects: '',
      immunisation_status: '',
      immunisation_country: '',
      medical_history_status: '',
      illness_diagnosis_history: '',
      dietary_status: '',
      special_dietary_requirements: '',
      dietary_side_effects: '',
      doctor_address: '',
      doctor_name: '',
      doctor_phone: '',
      health_visitor_name: '',
      health_visitor_clinic: '',
      health_visitor_phone: '',
      social_services_status: '',
      social_services_details: '',
      social_worker_name: '',
      social_worker_phone: '',
      social_worker_email: '',
      concern_walking: false,
      concern_speech_language: false,
      concern_hearing: false,
      concern_sight: false,
      concern_emotional_wellbeing: false,
      concern_behaviour: false,
      routine_care_notes: '',
    };
    this.step3 = {
      collection_password: '',
      collection_password_hint: '',
      national_insurance_number: '',
      funding_support_answer: '',
      applying_for_funding: false,
      early_years_pupil_premium: false,
      working_tax_credit: false,
      college_uni_paid_to_parent: false,
      college_uni_paid_to_nursery: false,
      funding_3yo_term_time: false,
      funding_2yo_term_time: false,
      parent1_address: '',
      parent1_work_address: '',
      parent1_has_responsibility: null,
      show_second_parent: false,
      second_parent_name: '',
      second_parent_relationship: '',
      second_parent_telephone: '',
      second_parent_email: '',
      second_parent_address: '',
      second_parent_work_address: '',
      second_parent_has_responsibility: null,
      other_benefits: '',
      other_funding_selected: false,
      has_funding_support: false,
    };
    this.step4 = {
      urgent_medical_treatment: false,
      plasters: false,
      safeguarding_reporting_acknowledgement: false,
      information_truthfulness_declaration: false,
      information_sharing_consent: false,
      gdpr_data_processing_consent: false,
      area_senco_liaison: false,
      health_visitor_liaison: false,
      transition_documents: false,
      local_outings: false,
      face_painting: false,
      parent_supplied_sun_cream: false,
      parent_supplied_nappy_cream: false,
      development_profile_photos: false,
      nursery_display_boards: false,
      promotional_literature: false,
      nursery_website: false,
      staff_student_coursework: false,
      social_media: false,
      urgent_medical_treatment_exceptions: null,
      notes_exceptions: null,
      signer_name: '',
      signed_date: this.todayIso,
      consent_change_reason: null,
    };
    this.step4NoReasons = {};
    this.patternEffectiveFrom = '';
    this.patternEffectiveTo = '';
    this.patternEntries = [];
    this.patternError = null;
    this.availableSessionTypes = [];
    this.sessionPatternLoading = false;
    this.sessionPatternLoadError = null;
    this.parentCarersDraft = [this.emptyContact('Mother')];
    this.emergencyContactsDraft = [this.emptyContact('Grandparent')];
    this.emergencyAuthorisedFlags = [true];
    this.emergencyContactAddresses = [''];
    this.referralsDraft = [];
    this.step1Touched = {};
    this.step1Submitted = false;
    this.step2Touched = {};
    this.step2Submitted = false;
    this.step3Touched = {};
    this.step3Submitted = false;
    this.consentsReviewed = {};
    this.consentAdvisories = [];
    this.originalStep4Snapshot = null;
    this.fieldErrors = {};
    this.errorMessage = null;
  }
}
