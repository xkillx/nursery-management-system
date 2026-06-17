import { CommonModule } from '@angular/common';
import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import { StaffApiService } from '../../data/staff-api.service';
import { AuthService } from '../../../../core/services/auth.service';
import { ChildRecord } from '../../models/children.models';
import { ChildProfile } from '../../models/child-profile.models';
import { formatSiteRate, formatHourlyRateGbp, missingRequirementLabel } from '../../utils/manager-list-formatters';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ChildGuardianLinkRecord, GuardianRecord } from '../../models/guardians.models';

@Component({
  selector: 'app-manager-child-detail',
  imports: [
    CommonModule,
    RouterLink,
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
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  readonly formatRate = formatHourlyRateGbp;
  readonly formatSiteRate = formatSiteRate;
  readonly requirementLabel = missingRequirementLabel;

  childId = '';
  child: ChildRecord | null = null;
  profile: ChildProfile | null = null;
  linkedGuardians: ChildGuardianLinkRecord[] = [];
  allGuardians: GuardianRecord[] = [];

  isLoading = false;
  errorMessage: string | null = null;

  ngOnInit(): void {
    this.childId = this.route.snapshot.paramMap.get('childId') ?? '';
    this.load();
  }

  load(): void {
    if (!this.childId) {
      this.errorMessage = 'Missing child id.';
      return;
    }
    this.isLoading = true;
    this.staffApi.getChild(this.childId).subscribe({
      next: (child) => {
        this.child = child;
        this.loadLinkedGuardians();
        this.loadProfile();
      },
      error: (err) => {
        this.errorMessage = err?.message ?? 'Failed to load child.';
        this.isLoading = false;
      },
    });
  }

  private loadLinkedGuardians(): void {
    this.staffApi.listChildGuardianLinks(this.childId).subscribe({
      next: (links) => {
        this.linkedGuardians = links;
        this.loadAllGuardians();
      },
      error: () => {
        this.linkedGuardians = [];
        this.isLoading = false;
      },
    });
  }

  private loadAllGuardians(): void {
    this.staffApi.listGuardians({ status: 'all', limit: 200, offset: 0 }).subscribe({
      next: (response) => {
        this.allGuardians = response;
        this.isLoading = false;
      },
      error: () => {
        this.allGuardians = [];
        this.isLoading = false;
      },
    });
  }

  private loadProfile(): void {
    this.staffApi.getChildProfile(this.childId).subscribe({
      next: (profile) => {
        this.profile = profile;
      },
      error: () => {
        this.profile = null;
      },
    });
  }
}
