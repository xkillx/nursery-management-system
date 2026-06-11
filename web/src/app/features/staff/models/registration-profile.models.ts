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
  otherLanguages: string[];
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
  socialWorkerContactDetails: string | null;
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
  sections: RegistrationProfileCompletenessSection[];
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

export interface CollectionPasswordPayload {
  password: string;
}

export interface OfficeChildSummary {
  id: string;
  fullName: string;
  dateOfBirth: string;
  startDate: string | null;
  endDate: string | null;
}

export interface OfficeChecklistMetadata {
  id: string;
  createdAt: string;
  updatedAt: string;
}

export interface OfficeUseChecklist {
  depositStatus: string | null;
  depositPaidDate: string | null;
  applicationDateStatus: string | null;
  applicationDate: string | null;
  startDateStatus: string | null;
  dateLeft: string | null;
  sessionsDaysRequestedStatus: string | null;
  sessionsDaysRequested: string | null;
  termTimeOnlySpaceStatus: string | null;
  contractStatus: string | null;
  contractDate: string | null;
  handbookStatus: string | null;
  handbookDate: string | null;
  redBookStatus: string | null;
  redBookCheckedDate: string | null;
  birthCertificatePassportStatus: string | null;
  birthCertificatePassportCheckedDate: string | null;
  proofOfAddressStatus: string | null;
  proofOfAddressCheckedDate: string | null;
  notes: string | null;
}

export interface OfficeCompletenessItem {
  code: string;
  status: 'complete' | 'incomplete';
  label: string;
  missingFields: string[];
}

export interface OfficeUseCompleteness {
  isComplete: boolean;
  missingFields: string[];
  items: OfficeCompletenessItem[];
}

export interface RegistrationOfficeUseChecklistResponse {
  child: OfficeChildSummary;
  checklistExists: boolean;
  checklist: OfficeChecklistMetadata | null;
  officeUseChecklist: OfficeUseChecklist;
  completeness: OfficeUseCompleteness;
}
