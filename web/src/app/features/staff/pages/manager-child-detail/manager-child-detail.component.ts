import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { ChildFormComponent } from '../../components/child-form/child-form.component';
import { StaffApiService } from '../../data/staff-api.service';
import { ChildRecord, ChildWritePayload, StatusFilter } from '../../models/children.models';
import { ChildGuardianLinkRecord, GuardianChildLinkWritePayload, GuardianRecord } from '../../models/guardians.models';
import { formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';

@Component({
  selector: 'app-manager-child-detail',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink,
    ChildFormComponent,
    PageHeaderComponent,
    ButtonComponent,
    AlertComponent,
    StatusBadgeComponent,
    EmptyStateComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-child-detail.component.html',
})
export class ManagerChildDetailComponent implements OnInit {
  private readonly staffApi = inject(StaffApiService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);

  readonly formatRate = formatHourlyRateGbp;
  readonly requirementLabel = missingRequirementLabel;

  childId = '';
  child: ChildRecord | null = null;
  linkedGuardians: ChildGuardianLinkRecord[] = [];
  allGuardians: GuardianRecord[] = [];

  isLoadingChild = false;
  isLoadingLinks = false;
  isSaving = false;
  isLinking = false;

  showEditForm = false;
  errorMessage: string | null = null;
  fieldErrors: Record<string, string> = {};

  selectedGuardianId = '';
  currentMonthLabel = '';

  ngOnInit(): void {
    this.currentMonthLabel = this.formatCurrentMonth();
    this.loadAll();
  }

  onEditChild(): void {
    this.fieldErrors = {};
    this.errorMessage = null;
    this.showEditForm = true;
  }

  closeEditForm(): void {
    this.showEditForm = false;
    this.fieldErrors = {};
    this.errorMessage = null;
  }

  saveChild(payload: ChildWritePayload): void {
    this.isSaving = true;
    this.fieldErrors = {};
    this.errorMessage = null;

    this.staffApi.updateChild(this.childId, payload).subscribe({
      next: () => {
        this.isSaving = false;
        this.showEditForm = false;
        this.loadChild();
      },
      error: (error) => {
        this.isSaving = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.fieldErrors = mapped.fieldErrors;
        this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
      },
    });
  }

  linkGuardian(): void {
    if (!this.selectedGuardianId) return;

    this.isLinking = true;
    this.errorMessage = null;

    const payload: GuardianChildLinkWritePayload = {
      guardian_id: this.selectedGuardianId,
      child_id: this.childId,
    };

    this.staffApi.createGuardianChildLink(payload).subscribe({
      next: () => {
        this.isLinking = false;
        this.selectedGuardianId = '';
        this.loadChild();
        this.loadLinkedGuardians();
      },
      error: (error) => {
        this.isLinking = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
      },
    });
  }

  get availableGuardians(): GuardianRecord[] {
    const linkedIds = new Set(this.linkedGuardians.map(l => l.guardianId));
    return this.allGuardians.filter(g => !linkedIds.has(g.id));
  }

  private loadAll(): void {
    this.childId = this.route.snapshot.paramMap.get('childId') ?? '';
    if (!this.childId) return;
    this.loadChild();
  }

  private loadChild(): void {

    this.isLoadingChild = true;
    this.errorMessage = null;

    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.isLoadingChild = false;
        this.loadLinkedGuardians();
        this.loadAllGuardians();
      },
      error: (error) => {
        this.isLoadingChild = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
      },
    });
  }

  private loadLinkedGuardians(): void {
    this.isLoadingLinks = true;
    this.staffApi.listChildGuardianLinks(this.childId).subscribe({
      next: (links) => {
        this.linkedGuardians = links;
        this.isLoadingLinks = false;
      },
      error: (error) => {
        this.isLoadingLinks = false;
        const mapped = this.errorMapper.mapAndHandle(error);
        this.errorMessage = this.messageWithRequestId(mapped.message, mapped.requestId);
      },
    });
  }

  private loadAllGuardians(): void {
    this.staffApi.listGuardians({ status: 'active' as StatusFilter, limit: 200, offset: 0 }).subscribe({
      next: (guardians) => {
        this.allGuardians = guardians;
      },
    });
  }

  private formatCurrentMonth(): string {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    return `${year}-${month}`;
  }

  private messageWithRequestId(message: string, requestId: string | null): string {
    if (!requestId) return message;
    return `${message} (Request: ${requestId})`;
  }
}
