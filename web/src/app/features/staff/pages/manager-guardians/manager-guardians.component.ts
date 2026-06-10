import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { GuardianFormComponent } from '../../components/guardian-form/guardian-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { StatusFilter } from '../../models/children.models';
import { GuardianRecord, GuardianWritePayload } from '../../models/guardians.models';
import { statusFilterLabel } from '../../utils/manager-list-formatters';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { TablePaginationComponent } from '../../../../shared/components/ui/table/table-pagination.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

@Component({
  selector: 'app-manager-guardians',
  imports: [
    CommonModule,
    RouterLink,
    GuardianFormComponent,
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    TableShellComponent,
    TablePaginationComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-guardians.component.html',
})
export class ManagerGuardiansComponent {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  readonly statusOptions: StatusFilter[] = ['active', 'inactive', 'all'];

  readonly statusLabel = statusFilterLabel;

  guardians: GuardianRecord[] = [];
  status: StatusFilter = 'active';
  limit = 10;
  offset = 0;
  isLoading = false;
  isSaving = false;

  selectedGuardian: GuardianRecord | null = null;
  formMode: 'create' | 'edit' = 'create';
  showForm = false;

  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};

  ngOnInit(): void {
    this.loadGuardians();
  }

  loadGuardians(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi
      .listGuardians({
        status: this.status,
        limit: this.limit,
        offset: this.offset,
      })
      .subscribe({
        next: (guardians) => {
          this.guardians = guardians;
          this.isLoading = false;
        },
        error: (error) => {
          this.isLoading = false;
          const mapped = this.errorMapper.mapAndHandle(error);
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.guardian'));
        },
      });
  }

  onStatusChange(nextStatus: string): void {
    this.status = nextStatus as StatusFilter;
    this.offset = 0;
    this.loadGuardians();
  }

  openCreate(): void {
    this.formMode = 'create';
    this.selectedGuardian = null;
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showForm = true;
  }

  openEdit(guardian: GuardianRecord): void {
    this.formMode = 'edit';
    this.selectedGuardian = guardian;
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showForm = true;
  }

  closeForm(): void {
    this.showForm = false;
    this.selectedGuardian = null;
    this.fieldErrors = {};
    this.errorMessage = null;
  }

  save(payload: GuardianWritePayload): void {
    this.isSaving = true;
    this.fieldErrors = {};
    this.errorMessage = null;

    const request$ =
      this.formMode === 'create'
        ? this.staffApi.createGuardian(payload)
        : this.staffApi.updateGuardian(this.selectedGuardian!.id, payload);

    request$.subscribe({
      next: () => {
        this.isSaving = false;
        this.closeForm();
        this.loadGuardians();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.guardian'));
      },
    });
  }

  nextPage(): void {
    if (this.guardians.length < this.limit) {
      return;
    }

    this.offset += this.limit;
    this.loadGuardians();
  }

  previousPage(): void {
    if (this.offset === 0) {
      return;
    }

    this.offset = Math.max(0, this.offset - this.limit);
    this.loadGuardians();
  }

}
