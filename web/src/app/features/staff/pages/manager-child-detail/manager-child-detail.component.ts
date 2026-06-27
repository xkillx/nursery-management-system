import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroExclamationTriangle,
  heroCalendarDays,
  heroIdentification,
  heroPencilSquare,
  heroCheckCircle,
  heroAcademicCap,
  heroHome,
  heroUserGroup,
  heroPhone,
  heroEnvelope,
  heroHeart,
  heroClipboardDocumentCheck,
  heroClipboardDocumentList,
  heroDocumentText,
  heroClock,
  heroEye,
  heroShieldCheck,
  heroPlus,
  heroLockClosed,
  heroXCircle,
  heroUser
} from '@ng-icons/heroicons/outline';
import { StaffApiService } from '../../data/staff-api.service';
import { ManagerInvoicesApiService } from '../../data/manager-invoices-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ChildRecord } from '../../models/children.models';
import {
  ChildProfile,
  ChildFundingRecord,
  ChildContact,
  ChildHealthProfile,
  ChildConsent,
  ChildSafeguardingProfile,
  ChildCollectionSettings,
  ChildRoomAssignment,
  ChildBillingProfile
} from '../../models/child-profile.models';
import { BookingPattern } from '../../models/booking-pattern.models';
import { ManagerInvoiceListItem } from '../../models/manager-invoices.models';
import { FundingProfileRecord } from '../../models/funding.models';
import { formatSiteRate, formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

@Component({
  selector: 'app-manager-child-detail',
  imports: [
    CommonModule,
    RouterLink,
    FormsModule,
    NgIcon,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  providers: [
    provideIcons({
      heroArrowLeft,
      heroExclamationTriangle,
      heroCalendarDays,
      heroIdentification,
      heroPencilSquare,
      heroCheckCircle,
      heroAcademicCap,
      heroHome,
      heroUserGroup,
      heroPhone,
      heroEnvelope,
      heroHeart,
      heroClipboardDocumentCheck,
      heroClipboardDocumentList,
      heroDocumentText,
      heroClock,
      heroEye,
      heroShieldCheck,
      heroPlus,
      heroLockClosed,
      heroXCircle,
      heroUser
    })
  ],
  templateUrl: './manager-child-detail.component.html',
})
export class ManagerChildDetailComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly invoicesApi = inject(ManagerInvoicesApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  readonly formatRate = formatHourlyRateGbp;
  readonly formatSiteRate = formatSiteRate;
  readonly requirementLabel = missingRequirementLabel;

  activeTab: 'overview' | 'attendance' | 'funding' | 'health' | 'contacts' = 'overview';

  childId = '';
  child: ChildRecord | null = null;
  profile: ChildProfile | null = null;
  parentCarers: ChildContact[] = [];
  emergencyContacts: ChildContact[] = [];
  authorisedCollectors: ChildContact[] = [];
  funding: ChildFundingRecord | null = null;
  healthProfile: ChildHealthProfile | null = null;
  consent: ChildConsent | null = null;
  safeguardingProfile: ChildSafeguardingProfile | null = null;
  collectionSettings: ChildCollectionSettings | null = null;
  roomAssignments: ChildRoomAssignment[] = [];
  billingProfile: ChildBillingProfile | null = null;
  bookingPatterns: BookingPattern[] = [];
  currentBookingPattern: BookingPattern | null = null;
  invoices: ManagerInvoiceListItem[] = [];

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

  selectTab(tab: 'overview' | 'attendance' | 'funding' | 'health' | 'contacts'): void {
    this.activeTab = tab;
  }

  load(): void {
    if (!this.childId) {
      this.errorMessage = 'Missing child id.';
      return;
    }
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.loadSubRecords();
      },
      error: (err) => {
        this.errorMessage = err?.message ?? 'Failed to load child.';
        this.isLoading = false;
      },
    });
  }

  private loadSubRecords(): void {
    const childId = this.childId;
    forkJoin({
      profile: this.staffApi.getChildProfile(childId).pipe(catchError(() => of(null))),
      contacts: this.staffApi.getChildContacts(childId).pipe(catchError(() => of({ parentCarers: [], emergencyContacts: [], authorisedCollectors: [] }))),
      funding: this.staffApi.getChildFunding(childId).pipe(catchError(() => of(null))),
      health: this.staffApi.getChildHealth(childId).pipe(catchError(() => of(null))),
      consent: this.staffApi.getChildConsent(childId).pipe(catchError(() => of(null))),
      safeguarding: this.staffApi.getChildSafeguarding(childId).pipe(catchError(() => of(null))),
      collectionSettings: this.staffApi.getChildCollectionSettings(childId).pipe(catchError(() => of(null))),
      roomAssignments: this.staffApi.listChildRoomAssignments(childId).pipe(catchError(() => of([]))),
      billingProfile: this.staffApi.getChildBillingProfile(childId).pipe(catchError(() => of(null))),
      bookingPatterns: this.staffApi.listChildBookingPatterns(childId).pipe(catchError(() => of([]))),
      currentBookingPattern: this.staffApi.getCurrentChildBookingPattern(childId).pipe(catchError(() => of(null))),
      invoices: this.invoicesApi.listInvoices({ childId, status: 'all', limit: 50, offset: 0 }).pipe(catchError(() => of({ items: [], total: 0 }))),
    }).subscribe({
      next: (res) => {
        this.profile = res.profile;
        this.parentCarers = res.contacts.parentCarers;
        this.emergencyContacts = res.contacts.emergencyContacts;
        this.authorisedCollectors = res.contacts.authorisedCollectors;
        this.funding = res.funding;
        this.healthProfile = res.health;
        this.consent = res.consent;
        this.safeguardingProfile = res.safeguarding;
        this.collectionSettings = res.collectionSettings;
        this.roomAssignments = res.roomAssignments;
        this.billingProfile = res.billingProfile;
        this.bookingPatterns = res.bookingPatterns;
        this.currentBookingPattern = res.currentBookingPattern;
        this.invoices = res.invoices.items;

        if (this.billingMonth) {
          this.loadMonthlyProfile();
        } else {
          this.isLoading = false;
        }
      },
      error: (err) => {
        this.errorMessage = err?.message ?? 'Failed to load child details.';
        this.isLoading = false;
      },
    });
  }

  private loadMonthlyProfile(): void {
    this.staffApi.getFundingProfile(this.childId, this.billingMonth).subscribe({
      next: (profile) => {
        this.monthlyProfile = profile;
        this.monthlyAllowanceMinutes = profile?.fundedAllowanceMinutes ?? 0;
        this.isLoading = false;
      },
      error: () => {
        this.monthlyProfile = null;
        this.monthlyAllowanceMinutes = 0;
        this.isLoading = false;
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
