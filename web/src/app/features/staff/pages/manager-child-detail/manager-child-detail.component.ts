import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroCalendarDays,
  heroCheck,
  heroClipboardDocumentCheck,
  heroDocumentText,
  heroEllipsisHorizontal,
  heroEnvelope,
  heroExclamationTriangle,
  heroHome,
  heroIdentification,
  heroKey,
  heroPencilSquare,
  heroPhone,
  heroShieldCheck,
  heroUserGroup,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { HttpErrorResponse } from '@angular/common/http';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ChildFormComponent } from '../../components/child-form/child-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ChildRecord, ChildWritePayload, StatusFilter } from '../../models/children.models';
import { FundingProfileRecord } from '../../models/funding.models';
import { ChildGuardianLinkRecord, GuardianChildLinkWritePayload, GuardianRecord } from '../../models/guardians.models';
import {
  ConsentRecord,
  ConsentWithCompletenessResponse,
  RegistrationContactEntry,
  RegistrationProfileCompleteness,
  RegistrationProfileResponse,
  RegistrationWorkflowStatus,
} from '../../models/registration-profile.models';
import { formatSiteRate, formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { formatCompletionStatus, getCompletionBadgeClass } from '../../utils/registration-profile-formatters';
import { mockChildDetailProfile } from './manager-child-detail.mock';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

@Component({
  selector: 'app-manager-child-detail',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    ChildFormComponent,
    NgIcon,
    SelectComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  providers: [
    provideIcons({
      heroArrowLeft,
      heroCalendarDays,
      heroCheck,
      heroClipboardDocumentCheck,
      heroDocumentText,
      heroEllipsisHorizontal,
      heroEnvelope,
      heroExclamationTriangle,
      heroHome,
      heroIdentification,
      heroKey,
      heroPencilSquare,
      heroPhone,
      heroShieldCheck,
      heroUserGroup,
      heroXMark,
    }),
  ],
  templateUrl: './manager-child-detail.component.html',
})
export class ManagerChildDetailComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly auth = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);

  readonly formatRate = formatHourlyRateGbp;
  readonly formatSiteRate = formatSiteRate;
  readonly requirementLabel = missingRequirementLabel;
  readonly formatCompletionStatus = formatCompletionStatus;
  readonly getCompletionBadgeClass = getCompletionBadgeClass;

  childId = '';
  child: ChildRecord | null = null;
  linkedGuardians: ChildGuardianLinkRecord[] = [];
  allGuardians: GuardianRecord[] = [];

  isLoadingChild = false;
  isLoadingLinks = false;
  isSaving = false;
  isLinking = false;
  profileCompleteness: RegistrationProfileCompleteness | null = null;
  workflowStatus: RegistrationWorkflowStatus | null = null;
  registrationProfile: RegistrationProfileResponse | null = null;
  consentSummary: ConsentWithCompletenessResponse | null = null;
  isLoadingRegistration = false;
  registrationLoadError: string | null = null;
  isLoadingProfile = false;
  profileLoadError: string | null = null;
  isLoadingConsents = false;
  consentsLoadError: string | null = null;

  showEditForm = false;
  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};

  selectedGuardianId = '';

  selectedBillingMonth = '';
  fundingProfile: FundingProfileRecord | null = null;
  isLoadingFunding = false;
  isSavingFunding = false;
  fundedHoursInput = '';
  fundedMinutesInput = '';
  fundingStatusMessage: string | null = null;
  fundingErrorMessage: string | null = null;
  fundingFieldErrors: Record<string, string> = {};

  // Temporary design-only fields. Replace with API-backed values before release.
  readonly mockProfile = mockChildDetailProfile;

  ngOnInit(): void {
    const queryMonth = this.route.snapshot.queryParamMap.get('billing_month');
    if (queryMonth && /^\d{4}-\d{2}$/.test(queryMonth)) {
      this.selectedBillingMonth = queryMonth;
    } else {
      this.selectedBillingMonth = this.formatCurrentMonth();
    }
    this.loadAll();
  }

  onEditChild(): void {
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showEditForm = true;
  }

  closeEditForm(): void {
    this.showEditForm = false;
    this.fieldErrors = {};
    this.errorMessage = null;
  }

  saveChild(payload: ChildWritePayload): void {
    this.isSaving = true;
    this.fieldErrors = {};
    this.errorMessage = null;

    this.staffApi.updateChild(this.childId, payload).subscribe({
      next: () => {
        this.isSaving = false;
        this.showEditForm = false;
        this.loadChild();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  linkGuardian(): void {
    if (!this.selectedGuardianId) return;

    this.isLinking = true;
    this.errorMessage = null;

    const payload: GuardianChildLinkWritePayload = {
      guardian_id: this.selectedGuardianId,
      child_id: this.childId,
    };

    this.staffApi.createGuardianChildLink(payload).subscribe({
      next: () => {
        this.isLinking = false;
        this.selectedGuardianId = '';
        this.loadChild();
        this.loadLinkedGuardians();
      },
      error: (error) => {
        this.isLinking = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  onBillingMonthChange(): void {
    this.clearFundingStatus();
    this.loadFundingProfile();
  }

  saveFundingAllowance(): void {
    this.clearFundingStatus();
    const validation = this.validateFundingInputs();
    if (validation) {
      this.fundingErrorMessage = validation;
      return;
    }

    const totalMinutes = this.totalFundingMinutesFromInputs();
    this.isSavingFunding = true;

    this.staffApi.upsertFundingProfile(this.childId, {
      billing_month: this.selectedBillingMonth,
      funded_allowance_minutes: totalMinutes,
    }).subscribe({
      next: (profile) => {
        this.fundingProfile = profile;
        this.isSavingFunding = false;
        this.populateInputsFromMinutes(profile.fundedAllowanceMinutes);
        this.fundingStatusMessage = 'Saved';
      },
      error: (error) => {
        this.isSavingFunding = false;
        this.handleFundingError(error);
      },
    });
  }

  get availableGuardians(): GuardianRecord[] {
    const linkedIds = new Set(this.linkedGuardians.map(l => l.guardianId));
    return this.allGuardians.filter(g => !linkedIds.has(g.id));
  }

  get guardianOptions(): Option[] {
    return [
      { value: '', label: 'Select guardian...' },
      ...this.availableGuardians.map(g => ({ value: g.id, label: g.fullName })),
    ];
  }

  roomOptions: Option[] = [];

  get fundingNotSet(): boolean {
    return !this.isLoadingFunding && this.fundingProfile === null && !this.fundingErrorMessage;
  }

  get initials(): string {
    const name = this.child?.fullName ?? '';
    const initials = name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join('');
    return initials || 'CH';
  }

  get childAgeLabel(): string {
    if (!this.child?.dateOfBirth) return 'Age not recorded';
    const dob = new Date(`${this.child.dateOfBirth}T00:00:00`);
    if (Number.isNaN(dob.getTime())) return 'Age not recorded';
    const today = new Date();
    let months = (today.getFullYear() - dob.getFullYear()) * 12 + today.getMonth() - dob.getMonth();
    if (today.getDate() < dob.getDate()) months -= 1;
    if (months < 0) return 'Age not recorded';
    const years = Math.floor(months / 12);
    const remainingMonths = months % 12;
    if (years === 0) return `${remainingMonths}m`;
    if (remainingMonths === 0) return `${years}y`;
    return `${years}y ${remainingMonths}m`;
  }

  get registrationEditLink(): string[] {
    return ['/staff/manager/children', this.childId, 'registration'];
  }

  get intakeLink(): string[] {
    return ['/staff/manager/registrations', this.childId, 'intake'];
  }

  get medicalDietaryAlerts(): string[] {
    const alerts: string[] = [];
    const medical: any = this.registrationProfile?.['medicalDietary'];
    if (medical) {
      if (this.isAffirmative(medical['medicalConditionsStatus']) && medical['medicalConditionsNotes']) {
        alerts.push(`Medical: ${medical['medicalConditionsNotes']}`);
      }
      if (this.isAffirmative(medical['prescribedMedicationStatus']) && medical['medicationNotes']) {
        alerts.push(`Medication: ${medical['medicationNotes']}`);
      }
      if (this.isAffirmative(medical['dietaryRequirementsStatus']) && medical['dietaryRequirementsNotes']) {
        alerts.push(`Dietary: ${medical['dietaryRequirementsNotes']}`);
      }
      if (medical['dietarySideEffects']) {
        alerts.push(`Dietary side effects: ${medical['dietarySideEffects']}`);
      }
    }
    return alerts.length > 0 ? alerts : this.mockProfile.alertChips;
  }

  get primaryContacts(): RegistrationContactEntry[] {
    return this.registrationProfile?.parentCarers?.length
      ? this.registrationProfile.parentCarers
      : this.linkedGuardians.map((link) => ({
          fullName: link.guardian.fullName,
          relationshipToChild: null,
          address: null,
          telephone: link.guardian.phone,
          email: link.guardian.email,
          workAddress: null,
          hasParentalResponsibility: null,
        }));
  }

  get emergencyContacts(): RegistrationContactEntry[] {
    return this.registrationProfile?.emergencyContacts ?? [];
  }

  get authorisedCollectors(): RegistrationContactEntry[] {
    return this.registrationProfile?.authorisedCollectors ?? [];
  }

  get currentConsent(): ConsentRecord | null {
    return this.consentSummary?.current ?? this.workflowStatus?.current_consent_record ?? null;
  }

  get keyConsentRows(): { label: string; granted: boolean | null }[] {
    const consent = this.currentConsent;
    return [
      { label: 'Safeguarding acknowledgement', granted: consent?.safeguarding_reporting_acknowledgement ?? null },
      { label: 'Urgent medical treatment', granted: consent?.urgent_medical_treatment ?? null },
      { label: 'Local outings', granted: consent?.local_outings ?? null },
      { label: 'Development photos', granted: consent?.development_profile_photos ?? null },
      { label: 'Social media', granted: consent?.social_media ?? null },
    ];
  }

  get collectionPasswordStatus(): string {
    return this.registrationProfile?.collection?.isSet ? 'Set' : 'Not set';
  }

  get sexLabel(): string {
    const sex = this.registrationProfile?.demographicsHome?.sex;
    if (!sex) return '-';
    return sex.charAt(0).toUpperCase() + sex.slice(1);
  }

  get fundingAllowanceLabel(): string {
    if (this.isLoadingFunding) return 'Loading';
    if (!this.fundingProfile) return 'Not set';
    const totalMinutes = this.fundingProfile.fundedAllowanceMinutes;
    const hours = Math.floor(totalMinutes / 60);
    const minutes = totalMinutes % 60;
    if (minutes === 0) return `${hours}h`;
    return `${hours}h ${minutes}m`;
  }

  get safeguardingLevel(): string {
    if (this.medicalDietaryAlerts.length > 0) return 'Enhanced';
    if (this.currentConsent === null) return 'Pending review';
    return 'Standard';
  }

  contactInitials(contact: RegistrationContactEntry): string {
    return contact.fullName
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase())
      .join('') || 'GC';
  }

  consentLabel(granted: boolean | null): string {
    if (granted === true) return 'Granted';
    if (granted === false) return 'Declined';
    return 'Not recorded';
  }

  booleanLabel(value: boolean | null | undefined): string {
    if (value === true) return 'Yes';
    if (value === false) return 'No';
    return 'Not recorded';
  }

  valueOrDash(value: string | number | null | undefined): string {
    if (value === null || value === undefined || value === '') return '-';
    return String(value);
  }

  formatAddress(address: Record<string, unknown> | null | undefined): string {
    if (!address) return '-';
    const values = Object.values(address)
      .filter((value): value is string | number => typeof value === 'string' || typeof value === 'number')
      .map((value) => String(value).trim())
      .filter(Boolean);
    return values.length ? values.join(', ') : '-';
  }

  consentBadgeClass(granted: boolean | null): string {
    if (granted === true) return 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-400';
    if (granted === false) return 'bg-error-50 text-error-700 dark:bg-error-500/15 dark:text-error-400';
    return 'bg-gray-100 text-gray-600 dark:bg-white/[0.05] dark:text-gray-400';
  }

  private clearFundingStatus(): void {
    this.fundingStatusMessage = null;
    this.fundingErrorMessage = null;
    this.fundingFieldErrors = {};
  }

  private validateFundingInputs(): string | null {
    const hours = this.fundedHoursInput.trim();
    const minutes = this.fundedMinutesInput.trim();

    if (hours === '' && minutes === '') {
      return 'Enter an allowance or enter 0 to save no funded hours.';
    }

    if (hours !== '' && (!Number.isInteger(Number(hours)) || Number(hours) < 0)) {
      return 'Hours must be a non-negative whole number.';
    }

    if (minutes !== '' && (!Number.isInteger(Number(minutes)) || Number(minutes) < 0 || Number(minutes) > 59)) {
      return 'Minutes must be a whole number between 0 and 59.';
    }

    if (!/^\d{4}-\d{2}$/.test(this.selectedBillingMonth)) {
      return 'Select a valid billing month.';
    }

    const totalMinutes = this.totalFundingMinutesFromInputs();
    if (totalMinutes > 44640) {
      return 'Total allowance cannot exceed 744 hours (44640 minutes).';
    }

    return null;
  }

  private totalFundingMinutesFromInputs(): number {
    const hours = Number(this.fundedHoursInput.trim() || '0');
    const minutes = Number(this.fundedMinutesInput.trim() || '0');
    return hours * 60 + minutes;
  }

  private handleFundingError(error: unknown): void {
    const mapped = this.errorMapper.mapAndHandle(error);
    if (mapped.code === 'funding_month_outside_enrollment_window') {
      this.fundingErrorMessage = 'This billing month does not overlap the child\'s enrollment window. Choose a month within the child\'s start and end dates.';
      return;
    }
    if (mapped.fieldErrors['funded_allowance_minutes']) {
      this.fundingFieldErrors = { funded_allowance_minutes: mapped.fieldErrors['funded_allowance_minutes'] };
      this.fundingErrorMessage = mapped.fieldErrors['funded_allowance_minutes'];
      return;
    }
    if (mapped.fieldErrors['billing_month']) {
      this.fundingFieldErrors = { billing_month: mapped.fieldErrors['billing_month'] };
      this.fundingErrorMessage = mapped.fieldErrors['billing_month'];
      return;
    }
    this.fundingErrorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
  }

  private populateInputsFromMinutes(totalMinutes: number): void {
    this.fundedHoursInput = String(Math.floor(totalMinutes / 60));
    this.fundedMinutesInput = String(totalMinutes % 60);
  }

  private loadAll(): void {
    this.childId = this.route.snapshot.paramMap.get('childId') ?? '';
    if (!this.childId) return;
    this.loadRoomOptions();
    this.loadChild();
  }

  private loadRoomOptions(): void {
    const branchId = this.auth.activeMembership()?.branch_id;
    if (!branchId) {
      this.roomOptions = [];
      return;
    }
    this.roomsApi
      .listRooms(branchId, { includeArchived: false })
      .subscribe({
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

  private loadChild(): void {
    this.isLoadingChild = true;
    this.errorMessage = null;

    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.isLoadingChild = false;
        this.loadLinkedGuardians();
        this.loadAllGuardians();
        this.loadFundingProfile();
        this.loadRegistrationSummary();
        this.loadRegistrationProfile();
        this.loadRegistrationConsents();
      },
      error: (error) => {
        this.isLoadingChild = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  private loadFundingProfile(): void {
    this.isLoadingFunding = true;
    this.fundingProfile = null;
    this.clearFundingStatus();

    this.staffApi.getFundingProfile(this.childId, this.selectedBillingMonth).subscribe({
      next: (profile) => {
        this.fundingProfile = profile;
        this.isLoadingFunding = false;
        this.populateInputsFromMinutes(profile.fundedAllowanceMinutes);
      },
      error: (error) => {
        this.isLoadingFunding = false;
        if (error instanceof HttpErrorResponse && error.status === 404) {
          this.fundingProfile = null;
          this.fundedHoursInput = '';
          this.fundedMinutesInput = '';
          return;
        }
        this.handleFundingError(error);
      },
    });
  }

  private loadRegistrationSummary(): void {
    if (!this.childId) return;
    this.isLoadingRegistration = true;
    this.registrationLoadError = null;

    this.staffApi.getRegistrationWorkflowStatus(this.childId).subscribe({
      next: (status) => {
        this.workflowStatus = status;
        this.profileCompleteness = status.profile_completeness as unknown as RegistrationProfileCompleteness;
        this.isLoadingRegistration = false;
      },
      error: (err) => {
        this.isLoadingRegistration = false;
        this.registrationLoadError = 'Could not load registration summary.';
        this.profileCompleteness = null;
        this.workflowStatus = null;
      },
    });
  }

  private loadRegistrationProfile(): void {
    if (!this.childId) return;
    this.isLoadingProfile = true;
    this.profileLoadError = null;

    this.staffApi.getRegistrationProfile(this.childId).subscribe({
      next: (profile) => {
        this.registrationProfile = profile;
        this.isLoadingProfile = false;
      },
      error: () => {
        this.registrationProfile = null;
        this.profileLoadError = 'Could not load registration profile.';
        this.isLoadingProfile = false;
      },
    });
  }

  private loadRegistrationConsents(): void {
    if (!this.childId) return;
    this.isLoadingConsents = true;
    this.consentsLoadError = null;

    this.staffApi.getRegistrationConsents(this.childId).subscribe({
      next: (summary) => {
        this.consentSummary = summary;
        this.isLoadingConsents = false;
      },
      error: () => {
        this.consentSummary = null;
        this.consentsLoadError = 'Could not load consent record.';
        this.isLoadingConsents = false;
      },
    });
  }

  private loadLinkedGuardians(): void {
    this.isLoadingLinks = true;
    this.staffApi.listChildGuardianLinks(this.childId).subscribe({
      next: (links) => {
        this.linkedGuardians = links;
        this.isLoadingLinks = false;
      },
      error: (error) => {
        this.isLoadingLinks = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  private loadAllGuardians(): void {
    this.staffApi.listGuardians({ status: 'active' as StatusFilter, limit: 200, offset: 0 }).subscribe({
      next: (guardians) => {
        this.allGuardians = guardians;
      },
    });
  }

  private formatCurrentMonth(): string {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    return `${year}-${month}`;
  }

  private isAffirmative(value: string | null | undefined): boolean {
    return ['yes', 'details'].includes((value ?? '').toLowerCase());
  }

}
