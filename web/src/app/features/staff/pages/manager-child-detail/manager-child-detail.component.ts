import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { forkJoin, of, Subject } from 'rxjs';
import { catchError, takeUntil } from 'rxjs/operators';

import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroExclamationTriangle,
  heroCalendarDays,
  heroCheck,
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
  heroEyeSlash,
  heroShieldCheck,
  heroPlus,
  heroLockClosed,
  heroXCircle,
  heroUser,
  heroCamera,
  heroTrash
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

export type ChildProfileTab = 'overview' | 'attendance' | 'funding' | 'health' | 'contacts';

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
      heroCheck,
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
      heroEyeSlash,
      heroShieldCheck,
      heroPlus,
      heroLockClosed,
      heroXCircle,
      heroUser,
      heroCamera,
      heroTrash
    })
  ],
  templateUrl: './manager-child-detail.component.html',
})
export class ManagerChildDetailComponent implements OnInit, OnDestroy {
  private readonly staffApi = inject(StaffApiService);
  private readonly invoicesApi = inject(ManagerInvoicesApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly destroy$ = new Subject<void>();

  private readonly validTabs: readonly ChildProfileTab[] = ['overview', 'attendance', 'funding', 'health', 'contacts'];

  readonly formatRate = formatHourlyRateGbp;
  readonly formatSiteRate = formatSiteRate;
  readonly requirementLabel = missingRequirementLabel;

  activeTab: ChildProfileTab = 'overview';

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

  isUploadingPhoto = false;
  photoErrorMessage: string | null = null;

  billingMonth = '';
  monthlyProfile: FundingProfileRecord | null = null;
  monthlyAllowanceMinutes = 0;
  isSavingMonthly = false;
  monthlySaveMessage: string | null = null;

  get suggestedMinutes(): number | null {
    if (!this.funding?.funded_hours_per_week) return null;
    return Math.round(this.funding.funded_hours_per_week * 4.33 * 60);
  }

  showPickupPassword = false;

  togglePickupPassword(): void {
    this.showPickupPassword = !this.showPickupPassword;
  }

  get childAge(): string {
    if (!this.child?.dateOfBirth) return '';
    const birthDate = new Date(this.child.dateOfBirth);
    if (isNaN(birthDate.getTime())) return '';
    const today = new Date();
    let years = today.getFullYear() - birthDate.getFullYear();
    let months = today.getMonth() - birthDate.getMonth();
    let days = today.getDate() - birthDate.getDate();
    
    if (days < 0) {
      months--;
      const prevMonth = new Date(today.getFullYear(), today.getMonth(), 0);
      days += prevMonth.getDate();
    }
    if (months < 0) {
      years--;
      months += 12;
    }
    
    if (years === 0) {
      if (months === 0) {
        return `${days}d`;
      }
      return `${months}m`;
    }
    return `${years}y ${months}m`;
  }

  get weeklySchedule(): { dayName: string; shortName: string; isActive: boolean; sessions: string[] }[] {
    const days = [
      { dayOfWeek: 1, dayName: 'Monday', shortName: 'Mon' },
      { dayOfWeek: 2, dayName: 'Tuesday', shortName: 'Tue' },
      { dayOfWeek: 3, dayName: 'Wednesday', shortName: 'Wed' },
      { dayOfWeek: 4, dayName: 'Thursday', shortName: 'Thu' },
      { dayOfWeek: 5, dayName: 'Friday', shortName: 'Fri' }
    ];
    
    return days.map(day => {
      const matchingEntries = this.currentBookingPattern?.entries.filter(e => e.day_of_week === day.dayOfWeek) ?? [];
      const sessions = matchingEntries.map(e => {
        if (e.session_type) {
          return `${e.session_type.name} (${e.session_type.start_time.slice(0, 5)} - ${e.session_type.end_time.slice(0, 5)})`;
        }
        return 'Scheduled';
      });
      return {
        dayName: day.dayName,
        shortName: day.shortName,
        isActive: matchingEntries.length > 0,
        sessions
      };
    });
  }

  ngOnInit(): void {
    this.childId = this.route.snapshot.paramMap.get('childId') ?? '';
    this.billingMonth = this.route.snapshot.queryParamMap.get('billing_month') ?? '';
    this.activeTab = this.resolveTab(this.route.snapshot.paramMap.get('tab'));
    this.load();

    this.route.paramMap.pipe(takeUntil(this.destroy$)).subscribe(params => {
      this.activeTab = this.resolveTab(params.get('tab'));
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  selectTab(tab: ChildProfileTab): void {
    const commands = tab === 'overview'
      ? ['/manager/children', this.childId]
      : ['/manager/children', this.childId, tab];
    this.router.navigate(commands, { queryParamsHandling: 'preserve' });
  }

  private resolveTab(tabParam: string | null): ChildProfileTab {
    if (tabParam && (this.validTabs as readonly string[]).includes(tabParam)) {
      return tabParam as ChildProfileTab;
    }
    return 'overview';
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

  onPhotoSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;

    const file = input.files[0];
    const maxSize = 5 * 1024 * 1024;
    if (file.size > maxSize) {
      this.photoErrorMessage = 'File exceeds maximum size of 5 MB.';
      input.value = '';
      return;
    }

    if (!['image/jpeg', 'image/png'].includes(file.type)) {
      this.photoErrorMessage = 'Invalid file type. Only JPEG and PNG are accepted.';
      input.value = '';
      return;
    }

    this.isUploadingPhoto = true;
    this.photoErrorMessage = null;

    this.staffApi.uploadPhoto(this.childId, file).subscribe({
      next: (result) => {
        if (this.child) {
          this.child = { ...this.child, photoUrl: result.photo_url };
        }
        this.isUploadingPhoto = false;
      },
      error: (err) => {
        this.photoErrorMessage = err?.message ?? 'Failed to upload photo.';
        this.isUploadingPhoto = false;
      },
    });

    input.value = '';
  }

  removePhoto(): void {
    this.isUploadingPhoto = true;
    this.photoErrorMessage = null;

    this.staffApi.removePhoto(this.childId).subscribe({
      next: () => {
        if (this.child) {
          this.child = { ...this.child, photoUrl: null };
        }
        this.isUploadingPhoto = false;
      },
      error: (err) => {
        this.photoErrorMessage = err?.message ?? 'Failed to remove photo.';
        this.isUploadingPhoto = false;
      },
    });
  }

  onPhotoError(event: Event): void {
    const img = event.target as HTMLImageElement;
    img.style.display = 'none';
    const parent = img.parentElement;
    if (parent) {
      const fallback = parent.querySelector('.photo-fallback') as HTMLElement;
      if (fallback) fallback.style.display = 'flex';
    }
  }
}
