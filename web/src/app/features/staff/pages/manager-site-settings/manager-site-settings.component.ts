import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowRightCircle,
  heroBanknotes,
  heroBuildingOffice2,
  heroChatBubbleLeftRight,
  heroCheckCircle,
  heroChevronRight,
  heroClock,
  heroCog6Tooth,
  heroReceiptPercent,
  heroScale,
  heroShieldCheck,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';
import { catchError, of } from 'rxjs';

import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StaffApiService } from '../../data/staff-api.service';
import { StaffRoom, StaffRoomsApiService } from '../../data/staff-rooms-api.service';

type CardTone = 'primary' | 'success' | 'warning' | 'info' | 'neutral';

interface SettingCardBase {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly icon: string;
  readonly iconBgClass: string;
  readonly iconClass: string;
  readonly tone: CardTone;
  readonly state: 'ready' | 'coming-soon';
  readonly link: string | null;
  readonly statusLabel: string;
  readonly statusTone: 'neutral' | 'success' | 'warning' | 'info' | 'primary';
  readonly statusItalic?: boolean;
  readonly badge?: { label: string; tone: 'success' | 'warning' | 'info' | 'primary' };
}

interface BadgeCard extends SettingCardBase {
  readonly kind: 'badge';
}

interface ToggleCard extends SettingCardBase {
  readonly kind: 'toggle';
  readonly toggles: { id: string; label: string; enabled: boolean }[];
}

interface PillCard extends SettingCardBase {
  readonly kind: 'pill';
  readonly pillIcon: string;
  readonly pillText: string;
}

interface HighlightCard extends SettingCardBase {
  readonly kind: 'highlight';
  readonly highlightLabel: string;
  readonly highlightValue: string;
}

interface RatioCard extends SettingCardBase {
  readonly kind: 'ratio';
  readonly thresholdPercent: number;
}

interface StatusListCard extends SettingCardBase {
  readonly kind: 'status-list';
  readonly items: { id: string; label: string; enabled: boolean }[];
}

interface KeyValueCard extends SettingCardBase {
  readonly kind: 'key-value';
  readonly items: { id: string; icon: string; text: string }[];
}

type SettingCard =
  | BadgeCard
  | ToggleCard
  | PillCard
  | HighlightCard
  | RatioCard
  | StatusListCard
  | KeyValueCard;

@Component({
  selector: 'app-manager-site-settings',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    NgIcon,
    PageHeaderComponent,
    LoadingStateComponent,
    AlertComponent,
  ],
  providers: [
    provideIcons({
      heroArrowRightCircle,
      heroBanknotes,
      heroBuildingOffice2,
      heroChatBubbleLeftRight,
      heroCheckCircle,
      heroChevronRight,
      heroClock,
      heroCog6Tooth,
      heroReceiptPercent,
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

  readonly routes = ROLE_ROUTES;

  readonly loading = signal(true);
  readonly pageError = signal<string | null>(null);
  readonly rooms = signal<StaffRoom[]>([]);
  readonly billing = signal<{ rateMinor: number | null; hasRate: boolean }>({ rateMinor: null, hasRate: false });
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
    standardLabel: 'EYFS 2024 Standards',
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

    const statusToneFor = (enabled: boolean): 'success' | 'warning' => (enabled ? 'success' : 'warning');

    return [
      {
        id: 'site-profile',
        kind: 'badge',
        title: 'Site profile',
        description: 'Manage nursery identity, contact information, and registration details.',
        icon: 'heroBuildingOffice2',
        iconBgClass: 'bg-brand-50 dark:bg-brand-500/15',
        iconClass: 'text-brand-500',
        tone: 'primary',
        state: 'coming-soon',
        link: null,
        statusLabel: 'Identity on file',
        statusTone: 'success',
        badge: { label: 'Configured', tone: 'success' },
      },
      {
        id: 'rooms',
        kind: 'pill',
        title: 'Rooms & capacity',
        description: 'Manage rooms, age groups, capacity limits, and EYFS ratio rules.',
        icon: 'heroCog6Tooth',
        iconBgClass: 'bg-gray-100 dark:bg-white/5',
        iconClass: 'text-gray-600 dark:text-gray-300',
        tone: 'neutral',
        state: 'ready',
        link: ROLE_ROUTES.managerRooms,
        statusLabel: activeCount === 0 ? 'No active rooms' : `${activeCount} active`,
        statusTone: activeCount === 0 ? 'warning' : 'success',
        pillIcon: 'heroUserGroup',
        pillText: ageGroups.length > 0 ? ageGroups.join(' · ') : 'No age groups yet',
      },
      {
        id: 'hours',
        kind: 'pill',
        title: 'Operating hours',
        description: 'Configure opening times, holidays, and closure windows.',
        icon: 'heroClock',
        iconBgClass: 'bg-warning-50 dark:bg-warning-500/15',
        iconClass: 'text-warning-600 dark:text-warning-400',
        tone: 'warning',
        state: 'coming-soon',
        link: null,
        statusLabel: hours.hasHolidays ? 'Holidays pending' : 'Holidays on file',
        statusTone: hours.hasHolidays ? 'warning' : 'success',
        pillIcon: 'heroClock',
        pillText: `${hours.days}, ${hours.window}`,
      },
      {
        id: 'attendance',
        kind: 'toggle',
        title: 'Attendance rules',
        description: 'Configure check-in, corrections, and attendance workflows for staff and parents.',
        icon: 'heroCheckCircle',
        iconBgClass: 'bg-blue-light-50 dark:bg-blue-light-500/15',
        iconClass: 'text-blue-light-500',
        tone: 'info',
        state: 'coming-soon',
        link: null,
        statusLabel: 'See workflow',
        statusTone: 'neutral',
        toggles: [
          { id: 'self-checkin', label: 'Parent self check-in', enabled: rules.parentSelfCheckin },
          { id: 'offline', label: 'Offline attendance', enabled: rules.offlineAttendance },
        ],
      },
      {
        id: 'ratio',
        kind: 'ratio',
        title: 'Ratio monitoring',
        description: 'Configure EYFS staffing ratio alerts and threshold guidance.',
        icon: 'heroScale',
        iconBgClass: 'bg-warning-50 dark:bg-warning-500/15',
        iconClass: 'text-warning-600 dark:text-warning-400',
        tone: 'warning',
        state: 'coming-soon',
        link: null,
        statusLabel: ratio.standardLabel,
        statusTone: 'neutral',
        thresholdPercent: ratio.thresholdPercent,
        badge: { label: ratio.isActive ? 'Active' : 'Paused', tone: ratio.isActive ? 'success' : 'warning' },
      },
      {
        id: 'funding',
        kind: 'highlight',
        title: 'Funding configuration',
        description: 'Manage funded hours, entitlement rules, and term settings for EYPP/FSM.',
        icon: 'heroBanknotes',
        iconBgClass: 'bg-brand-50 dark:bg-brand-500/15',
        iconClass: 'text-brand-500',
        tone: 'primary',
        state: 'ready',
        link: ROLE_ROUTES.managerFunding,
        statusLabel: 'Active',
        statusTone: 'success',
        highlightLabel: 'Current focus',
        highlightValue: term,
      },
      {
        id: 'billing',
        kind: 'key-value',
        title: 'Fees & billing',
        description: 'Manage sessions, prices, discounts, and auto-invoicing rules.',
        icon: 'heroReceiptPercent',
        iconBgClass: 'bg-success-50 dark:bg-success-500/15',
        iconClass: 'text-success-600 dark:text-success-500',
        tone: 'success',
        state: 'ready',
        link: ROLE_ROUTES.managerBillingSetup,
        statusLabel: hasBilling ? 'Auto-invoicing: ON' : 'Hourly rate pending',
        statusTone: hasBilling ? 'success' : 'warning',
        items: [
          { id: 'rate', icon: 'heroReceiptPercent', text: `Core rate: ${billing}` },
          { id: 'discount', icon: 'heroReceiptPercent', text: 'Sibling discount rules' },
        ],
      },
      {
        id: 'comms',
        kind: 'status-list',
        title: 'Parent communication',
        description: 'Manage messages, notifications, and the daily parent experience.',
        icon: 'heroChatBubbleLeftRight',
        iconBgClass: 'bg-gray-100 dark:bg-white/5',
        iconClass: 'text-gray-600 dark:text-gray-300',
        tone: 'neutral',
        state: 'coming-soon',
        link: null,
        statusLabel: `Last activity: ${comms.lastActivity}`,
        statusTone: 'neutral',
        items: [
          { id: 'messaging', label: 'Messaging', enabled: comms.messagingEnabled },
          { id: 'media', label: 'Photo & media sharing', enabled: comms.mediaEnabled },
        ],
      },
      {
        id: 'compliance',
        kind: 'status-list',
        title: 'Compliance',
        description: 'Manage GDPR, audit logs, retention, and security settings.',
        icon: 'heroShieldCheck',
        iconBgClass: 'bg-blue-light-50 dark:bg-blue-light-500/15',
        iconClass: 'text-blue-light-500',
        tone: 'info',
        state: 'coming-soon',
        link: null,
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
  trackItem = (_index: number, item: { id: string }): string => item.id;

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
  }
}
