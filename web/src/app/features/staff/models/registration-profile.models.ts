export interface RegistrationChildSummary {
  id: string;
  fullName: string;
  dateOfBirth: string;
}

export interface RegistrationProfileMetadata {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface RegistrationContactEntry {
  fullName: string;
  relationshipToChild: string | null;
  address: Record<string, unknown> | null;
  telephone: string | null;
  email: string | null;
  workAddress: Record<string, unknown> | null;
  hasParentalResponsibility: boolean | null;
}

export interface RegistrationProfileDemographicsHome {
  sex: string | null;
  religion: string | null;
  ethnicOrigin: string | null;
  firstLanguage: string | null;
  otherLanguages: string | null;
  homeAddress: Record<string, unknown> | null;
  homePostcode: string | null;
  homeTelephone: string | null;
  disabilityStatus: string | null;
  disabilityNotes: string | null;
  accessRequirements: string | null;
  demographicsHomeReviewed: boolean;
}

export interface RegistrationProfileMedicalDietary {
  medicalConditionsStatus: string | null;
  medicalConditionsNotes: string | null;
  prescribedMedicationStatus: string | null;
  medicationNotes: string | null;
  immunisationStatus: string | null;
  immunisationCountry: string | null;
  illnessDiagnosisHistory: string | null;
  dietaryRequirementsStatus: string | null;
  dietaryRequirementsNotes: string | null;
  dietarySideEffects: string | null;
  medicalDietaryReviewed: boolean;
}

export interface RegistrationProfileHealthContacts {
  doctorName: string | null;
  doctorAddress: string | null;
  doctorPhone: string | null;
  healthVisitorName: string | null;
  healthVisitorAddress: string | null;
  healthVisitorPhone: string | null;
  healthContactsReviewed: boolean;
}

export interface RegistrationProfileSocialDevelopment {
  socialServicesStatus: string | null;
  socialServicesNotes: string | null;
  socialWorkerName: string | null;
  socialWorkerPhone: string | null;
  socialWorkerEmail: string | null;
  concernWalking: string | null;
  concernSpeechLanguage: string | null;
  concernHearing: string | null;
  concernSight: string | null;
  concernEmotionalWellbeing: string | null;
  concernBehaviour: string | null;
  professionalReferrals: ProfessionalReferral[];
  socialDevelopmentReviewed: boolean;
}

export interface ProfessionalReferral {
  type: string;
  referredDate: string | null;
  referredBy: string | null;
  waitingListStatus: string;
  notes: string | null;
}

export interface RegistrationProfileCollection {
  isSet: boolean;
  lastUpdatedAt: string | null;
  lastUpdatedByUserId: string | null;
  lastUpdatedByMembershipId: string | null;
  over18CollectionAcknowledged: boolean;
  emergencyCollectionReviewed: boolean;
}

export interface RegistrationProfileFundingSupport {
  benefitsContributeToFees: string | null;
  workingTaxCredit: string | null;
  collegeUniPaidToParent: string | null;
  collegeUniPaidToNursery: string | null;
  funding3yoTermTime: string | null;
  funding2yoTermTime: string | null;
  fundingSupportNotes: string | null;
  fundingSupportReviewed: boolean;
}

export interface RegistrationProfileRoutineCare {
  routineCareNotes: string | null;
  routineCareReviewed: boolean;
}

export interface RegistrationProfileGDPRDeclaration {
  gdprDeclaredByName: string | null;
  gdprDeclaredAt: string | null;
  gdprDeclarationDate: string | null;
}

export interface RegistrationProfileCompletenessSection {
  code: string;
  status: 'complete' | 'incomplete';
  missingFields: string[];
}

export interface RegistrationProfileCompleteness {
  isComplete: boolean;
  missingSections: string[];
  sections?: { code: string; status: string; missingFields: string[] }[];
}

export interface ConsentRecord {
  id: string;
  child_id: string;
  version: number;
  source: string;
  paper_form_on_file: boolean;
  urgent_medical_treatment: boolean;
  urgent_medical_treatment_exceptions: string | null;
  plasters: boolean;
  safeguarding_reporting_acknowledgement: boolean;
  information_truthfulness_declaration: boolean;
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
  social_media_channel_notes: string | null;
  notes_exceptions: string | null;
  entered_by_user_id: string;
  entered_by_membership_id: string;
  created_at: string;
}

export interface ConsentWritePayload {
  paper_form_on_file: boolean;
  urgent_medical_treatment: boolean;
  urgent_medical_treatment_exceptions?: string | null;
  plasters: boolean;
  safeguarding_reporting_acknowledgement: boolean;
  information_truthfulness_declaration: boolean;
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
}

export interface ConsentWithCompletenessResponse {
  child: { id: string; first_name: string; middle_name?: string | null; last_name?: string | null; date_of_birth: string };
  current: ConsentRecord | null;
  history: ConsentRecord[];
  completeness: { is_complete: boolean; missing_decisions: string[] };
}

export interface RegistrationWorkflowStatus {
  child: { id: string; first_name: string; middle_name?: string | null; last_name?: string | null; date_of_birth: string };
  profile_completeness: {
    is_complete: boolean;
    missing_sections: string[];
    sections?: { code: string; status: string; missing_fields: string[] }[];
  };
  consent_completeness: { is_complete: boolean; missing_decisions: string[] };
  current_consent_record?: ConsentRecord | null;
  latest_attestation?: {
    id: string;
    consent_record_id?: string | null;
    attested_by_user_id: string;
    attested_by_membership_id: string;
    attested_at: string;
  } | null;
  can_mark_complete: boolean;
  is_reviewed_complete: boolean;
  needs_review: boolean;
  missing_groups: string[];
}

export interface RegistrationProfileResponse {
  child: RegistrationChildSummary;
  profileExists: boolean;
  profile: RegistrationProfileMetadata | null;
  demographicsHome: RegistrationProfileDemographicsHome | null;
  medicalDietary: RegistrationProfileMedicalDietary | null;
  healthContacts: RegistrationProfileHealthContacts | null;
  socialDevelopment: RegistrationProfileSocialDevelopment | null;
  parentCarers: RegistrationContactEntry[];
  emergencyContacts: RegistrationContactEntry[];
  authorisedCollectors: RegistrationContactEntry[];
  collection: RegistrationProfileCollection | null;
  fundingSupport: RegistrationProfileFundingSupport | null;
  routineCare: RegistrationProfileRoutineCare | null;
  gdprDeclaration: RegistrationProfileGDPRDeclaration | null;
  completeness: RegistrationProfileCompleteness;
}

export interface CompleteRegistrationChildPayload {
  first_name: string;
  middle_name?: string | null;
  last_name?: string | null;
  date_of_birth: string;
  start_date: string;
  notes?: string;
}

export interface CompleteRegistrationPayload {
  child: CompleteRegistrationChildPayload;
  registration_profile?: Record<string, unknown>;
  consents?: ConsentWritePayload;
  collection_password?: string;
}

export interface CompleteRegistrationResponse {
  id: string;
  first_name: string;
  middle_name?: string | null;
  last_name?: string | null;
  start_date: string;
}

export interface CollectionPasswordPayload {
  password: string;
}
