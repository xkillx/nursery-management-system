import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import { HttpErrorResponse } from '@angular/common/http';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ChildFormComponent } from '../../components/child-form/child-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord, ChildWritePayload, StatusFilter } from '../../models/children.models';
import { FundingProfileRecord } from '../../models/funding.models';
import { ChildGuardianLinkRecord, GuardianChildLinkWritePayload, GuardianRecord } from '../../models/guardians.models';
import { RegistrationProfileCompleteness, OfficeUseCompleteness } from '../../models/registration-profile.models';
import { formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { formatCompletionStatus, getCompletionBadgeClass } from '../../utils/registration-profile-formatters';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
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
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-child-detail.component.html',
})
export class ManagerChildDetailComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);

  readonly formatRate = formatHourlyRateGbp;
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
  officeCompleteness: OfficeUseCompleteness | null = null;
  isLoadingRegistration = false;
  registrationLoadError: string | null = null;

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

  get fundingNotSet(): boolean {
    return !this.isLoadingFunding && this.fundingProfile === null && !this.fundingErrorMessage;
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
    this.loadChild();
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

    this.staffApi.getRegistrationProfile(this.childId).subscribe({
      next: (profile) => {
        this.profileCompleteness = profile.completeness;
        this.isLoadingRegistration = false;
        this.loadOfficeChecklistSummary();
      },
      error: (err) => {
        this.isLoadingRegistration = false;
        this.registrationLoadError = 'Could not load registration summary.';
        this.profileCompleteness = null;
      },
    });
  }

  private loadOfficeChecklistSummary(): void {
    if (!this.childId) return;

    this.staffApi.getRegistrationOfficeUseChecklist(this.childId).subscribe({
      next: (checklist) => {
        this.officeCompleteness = checklist.completeness;
      },
      error: () => {
        this.officeCompleteness = null;
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

}
