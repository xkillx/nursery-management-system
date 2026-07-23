import { Component, inject, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ParentsApiService } from '../../data/parents-api.service';
import { ParentRecord } from '../../models/parents.models';
import { ToastService } from '../../../../shared/services/toast.service';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ManagerParentFormComponent, ParentFormData } from '../manager-parent-form/manager-parent-form.component';

@Component({
  selector: 'app-manager-parent-edit',
  imports: [ManagerParentFormComponent, LoadingStateComponent],
  template: `
    @if (isLoading) {
      <app-loading-state [label]="'Loading parent details...'" />
    } @else {
      <app-manager-parent-form
        [parent]="parent"
        [isEditMode]="true"
        [isSubmitting]="isSubmitting"
        [errorMessage]="errorMessage"
        (saved)="onSave($event)"
      />
    }
  `,
})
export class ManagerParentEditComponent implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly parentsApi = inject(ParentsApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);
  private readonly destroy$ = new Subject<void>();

  parent: ParentRecord | null = null;
  isLoading = true;
  isSubmitting = false;
  errorMessage: string | null = null;

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

  private loadParent(parentId: string): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.parentsApi.get(parentId).subscribe({
      next: (data) => {
        this.parent = data;
        this.isLoading = false;
      },
      error: (error) => {
        this.isLoading = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }

  onSave(data: ParentFormData): void {
    if (!this.parent) return;

    this.isSubmitting = true;
    this.errorMessage = null;

    this.parentsApi.update(this.parent.id, {
      first_name: data.first_name.trim(),
      last_name: data.last_name.trim() || null,
      email: data.email.trim() || null,
      phone: data.phone.trim() || null,
      address_line1: data.address_line1.trim() || null,
      address_line2: data.address_line2.trim() || null,
      address_city: data.address_city.trim() || null,
      address_postcode: data.address_postcode.trim() || null,
      has_parental_responsibility: data.has_parental_responsibility,
      can_pick_up: data.can_pick_up,
      is_emergency_contact: data.is_emergency_contact,
      notes: data.notes.trim() || null,
    }).subscribe({
      next: (parent) => {
        this.isSubmitting = false;
        this.toast.success('Parent updated successfully.');
        this.router.navigate(['/manager/parents', parent.id]);
      },
      error: (error) => {
        this.isSubmitting = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.parent'));
      },
    });
  }
}
