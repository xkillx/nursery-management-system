import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ParentsApiService } from '../../data/parents-api.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { ManagerParentFormComponent, ParentFormData } from '../manager-parent-form/manager-parent-form.component';

@Component({
  selector: 'app-manager-parent-new',
  imports: [ManagerParentFormComponent],
  template: `
    <app-manager-parent-form
      [isSubmitting]="isSubmitting"
      [errorMessage]="errorMessage"
      (saved)="onSave($event)"
    />
  `,
})
export class ManagerParentNewComponent {
  private readonly parentsApi = inject(ParentsApiService);
  private readonly router = inject(Router);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly toast = inject(ToastService);

  isSubmitting = false;
  errorMessage: string | null = null;

  onSave(data: ParentFormData): void {
    this.isSubmitting = true;
    this.errorMessage = null;

    this.parentsApi.create({
      first_name: data.first_name.trim(),
      last_name: data.last_name.trim() || undefined,
      email: data.email.trim() || undefined,
      phone: data.phone.trim() || undefined,
      address_line1: data.address_line1.trim() || undefined,
      address_line2: data.address_line2.trim() || undefined,
      address_city: data.address_city.trim() || undefined,
      address_postcode: data.address_postcode.trim() || undefined,
      has_parental_responsibility: data.has_parental_responsibility,
      can_pick_up: data.can_pick_up,
      is_emergency_contact: data.is_emergency_contact,
      notes: data.notes.trim() || undefined,
    }).subscribe({
      next: (parent) => {
        this.isSubmitting = false;
        this.toast.success('Parent created successfully.');
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
