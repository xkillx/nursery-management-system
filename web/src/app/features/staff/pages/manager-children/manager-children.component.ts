import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ChildFormComponent } from '../../components/child-form/child-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord, ChildWritePayload, StatusFilter } from '../../models/children.models';

@Component({
  selector: 'app-manager-children',
  imports: [CommonModule, ChildFormComponent],
  templateUrl: './manager-children.component.html',
})
export class ManagerChildrenComponent {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);

  readonly statusOptions: StatusFilter[] = ['active', 'inactive', 'all'];

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
          this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
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
        this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
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

  private messageWithRequestId(message: string, requestId: string | null): string {
    if (!requestId) {
      return message;
    }

    return `${message} (Request: ${requestId})`;
  }
}
