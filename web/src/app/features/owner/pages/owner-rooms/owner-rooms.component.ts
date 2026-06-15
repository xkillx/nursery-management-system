import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule } from '@angular/router';

import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { OwnerApiService } from '../../data/owner-api.service';
import { Room, OwnerSiteSummary } from '../../models/owner.models';
import { AuthService } from '../../../../core/services/auth.service';
import { ROLES } from '../../../../core/constants/roles';

@Component({
  selector: 'app-owner-rooms',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    PageHeaderComponent,
    SelectComponent,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    ButtonComponent,
  ],
  templateUrl: './owner-rooms.component.html',
})
export class OwnerRoomsComponent implements OnInit {
  private readonly api = inject(OwnerApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly authService = inject(AuthService);

  loadingSites = true;
  loadingRooms = false;
  submitting = false;
  pageError: string | null = null;
  successMessage: string | null = null;

  sites: OwnerSiteSummary[] = [];
  selectedSiteId: string | null = null;
  rooms: Room[] = [];
  showArchived = false;
  isEditing = false;
  editingRoomId: string | null = null;

  formName = '';
  formAgeGroup = '';
  formCapacity: number | null = null;
  formDescription = '';
  formError: string | null = null;
  formSubmitting = false;

  ROLES = ROLES;

  get siteOptions(): Option[] {
    return this.sites.map(s => ({ value: s.siteId, label: s.siteName }));
  }

  readonly ageGroupOptions: Option[] = [
    { value: 'baby', label: 'Baby' },
    { value: 'toddler', label: 'Toddler' },
    { value: 'preschool', label: 'Preschool' },
    { value: 'mixed', label: 'Mixed' },
  ];

  get currentRole(): string | null {
    return this.authService.currentRole();
  }

  get isOwner(): boolean {
    return this.currentRole === ROLES.owner;
  }

  get isManager(): boolean {
    return this.currentRole === ROLES.manager;
  }

  get isPractitioner(): boolean {
    return this.currentRole === ROLES.practitioner;
  }

  get canWrite(): boolean {
    return this.isOwner || this.isManager;
  }

  get filteredRooms(): Room[] {
    if (this.showArchived) {
      return this.rooms;
    }
    return this.rooms.filter(r => r.isActive);
  }

  get hasFormErrors(): boolean {
    return !this.formName.trim()
      || !this.formAgeGroup
      || this.formCapacity === null || this.formCapacity <= 0;
  }

  ngOnInit(): void {
    if (this.isOwner) {
      this.loadSites();
    } else {
      const membership = this.authService.activeMembership();
      if (membership?.branch_id) {
        this.selectedSiteId = membership.branch_id;
        this.loadRooms();
      } else {
        this.pageError = 'Could not determine your site.';
        this.loadingSites = false;
      }
    }
  }

  onSiteValueChange(value: string): void {
    this.selectedSiteId = value || null;
    if (this.selectedSiteId) {
      this.loadRooms();
    } else {
      this.rooms = [];
    }
  }

  startCreate(): void {
    this.isEditing = true;
    this.editingRoomId = null;
    this.resetForm();
  }

  startEdit(room: Room): void {
    this.isEditing = true;
    this.editingRoomId = room.id;
    this.formName = room.name;
    this.formAgeGroup = room.ageGroup;
    this.formCapacity = room.capacity;
    this.formDescription = room.description ?? '';
    this.formError = null;
  }

  cancelForm(): void {
    this.isEditing = false;
    this.editingRoomId = null;
    this.resetForm();
  }

  onSubmitForm(): void {
    this.formError = null;

    if (this.hasFormErrors) {
      this.formError = 'Please fill in all required fields.';
      return;
    }

    if (!this.selectedSiteId) return;

    this.formSubmitting = true;

    const body = {
      name: this.formName.trim(),
      age_group: this.formAgeGroup,
      capacity: this.formCapacity!,
      description: this.formDescription.trim() || undefined,
    };

    if (this.editingRoomId) {
      this.api.updateRoom(this.selectedSiteId, this.editingRoomId, body).subscribe({
        next: () => {
          this.formSubmitting = false;
          this.cancelForm();
          this.successMessage = 'Room updated successfully.';
          this.loadRooms();
        },
        error: (err) => {
          this.formSubmitting = false;
          this.formError = this.mapFormError(err);
        },
      });
    } else {
      this.api.createRoom(this.selectedSiteId, body).subscribe({
        next: () => {
          this.formSubmitting = false;
          this.cancelForm();
          this.successMessage = 'Room created successfully.';
          this.loadRooms();
        },
        error: (err) => {
          this.formSubmitting = false;
          this.formError = this.mapFormError(err);
        },
      });
    }
  }

  onArchive(room: Room): void {
    if (!this.selectedSiteId) return;
    if (!confirm(`Archive "${room.name}"? Children must be reassigned first.`)) return;

    this.pageError = null;
    this.successMessage = null;
    this.api.archiveRoom(this.selectedSiteId, room.id).subscribe({
      next: () => {
        this.successMessage = `"${room.name}" archived.`;
        this.loadRooms();
      },
      error: (err) => {
        this.pageError = this.mapError(err);
      },
    });
  }

  onReactivate(room: Room): void {
    if (!this.selectedSiteId) return;

    this.pageError = null;
    this.successMessage = null;
    this.api.reactivateRoom(this.selectedSiteId, room.id).subscribe({
      next: () => {
        this.successMessage = `"${room.name}" reactivated.`;
        this.loadRooms();
      },
      error: (err) => {
        this.pageError = this.mapError(err);
      },
    });
  }

  toggleShowArchived(): void {
    this.showArchived = !this.showArchived;
  }

  private loadSites(): void {
    this.api.getSiteSummaries().subscribe({
      next: (res) => {
        this.sites = res.sites;
        this.loadingSites = false;
        const qSiteId = this.route.snapshot.queryParamMap.get('site_id');
        if (qSiteId && this.sites.some((s) => s.siteId === qSiteId)) {
          this.selectedSiteId = qSiteId;
          this.loadRooms();
        }
      },
      error: () => {
        this.pageError = 'Failed to load sites.';
        this.loadingSites = false;
      },
    });
  }

  private loadRooms(): void {
    if (!this.selectedSiteId) return;

    this.loadingRooms = true;
    this.pageError = null;
    this.api.listRooms(this.selectedSiteId, true).subscribe({
      next: (rooms) => {
        this.rooms = rooms;
        this.loadingRooms = false;
      },
      error: () => {
        this.pageError = 'Failed to load rooms.';
        this.loadingRooms = false;
      },
    });
  }

  private resetForm(): void {
    this.formName = '';
    this.formAgeGroup = '';
    this.formCapacity = null;
    this.formDescription = '';
    this.formError = null;
  }

  private mapFormError(err: any): string {
    const code = err?.error?.code;
    if (code === 'room_name_duplicate') return 'A room with this name already exists in this site.';
    if (code === 'invalid_age_group') return 'Invalid age group selected.';
    if (code === 'validation_error') {
      const field = err?.error?.details?.field;
      if (field === 'name') return 'Room name is required.';
      if (field === 'capacity') return 'Capacity must be greater than 0.';
      if (field === 'age_group') return 'Age group is required.';
    }
    return 'An unexpected error occurred. Please try again.';
  }

  private mapError(err: any): string {
    const code = err?.error?.code;
    if (code === 'room_not_found') return 'Room not found. It may have been deleted.';
    if (code === 'room_has_children') return err?.error?.message || 'Room has active children assigned.';
    if (code === 'room_not_active') return 'Room is already archived.';
    return 'An unexpected error occurred. Please try again.';
  }
}
