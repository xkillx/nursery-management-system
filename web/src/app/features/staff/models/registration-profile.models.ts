// Compatibility shim. The registration model has been split into per-resource
// models in child-profile.models.ts. This file re-exports the legacy types
// referenced by the legacy manager-registration-intake component, which is
// being progressively replaced by manager-child-edit.
import {
  ChildProfile as ChildProfileRaw,
  ChildProfileInput as ChildProfileInputRaw,
  ChildHealthProfile as ChildHealthProfileRaw,
  ChildHealthProfileInput as ChildHealthProfileInputRaw,
  ChildSafeguardingProfile as ChildSafeguardingProfileRaw,
  ChildSafeguardingProfileInput as ChildSafeguardingProfileInputRaw,
  ChildContact as ChildContactRaw,
  ChildConsent as ChildConsentRaw,
  ChildConsentInput as ChildConsentInputRaw,
  ChildFundingRecord as ChildFundingRecordRaw,
  ChildCollectionSettings as ChildCollectionSettingsRaw,
  CreateChildPayload as CreateChildPayloadRaw,
  CreateChildResponse as CreateChildResponseRaw,
} from './child-profile.models';

export type RegistrationProfileResponse = ChildProfileRaw & {
  child: { id: string; fullName: string; dateOfBirth: string };
  profileExists: boolean;
  profile: { id: string; createdAt: string; updatedAt: string } | null;
  medicalDietary: { [key: string]: any } | null;
  healthContacts: { [key: string]: any } | null;
  socialDevelopment: { [key: string]: any } | null;
  parentCarers: RegistrationContactEntry[];
  emergencyContacts: RegistrationContactEntry[];
  authorisedCollectors: RegistrationContactEntry[];
  collection: { [key: string]: any } | null;
  fundingSupport: { [key: string]: any } | null;
  routineCare: { [key: string]: any } | null;
  gdprDeclaration: { [key: string]: any } | null;
  registrationDate: string | null;
  completeness: RegistrationProfileCompleteness;
  demographicsHome?: { [key: string]: any } | null;
};

export type RegistrationProfileCompleteness = {
  isComplete: boolean;
  missingSections: string[];
  sections: { code: string; status: 'complete' | 'incomplete'; missingFields: string[] }[];
  [key: string]: any;
};

export type RegistrationContactEntry = ChildContactRaw;

export type ConsentRecord = ChildConsentRaw;
export type ConsentWritePayload = ChildConsentInputRaw;

export type ChildHealthRecord = ChildHealthProfileRaw;
export type ChildHealthWritePayload = ChildHealthProfileInputRaw;
export type ChildSafeguardingRecord = ChildSafeguardingProfileRaw;
export type ChildSafeguardingWritePayload = ChildSafeguardingProfileInputRaw;

export type ChildFundingWritePayload = ChildProfileInputRaw;

export type CollectionPasswordPayload = {
  password: string;
};

export type RegistrationWorkflowStatus = {
  child_summary: { id: string; full_name: string; date_of_birth: string };
  is_reviewed_complete: boolean;
  can_mark_complete: boolean;
  needs_review: boolean;
  missing_groups: string[];
  current_consent_record?: ChildConsentRaw | null;
  profile_completeness?: any;
  consent_completeness?: any;
  latest_attestation?: any;
  [key: string]: any;
};

export type ConsentWithCompletenessResponse = {
  current: ChildConsentRaw | null;
  history: ChildConsentRaw[];
  completeness: {
    is_complete: boolean;
    missing_decisions: string[];
  };
};

export type CompleteRegistrationPayload = CreateChildPayloadRaw;
export type CompleteRegistrationResponse = CreateChildResponseRaw;

export type ChildFundingSupportRecord = ChildFundingRecordRaw;
export type ChildCollectionSettingsRecord = ChildCollectionSettingsRaw;
