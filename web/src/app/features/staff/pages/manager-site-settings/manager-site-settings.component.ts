import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroBanknotes,
  heroBuildingOffice2,
  heroCalendarDays,
  heroChatBubbleLeftRight,
  heroCheckCircle,
  heroChevronRight,
  heroClock,
  heroCog6Tooth,
  heroReceiptPercent,
  heroRectangleStack,
  heroScale,
  heroShieldCheck,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';
import { catchError, of } from 'rxjs';

import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoom, StaffRoomsApiService } from '../../data/staff-rooms-api.service';
import { StaffSessionType, StaffSessionTypesApiService } from '../../data/session-types-api.service';
import { StaffSiteProfileApiService } from '../../data/staff-site-profile-api.service';
import { SiteProfile } from '../../models/site-profile.models';

type CardTone = 'brand' | 'success' | 'warning' | 'info' | 'neutral';
type PillTone = 'success' | 'warning' | 'info' | 'brand' | 'neutral';
type StatusTone = 'success' | 'warning' | 'info' | 'brand' | 'neutral';
type CardKind = 'simple' | 'toggle' | 'ratio' | 'list';

interface CardItem {
  id: string;
  label: string;
  enabled: boolean;
}

interface SettingCard {
  readonly id: string;
  readonly title: string;
  readonly headline: string;
  readonly detail?: string;
  readonly icon: string;
  readonly tone: CardTone;
  readonly kind: CardKind;
  readonly state: 'ready' | 'coming-soon';
  readonly link: string | null;
  readonly pillLabel?: string;
  readonly pillTone?: PillTone;
  readonly statusLabel: string;
  readonly statusTone: StatusTone;
  readonly toggles?: CardItem[];
  readonly thresholdPercent?: number;
  readonly items?: CardItem[];
}

@Component({
  selector: 'app-manager-site-settings',
  standalone: true,
  imports: [CommonModule, RouterLink, NgIcon, AlertComponent],
  providers: [
    provideIcons({
      heroBanknotes,
      heroBuildingOffice2,
      heroCalendarDays,
      heroChatBubbleLeftRight,
      heroCheckCircle,
      heroChevronRight,
      heroClock,
      heroCog6Tooth,
      heroReceiptPercent,
      heroRectangleStack,
      heroScale,
      heroShieldCheck,
      heroUserGroup,
    }),
  ],
  templateUrl: './manager-site-settings.component.html',
})
export class ManagerSiteSettingsComponent implements OnInit {
  private readonly auth = inject(AuthService);
  private readonly staffApi = inject(StaffApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly siteProfileApi = inject(StaffSiteProfileApiService);

  readonly routes = ROLE_ROUTES;

  readonly loading = signal(true);
  readonly pageError = signal<string | null>(null);
  readonly rooms = signal<StaffRoom[]>([]);
  readonly sessionTypes = signal<StaffSessionType[]>([]);
  readonly billing = signal<{ rateMinor: number | null; hasRate: boolean }>({ rateMinor: null, hasRate: false });
  readonly siteProfile = signal<{ hasProfile: boolean; profile: SiteProfile | null } | null>(null);
  readonly attendanceRules = signal({
    parentSelfCheckin: true,
    offlineAttendance: true,
  });
  readonly currentTerm = signal('Spring Term 2026');
  readonly operatingHours = signal({
    days: 'Mon-Fri',
    window: '07:30 - 18:30',
    hasHolidays: false,
  });
  readonly ratioMonitor = signal({
    thresholdPercent: 80,
    isActive: true,
  });
  readonly compliance = signal({
    mfaEnabled: true,
    auditLogActive: true,
    securityScore: 98,
  });
  readonly parentComms = signal({
    messagingEnabled: true,
    mediaEnabled: true,
    lastActivity: '1h ago',
  });

  readonly siteName = computed(() => {
    const membership = this.auth.activeMembership();
    return membership?.branch_name ?? 'Your nursery';
  });

  readonly activeSessionTypes = computed(() => this.sessionTypes().filter((st) => st.isActive));

  readonly activeRooms = computed(() => this.rooms().filter((room) => room.isActive));
  readonly ageGroupLabels = computed(() => {
    const groups = new Set<string>();
    for (const room of this.activeRooms()) {
      groups.add(this.formatAgeGroup(room.ageGroup));
    }
    return Array.from(groups);
  });

  readonly billingDisplay = computed(() => {
    const data = this.billing();
    if (!data.hasRate || data.rateMinor === null) {
      return 'Hourly rate not set';
    }
    return '£' + (data.rateMinor / 100).toFixed(2) + '/hr';
  });

  readonly cards = computed<SettingCard[]>(() => {
    const billing = this.billingDisplay();
    const hours = this.operatingHours();
    const ratio = this.ratioMonitor();
    const rules = this.attendanceRules();
    const comms = this.parentComms();
    const compliance = this.compliance();
    const term = this.currentTerm();
    const ageGroups = this.ageGroupLabels();
    const activeCount = this.activeRooms().length;
    const hasBilling = this.billing().hasRate;
    const sessionActiveCount = this.activeSessionTypes().length;

    const spState = this.siteProfile();

    const siteProfileCard: SettingCard = spState === null
      ? {
          id: 'site-profile',
          title: 'Site profile',
          headline: 'Loading...',
          detail: 'Fetching site profile data.',
          icon: 'heroBuildingOffice2',
          tone: 'brand',
          kind: 'simple',
          state: 'ready',
          link: null,
          pillLabel: 'Loading',
          pillTone: 'neutral',
          statusLabel: 'Loading...',
          statusTone: 'neutral',
        }
      : !spState.hasProfile
        ? {
            id: 'site-profile',
            title: 'Site profile',
            headline: 'Setup needed',
            detail: 'Add a Site Profile to appear on parent invoices.',
            icon: 'heroBuildingOffice2',
            tone: 'brand',
            kind: 'simple',
            state: 'ready',
            link: ROLE_ROUTES.managerSiteProfile,
            pillLabel: 'Not configured',
            pillTone: 'warning',
            statusLabel: 'Not configured',
            statusTone: 'warning',
          }
        : {
            id: 'site-profile',
            title: 'Site profile',
            headline: spState.profile!.nursery_name,
            detail: `${spState.profile!.address_street}, ${spState.profile!.address_city}`,
            icon: 'heroBuildingOffice2',
            tone: 'brand',
            kind: 'simple',
            state: 'ready',
            link: ROLE_ROUTES.managerSiteProfile,
            pillLabel: 'Configured',
            pillTone: 'success',
            statusLabel: 'Identity on file',
            statusTone: 'success',
          };

    return [
      siteProfileCard,
      {
        id: 'rooms',
        title: 'Rooms & capacity',
        headline: activeCount === 0 ? 'No active rooms' : `${activeCount} active room${activeCount === 1 ? '' : 's'}`,
        detail: ageGroups.length > 0 ? ageGroups.join(' · ') : 'Add a room to define age groups.',
        icon: 'heroCog6Tooth',
        tone: 'neutral',
        kind: 'simple',
        state: 'ready',
        link: ROLE_ROUTES.managerRooms,
        pillLabel: activeCount === 0 ? 'Setup needed' : 'Live',
        pillTone: activeCount === 0 ? 'warning' : 'success',
        statusLabel: `${activeCount} active`,
        statusTone: activeCount === 0 ? 'warning' : 'success',
      },
      {
        id: 'session-types',
        title: 'Session types',
        headline: sessionActiveCount === 0 ? 'No session types' : `${sessionActiveCount} active type${sessionActiveCount === 1 ? '' : 's'}`,
        detail: 'Define reusable time blocks for bookings.',
        icon: 'heroRectangleStack',
        tone: 'brand',
        kind: 'simple',
        state: 'ready',
        link: ROLE_ROUTES.managerSessionTypes,
        pillLabel: sessionActiveCount === 0 ? 'Setup needed' : 'Configured',
        pillTone: sessionActiveCount === 0 ? 'warning' : 'success',
        statusLabel: `${sessionActiveCount} active`,
        statusTone: sessionActiveCount === 0 ? 'warning' : 'success',
      },
      {
        id: 'hours',
        title: 'Operating hours',
        headline: `${hours.days}, ${hours.window}`,
        detail: 'Holiday windows and closure days are tracked here.',
        icon: 'heroClock',
        tone: 'warning',
        kind: 'simple',
        state: 'coming-soon',
        link: null,
        pillLabel: hours.hasHolidays ? 'Holidays pending' : 'Holidays on file',
        pillTone: hours.hasHolidays ? 'warning' : 'success',
        statusLabel: hours.hasHolidays ? 'Holidays pending' : 'Holidays on file',
        statusTone: hours.hasHolidays ? 'warning' : 'success',
      },
      {
        id: 'attendance',
        title: 'Attendance rules',
        headline: 'Workflow defaults',
        detail: 'Check-in, corrections, and offline capture.',
        icon: 'heroCheckCircle',
        tone: 'info',
        kind: 'toggle',
        state: 'coming-soon',
        link: null,
        pillLabel: 'Defaults',
        pillTone: 'info',
        statusLabel: 'Review workflow',
        statusTone: 'neutral',
        toggles: [
          { id: 'self-checkin', label: 'Parent self check-in', enabled: rules.parentSelfCheckin },
          { id: 'offline', label: 'Offline attendance', enabled: rules.offlineAttendance },
        ],
      },
      {
        id: 'ratio',
        title: 'Ratio monitoring',
        headline: ratio.isActive ? 'Live EYFS alerts' : 'Paused',
        detail: 'Threshold guidance against EYFS 2024 standards.',
        icon: 'heroScale',
        tone: 'warning',
        kind: 'ratio',
        state: 'coming-soon',
        link: null,
        pillLabel: ratio.isActive ? 'Active' : 'Paused',
        pillTone: ratio.isActive ? 'success' : 'warning',
        statusLabel: 'EYFS 2024 standards',
        statusTone: 'neutral',
        thresholdPercent: ratio.thresholdPercent,
      },
      {
        id: 'funding',
        title: 'Funding configuration',
        headline: term,
        detail: 'Funded hours, entitlement rules, and term settings.',
        icon: 'heroBanknotes',
        tone: 'brand',
        kind: 'simple',
        state: 'coming-soon',
        link: null,
        statusLabel: term,
        statusTone: 'neutral',
      },
      {
        id: 'billing',
        title: 'Fees & billing',
        headline: billing,
        detail: 'Sessions, sibling discounts, and auto-invoicing.',
        icon: 'heroReceiptPercent',
        tone: 'success',
        kind: 'simple',
        state: 'ready',
        link: ROLE_ROUTES.managerBillingSetup,
        pillLabel: hasBilling ? 'Live' : 'Setup',
        pillTone: hasBilling ? 'success' : 'warning',
        statusLabel: hasBilling ? 'Auto-invoicing: ON' : 'Hourly rate pending',
        statusTone: hasBilling ? 'success' : 'warning',
      },
      {
        id: 'term-calendar',
        title: 'Term calendar',
        headline: 'Academic terms',
        detail: 'Manage autumn, spring, and summer term dates for term-time billing.',
        icon: 'heroCalendarDays',
        tone: 'info',
        kind: 'simple',
        state: 'ready',
        link: '/manager/site-settings/term-calendar',
        statusLabel: 'Manage terms',
        statusTone: 'info',
      },
      {
        id: 'closure-days',
        title: 'Closure days',
        headline: 'Inset days & bank holidays',
        detail: 'Exclude closure dates from billing calculations automatically.',
        icon: 'heroCalendarDays',
        tone: 'warning',
        kind: 'simple',
        state: 'ready',
        link: '/manager/site-settings/closure-days',
        statusLabel: 'Manage closures',
        statusTone: 'warning',
      },
      {
        id: 'comms',
        title: 'Parent communication',
        headline: 'Messaging & media',
        detail: 'Notifications, daily updates, and parent experience.',
        icon: 'heroChatBubbleLeftRight',
        tone: 'neutral',
        kind: 'list',
        state: 'coming-soon',
        link: null,
        pillLabel: 'Enabled',
        pillTone: 'success',
        statusLabel: `Last activity: ${comms.lastActivity}`,
        statusTone: 'neutral',
        items: [
          { id: 'messaging', label: 'Messaging', enabled: comms.messagingEnabled },
          { id: 'media', label: 'Photo & media sharing', enabled: comms.mediaEnabled },
        ],
      },
      {
        id: 'compliance',
        title: 'Compliance',
        headline: `Security score ${compliance.securityScore}%`,
        detail: 'GDPR, audit logs, retention, and security.',
        icon: 'heroShieldCheck',
        tone: 'info',
        kind: 'list',
        state: 'coming-soon',
        link: null,
        pillLabel: `${compliance.securityScore}%`,
        pillTone: 'success',
        statusLabel: `Security score: ${compliance.securityScore}%`,
        statusTone: 'success',
        items: [
          { id: 'mfa', label: 'MFA enabled', enabled: compliance.mfaEnabled },
          { id: 'audit', label: 'Audit log active', enabled: compliance.auditLogActive },
        ],
      },
    ];
  });

  ngOnInit(): void {
    this.load();
  }

  iconWrapClass(card: SettingCard): string {
    const map: Record<CardTone, string> = {
      brand: 'bg-brand-50 text-brand-600 dark:bg-brand-500/15 dark:text-brand-300',
      success: 'bg-success-50 text-success-600 dark:bg-success-500/15 dark:text-success-300',
      warning: 'bg-warning-50 text-warning-600 dark:bg-warning-500/15 dark:text-warning-300',
      info: 'bg-blue-light-50 text-blue-light-700 dark:bg-blue-light-500/15 dark:text-blue-light-300',
      neutral: 'bg-gray-100 text-gray-600 dark:bg-white/[0.08] dark:text-gray-300',
    };
    return map[card.tone];
  }

  pillTextClass(card: SettingCard): string {
    const tone = card.pillTone ?? 'neutral';
    const map: Record<PillTone, string> = {
      success: 'text-success-600 dark:text-success-400',
      warning: 'text-warning-700 dark:text-warning-300',
      info: 'text-blue-light-700 dark:text-blue-light-300',
      brand: 'text-brand-600 dark:text-brand-300',
      neutral: 'text-gray-700 dark:text-gray-300',
    };
    return map[tone];
  }

  statusToneClass(card: SettingCard): string {
    const map: Record<StatusTone, string> = {
      success: 'text-success-600 dark:text-success-400',
      warning: 'text-warning-700 dark:text-warning-300',
      info: 'text-blue-light-700 dark:text-blue-light-300',
      brand: 'text-brand-600 dark:text-brand-300',
      neutral: 'text-gray-700 dark:text-gray-300',
    };
    return map[card.statusTone];
  }

  formatAgeGroup(ageGroup: string): string {
    switch (ageGroup) {
      case 'baby':
        return 'Baby';
      case 'toddler':
        return 'Toddler';
      case 'preschool':
        return 'Preschool';
      case 'mixed':
        return 'Mixed';
      default:
        return ageGroup;
    }
  }

  trackCard = (_index: number, card: SettingCard): string => card.id;
  trackItem = (_index: number, item: CardItem): string => item.id;

  private load(): void {
    const membership = this.auth.activeMembership();
    const branchId = membership?.branch_id ?? null;

    this.loading.set(true);
    this.pageError.set(null);

    this.staffApi
      .getSiteRate()
      .pipe(catchError(() => of({ core_hourly_rate_minor: 0, has_rate: false })))
      .subscribe((billing) => {
        this.billing.set({
          rateMinor: billing.core_hourly_rate_minor,
          hasRate: billing.has_rate,
        });
      });

    if (!branchId) {
      this.loading.set(false);
      return;
    }

    this.roomsApi.listRooms(branchId, { includeArchived: false }).subscribe({
      next: (rooms) => {
        this.rooms.set(rooms);
        this.loading.set(false);
      },
      error: () => {
        this.pageError.set('Failed to load site configuration. Please try again.');
        this.loading.set(false);
      },
    });

    this.sessionTypesApi.listSessionTypes(branchId, { includeArchived: true }).subscribe({
      next: (types) => this.sessionTypes.set(types),
    });

    this.siteProfileApi.getSiteProfile().pipe(
      catchError(() => of({ site_profile: null })),
    ).subscribe((resp) => {
      this.siteProfile.set({
        hasProfile: resp.site_profile !== null,
        profile: resp.site_profile,
      });
    });
  }
}
