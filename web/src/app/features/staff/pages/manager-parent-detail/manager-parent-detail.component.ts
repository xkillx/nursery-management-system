import { CommonModule } from '@angular/common';
import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroCheckCircle,
  heroEnvelope,
  heroLink,
  heroPencilSquare,
  heroTrash,
  heroUserGroup,
  heroXCircle,
} from '@ng-icons/heroicons/outline';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ParentsApiService } from '../../data/parents-api.service';
import { ParentRecord, ParentChildLink } from '../../models/parents.models';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ToastService } from '../../../../shared/services/toast.service';

@Component({
  selector: 'app-manager-parent-detail',
  imports: [
    CommonModule,
    RouterLink,
    AlertComponent,
    EmptyStateComponent,
    LoadingStateComponent,
    NgIcon,
  ],
  templateUrl: './manager-parent-detail.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroCheckCircle,
      heroEnvelope,
      heroLink,
      heroPencilSquare,
      heroTrash,
      heroUserGroup,
      heroXCircle,
    }),
  ],
})
export class ManagerParentDetailComponent implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly parentsApi = inject(ParentsApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);
  private readonly destroy$ = new Subject<void>();

  parent: ParentRecord | null = null;
  children: ParentChildLink[] = [];
  isLoading = true;
  errorMessage: string | null = null;
  actionInProgress = false;

  ngOnInit(): void {
    this.route.paramMap.pipe(takeUntil(this.destroy$)).subscribe(params => {
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
        this.isLoading = false;
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

  revokeAccess(): void {
    if (!this.parent || this.actionInProgress) return;
    this.actionInProgress = true;

    this.parentsApi.revokeAccess(this.parent.id).subscribe({
      next: () => {
        this.actionInProgress = false;
        this.toast.success('Portal access revoked.');
        this.loadParent(this.parent!.id);
      },
      error: (error) => {
        this.actionInProgress = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  deactivateParent(): void {
    if (!this.parent || this.actionInProgress) return;
    this.actionInProgress = true;

    this.parentsApi.delete(this.parent.id).subscribe({
      next: () => {
        this.actionInProgress = false;
        this.toast.success('Parent deactivated.');
        this.loadParent(this.parent!.id);
      },
      error: (error) => {
        this.actionInProgress = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  unlinkChild(childId: string): void {
    if (!this.parent || this.actionInProgress) return;
    this.actionInProgress = true;

    this.parentsApi.unlinkChild(this.parent.id, childId, 'contact_update', 'Unlinked from parent detail page').subscribe({
      next: () => {
        this.actionInProgress = false;
        this.toast.success('Child unlinked successfully.');
        this.loadParent(this.parent!.id);
      },
      error: (error) => {
        this.actionInProgress = false;
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
