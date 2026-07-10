import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import { heroEnvelope, heroPaperAirplane, heroXCircle } from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { StaffApiService } from '../../data/staff-api.service';
import {
  InviteRecord,
  InviteRole,
  InviteStatusFilter,
} from '../../models/invites.models';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { ToastService } from '../../../../shared/services/toast.service';

const STATUS_OPTIONS: InviteStatusFilter[] = ['pending', 'all', 'accepted', 'revoked', 'expired'];
const STATUS_FILTER_OPTIONS: Option[] = [
  { value: 'all', label: 'All statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'accepted', label: 'Accepted' },
  { value: 'revoked', label: 'Revoked' },
  { value: 'expired', label: 'Expired' },
];
const ROLE_OPTIONS: Option[] = [
  { value: 'practitioner', label: 'Practitioner' },
  { value: 'parent', label: 'Parent' },
];

@Component({
  selector: 'app-manager-invites',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    PageHeaderComponent,
    SelectComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    ConfirmationDialogComponent,
  ],
  templateUrl: './manager-invites.component.html',
  providers: [
    provideIcons({ heroEnvelope, heroPaperAirplane, heroXCircle }),
  ],
})
export class ManagerInvitesComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);

  readonly statusOptions = STATUS_OPTIONS;
  readonly statusFilterOptions = STATUS_FILTER_OPTIONS;
  readonly roleOptions = ROLE_OPTIONS;

  invites: InviteRecord[] = [];
  status: InviteStatusFilter = 'pending';
  email = '';
  role: InviteRole = 'practitioner';

  isLoading = false;
  isSubmitting = false;
  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};

  rowErrors: Record<string, string> = {};
  pendingInviteIds = new Set<string>();
  inviteToRevoke: InviteRecord | null = null;
  isRevoking = false;

  ngOnInit(): void {
    this.loadInvites();
  }

  canAct(invite: InviteRecord): boolean {
    return invite.status === 'pending';
  }

  isRowPending(inviteId: string): boolean {
    return this.pendingInviteIds.has(inviteId);
  }

  setStatusFilter(status: InviteStatusFilter): void {
    this.status = status;
    this.errorMessage = null;
    this.loadInvites();
  }

  submitInvite(): void {
    const trimmed = this.email.trim();
    if (!trimmed || !this.role) return;

    this.clearFormErrors();
    this.isSubmitting = true;

    this.staffApi.createInvite({ email: trimmed, role: this.role }).subscribe({
      next: () => {
        this.isSubmitting = false;
        this.toast.success(`Invitation pending for ${trimmed}.`);
        this.email = '';
        this.role = 'practitioner';
        this.loadInvites();
      },
      error: (err) => {
        this.isSubmitting = false;
        const mapped = this.errorMapper.mapAndHandle(err);
        const presented = presentApiError(mapped, 'auth.managerInvites');
        if (presented.fieldErrors['email'] || presented.fieldErrors['role']) {
          this.fieldErrors = { ...presented.fieldErrors };
        } else {
          this.errorMessage = formatPresentedApiError(presented);
        }
      },
    });
  }

  resend(invite: InviteRecord): void {
    if (!this.canAct(invite)) return;

    delete this.rowErrors[invite.id];
    this.pendingInviteIds.add(invite.id);

    this.staffApi.resendInvite(invite.id).subscribe({
      next: () => {
        this.pendingInviteIds.delete(invite.id);
        this.toast.success(`Invitation resent to ${invite.email}.`);
        this.loadInvites();
      },
      error: (err) => {
        this.pendingInviteIds.delete(invite.id);
        const mapped = this.errorMapper.mapAndHandle(err);
        this.rowErrors[invite.id] = formatPresentedApiError(presentApiError(mapped, 'auth.managerInvites'));
      },
    });
  }

  openRevoke(invite: InviteRecord): void {
    if (!this.canAct(invite)) return;
    this.inviteToRevoke = invite;
  }

  cancelRevoke(): void {
    this.inviteToRevoke = null;
  }

  confirmRevoke(): void {
    if (!this.inviteToRevoke) return;

    this.isRevoking = true;
    const inviteId = this.inviteToRevoke.id;

    this.staffApi.revokeInvite(inviteId).subscribe({
      next: () => {
        this.isRevoking = false;
        this.toast.success(`Invitation revoked for ${this.inviteToRevoke!.email}.`);
        this.inviteToRevoke = null;
        this.loadInvites();
      },
      error: (err) => {
        this.isRevoking = false;
        const mapped = this.errorMapper.mapAndHandle(err);
        this.rowErrors[inviteId] = formatPresentedApiError(presentApiError(mapped, 'auth.managerInvites'));
        this.inviteToRevoke = null;
      },
    });
  }

  formatDateTime(iso: string | null): string {
    if (!iso) return '-';
    return new Intl.DateTimeFormat('en-GB', {
      dateStyle: 'medium',
      timeStyle: 'short',
      timeZone: 'Europe/London',
    }).format(new Date(iso));
  }

  private loadInvites(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi.listInvites(this.status).subscribe({
      next: (invites) => {
        this.invites = invites;
        this.isLoading = false;
      },
      error: (err) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'auth.managerInvites'));
      },
    });
  }

  private clearFormErrors(): void {
    this.errorMessage = null;
    this.fieldErrors = {};
  }
}
