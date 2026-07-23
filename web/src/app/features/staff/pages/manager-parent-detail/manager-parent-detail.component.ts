import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroCheck,
  heroCheckCircle,
  heroChevronRight,
  heroEnvelope,
  heroExclamationCircle,
  heroLink,
  heroMapPin,
  heroPencilSquare,
  heroPhone,
  heroPlus,
  heroShieldCheck,
  heroTrash,
  heroUser,
  heroUserGroup,
  heroXCircle,
} from '@ng-icons/heroicons/outline';
import { forkJoin, of, Subject } from 'rxjs';
import { catchError, takeUntil } from 'rxjs/operators';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ParentsApiService } from '../../data/parents-api.service';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord } from '../../models/children.models';
import { ParentRecord, ParentChildLink } from '../../models/parents.models';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ToastService } from '../../../../shared/services/toast.service';

export interface LinkedChildDetail {
  link: ParentChildLink;
  child: ChildRecord | null;
}

@Component({
  selector: 'app-manager-parent-detail',
  imports: [
    CommonModule,
    RouterLink,
    AlertComponent,
    BadgeComponent,
    StatusBadgeComponent,
    ConfirmationDialogComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    NgIcon,
  ],
  templateUrl: './manager-parent-detail.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroCheck,
      heroCheckCircle,
      heroChevronRight,
      heroEnvelope,
      heroExclamationCircle,
      heroLink,
      heroMapPin,
      heroPencilSquare,
      heroPhone,
      heroPlus,
      heroShieldCheck,
      heroTrash,
      heroUser,
      heroUserGroup,
      heroXCircle,
    }),
  ],
})
export class ManagerParentDetailComponent implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly parentsApi = inject(ParentsApiService);
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);
  private readonly destroy$ = new Subject<void>();

  parent: ParentRecord | null = null;
  children: ParentChildLink[] = [];
  linkedChildrenDetails: LinkedChildDetail[] = [];
  isLoading = true;
  errorMessage: string | null = null;
  actionInProgress = false;

  // Confirmation Modals
  showUnlinkDialog = false;
  childIdToUnlink: string | null = null;
  childNameToUnlink = '';

  showDeactivateDialog = false;
  showRevokeDialog = false;

  ngOnInit(): void {
    this.route.paramMap.pipe(takeUntil(this.destroy$)).subscribe((params) => {
      const parentId = params.get('parentId');
      if (parentId) {
        this.loadParent(parentId);
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  loadParent(parentId: string): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.parentsApi.get(parentId).subscribe({
      next: (data) => {
        this.parent = data;
        this.children = data.children || [];

        if (this.children.length > 0) {
          const childRequests = this.children.map((link) =>
            this.staffApi.getChild(link.child_id).pipe(catchError(() => of(null))),
          );
          forkJoin(childRequests).subscribe((childRecords) => {
            this.linkedChildrenDetails = this.children.map((link, idx) => ({
              link,
              child: childRecords[idx],
            }));
            this.isLoading = false;
          });
        } else {
          this.linkedChildrenDetails = [];
          this.isLoading = false;
        }
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  parentName(): string {
    if (!this.parent) return '';
    return [this.parent.first_name, this.parent.last_name].filter(Boolean).join(' ');
  }

  parentInitials(): string {
    if (!this.parent) return '';
    const f = this.parent.first_name ? this.parent.first_name[0] : '';
    const l = this.parent.last_name ? this.parent.last_name[0] : '';
    return (f + l).toUpperCase() || 'P';
  }

  childInitials(child: ChildRecord | null): string {
    if (!child) return 'C';
    const f = child.firstName ? child.firstName[0] : '';
    const l = child.lastName ? child.lastName[0] : '';
    return (f + l).toUpperCase() || 'C';
  }

  formatAddress(): string {
    if (!this.parent) return '';
    const parts = [
      this.parent.address_line1,
      this.parent.address_line2,
      this.parent.address_city,
      this.parent.address_postcode,
    ].filter(Boolean);
    return parts.length > 0 ? parts.join(', ') : '—';
  }

  promptInviteToPortal(): void {
    if (!this.parent || this.actionInProgress) return;
    this.inviteToPortal();
  }

  inviteToPortal(): void {
    if (!this.parent || this.actionInProgress) return;
    this.actionInProgress = true;

    this.parentsApi.inviteToPortal(this.parent.id).subscribe({
      next: () => {
        this.actionInProgress = false;
        this.toast.success('Portal invite sent successfully.');
        this.loadParent(this.parent!.id);
      },
      error: (error) => {
        this.actionInProgress = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  promptRevokeAccess(): void {
    this.showRevokeDialog = true;
  }

  confirmRevokeAccess(): void {
    if (!this.parent || this.actionInProgress) return;
    this.actionInProgress = true;

    this.parentsApi.revokeAccess(this.parent.id).subscribe({
      next: () => {
        this.actionInProgress = false;
        this.showRevokeDialog = false;
        this.toast.success('Portal access revoked.');
        this.loadParent(this.parent!.id);
      },
      error: (error) => {
        this.actionInProgress = false;
        this.showRevokeDialog = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  promptDeactivateParent(): void {
    this.showDeactivateDialog = true;
  }

  confirmDeactivateParent(): void {
    if (!this.parent || this.actionInProgress) return;
    this.actionInProgress = true;

    this.parentsApi.delete(this.parent.id).subscribe({
      next: () => {
        this.actionInProgress = false;
        this.showDeactivateDialog = false;
        this.toast.success('Parent deactivated.');
        this.loadParent(this.parent!.id);
      },
      error: (error) => {
        this.actionInProgress = false;
        this.showDeactivateDialog = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  promptUnlinkChild(childId: string, childName?: string): void {
    this.childIdToUnlink = childId;
    this.childNameToUnlink = childName || childId;
    this.showUnlinkDialog = true;
  }

  confirmUnlinkChild(): void {
    if (!this.parent || !this.childIdToUnlink || this.actionInProgress) return;
    this.actionInProgress = true;
    const targetChildId = this.childIdToUnlink;

    this.parentsApi
      .unlinkChild(this.parent.id, targetChildId, 'contact_update', 'Unlinked from parent detail page')
      .subscribe({
        next: () => {
          this.actionInProgress = false;
          this.showUnlinkDialog = false;
          this.childIdToUnlink = null;
          this.toast.success('Child unlinked successfully.');
          this.loadParent(this.parent!.id);
        },
        error: (error) => {
          this.actionInProgress = false;
          this.showUnlinkDialog = false;
          this.childIdToUnlink = null;
          const mapped = this.errorMapper.mapAndHandle(error);
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
        },
      });
  }

  formatDate(value: string | null): string {
    if (!value) return '—';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value || '—';
    return new Intl.DateTimeFormat('en-GB', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    }).format(date);
  }

  goBack(): void {
    this.router.navigate(['/manager/parents']);
  }
}
