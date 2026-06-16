import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroBuildingOffice,
  heroInformationCircle,
  heroPencilSquare,
  heroPlus,
  heroShieldCheck,
} from '@ng-icons/heroicons/outline';

import { ROLES, ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { OwnerApiService } from '../../data/owner-api.service';
import { OwnerSiteSummary, Room } from '../../models/owner.models';

interface RoomFormModel {
  name: string;
  ageGroup: string;
  capacity: number | null;
  description: string;
}

@Component({
  selector: 'app-owner-room-form',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    LoadingStateComponent,
    AlertComponent,
    SelectComponent,
    NgIcon,
  ],
  templateUrl: './owner-room-form.component.html',
  styleUrl: './owner-room-form.component.css',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroBuildingOffice,
      heroInformationCircle,
      heroPencilSquare,
      heroPlus,
      heroShieldCheck,
    }),
  ],
})
export class OwnerRoomFormComponent implements OnInit {
  private readonly api = inject(OwnerApiService);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  loadingSites = false;
  loadingRoom = false;
  submitting = false;
  pageError: string | null = null;
  fieldErrors: Partial<Record<keyof RoomFormModel, string>> = {};

  sites: OwnerSiteSummary[] = [];
  selectedSiteId: string | null = null;
  selectedSiteName = '';
  roomId: string | null = null;
  room: Room | null = null;

  model: RoomFormModel = {
    name: '',
    ageGroup: '',
    capacity: null,
    description: '',
  };

  readonly ageGroupOptions: Option[] = [
    { value: 'baby', label: 'Baby (0-2 years)' },
    { value: 'toddler', label: 'Toddler (2-3 years)' },
    { value: 'preschool', label: 'Preschool (3-5 years)' },
    { value: 'mixed', label: 'Mixed' },
  ];

  ngOnInit(): void {
    this.roomId = this.route.snapshot.paramMap.get('roomId');

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
    this.loadRoomForEdit();
  }

  get isOwner(): boolean {
    return this.auth.currentRole() === ROLES.owner;
  }

  get isEditMode(): boolean {
    return !!this.roomId;
  }

  get siteOptions(): Option[] {
    return this.sites.map((site) => ({ value: site.siteId, label: site.siteName }));
  }

  get listRoute(): string {
    return this.isOwner ? ROLE_ROUTES.ownerRooms : ROLE_ROUTES.managerRooms;
  }

  get title(): string {
    return this.isEditMode ? 'Edit Room' : 'Add New Room';
  }

  get description(): string {
    const site = this.selectedSiteName || 'this site';
    return this.isEditMode
      ? `Update room details for ${site}.`
      : `Configure a new educational space for ${site}.`;
  }

  onSiteValueChange(siteId: string): void {
    this.selectedSiteId = siteId || null;
    this.selectedSiteName = this.sites.find((site) => site.siteId === siteId)?.siteName ?? '';
    if (this.isEditMode) {
      this.loadRoomForEdit();
    }
  }

  submit(form: NgForm): void {
    this.fieldErrors = {};
    this.pageError = null;

    if (!this.selectedSiteId) {
      this.pageError = 'Select a site before saving this room.';
      return;
    }

    if (!this.validate(form)) return;

    const payload = {
      name: this.model.name.trim(),
      age_group: this.model.ageGroup,
      capacity: Number(this.model.capacity),
      description: this.model.description.trim(),
    };

    this.submitting = true;
    const request = this.isEditMode && this.roomId
      ? this.api.updateRoom(this.selectedSiteId, this.roomId, payload)
      : this.api.createRoom(this.selectedSiteId, payload);

    request.subscribe({
      next: () => {
        this.submitting = false;
        this.router.navigate([this.listRoute], {
          queryParams: this.isOwner ? { site_id: this.selectedSiteId } : undefined,
        });
      },
      error: (err) => {
        this.submitting = false;
        this.applyApiError(err);
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

    if (!site) {
      this.pageError = 'No active sites are available for room setup.';
      return;
    }

    this.selectedSiteId = site.siteId;
    this.selectedSiteName = site.siteName;
    this.loadRoomForEdit();
  }

  private loadRoomForEdit(): void {
    if (!this.isEditMode || !this.roomId || !this.selectedSiteId) return;

    this.loadingRoom = true;
    this.pageError = null;
    this.api.getRoom(this.selectedSiteId, this.roomId).subscribe({
      next: (room) => {
        this.room = room;
        this.loadingRoom = false;
        if (!room.isActive) {
          this.pageError = 'Reactivate this room before editing.';
          return;
        }
        this.model = {
          name: room.name,
          ageGroup: room.ageGroup,
          capacity: room.capacity,
          description: room.description ?? '',
        };
      },
      error: (err) => {
        this.loadingRoom = false;
        this.applyApiError(err);
      },
    });
  }

  private validate(form: NgForm): boolean {
    form.control.markAllAsTouched();

    if (!this.model.name.trim()) {
      this.fieldErrors.name = 'Room name is required.';
    } else if (this.model.name.trim().length > 120) {
      this.fieldErrors.name = 'Room name must be 120 characters or fewer.';
    }

    if (!this.model.ageGroup) {
      this.fieldErrors.ageGroup = 'Age group is required.';
    }

    const capacity = Number(this.model.capacity);
    if (!Number.isInteger(capacity) || capacity <= 0) {
      this.fieldErrors.capacity = 'Capacity must be a positive whole number.';
    }

    if (this.model.description.length > 500) {
      this.fieldErrors.description = 'Description must be 500 characters or fewer.';
    }

    return Object.keys(this.fieldErrors).length === 0;
  }

  private applyApiError(err: any): void {
    const code = err?.error?.code;
    const field = err?.error?.field;

    if (code === 'room_name_duplicate') {
      this.fieldErrors.name = 'An active room with this name already exists in this site.';
      return;
    }
    if (code === 'invalid_age_group') {
      this.fieldErrors.ageGroup = 'Select a valid age group.';
      return;
    }
    if (code === 'validation_error' && field) {
      this.fieldErrors[this.mapApiField(field)] = err?.error?.message ?? 'Check this field.';
      return;
    }
    if (code === 'site_not_found') {
      this.pageError = 'Site not found or no longer active.';
      return;
    }
    if (code === 'room_not_found') {
      this.pageError = 'Room not found.';
      return;
    }

    this.pageError = 'Failed to save room. Please try again.';
  }

  private mapApiField(field: string): keyof RoomFormModel {
    if (field === 'age_group') return 'ageGroup';
    if (field === 'capacity') return 'capacity';
    if (field === 'description') return 'description';
    return 'name';
  }
}
