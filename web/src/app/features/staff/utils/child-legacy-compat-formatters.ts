const PROFILE_SECTION_LABELS: Record<string, string> = {
  child_demographics_home: 'Demographics and home',
  medical_dietary: 'Medical and dietary',
  health_contacts: 'Doctor and health visitor',
  social_development: 'Social services and development',
  parent_responsibility: 'Parent/carer responsibility',
  emergency_collection: 'Emergency contacts and collection',
  funding_benefits: 'Funding and benefits',
  routine_care: 'Routine care',
  gdpr_declaration: 'GDPR declaration',
};

const PROFILE_MISSING_FIELD_LABELS: Record<string, string> = {
  review_required: 'Review required',
  medical_conditions_status_unknown: 'Medical conditions not answered',
  prescribed_medication_status_unknown: 'Prescribed medication not answered',
  dietary_requirements_status_unknown: 'Dietary requirements not answered',
  immunisation_status_unknown: 'Immunisation status not answered',
  medical_conditions_notes_required: 'Medical conditions notes required',
  medication_notes_required: 'Medication notes required',
  dietary_requirements_notes_required: 'Dietary requirements notes required',
  social_services_status_unknown: 'Social services status not answered',
  social_services_notes_or_worker_required: 'Social services notes or worker required',
  concern_walking_unknown: 'Walking concern not answered',
  concern_speech_language_unknown: 'Speech/language concern not answered',
  concern_hearing_unknown: 'Hearing concern not answered',
  concern_sight_unknown: 'Sight concern not answered',
  concern_emotional_wellbeing_unknown: 'Emotional wellbeing concern not answered',
  concern_behaviour_unknown: 'Behaviour concern not answered',
  parent_carer_or_notes_required: 'Parent/carer or notes required',
  emergency_contact_missing: 'Emergency contact missing',
  authorised_collector_missing: 'Authorised collector missing',
  over18_collection_acknowledgement_required: 'Over-18 collection acknowledgement required',
  collection_password_missing: 'Collection password missing',
  gdpr_declared_by_name_missing: 'Declaring person name missing',
  gdpr_declared_at_missing: 'Confirmation timestamp missing',
  gdpr_declaration_date_missing: 'Declaration date missing',
};

const OFFICE_CHECK_STATUS_LABELS: Record<string, string> = {
  unknown: 'Unknown',
  complete: 'Complete',
  missing: 'Still needed',
  not_applicable: 'Not applicable',
};

const OFFICE_ITEM_LABELS: Record<string, string> = {
  deposit: 'Deposit',
  application_date: 'Application date',
  start_date_check: 'Start date check',
  sessions_days_requested: 'Sessions/days requested',
  term_time_only_space: 'Term-time-only space',
  contract: 'Contract/signature',
  handbook: 'Handbook',
  red_book: 'Red Book',
  birth_certificate_passport: 'Birth certificate/passport',
  proof_of_address: 'Proof of address',
};

export function formatProfileSectionLabel(code: string): string {
  return PROFILE_SECTION_LABELS[code] ?? code;
}

export function formatProfileMissingFieldLabel(code: string): string {
  return PROFILE_MISSING_FIELD_LABELS[code] ?? code.replace(/_/g, ' ');
}

export function formatOfficeItemLabel(code: string): string {
  return OFFICE_ITEM_LABELS[code] ?? code.replace(/_/g, ' ');
}

export function formatOfficeCheckStatusLabel(status: string): string {
  return OFFICE_CHECK_STATUS_LABELS[status] ?? status;
}

export function getCompletionBadgeClass(isComplete: boolean): string {
  return isComplete
    ? 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-400'
    : 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-400';
}

export function formatCompletionStatus(isComplete: boolean): string {
  return isComplete ? 'Complete' : 'Incomplete';
}
