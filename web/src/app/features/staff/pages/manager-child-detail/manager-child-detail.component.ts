import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroExclamationTriangle } from '@ng-icons/heroicons/outline';

import { StaffApiService } from '../../data/staff-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ChildRecord } from '../../models/children.models';
import { ChildProfile, ChildFundingRecord } from '../../models/child-profile.models';
import { FundingProfileRecord } from '../../models/funding.models';
import { formatSiteRate, formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ChildContact } from '../../models/child-profile.models';

@Component({
  selector: 'app-manager-child-detail',
  imports: [
    CommonModule,
    RouterLink,
    FormsModule,
    NgIcon,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  providers: [provideIcons({ heroExclamationTriangle })],
  templateUrl: './manager-child-detail.component.html',
})
export class ManagerChildDetailComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  readonly formatRate = formatHourlyRateGbp;
  readonly formatSiteRate = formatSiteRate;
  readonly requirementLabel = missingRequirementLabel;

  childId = '';
  child: ChildRecord | null = null;
  profile: ChildProfile | null = null;
  parentCarers: ChildContact[] = [];
  funding: ChildFundingRecord | null = null;

  isLoading = false;
  errorMessage: string | null = null;

  billingMonth = '';
  monthlyProfile: FundingProfileRecord | null = null;
  monthlyAllowanceMinutes = 0;
  isSavingMonthly = false;
  monthlySaveMessage: string | null = null;

  get suggestedMinutes(): number | null {
    if (!this.funding?.funded_hours_per_week) return null;
    return Math.round(this.funding.funded_hours_per_week * 4.33 * 60);
  }

  ngOnInit(): void {
    this.childId = this.route.snapshot.paramMap.get('childId') ?? '';
    this.billingMonth = this.route.snapshot.queryParamMap.get('billing_month') ?? '';
    this.load();
  }

  load(): void {
    if (!this.childId) {
      this.errorMessage = 'Missing child id.';
      return;
    }
    this.isLoading = true;
    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.loadParentCarers();
        this.loadProfile();
        this.loadFunding();
      },
      error: (err) => {
        this.errorMessage = err?.message ?? 'Failed to load child.';
        this.isLoading = false;
      },
    });
  }

  private loadParentCarers(): void {
    this.staffApi.getChildContacts(this.childId).subscribe({
      next: (contacts) => {
        this.parentCarers = contacts.parentCarers;
        this.isLoading = false;
      },
      error: () => {
        this.parentCarers = [];
        this.isLoading = false;
      },
    });
  }

  private loadProfile(): void {
    this.staffApi.getChildProfile(this.childId).subscribe({
      next: (profile) => {
        this.profile = profile;
      },
      error: () => {
        this.profile = null;
      },
    });
  }

  private loadFunding(): void {
    this.staffApi.getChildFunding(this.childId).subscribe({
      next: (funding) => {
        this.funding = funding;
        if (this.billingMonth) {
          this.loadMonthlyProfile();
        }
      },
      error: () => {
        this.funding = null;
      },
    });
  }

  private loadMonthlyProfile(): void {
    this.staffApi.getFundingProfile(this.childId, this.billingMonth).subscribe({
      next: (profile) => {
        this.monthlyProfile = profile;
        this.monthlyAllowanceMinutes = profile?.fundedAllowanceMinutes ?? 0;
      },
      error: () => {
        this.monthlyProfile = null;
        this.monthlyAllowanceMinutes = 0;
      },
    });
  }

  useSuggestedValue(): void {
    if (this.suggestedMinutes !== null) {
      this.monthlyAllowanceMinutes = this.suggestedMinutes;
    }
  }

  saveMonthlyProfile(): void {
    if (!this.billingMonth) return;
    this.isSavingMonthly = true;
    this.monthlySaveMessage = null;
    this.staffApi.upsertFundingProfile(this.childId, {
      billing_month: this.billingMonth,
      funded_allowance_minutes: this.monthlyAllowanceMinutes,
    }).subscribe({
      next: () => {
        this.isSavingMonthly = false;
        this.monthlySaveMessage = 'Monthly funding profile saved.';
      },
      error: (err) => {
        this.isSavingMonthly = false;
        this.errorMessage = err?.message ?? 'Failed to save monthly funding profile.';
      },
    });
  }
}
