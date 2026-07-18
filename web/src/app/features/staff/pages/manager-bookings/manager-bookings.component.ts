import { CommonModule } from '@angular/common';
import { Component, inject, OnInit, OnDestroy, HostListener } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendarDays,
  heroChevronDown,
  heroChevronLeft,
  heroChevronRight,
  heroClock,
  heroFunnel,
  heroMagnifyingGlass,
  heroPlus,
  heroXMark,
} from '@ng-icons/heroicons/outline';
import { Subject, debounceTime, takeUntil } from 'rxjs';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { AuthService } from '../../../../core/services/auth.service';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { BookingsApiService } from '../../data/bookings-api.service';
import { StaffRoomsApiService, StaffRoom } from '../../data/staff-rooms-api.service';
import { StaffSessionTypesApiService } from '../../data/session-types-api.service';
import { StaffSessionTemplatesApiService } from '../../data/session-templates-api.service';
import { ToastService } from '../../../../shared/services/toast.service';
import {
  UnifiedBooking,
  BookingType,
  BookingStatus,
  BookingListFilters,
} from '../../models/booking.models';
import { BookingDetailDrawerComponent } from './booking-detail-drawer/booking-detail-drawer.component';

const BOOKING_TYPE_OPTIONS: { value: BookingType; label: string }[] = [
  { value: 'recurring', label: 'Recurring' },
  { value: 'ad_hoc', label: 'Ad-hoc' },
  { value: 'hourly', label: 'Hourly' },
];

const STATUS_OPTIONS: { value: BookingStatus; label: string }[] = [
  { value: 'active', label: 'Active' },
  { value: 'paused', label: 'Paused' },
  { value: 'cancelled', label: 'Cancelled' },
];

const LIMIT = 50;
const LS_KEY = 'nursery.booking_filters';

type DatePreset = 'this_month' | 'next_month' | 'custom' | '';

interface FilterState {
  types: BookingType[];
  statuses: BookingStatus[];
  roomId: string;
  dateFrom: string;
  dateTo: string;
  datePreset: DatePreset;
  q: string;
}

// eslint-disable-next-line @typescript-eslint/consistent-indexed-object-style
interface SessionLookup {
  [id: string]: string;
}

@Component({
  selector: 'app-manager-bookings',
  imports: [
    CommonModule,
    FormsModule,
    EmptyStateComponent,
    AlertComponent,
    StatusBadgeComponent,
    NgIcon,
    BookingDetailDrawerComponent,
  ],
  templateUrl: './manager-bookings.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroChevronDown,
      heroChevronLeft,
      heroChevronRight,
      heroClock,
      heroFunnel,
      heroMagnifyingGlass,
      heroPlus,
      heroXMark,
    }),
  ],
})
export class ManagerBookingsComponent implements OnInit, OnDestroy {
  private readonly apiService = inject(BookingsApiService);
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly sessionTemplatesApi = inject(StaffSessionTemplatesApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly toast = inject(ToastService);
  private readonly auth = inject(AuthService);
  private readonly destroy$ = new Subject<void>();
  private readonly searchChanged$ = new Subject<string>();

  readonly bookingTypeOptions = BOOKING_TYPE_OPTIONS;
  readonly statusOptions = STATUS_OPTIONS;
  readonly datePresetOptions: { value: DatePreset; label: string }[] = [
    { value: 'this_month', label: 'This Month' },
    { value: 'next_month', label: 'Next Month' },
    { value: 'custom', label: 'Custom' },
    { value: '', label: 'All Time' },
  ];

  siteId: string | null = null;

  selectedTypes: BookingType[] = [];
  selectedStatuses: BookingStatus[] = [];
  selectedRoomId = '';
  datePreset: DatePreset = 'this_month';
  dateFrom = '';
  dateTo = '';
  searchQuery = '';
  offset = 0;

  items: UnifiedBooking[] = [];
  total = 0;
  rooms: StaffRoom[] = [];
  sessionLookup: SessionLookup = {};
  isLoading = false;
  errorMessage: string | null = null;

  isCreateDropdownOpen = false;
  selectedBooking: UnifiedBooking | null = null;

  get hasPrevious(): boolean {
    return this.offset > 0;
  }

  get hasNext(): boolean {
    return this.offset + LIMIT < this.total;
  }

  get currentPage(): number {
    return Math.floor(this.offset / LIMIT) + 1;
  }

  get totalPages(): number {
    return Math.max(1, Math.ceil(this.total / LIMIT));
  }

  get showingFrom(): number {
    return this.total === 0 ? 0 : this.offset + 1;
  }

  get showingTo(): number {
    return Math.min(this.offset + this.items.length, this.total);
  }

  get hasActiveFilters(): boolean {
    return (
      this.selectedTypes.length > 0 ||
      this.selectedStatuses.length > 0 ||
      this.selectedRoomId !== '' ||
      this.datePreset !== '' ||
      this.searchQuery.trim() !== ''
    );
  }

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.errorMessage = 'No site is attached to this manager session.';
      return;
    }
    this.siteId = membership.branch_id;

    this.searchChanged$.pipe(debounceTime(300), takeUntil(this.destroy$)).subscribe(() => {
      this.offset = 0;
      this.loadList();
    });

    this.route.queryParams.pipe(takeUntil(this.destroy$)).subscribe((params) => {
      const hasUrlParams = Object.keys(params).length > 0;
      if (hasUrlParams) {
        this.applyQueryParams(params);
      } else {
        this.restoreFromLocalStorage();
        // Apply default date preset on first visit
        if (this.datePreset && !this.dateFrom && !this.dateTo) {
          this.setDatePreset(this.datePreset);
          return;
        }
      }
      this.loadList();
    });

    this.loadRooms();
    this.loadSessionLookups();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onTypeToggle(type: BookingType): void {
    const idx = this.selectedTypes.indexOf(type);
    if (idx >= 0) {
      this.selectedTypes = this.selectedTypes.filter((t) => t !== type);
    } else {
      this.selectedTypes = [...this.selectedTypes, type];
    }
    this.offset = 0;
    this.onFilterChange();
  }

  isTypeSelected(type: BookingType): boolean {
    return this.selectedTypes.includes(type);
  }

  onStatusToggle(status: BookingStatus): void {
    const idx = this.selectedStatuses.indexOf(status);
    if (idx >= 0) {
      this.selectedStatuses = this.selectedStatuses.filter((s) => s !== status);
    } else {
      this.selectedStatuses = [...this.selectedStatuses, status];
    }
    this.offset = 0;
    this.onFilterChange();
  }

  isStatusSelected(status: BookingStatus): boolean {
    return this.selectedStatuses.includes(status);
  }

  onRoomChange(): void {
    this.offset = 0;
    this.onFilterChange();
  }

  onDateFromChange(): void {
    this.offset = 0;
    this.onFilterChange();
  }

  onDateToChange(): void {
    this.offset = 0;
    this.onFilterChange();
  }

  onSearchInput(value: string): void {
    this.searchQuery = value;
    this.searchChanged$.next(value);
  }

  clearSearch(): void {
    this.searchQuery = '';
    this.offset = 0;
    this.onFilterChange();
  }

  clearAllFilters(): void {
    this.selectedTypes = [];
    this.selectedStatuses = [];
    this.selectedRoomId = '';
    this.datePreset = '';
    this.dateFrom = '';
    this.dateTo = '';
    this.searchQuery = '';
    this.offset = 0;
    localStorage.removeItem(LS_KEY);
    this.router.navigate([], { queryParams: {} });
    this.loadList();
  }

  setDatePreset(preset: DatePreset): void {
    if (this.datePreset === preset) return;
    this.datePreset = preset;
    const now = new Date();
    switch (preset) {
      case 'this_month': {
        const first = new Date(now.getFullYear(), now.getMonth(), 1);
        const last = new Date(now.getFullYear(), now.getMonth() + 1, 0);
        this.dateFrom = this.toIsoDate(first);
        this.dateTo = this.toIsoDate(last);
        break;
      }
      case 'next_month': {
        const first = new Date(now.getFullYear(), now.getMonth() + 1, 1);
        const last = new Date(now.getFullYear(), now.getMonth() + 2, 0);
        this.dateFrom = this.toIsoDate(first);
        this.dateTo = this.toIsoDate(last);
        break;
      }
      case 'custom':
        break;
      default:
        this.dateFrom = '';
        this.dateTo = '';
        break;
    }
    this.offset = 0;
    this.onFilterChange();
  }

  @HostListener('document:keydown', ['$event'])
  handleKeyboardShortcut(event: KeyboardEvent): void {
    const target = event.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.tagName === 'SELECT' || target.isContentEditable;
    if (isInput) return;

    if (event.key === 'n' && !event.ctrlKey && !event.metaKey && !event.altKey) {
      event.preventDefault();
      this.openCreateDropdown();
    }
  }

  previousPage(): void {
    this.offset = Math.max(0, this.offset - LIMIT);
    this.loadList();
  }

  nextPage(): void {
    this.offset += LIMIT;
    this.loadList();
  }

  openCreateDropdown(): void {
    this.isCreateDropdownOpen = !this.isCreateDropdownOpen;
  }

  closeCreateDropdown(): void {
    this.isCreateDropdownOpen = false;
  }

  navigateToCreate(type: BookingType): void {
    this.isCreateDropdownOpen = false;
    this.router.navigate(['/manager/bookings/new', type]);
  }

  openBookingDetail(booking: UnifiedBooking, event: Event): void {
    const target = event.target as HTMLElement;
    if (target.closest('button')) return;
    this.selectedBooking = booking;
  }

  closeDetailDrawer(): void {
    this.selectedBooking = null;
  }

  onBookingCancelled(): void {
    this.closeDetailDrawer();
    this.loadList();
    this.toast.success('Booking cancelled successfully.');
  }

  onBookingUpdated(): void {
    this.closeDetailDrawer();
    this.loadList();
    this.toast.success('Booking updated successfully.');
  }

  sessionName(id: string): string {
    return this.sessionLookup[id] ?? '—';
  }

  bookingTypeLabel(type: BookingType): string {
    switch (type) {
      case 'recurring': return 'Recurring';
      case 'ad_hoc': return 'Ad-hoc';
      case 'hourly': return 'Hourly';
    }
  }

  bookingTypeClasses(type: BookingType): string {
    switch (type) {
      case 'recurring':
        return 'bg-brand-50 text-brand-700 dark:bg-brand-500/15 dark:text-brand-300';
      case 'ad_hoc':
        return 'bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-warning-300';
      case 'hourly':
        return 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-300';
    }
  }

  formatDateRange(booking: UnifiedBooking): string {
    if (!booking.endDate || booking.startDate === booking.endDate) {
      return this.formatDate(booking.startDate);
    }
    return `${this.formatDate(booking.startDate)} – ${this.formatDate(booking.endDate)}`;
  }

  formatDate(iso: string): string {
    if (!iso) return '';
    const d = new Date(iso);
    return new Intl.DateTimeFormat('en-GB', {
      timeZone: 'Europe/London',
      dateStyle: 'medium',
    }).format(d);
  }

  childFullName(booking: UnifiedBooking): string {
    return `${booking.childFirstName} ${booking.childLastName}`.trim();
  }

  private onFilterChange(): void {
    this.saveToLocalStorage();
    this.syncUrlParams();
    this.loadList();
  }

  private applyQueryParams(params: Record<string, string>): void {
    if (params['type']) {
      this.selectedTypes = params['type']
        .split(',')
        .filter((t): t is BookingType => ['recurring', 'ad_hoc', 'hourly'].includes(t));
    } else {
      this.selectedTypes = [];
    }

    if (params['status']) {
      this.selectedStatuses = params['status']
        .split(',')
        .filter((s): s is BookingStatus => ['active', 'paused', 'cancelled'].includes(s));
    } else {
      this.selectedStatuses = [];
    }

    if (params['room_id']) this.selectedRoomId = params['room_id'];
    if (params['preset']) this.datePreset = params['preset'] as DatePreset;
    if (params['from']) this.dateFrom = params['from'];
    if (params['to']) this.dateTo = params['to'];
    if (params['q']) this.searchQuery = params['q'];
    if (params['offset']) {
      const o = parseInt(params['offset'], 10);
      if (!isNaN(o) && o >= 0) this.offset = o;
    }
  }

  private restoreFromLocalStorage(): void {
    try {
      const raw = localStorage.getItem(LS_KEY);
      if (!raw) return;
      const state: FilterState = JSON.parse(raw);
      if (state.types) this.selectedTypes = state.types;
      if (state.statuses) this.selectedStatuses = state.statuses;
      if (state.roomId) this.selectedRoomId = state.roomId;
      if (state.datePreset) this.datePreset = state.datePreset;
      if (state.dateFrom) this.dateFrom = state.dateFrom;
      if (state.dateTo) this.dateTo = state.dateTo;
      if (state.q) this.searchQuery = state.q;
    } catch {
      // corrupted localStorage — ignore
    }
  }

  private saveToLocalStorage(): void {
    const state: FilterState = {
      types: this.selectedTypes,
      statuses: this.selectedStatuses,
      roomId: this.selectedRoomId,
      datePreset: this.datePreset,
      dateFrom: this.dateFrom,
      dateTo: this.dateTo,
      q: this.searchQuery,
    };
    try {
      localStorage.setItem(LS_KEY, JSON.stringify(state));
    } catch {
      // localStorage full or unavailable — ignore
    }
  }

  private syncUrlParams(): void {
    const queryParams: Record<string, string> = {};
    if (this.selectedTypes.length > 0) queryParams['type'] = this.selectedTypes.join(',');
    if (this.selectedStatuses.length > 0) queryParams['status'] = this.selectedStatuses.join(',');
    if (this.selectedRoomId) queryParams['room_id'] = this.selectedRoomId;
    if (this.datePreset) queryParams['preset'] = this.datePreset;
    if (this.dateFrom) queryParams['from'] = this.dateFrom;
    if (this.dateTo) queryParams['to'] = this.dateTo;
    if (this.searchQuery.trim()) queryParams['q'] = this.searchQuery.trim();
    if (this.offset > 0) queryParams['offset'] = String(this.offset);
    this.router.navigate([], { queryParams, queryParamsHandling: 'merge' });
  }

  private toIsoDate(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  private loadList(): void {
    if (!this.siteId) return;
    this.isLoading = true;
    this.errorMessage = null;

    const filters: BookingListFilters = {};
    if (this.selectedTypes.length === 1) filters.status = undefined;
    if (this.selectedStatuses.length === 1) filters.status = this.selectedStatuses[0];
    if (this.selectedRoomId) filters.roomId = this.selectedRoomId;
    if (this.searchQuery.trim()) filters.search = this.searchQuery.trim();
    if (this.dateFrom) filters.from = this.dateFrom;
    if (this.dateTo) filters.to = this.dateTo;

    this.apiService.listBookings(this.siteId, filters, Math.floor(this.offset / LIMIT) + 1, LIMIT).subscribe({
      next: (result) => {
        let filtered = result.items;
        if (this.selectedTypes.length > 0) {
          filtered = filtered.filter((b) => this.selectedTypes.includes(b.bookingType));
        }
        if (this.selectedStatuses.length > 0) {
          filtered = filtered.filter((b) => this.selectedStatuses.includes(b.status));
        }
        this.items = filtered;
        this.total = result.total;
        this.isLoading = false;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = mapped.message ?? 'Failed to load bookings.';
        this.isLoading = false;
      },
    });
  }

  private loadRooms(): void {
    if (!this.siteId) return;
    this.roomsApi.listRooms(this.siteId, { includeArchived: false }).subscribe({
      next: (rooms) => this.rooms = rooms.filter((r) => r.isActive),
      error: () => { /* Rooms load failure handled by template defaults */ },
    });
  }

  private loadSessionLookups(): void {
    if (!this.siteId) return;

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => {
        for (const t of types.filter((st) => st.isActive)) {
          this.sessionLookup[t.id] = t.name;
        }
      },
      error: () => { /* Session types load failure handled by template defaults */ },
    });

    this.sessionTemplatesApi.listSessionTemplates(this.siteId, { includeArchived: false }).subscribe({
      next: (templates) => {
        for (const t of templates.filter((st) => st.isActive)) {
          this.sessionLookup[t.id] = t.name;
        }
      },
      error: () => { /* Session templates load failure handled by template defaults */ },
    });
  }
}
