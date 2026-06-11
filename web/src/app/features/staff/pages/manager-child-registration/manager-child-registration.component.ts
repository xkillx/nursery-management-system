import { Component, OnInit, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';

import { StaffApiService } from '../../data/staff-api.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';

import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import {
  RegistrationProfileResponse, RegistrationOfficeUseChecklistResponse,
  OfficeUseChecklist, RegistrationContactEntry,
  RegistrationProfileDemographicsHome, RegistrationProfileMedicalDietary,
  RegistrationProfileHealthContacts, RegistrationProfileSocialDevelopment,
  RegistrationProfileFundingSupport, RegistrationProfileRoutineCare,
} from '../../models/registration-profile.models';
import {
  formatProfileSectionLabel, formatProfileMissingFieldLabel,
  formatOfficeItemLabel, formatOfficeCheckStatusLabel,
  getCompletionBadgeClass, formatCompletionStatus,
} from '../../utils/registration-profile-formatters';

@Component({
  selector: 'app-manager-child-registration',
  standalone: true,
  imports: [
    RouterLink, DatePipe, FormsModule,
    AlertComponent, LoadingStateComponent, PageHeaderComponent,
    ButtonComponent,
  ],
  templateUrl: './manager-child-registration.component.html',
})
export class ManagerChildRegistrationComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private api = inject(StaffApiService);
  private errorMapper = inject(ApiErrorMapper);

  childId: string | null = null;

  isLoading = true;
  isSavingSection: string | null = null;
  errorMessage: string | null = null;

  profile: RegistrationProfileResponse | null = null;
  checklist: RegistrationOfficeUseChecklistResponse | null = null;

  collectionPassword = '';
  collectionPasswordMessage: string | null = null;
  collectionPasswordError: string | null = null;

  sectionMessages: Record<string, string> = {};
  sectionErrors: Record<string, string> = {};

  demoHomeDraft: RegistrationProfileDemographicsHome | null = null;
  medicalDietaryDraft: RegistrationProfileMedicalDietary | null = null;
  healthContactsDraft: RegistrationProfileHealthContacts | null = null;
  socialDevDraft: RegistrationProfileSocialDevelopment | null = null;
  parentCarersDraft: RegistrationContactEntry[] = [];
  emergencyContactsDraft: RegistrationContactEntry[] = [];
  authorisedCollectorsDraft: RegistrationContactEntry[] = [];
  collectionOver18 = false;
  collectionEmergencyReviewed = false;
  fundingSupportDraft: RegistrationProfileFundingSupport | null = null;
  routineCareDraft: RegistrationProfileRoutineCare | null = null;
  gdprName = '';
  gdprDate = '';
  officeDraft: OfficeUseChecklist | null = null;

  ngOnInit(): void {
    this.childId = this.route.snapshot.paramMap.get('childId');
    if (!this.childId) {
      this.errorMessage = 'Child ID is required.';
      this.isLoading = false;
      return;
    }
    this.loadAll();
  }

  private loadAll(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.api.getRegistrationProfile(this.childId!).subscribe({
      next: (profile) => {
        this.profile = profile;
        this.initDraftsFromProfile(profile);
        this.loadChecklist();
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isLoading = false;
      },
    });
  }

  private initDraftsFromProfile(profile: RegistrationProfileResponse): void {
    this.demoHomeDraft = profile.demographicsHome ? { ...profile.demographicsHome } : null;
    this.medicalDietaryDraft = profile.medicalDietary ? { ...profile.medicalDietary } : null;
    this.healthContactsDraft = profile.healthContacts ? { ...profile.healthContacts } : null;
    this.socialDevDraft = profile.socialDevelopment ? { ...profile.socialDevelopment } : null;
    this.parentCarersDraft = profile.parentCarers ? profile.parentCarers.map(c => ({ ...c })) : [];
    this.emergencyContactsDraft = profile.emergencyContacts ? profile.emergencyContacts.map(c => ({ ...c })) : [];
    this.authorisedCollectorsDraft = profile.authorisedCollectors ? profile.authorisedCollectors.map(c => ({ ...c })) : [];
    this.fundingSupportDraft = profile.fundingSupport ? { ...profile.fundingSupport } : null;
    this.routineCareDraft = profile.routineCare ? { ...profile.routineCare } : null;
    this.gdprName = profile.gdprDeclaration?.gdprDeclaredByName ?? '';
    this.gdprDate = profile.gdprDeclaration?.gdprDeclarationDate ?? '';
    this.collectionOver18 = profile.collection?.over18CollectionAcknowledged ?? false;
    this.collectionEmergencyReviewed = profile.collection?.emergencyCollectionReviewed ?? false;
  }

  private initDraftsFromChecklist(checklist: RegistrationOfficeUseChecklistResponse): void {
    this.officeDraft = checklist.officeUseChecklist ? { ...checklist.officeUseChecklist } : null;
  }

  private loadChecklist(): void {
    this.api.getRegistrationOfficeUseChecklist(this.childId!).subscribe({
      next: (checklist) => {
        this.checklist = checklist;
        this.initDraftsFromChecklist(checklist);
        this.isLoading = false;
      },
      error: () => {
        this.checklist = null;
        this.isLoading = false;
      },
    });
  }

  get childName(): string {
    return this.profile?.child?.fullName ?? '';
  }

  get childDob(): string {
    return this.profile?.child?.dateOfBirth ?? '';
  }

  get profileCompleteness() {
    return this.profile?.completeness ?? null;
  }

  get officeCompleteness() {
    return this.checklist?.completeness ?? null;
  }

  get profileCompletionBadge(): string {
    return getCompletionBadgeClass(this.profileCompleteness?.isComplete ?? false);
  }

  get officeCompletionBadge(): string {
    return getCompletionBadgeClass(this.officeCompleteness?.isComplete ?? false);
  }

  private patchProfile(sectionKey: string, patch: Record<string, unknown>): void {
    if (!this.childId) return;

    this.isSavingSection = sectionKey;
    this.sectionMessages[sectionKey] = '';
    this.sectionErrors[sectionKey] = '';

    this.api.patchRegistrationProfile(this.childId, patch).subscribe({
      next: (updated) => {
        this.profile = updated;
        this.initDraftsFromProfile(updated);
        this.isSavingSection = null;
        this.sectionMessages[sectionKey] = 'Section saved.';
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.sectionErrors[sectionKey] = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  protected saveDemographicsHome(): void {
    this.patchProfile('demographics_home', this.buildDemographicsHomePatch());
  }

  protected saveMedicalDietary(): void {
    this.patchProfile('medical_dietary', this.buildMedicalDietaryPatch());
  }

  protected saveHealthContacts(): void {
    this.patchProfile('health_contacts', this.buildHealthContactsPatch());
  }

  protected saveSocialDevelopment(): void {
    this.patchProfile('social_development', this.buildSocialDevelopmentPatch());
  }

  protected saveCollectionFlags(): void {
    this.patchProfile('collection', this.buildCollectionPatch());
  }

  protected saveFundingSupport(): void {
    this.patchProfile('funding_support', this.buildFundingSupportPatch());
  }

  protected saveRoutineCare(): void {
    this.patchProfile('routine_care', this.buildRoutineCarePatch());
  }

  protected saveGdprDeclaration(): void {
    this.patchProfile('gdpr_declaration', this.buildGdprPatch());
  }

  protected saveContacts(contactType: string): void {
    let entries: RegistrationContactEntry[];
    switch (contactType) {
      case 'parent_carers': entries = this.parentCarersDraft; break;
      case 'emergency_contacts': entries = this.emergencyContactsDraft; break;
      case 'authorised_collectors': entries = this.authorisedCollectorsDraft; break;
      default: return;
    }
    this.patchProfile(contactType, { [contactType]: entries });
  }

  protected saveOfficeChecklist(): void {
    if (!this.childId) return;

    this.isSavingSection = 'office_use_checklist';
    this.sectionMessages['office_use_checklist'] = '';
    this.sectionErrors['office_use_checklist'] = '';

    const patch = this.buildOfficeChecklistPatch();
    if (Object.keys(patch).length === 0) {
      this.sectionErrors['office_use_checklist'] = 'Office-use checklist data not loaded.';
      this.isSavingSection = null;
      return;
    }

    this.api.patchRegistrationOfficeUseChecklist(this.childId, patch).subscribe({
      next: (updated) => {
        this.checklist = updated;
        this.initDraftsFromChecklist(updated);
        this.isSavingSection = null;
        this.sectionMessages['office_use_checklist'] = 'Office-use checklist saved.';
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.sectionErrors['office_use_checklist'] = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  protected setCollectionPassword(): void {
    if (!this.childId || !this.collectionPassword) return;

    this.collectionPasswordMessage = '';
    this.collectionPasswordError = '';
    this.isSavingSection = 'collection';

    this.api.setRegistrationCollectionPassword(this.childId, this.collectionPassword).subscribe({
      next: (updated) => {
        this.profile = updated;
        this.initDraftsFromProfile(updated);
        this.collectionPassword = '';
        this.collectionPasswordMessage = 'Collection password set.';
        this.isSavingSection = null;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.collectionPasswordError = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  protected readonly formatProfileSectionLabel = formatProfileSectionLabel;
  protected readonly formatProfileMissingFieldLabel = formatProfileMissingFieldLabel;
  protected readonly formatOfficeItemLabel = formatOfficeItemLabel;
  protected readonly formatOfficeCheckStatusLabel = formatOfficeCheckStatusLabel;
  protected readonly formatCompletionStatus = formatCompletionStatus;

  /* Conversion helpers */

  protected nullableStr(v: string | null): string {
    return v ?? '';
  }

  protected toNullWhenEmpty(v: string): string | null {
    const t = v.trim();
    return t || null;
  }

  protected addressToString(addr: Record<string, unknown> | null | undefined): string {
    if (!addr) return '';
    if (typeof addr['text'] === 'string' && addr['text']) return addr['text'] as string;
    return Object.values(addr).filter(v => typeof v === 'string').join(', ');
  }

  protected stringToAddress(s: string): Record<string, unknown> | null {
    const t = s.trim();
    return t ? { text: t } : null;
  }

  protected languagesToString(langs: string[] | null | undefined): string {
    if (!langs || langs.length === 0) return '';
    return langs.join(', ');
  }

  protected stringToLanguages(s: string): string[] {
    const t = s.trim();
    if (!t) return [];
    return t.split(',').map(l => l.trim()).filter(l => l.length > 0);
  }

  protected addContactRow(arr: RegistrationContactEntry[]): void {
    arr.push({ fullName: '', relationshipToChild: null, address: null, telephone: null, email: null, workAddress: null, hasParentalResponsibility: null });
  }

  protected removeContactRow(arr: RegistrationContactEntry[], index: number): void {
    arr.splice(index, 1);
  }

  protected trackByIndex(index: number): number {
    return index;
  }

  /* Builder for contact patches */

  private buildContactPatch(contactType: string, entries: RegistrationContactEntry[]): Record<string, unknown> {
    return { [contactType]: entries };
  }

  protected buildDemographicsHomePatch(): Record<string, unknown> {
    const d = this.demoHomeDraft;
    if (!d) return {};
    return {
      demographics_home: {
        sex: this.toNullWhenEmpty(d.sex ?? ''),
        religion: this.toNullWhenEmpty(d.religion ?? ''),
        ethnic_origin: this.toNullWhenEmpty(d.ethnicOrigin ?? ''),
        first_language: this.toNullWhenEmpty(d.firstLanguage ?? ''),
        other_languages: d.otherLanguages && d.otherLanguages.length > 0 ? d.otherLanguages : null,
        home_address: this.stringToAddress(d.homeAddress ? this.addressToString(d.homeAddress) : ''),
        home_postcode: this.toNullWhenEmpty(d.homePostcode ?? ''),
        home_telephone: this.toNullWhenEmpty(d.homeTelephone ?? ''),
        disability_status: this.toNullWhenEmpty(d.disabilityStatus ?? ''),
        disability_notes: this.toNullWhenEmpty(d.disabilityNotes ?? ''),
        access_requirements: this.toNullWhenEmpty(d.accessRequirements ?? ''),
        demographics_home_reviewed: d.demographicsHomeReviewed,
      },
    };
  }

  protected buildMedicalDietaryPatch(): Record<string, unknown> {
    const d = this.medicalDietaryDraft;
    if (!d) return {};
    return {
      medical_dietary: {
        medical_conditions_status: this.toNullWhenEmpty(d.medicalConditionsStatus ?? ''),
        medical_conditions_notes: this.toNullWhenEmpty(d.medicalConditionsNotes ?? ''),
        prescribed_medication_status: this.toNullWhenEmpty(d.prescribedMedicationStatus ?? ''),
        medication_notes: this.toNullWhenEmpty(d.medicationNotes ?? ''),
        immunisation_status: this.toNullWhenEmpty(d.immunisationStatus ?? ''),
        immunisation_country: this.toNullWhenEmpty(d.immunisationCountry ?? ''),
        illness_diagnosis_history: this.toNullWhenEmpty(d.illnessDiagnosisHistory ?? ''),
        dietary_requirements_status: this.toNullWhenEmpty(d.dietaryRequirementsStatus ?? ''),
        dietary_requirements_notes: this.toNullWhenEmpty(d.dietaryRequirementsNotes ?? ''),
        dietary_side_effects: this.toNullWhenEmpty(d.dietarySideEffects ?? ''),
        medical_dietary_reviewed: d.medicalDietaryReviewed,
      },
    };
  }

  protected buildHealthContactsPatch(): Record<string, unknown> {
    const d = this.healthContactsDraft;
    if (!d) return {};
    return {
      health_contacts: {
        doctor_name: this.toNullWhenEmpty(d.doctorName ?? ''),
        doctor_address: this.toNullWhenEmpty(d.doctorAddress ?? ''),
        doctor_phone: this.toNullWhenEmpty(d.doctorPhone ?? ''),
        health_visitor_name: this.toNullWhenEmpty(d.healthVisitorName ?? ''),
        health_visitor_address: this.toNullWhenEmpty(d.healthVisitorAddress ?? ''),
        health_visitor_phone: this.toNullWhenEmpty(d.healthVisitorPhone ?? ''),
        health_contacts_reviewed: d.healthContactsReviewed,
      },
    };
  }

  protected buildSocialDevelopmentPatch(): Record<string, unknown> {
    const d = this.socialDevDraft;
    if (!d) return {};
    return {
      social_development: {
        social_services_status: this.toNullWhenEmpty(d.socialServicesStatus ?? ''),
        social_services_notes: this.toNullWhenEmpty(d.socialServicesNotes ?? ''),
        social_worker_contact_details: this.toNullWhenEmpty(d.socialWorkerContactDetails ?? ''),
        concern_walking: this.toNullWhenEmpty(d.concernWalking ?? ''),
        concern_speech_language: this.toNullWhenEmpty(d.concernSpeechLanguage ?? ''),
        concern_hearing: this.toNullWhenEmpty(d.concernHearing ?? ''),
        concern_sight: this.toNullWhenEmpty(d.concernSight ?? ''),
        concern_emotional_wellbeing: this.toNullWhenEmpty(d.concernEmotionalWellbeing ?? ''),
        concern_behaviour: this.toNullWhenEmpty(d.concernBehaviour ?? ''),
        professional_referrals: d.professionalReferrals || null,
        social_development_reviewed: d.socialDevelopmentReviewed,
      },
    };
  }

  protected buildCollectionPatch(): Record<string, unknown> {
    return {
      collection: {
        over18_collection_acknowledged: this.collectionOver18,
        emergency_collection_reviewed: this.collectionEmergencyReviewed,
      },
    };
  }

  protected buildFundingSupportPatch(): Record<string, unknown> {
    const d = this.fundingSupportDraft;
    if (!d) return {};
    return {
      funding_support: {
        benefits_contribute_to_fees: this.toNullWhenEmpty(d.benefitsContributeToFees ?? ''),
        working_tax_credit: this.toNullWhenEmpty(d.workingTaxCredit ?? ''),
        college_uni_paid_to_parent: this.toNullWhenEmpty(d.collegeUniPaidToParent ?? ''),
        college_uni_paid_to_nursery: this.toNullWhenEmpty(d.collegeUniPaidToNursery ?? ''),
        funding_3yo_term_time: this.toNullWhenEmpty(d.funding3yoTermTime ?? ''),
        funding_2yo_term_time: this.toNullWhenEmpty(d.funding2yoTermTime ?? ''),
        funding_support_notes: this.toNullWhenEmpty(d.fundingSupportNotes ?? ''),
        funding_support_reviewed: d.fundingSupportReviewed,
      },
    };
  }

  protected buildRoutineCarePatch(): Record<string, unknown> {
    const d = this.routineCareDraft;
    if (!d) return {};
    return {
      routine_care: {
        routine_care_notes: this.toNullWhenEmpty(d.routineCareNotes ?? ''),
        routine_care_reviewed: d.routineCareReviewed,
      },
    };
  }

  protected buildGdprPatch(): Record<string, unknown> {
    return {
      gdpr_declaration: {
        gdpr_declared_by_name: this.toNullWhenEmpty(this.gdprName),
        gdpr_declaration_date: this.toNullWhenEmpty(this.gdprDate),
      },
    };
  }

  protected buildOfficeChecklistPatch(): Partial<OfficeUseChecklist> {
    const d = this.officeDraft;
    if (!d) return {};
    return {
      depositStatus: this.toNullWhenEmpty(d.depositStatus ?? ''),
      depositPaidDate: this.toNullWhenEmpty(d.depositPaidDate ?? ''),
      applicationDateStatus: this.toNullWhenEmpty(d.applicationDateStatus ?? ''),
      applicationDate: this.toNullWhenEmpty(d.applicationDate ?? ''),
      startDateStatus: this.toNullWhenEmpty(d.startDateStatus ?? ''),
      dateLeft: this.toNullWhenEmpty(d.dateLeft ?? ''),
      sessionsDaysRequestedStatus: this.toNullWhenEmpty(d.sessionsDaysRequestedStatus ?? ''),
      sessionsDaysRequested: this.toNullWhenEmpty(d.sessionsDaysRequested ?? ''),
      termTimeOnlySpaceStatus: this.toNullWhenEmpty(d.termTimeOnlySpaceStatus ?? ''),
      contractStatus: this.toNullWhenEmpty(d.contractStatus ?? ''),
      contractDate: this.toNullWhenEmpty(d.contractDate ?? ''),
      handbookStatus: this.toNullWhenEmpty(d.handbookStatus ?? ''),
      handbookDate: this.toNullWhenEmpty(d.handbookDate ?? ''),
      redBookStatus: this.toNullWhenEmpty(d.redBookStatus ?? ''),
      redBookCheckedDate: this.toNullWhenEmpty(d.redBookCheckedDate ?? ''),
      birthCertificatePassportStatus: this.toNullWhenEmpty(d.birthCertificatePassportStatus ?? ''),
      birthCertificatePassportCheckedDate: this.toNullWhenEmpty(d.birthCertificatePassportCheckedDate ?? ''),
      proofOfAddressStatus: this.toNullWhenEmpty(d.proofOfAddressStatus ?? ''),
      proofOfAddressCheckedDate: this.toNullWhenEmpty(d.proofOfAddressCheckedDate ?? ''),
      notes: this.toNullWhenEmpty(d.notes ?? ''),
    };
  }

}
