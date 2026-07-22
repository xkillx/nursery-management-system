import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { forkJoin, of, Subject, Subscription } from 'rxjs';
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
import { BookingsApiService } from '../../data/bookings-api.service';
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
import { UnifiedBooking, SessionEntry } from '../../models/booking.models';
import { StaffSessionType, StaffSessionTypesApiService } from '../../data/session-types-api.service';
import { ManagerInvoiceListItem } from '../../models/manager-invoices.models';
import { FundingRecordDetail, FundingRecordWritePayload } from '../../models/funding.models';
import { formatSiteRate, formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ChildPhotoService } from '../../../../shared/services/child-photo.service';

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
  private readonly bookingsApi = inject(BookingsApiService);
  private readonly invoicesApi = inject(ManagerInvoicesApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly photoService = inject(ChildPhotoService);
  private readonly destroy$ = new Subject<void>();
  private photoSub: Subscription | null = null;

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
  currentBooking: UnifiedBooking | null = null;
  currentBookingEntries: SessionEntry[] | null = null;
  sessionTypes: StaffSessionType[] = [];
  invoices: ManagerInvoiceListItem[] = [];

  isLoading = false;
  errorMessage: string | null = null;

  isUploadingPhoto = false;
  photoErrorMessage: string | null = null;
  resolvedPhotoUrl: string | null = null;

  billingMonth = '';
  fundingDetail: FundingRecordDetail | null = null;
  isSavingFunding = false;
  fundingSaveMessage: string | null = null;

  get fundedAllowanceMinutes(): number {
    return this.fundingDetail?.fundedAllowanceMinutes ?? 0;
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
      const matchingEntries = this.currentBookingEntries?.filter(e => e.day_of_week === day.dayOfWeek) ?? [];
      const sessions = matchingEntries.map(e => {
        const st = this.sessionTypes.find(t => t.id === e.session_type_id);
        if (st) {
          return `${st.name} (${st.startTime.slice(0, 5)} - ${st.endTime.slice(0, 5)})`;
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
    this.photoSub?.unsubscribe();
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
        this.resolvePhotoUrl();
        this.loadSubRecords();
      },
      error: (err) => {
        this.errorMessage = err?.message ?? 'Failed to load child.';
        this.isLoading = false;
      },
    });
  }

  private resolvePhotoUrl(): void {
    this.photoSub?.unsubscribe();
    this.resolvedPhotoUrl = null;
    const url = this.child?.photoUrl;
    if (!url) return;
    this.photoSub = this.photoService.getPhotoUrl(url).subscribe({
      next: (blobUrl) => (this.resolvedPhotoUrl = blobUrl),
      error: () => (this.resolvedPhotoUrl = null),
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
        this.invoices = res.invoices.items;

        this.loadCurrentBooking();

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
    this.staffApi.getFundingRecord(this.childId, this.billingMonth).subscribe({
      next: (detail) => {
        this.fundingDetail = detail;
        this.isLoading = false;
      },
      error: () => {
        this.fundingDetail = null;
        this.isLoading = false;
      },
    });
  }

  private loadCurrentBooking(): void {
    const membership = this.auth.activeMembership();
    const siteId = membership?.branch_id;
    if (!siteId) return;

    this.sessionTypesApi.listSessionTypes(siteId, { includeArchived: false }).subscribe({
      next: (types) => (this.sessionTypes = types),
      error: () => { /* Session types load failure handled gracefully */ },
    });

    this.bookingsApi.listBookings(siteId, { childId: this.childId, status: 'active' }, 1, 50).subscribe({
      next: (result) => {
        const recurring = result.items.find(b => b.bookingType === 'recurring' && b.status === 'active');
        if (recurring) {
          this.currentBooking = recurring;
          this.bookingsApi.getBooking(siteId, recurring.id).subscribe({
            next: (detail) => {
              this.currentBookingEntries = (detail as unknown as { session_entries?: SessionEntry[] }).session_entries ?? [];
            },
            error: () => (this.currentBookingEntries = []),
          });
        } else {
          this.currentBooking = null;
          this.currentBookingEntries = [];
        }
      },
      error: () => {
        this.currentBooking = null;
        this.currentBookingEntries = [];
      },
    });
  }

  saveFundingRecord(): void {
    if (!this.fundingDetail?.record) return;
    this.isSavingFunding = true;
    this.fundingSaveMessage = null;
    const r = this.fundingDetail.record;
    const payload: FundingRecordWritePayload = {
      funding_enabled: r.fundingEnabled,
      funding_type: r.fundingType,
      funding_model: r.fundingModel,
      funded_hours_per_week: r.fundedHoursPerWeek,
      funding_start_date: r.fundingStartDate,
      funding_end_date: r.fundingEndDate,
      eligibility_code: r.eligibilityCode,
      eligibility_code_validated: r.eligibilityCodeValidated,
      evidence_received: r.evidenceReceived,
    };
    this.staffApi.upsertFundingRecord(this.childId, payload).subscribe({
      next: () => {
        this.isSavingFunding = false;
        this.fundingSaveMessage = 'Funding record saved.';
        this.loadMonthlyProfile();
      },
      error: (err) => {
        this.isSavingFunding = false;
        this.errorMessage = err?.message ?? 'Failed to save funding record.';
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
          this.resolvePhotoUrl();
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
          this.resolvedPhotoUrl = null;
          this.photoService.invalidate(`/api/v1/children/${this.childId}/photo`);
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
