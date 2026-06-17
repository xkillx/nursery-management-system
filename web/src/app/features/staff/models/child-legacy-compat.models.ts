// Legacy alias file. The registration model has been split into per-resource
// models in child-profile.models.ts. This file re-exports the new types under
// names the manager-child-edit stepper still references. Once the stepper is
// rewritten to call the per-resource endpoints directly, this file can be
// removed.
import {
  ChildContact as RegistrationContactEntry,
  ChildConsent as ConsentRecord,
  ChildConsentInput as ConsentWritePayload,
  ChildCollectionSettings,
  ChildFundingRecord,
  ChildProfile,
  ChildHealthProfile,
  ChildSafeguardingProfile,
  CreateChildPayload as CompleteRegistrationPayload,
  CreateChildResponse as CompleteRegistrationResponse,
} from './child-profile.models';

export type {
  RegistrationContactEntry,
  ConsentRecord,
  ConsentWritePayload,
  ChildCollectionSettings,
  ChildFundingRecord,
  ChildProfile,
  ChildHealthProfile,
  ChildSafeguardingProfile,
  CompleteRegistrationPayload,
  CompleteRegistrationResponse,
};

// Aggregated view of a child + all its sub-records, used by the stepper for
// load+populate. Built by the StaffApiService loader that fans out to the
// per-resource endpoints.
export type StepperProfileView = {
  child: ChildProfile;
  profile: ChildProfile | null;
  health: ChildHealthProfile | null;
  safeguarding: ChildSafeguardingProfile | null;
  parentCarers: RegistrationContactEntry[];
  emergencyContacts: RegistrationContactEntry[];
  authorisedCollectors: RegistrationContactEntry[];
  collection: { over18CollectionAcknowledged: boolean; hasCollectionPassword: boolean } | null;
  funding: ChildFundingRecord | null;
  consent: ConsentRecord | null;
  completeness: {
    isComplete: boolean;
    missingSections: string[];
  };
};

// Aggregated completion state. The old "workflow status" endpoint was dropped;
// the stepper now derives this from the per-resource responses it has already
// loaded.
export type StepperCompletionStatus = {
  isReviewedComplete: boolean;
  canMarkComplete: boolean;
  needsReview: boolean;
  missingGroups: string[];
  currentConsent: ConsentRecord | null;
};
