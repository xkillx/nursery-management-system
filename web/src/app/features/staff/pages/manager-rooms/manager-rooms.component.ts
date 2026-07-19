import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArchiveBoxXMark,
  heroArrowUturnLeft,
  heroBuildingOffice2,
  heroChartBar,
  heroChevronDoubleDown,
  heroFunnel,
  heroPencilSquare,
  heroPlus,
  heroUserCircle,
  heroUserGroup,
} from '@ng-icons/heroicons/outline';

import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { StaffRoomsApiService, StaffRoom } from '../../data/staff-rooms-api.service';

type RoomStatusFilter = 'all' | 'active' | 'archived';

interface RoomOccupancy {
  current: number;
  percent: number;
}

interface RoomRow {
  room: StaffRoom;
  occupancy: RoomOccupancy;
}

@Component({
  selector: 'app-manager-rooms',
  imports: [
    CommonModule,
    RouterModule,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    SelectComponent,
    NgIcon,
  ],
  templateUrl: './manager-rooms.component.html',
  providers: [
    provideIcons({
      heroArchiveBoxXMark,
      heroArrowUturnLeft,
      heroBuildingOffice2,
      heroChartBar,
      heroChevronDoubleDown,
      heroFunnel,
      heroPencilSquare,
      heroPlus,
      heroUserCircle,
      heroUserGroup,
    }),
  ],
})
export class ManagerRoomsComponent implements OnInit {
  private readonly roomsApi = inject(StaffRoomsApiService);
  private readonly auth = inject(AuthService);

  readonly limit = 25;
  readonly listRoute = ROLE_ROUTES.managerRooms;
  readonly newRoomRoute = `${ROLE_ROUTES.managerRooms}/new`;

  loadingRooms = false;
  archivingRoomId: string | null = null;
  pageError: string | null = null;

  selectedSiteId: string | null = null;
  selectedSiteName = '';
  rooms: StaffRoom[] = [];
  statusFilter: RoomStatusFilter = 'all';
  searchTerm = '';
  visibleCount = this.limit;

  readonly statusOptions: Option[] = [
    { value: 'all', label: 'All rooms' },
    { value: 'active', label: 'Active only' },
    { value: 'archived', label: 'Archived only' },
  ];

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.pageError = 'No site is attached to this manager session.';
      return;
    }

    this.selectedSiteId = membership.branch_id;
    this.selectedSiteName = membership.branch_name ?? 'Assigned site';
    this.loadRooms();
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

  get totalRoomsPill(): string {
    return this.totalRooms === 0 ? 'No rooms yet' : 'Live snapshot';
  }

  readonly totalCapacityPill = 'Across displayed rooms';

  get occupancyPill(): string {
    return this.activeRows.length === 0 ? 'Awaiting rooms' : 'Live snapshot';
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

  editRoute(room: StaffRoom): string {
    return `${ROLE_ROUTES.managerRooms}/${room.id}/edit`;
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

  archiveRoom(room: StaffRoom): void {
    if (!this.selectedSiteId || !room.isActive) return;
    if (!confirm(`Archive ${room.name}? Children must be reassigned first.`)) return;

    this.archivingRoomId = room.id;
    this.pageError = null;
    this.roomsApi.archiveRoom(this.selectedSiteId, room.id).subscribe({
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

  reactivateRoom(room: StaffRoom): void {
    if (!this.selectedSiteId || room.isActive) return;

    this.archivingRoomId = room.id;
    this.pageError = null;
    this.roomsApi.reactivateRoom(this.selectedSiteId, room.id).subscribe({
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

  private loadRooms(): void {
    if (!this.selectedSiteId) {
      this.rooms = [];
      return;
    }

    this.loadingRooms = true;
    this.pageError = null;
    this.roomsApi
      .listRooms(this.selectedSiteId, { includeArchived: true, includeOccupancy: true, pageSize: 200 })
      .subscribe({
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

  private computeOccupancy(room: StaffRoom): RoomOccupancy {
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
