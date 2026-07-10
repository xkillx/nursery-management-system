export interface ChildProfile {
  id: string;
  child_id: string;
  sex?: string | null;
  religion?: string | null;
  ethnic_origin?: string | null;
  first_language?: string | null;
  other_languages?: string | null;
  home_address: Record<string, unknown>;
  home_postcode?: string | null;
  home_telephone?: string | null;
  disability_status: string;
  disability_notes?: string | null;
  access_requirements?: string | null;
  routine_care_notes?: string | null;
  gdpr_declared_by_name?: string | null;
  gdpr_declared_at?: string | null;
  gdpr_declaration_date?: string | null;
  registration_date?: string | null;
  demographics_home_reviewed: boolean;
  medical_dietary_reviewed: boolean;
  health_contacts_reviewed: boolean;
  social_development_reviewed: boolean;
  parent_responsibility_reviewed: boolean;
  emergency_collection_reviewed: boolean;
  routine_care_reviewed: boolean;
  created_at: string;
  updated_at: string;
  // Legacy camelCase aliases for the still-imported manager-registration-intake
  // stepper. The new stepper uses snake_case fields directly.
  firstLanguage?: string;
  lastName?: string;
  medicalConditionsStatus?: string;
  medicalConditionsNotes?: string;
  prescribedMedicationStatus?: string;
  medicationNotes?: string;
  immunisationStatus?: string;
  immunisationCountry?: string;
  illnessDiagnosisHistory?: string;
  dietaryRequirementsStatus?: string;
  dietaryRequirementsNotes?: string;
  dietarySideEffects?: string;
  socialServicesStatus?: string;
  socialServicesNotes?: string;
  socialWorkerName?: string;
  socialWorkerPhone?: string;
  socialWorkerEmail?: string;
  concernWalking?: string;
  concernSpeechLanguage?: string;
  concernHearing?: string;
  concernSight?: string;
  concernEmotionalWellbeing?: string;
  concernBehaviour?: string;
  professionalReferrals?: ProfessionalReferral[];
  benefitsContributeToFees?: string;
  workingTaxCredit?: string;
  collegeUniPaidToParent?: string;
  collegeUniPaidToNursery?: string;
  funding3yoTermTime?: string;
  funding2yoTermTime?: string;
  fundingSupportNotes?: string;
  fundingSupportReviewed?: boolean;
  demographicsHome?: unknown;
  emergencyCollectionReviewed?: boolean;
  fundingSupport?: unknown;
  hasFundingSupport?: boolean;
  medicalHistoryStatus?: string;
}

export interface ChildProfileInput {
  sex?: string | null;
  religion?: string | null;
  ethnic_origin?: string | null;
  first_language?: string | null;
  other_languages?: string | null;
  home_address?: Record<string, unknown>;
  home_postcode?: string | null;
  home_telephone?: string | null;
  disability_status: string;
  disability_notes?: string | null;
  access_requirements?: string | null;
  routine_care_notes?: string | null;
  gdpr_declared_by_name?: string | null;
  gdpr_declared_at?: string | null;
  gdpr_declaration_date?: string | null;
  registration_date?: string | null;
  demographics_home_reviewed: boolean;
  medical_dietary_reviewed: boolean;
  health_contacts_reviewed: boolean;
  social_development_reviewed: boolean;
  parent_responsibility_reviewed: boolean;
  emergency_collection_reviewed: boolean;
  routine_care_reviewed: boolean;
}

export interface ChildHealthProfile {
  id: string;
  child_id: string;
  medical_conditions_status: string;
  medical_conditions_notes?: string | null;
  prescribed_medication_status: string;
  medication_notes?: string | null;
  immunisation_status: string;
  immunisation_country?: string | null;
  illness_diagnosis_history?: string | null;
  dietary_requirements_status: string;
  dietary_requirements_notes?: string | null;
  dietary_side_effects?: string | null;
  doctor_name?: string | null;
  doctor_address?: string | null;
  doctor_phone?: string | null;
  health_visitor_name?: string | null;
  health_visitor_address?: string | null;
  health_visitor_phone?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ChildHealthProfileInput {
  medical_conditions_status: string;
  medical_conditions_notes?: string | null;
  prescribed_medication_status: string;
  medication_notes?: string | null;
  immunisation_status: string;
  immunisation_country?: string | null;
  illness_diagnosis_history?: string | null;
  dietary_requirements_status: string;
  dietary_requirements_notes?: string | null;
  dietary_side_effects?: string | null;
  doctor_name?: string | null;
  doctor_address?: string | null;
  doctor_phone?: string | null;
  health_visitor_name?: string | null;
  health_visitor_address?: string | null;
  health_visitor_phone?: string | null;
}

export interface ProfessionalReferral {
  type: string;
  referred_date?: string | null;
  referred_by?: string | null;
  waiting_list_status: string;
  notes?: string | null;
}

export interface ChildSafeguardingProfile {
  id: string;
  child_id: string;
  social_services_status: string;
  social_services_notes?: string | null;
  social_worker_name?: string | null;
  social_worker_phone?: string | null;
  social_worker_email?: string | null;
  concern_walking: string;
  concern_speech_language: string;
  concern_hearing: string;
  concern_sight: string;
  concern_emotional_wellbeing: string;
  concern_behaviour: string;
  professional_referrals: ProfessionalReferral[];
  restricted_notes?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ChildSafeguardingProfileInput {
  social_services_status: string;
  social_services_notes?: string | null;
  social_worker_name?: string | null;
  social_worker_phone?: string | null;
  social_worker_email?: string | null;
  concern_walking: string;
  concern_speech_language: string;
  concern_hearing: string;
  concern_sight: string;
  concern_emotional_wellbeing: string;
  concern_behaviour: string;
  professional_referrals: ProfessionalReferral[];
  restricted_notes?: string | null;
}

export interface ChildContact {
  id?: string;
  contact_type?: 'parent_carer' | 'emergency_contact' | 'authorised_collector';
  sort_order?: number;
  full_name?: string;
  fullName: string;
  relationship_to_child?: string | null;
  relationshipToChild?: string | null;
  address?: Record<string, unknown> | null;
  telephone?: string | null;
  email?: string | null;
  work_address?: Record<string, unknown> | null;
  workAddress?: Record<string, unknown> | null;
  has_parental_responsibility?: boolean | null;
  hasParentalResponsibility?: boolean | null;
}

export interface ChildConsent {
  id: string;
  child_id: string;
  urgent_medical_treatment: boolean;
  urgent_medical_treatment_exceptions?: string | null;
  plasters: boolean;
  safeguarding_reporting_acknowledgement: boolean;
  information_sharing_consent: boolean;
  information_truthfulness_declaration?: boolean;
  gdpr_data_processing_consent: boolean;
  area_senco_liaison: boolean;
  health_visitor_liaison: boolean;
  transition_documents: boolean;
  local_outings: boolean;
  face_painting: boolean;
  parent_supplied_sun_cream: boolean;
  parent_supplied_nappy_cream: boolean;
  development_profile_photos: boolean;
  nursery_display_boards: boolean;
  promotional_literature: boolean;
  nursery_website: boolean;
  staff_student_coursework: boolean;
  social_media: boolean;
  social_media_channel_notes?: string | null;
  notes_exceptions?: string | null;
  signer_name: string;
  signed_date: string;
  paper_form_on_file: boolean;
  created_at: string;
  updated_at: string;
}

export interface ChildConsentInput {
  urgent_medical_treatment: boolean;
  urgent_medical_treatment_exceptions?: string | null;
  plasters: boolean;
  safeguarding_reporting_acknowledgement: boolean;
  information_truthfulness_declaration?: boolean;
  information_sharing_consent: boolean;
  gdpr_data_processing_consent: boolean;
  area_senco_liaison: boolean;
  health_visitor_liaison: boolean;
  transition_documents: boolean;
  local_outings: boolean;
  face_painting: boolean;
  parent_supplied_sun_cream: boolean;
  parent_supplied_nappy_cream: boolean;
  development_profile_photos: boolean;
  nursery_display_boards: boolean;
  promotional_literature: boolean;
  nursery_website: boolean;
  staff_student_coursework: boolean;
  social_media: boolean;
  social_media_channel_notes?: string | null;
  notes_exceptions?: string | null;
  signer_name: string;
  signed_date: string;
  paper_form_on_file?: boolean;
  consent_change_reason?: string | null;
}

export type FundingType = 'none' | 'fifteen_hours' | 'thirty_hours' | 'two_year_old' | 'custom' | 'unknown';
export type FundingModel = 'term_time_only' | 'stretched' | 'unknown';
export type BenefitsStatus = 'no' | 'yes' | 'unknown';

export interface ChildFundingRecord {
  id: string;
  child_id: string;
  funding_enabled: boolean;
  funding_type: FundingType;
  funding_model: FundingModel;
  funded_hours_per_week: number | null;
  funding_start_date: string | null;
  funding_end_date: string | null;
  eligibility_code: string | null;
  eligibility_code_validated: boolean;
  evidence_received: boolean;
  benefits_status: BenefitsStatus;
  benefits: string[];
  other_benefit_name: string | null;
  benefit_notes: string | null;
  manager_notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface ChildFundingRecordInput {
  funding_enabled: boolean;
  funding_type: FundingType;
  funding_model: FundingModel;
  funded_hours_per_week: number | null;
  funding_start_date: string | null;
  funding_end_date: string | null;
  eligibility_code: string | null;
  eligibility_code_validated: boolean;
  evidence_received: boolean;
  benefits_status: BenefitsStatus;
  benefits: string[];
  other_benefit_name: string | null;
  benefit_notes: string | null;
  manager_notes: string | null;
}

export interface ChildCollectionSettings {
  id: string;
  child_id: string;
  over_18_collection_acknowledged: boolean;
  collection_password_set: boolean;
  collection_password?: string | null;
  collection_password_hint?: string | null;
  collection_password_updated_at?: string | null;
  collection_password_updated_by_user_id?: string | null;
  collection_password_updated_by_membership_id?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ChildCollectionSettingsInput {
  password?: string;
  password_hint?: string;
  over_18_collection_acknowledged?: boolean;
}

export interface ChildRoomAssignment {
  id: string;
  child_id: string;
  room_id: string;
  start_date: string;
  end_date?: string | null;
  is_current: boolean;
  created_at: string;
}

export interface ChildRoomAssignmentInput {
  room_id: string;
  start_date: string;
}

export interface ChildBillingProfile {
  id: string;
  child_id: string;
  billing_basis: 'site_rate' | 'custom';
  custom_rate_minor?: number | null;
  effective_from: string;
  created_at: string;
  updated_at: string;
}

export interface ChildBillingProfileInput {
  billing_basis: 'site_rate' | 'custom';
  custom_rate_minor?: number | null;
  effective_from?: string;
}

export interface ChildLeavingRecord {
  id: string;
  child_id: string;
  left_at: string;
  reason_code: string;
  reason_note?: string | null;
  created_at: string;
}

export interface CreateChildPayload {
  child: {
    first_name: string;
    middle_name?: string | null;
    last_name?: string | null;
    date_of_birth: string;
    start_date: string;
    end_date?: string;
    notes?: string;
  };
  profile?: ChildProfileInput;
  health?: ChildHealthProfileInput;
  safeguarding?: ChildSafeguardingProfileInput;
  contacts?: {
    parent_carers?: Record<string, unknown>[];
    emergency_contacts?: Record<string, unknown>[];
    authorised_collectors?: Record<string, unknown>[];
  };
  consent: ChildConsentInput;
  funding?: ChildFundingRecordInput;
  collection_settings?: ChildCollectionSettingsInput;
  room: {
    room_id: string;
    start_date: string;
  };
  booking_pattern?: {
    effective_from: string;
    effective_to?: string;
    entries: { day_of_week: number; session_type_id: string }[];
  };
}

export interface CreateChildResponse {
  id: string;
  first_name: string;
  middle_name?: string | null;
  last_name?: string | null;
  start_date: string;
  created_sub_records: string[];
}
