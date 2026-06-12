import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ChildFormComponent } from '../../components/child-form/child-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord, ChildWritePayload, StatusFilter } from '../../models/children.models';
import { formatHourlyRateGbp, missingRequirementLabel, statusFilterLabel } from '../../utils/manager-list-formatters';
import { SelectComponent, Option } from '../../../../shared/components/form/select/select.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { TableShellComponent } from '../../../../shared/components/ui/table/table-shell.component';
import { TablePaginationComponent } from '../../../../shared/components/ui/table/table-pagination.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

@Component({
  selector: 'app-manager-children',
  imports: [
    CommonModule,
    RouterLink,
    SelectComponent,
    ChildFormComponent,
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    TableShellComponent,
    TablePaginationComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-children.component.html',
})
export class ManagerChildrenComponent {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  readonly statusOptions: StatusFilter[] = ['active', 'inactive', 'all'];

  readonly statusLabel = statusFilterLabel;
  readonly formatRate = formatHourlyRateGbp;
  get statusSelectOptions(): Option[] {
    return this.statusOptions.map(s => ({ value: s, label: statusFilterLabel(s) }));
  }
  readonly requirementLabel = missingRequirementLabel;

  children: ChildRecord[] = [];
  status: StatusFilter = 'active';
  limit = 10;
  offset = 0;
  isLoading = false;
  isSaving = false;

  selectedChild: ChildRecord | null = null;
  formMode: 'create' | 'edit' = 'create';
  showForm = false;

  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};

  ngOnInit(): void {
    this.loadChildren();
  }

  loadChildren(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.staffApi
      .listChildren({
        status: this.status,
        limit: this.limit,
        offset: this.offset,
      })
      .subscribe({
        next: (children) => {
          this.children = children;
          this.isLoading = false;
        },
        error: (error) => {
          this.isLoading = false;
          const mapped = this.errorMapper.mapAndHandle(error);
          this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        },
      });
  }

  onStatusChange(nextStatus: string): void {
    this.status = nextStatus as StatusFilter;
    this.offset = 0;
    this.loadChildren();
  }

  openCreate(): void {
    this.formMode = 'create';
    this.selectedChild = null;
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showForm = true;
  }

  openEdit(child: ChildRecord): void {
    this.formMode = 'edit';
    this.selectedChild = child;
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showForm = true;
  }

  closeForm(): void {
    this.showForm = false;
    this.selectedChild = null;
    this.fieldErrors = {};
    this.errorMessage = null;
  }

  save(payload: ChildWritePayload): void {
    this.isSaving = true;
    this.fieldErrors = {};
    this.errorMessage = null;

    const request$ =
      this.formMode === 'create'
        ? this.staffApi.createChild(payload)
        : this.staffApi.updateChild(this.selectedChild!.id, payload);

    request$.subscribe({
      next: () => {
        this.isSaving = false;
        this.closeForm();
        this.loadChildren();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
      },
    });
  }

  nextPage(): void {
    if (this.children.length < this.limit) {
      return;
    }

    this.offset += this.limit;
    this.loadChildren();
  }

  previousPage(): void {
    if (this.offset === 0) {
      return;
    }

    this.offset = Math.max(0, this.offset - this.limit);
    this.loadChildren();
  }

}
