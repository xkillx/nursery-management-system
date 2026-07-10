import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroBuildingOffice2,
  heroChartBar,
  heroFunnel,
  heroPlus,
  heroUserCircle,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';

import { ROLES, ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import {
  OverCapacityBannerComponent,
  OverCapacityRoom,
} from '../../../../shared/components/over-capacity-banner/over-capacity-banner.component';
import { OwnerApiService } from '../../data/owner-api.service';
import { OwnerSiteSummary, Room } from '../../models/owner.models';

type RoomStatusFilter = 'all' | 'active' | 'archived';

interface RoomOccupancy {
  current: number;
  percent: number;
}

interface RoomRow {
  room: Room;
  occupancy: RoomOccupancy;
}

@Component({
  selector: 'app-owner-rooms',
  imports: [
    CommonModule,
    RouterModule,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    SelectComponent,
    OverCapacityBannerComponent,
    NgIcon,
  ],
  templateUrl: './owner-rooms.component.html',
  providers: [
    provideIcons({
      heroBuildingOffice2,
      heroChartBar,
      heroFunnel,
      heroPlus,
      heroUserCircle,
      heroUserGroup,
    }),
  ],
})
export class OwnerRoomsComponent implements OnInit {
  private readonly api = inject(OwnerApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  readonly limit = 25;

  loadingSites = false;
  loadingRooms = false;
  archivingRoomId: string | null = null;
  pageError: string | null = null;

  sites: OwnerSiteSummary[] = [];
  selectedSiteId: string | null = null;
  selectedSiteName = '';
  rooms: Room[] = [];
  statusFilter: RoomStatusFilter = 'all';
  searchTerm = '';
  visibleCount = this.limit;

  readonly statusOptions: Option[] = [
    { value: 'all', label: 'All rooms' },
    { value: 'active', label: 'Active only' },
    { value: 'archived', label: 'Archived only' },
  ];

  ngOnInit(): void {
    if (this.isOwner) {
      this.loadOwnerSites();
      return;
    }

    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.pageError = 'No site is attached to this manager session.';
      return;
    }

    this.selectedSiteId = membership.branch_id;
    this.selectedSiteName = membership.branch_name ?? 'Assigned site';
    this.loadRooms();
  }

  get isOwner(): boolean {
    return this.auth.currentRole() === ROLES.owner;
  }

  get siteOptions(): Option[] {
    return this.sites.map((site) => ({ value: site.siteId, label: site.siteName }));
  }

  get listRoute(): string {
    return this.isOwner ? ROLE_ROUTES.ownerRooms : ROLE_ROUTES.managerRooms;
  }

  get newRoomRoute(): string {
    return this.isOwner ? `${ROLE_ROUTES.ownerRooms}/new` : `${ROLE_ROUTES.managerRooms}/new`;
  }

  get statusFilteredRows(): RoomRow[] {
    return this.rooms
      .filter((room) => {
        if (this.statusFilter === 'active') return room.isActive;
        if (this.statusFilter === 'archived') return !room.isActive;
        return true;
      })
      .map((room) => ({ room, occupancy: this.computeOccupancy(room) }))
      .sort((a, b) => this.compareRows(a, b));
  }

  get filteredRows(): RoomRow[] {
    const term = this.searchTerm.trim().toLowerCase();
    if (!term) {
      return this.statusFilteredRows;
    }

    return this.statusFilteredRows.filter((row) => {
      const haystack = [row.room.name, row.room.description ?? ''].join(' ').toLowerCase();
      return haystack.includes(term);
    });
  }

  get visibleRows(): RoomRow[] {
    return this.filteredRows.slice(0, this.visibleCount);
  }

  get canLoadMore(): boolean {
    return this.visibleCount < this.filteredRows.length && !this.loadingRooms;
  }

  get activeRows(): RoomRow[] {
    return this.statusFilteredRows.filter((row) => row.room.isActive);
  }

  get totalRooms(): number {
    return this.statusFilteredRows.length;
  }

  get totalCapacity(): number {
    return this.statusFilteredRows.reduce((sum, row) => sum + row.room.capacity, 0);
  }

  get averageOccupancy(): number {
    if (this.activeRows.length === 0) return 0;
    return Math.round(this.activeRows.reduce((sum, row) => sum + row.occupancy.percent, 0) / this.activeRows.length);
  }

  get staffRatio(): string {
    if (this.averageOccupancy >= 88) return '1:4';
    if (this.averageOccupancy >= 72) return '1:5';
    return '1:6';
  }

  get highestOccupancyRoom(): RoomRow | null {
    return this.activeRows.reduce<RoomRow | null>((highest, row) => {
      if (!highest || row.occupancy.percent > highest.occupancy.percent) return row;
      return highest;
    }, null);
  }

  get overCapacityRooms(): OverCapacityRoom[] {
    return this.activeRows
      .filter((row) => row.room.isOverCapacity === true)
      .map((row) => ({
        id: row.room.id,
        name: row.room.name,
        assigned: row.occupancy.current,
        capacity: row.room.capacity,
      }));
  }

  get totalRoomsPill(): string {
    return this.totalRooms === 0 ? 'No rooms yet' : 'Demo-ready room setup';
  }

  readonly totalCapacityPill = 'Across displayed rooms';

  get occupancyPill(): string {
    return this.activeRows.length === 0 ? 'Awaiting rooms' : 'Live snapshot';
  }

  get staffRatioPill(): string {
    const room = this.highestOccupancyRoom;
    if (!room || room.occupancy.percent < 88) return 'Demo ratio';
    return `Check ${room.room.name}`;
  }

  onSiteValueChange(siteId: string): void {
    this.selectedSiteId = siteId || null;
    this.selectedSiteName = this.sites.find((site) => site.siteId === siteId)?.siteName ?? '';
    this.visibleCount = this.limit;
    this.searchTerm = '';
    this.loadRooms();
  }

  onStatusFilterChange(value: string): void {
    this.statusFilter = (value as RoomStatusFilter) || 'all';
    this.visibleCount = this.limit;
  }

  onSearchChange(event: Event): void {
    this.searchTerm = (event.target as HTMLInputElement).value;
    this.visibleCount = this.limit;
  }

  loadMore(): void {
    if (!this.canLoadMore) return;
    this.visibleCount = Math.min(this.visibleCount + this.limit, this.filteredRows.length);
  }

  editRoute(room: Room): string {
    const base = this.isOwner ? ROLE_ROUTES.ownerRooms : ROLE_ROUTES.managerRooms;
    return `${base}/${room.id}/edit`;
  }

  ageGroupLabel(ageGroup: string): string {
    switch (ageGroup) {
      case 'baby':
        return '0 - 18 Months';
      case 'toddler':
        return '18m - 3 Years';
      case 'preschool':
        return '3 - 5 Years';
      case 'mixed':
        return 'All Ages';
      default:
        return ageGroup;
    }
  }

  ageGroupClasses(ageGroup: string): string {
    switch (ageGroup) {
      case 'baby':
        return 'bg-blue-100 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300';
      case 'toddler':
        return 'bg-orange-100 text-orange-700 dark:bg-orange-500/15 dark:text-orange-300';
      case 'preschool':
        return 'bg-purple-100 text-purple-700 dark:bg-purple-500/15 dark:text-purple-300';
      case 'mixed':
        return 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-200';
      default:
        return 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-200';
    }
  }

  occupancyBarClasses(row: RoomRow): string {
    if (!row.room.isActive) return 'bg-gray-300 dark:bg-gray-600';
    if (row.room.isOverCapacity) return 'bg-red-500';
    if (row.occupancy.percent >= 90) return 'bg-orange-500';
    if (row.occupancy.percent >= 80) return 'bg-emerald-500';
    return 'bg-brand-500';
  }

  archiveRoom(room: Room): void {
    if (!this.selectedSiteId || !room.isActive) return;
    if (!confirm(`Archive ${room.name}? Children must be reassigned first.`)) return;

    this.archivingRoomId = room.id;
    this.pageError = null;
    this.api.archiveRoom(this.selectedSiteId, room.id).subscribe({
      next: () => {
        this.archivingRoomId = null;
        this.loadRooms();
      },
      error: (err) => {
        this.archivingRoomId = null;
        this.pageError = this.mapError(err);
      },
    });
  }

  reactivateRoom(room: Room): void {
    if (!this.selectedSiteId || room.isActive) return;

    this.archivingRoomId = room.id;
    this.pageError = null;
    this.api.reactivateRoom(this.selectedSiteId, room.id).subscribe({
      next: () => {
        this.archivingRoomId = null;
        this.loadRooms();
      },
      error: (err) => {
        this.archivingRoomId = null;
        this.pageError = this.mapError(err);
      },
    });
  }

  private loadOwnerSites(): void {
    this.loadingSites = true;
    this.pageError = null;

    this.api.getSiteSummaries().subscribe({
      next: (res) => {
        this.sites = res.sites;
        this.loadingSites = false;
        this.applyInitialOwnerSite();
      },
      error: () => {
        this.loadingSites = false;
        this.pageError = 'Failed to load sites.';
      },
    });
  }

  private applyInitialOwnerSite(): void {
    const querySiteId = this.route.snapshot.queryParamMap.get('site_id');
    const site = this.sites.find((s) => s.siteId === querySiteId) ?? this.sites[0];
    if (!site) return;

    this.selectedSiteId = site.siteId;
    this.selectedSiteName = site.siteName;
    this.loadRooms();
  }

  private loadRooms(): void {
    if (!this.selectedSiteId) {
      this.rooms = [];
      return;
    }

    this.loadingRooms = true;
    this.pageError = null;
    this.api.listRooms(this.selectedSiteId, true, true).subscribe({
      next: (rooms) => {
        this.rooms = rooms;
        this.loadingRooms = false;
        this.visibleCount = this.limit;
      },
      error: (err) => {
        this.loadingRooms = false;
        this.pageError = this.mapError(err);
      },
    });
  }

  private computeOccupancy(room: Room): RoomOccupancy {
    if (!room.isActive || room.capacity <= 0) {
      return { current: 0, percent: 0 };
    }

    const assigned = room.assignedCount ?? 0;
    return {
      current: assigned,
      percent: Math.round((assigned / room.capacity) * 100),
    };
  }

  private compareRows(a: RoomRow, b: RoomRow): number {
    if (b.occupancy.percent !== a.occupancy.percent) {
      return b.occupancy.percent - a.occupancy.percent;
    }
    return a.room.name.localeCompare(b.room.name);
  }

  private mapError(err: unknown): string {
    const body = (err as { error?: { code?: string; details?: { assigned_count?: number }; message?: string } })?.error;
    const code = body?.code;
    if (code === 'site_not_found') return 'Site not found or no longer active.';
    if (code === 'room_has_children') {
      const assigned = body?.details?.assigned_count;
      if (typeof assigned === 'number') {
        return `${assigned} ${assigned === 1 ? 'child is' : 'children are'} still assigned — reassign them before archiving.`;
      }
      return body?.message ?? 'Children must be reassigned before archiving this room.';
    }
    if (code === 'room_not_found') return 'Room not found. The list has been refreshed.';
    return 'Failed to update rooms. Please try again.';
  }
}
