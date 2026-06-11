import { Component, OnInit, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';

import { StaffApiService } from '../../data/staff-api.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { PageHeaderComponent } from '../../../../shared/components/common/page-header/page-header.component';
import { StatusBadgeComponent } from '../../../../shared/components/ui/badge/status-badge.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import {
  RegistrationProfileResponse, RegistrationOfficeUseChecklistResponse,
  OfficeUseChecklist, RegistrationContactEntry,
} from '../../models/registration-profile.models';
import {
  formatProfileSectionLabel, formatProfileMissingFieldLabel,
  formatOfficeItemLabel, formatOfficeCheckStatusLabel,
  getCompletionBadgeClass, formatCompletionStatus,
} from '../../utils/registration-profile-formatters';

@Component({
  selector: 'app-manager-child-registration',
  standalone: true,
  imports: [
    RouterLink, DatePipe, FormsModule,
    AlertComponent, LoadingStateComponent, PageHeaderComponent,
    StatusBadgeComponent, ButtonComponent,
  ],
  templateUrl: './manager-child-registration.component.html',
})
export class ManagerChildRegistrationComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private api = inject(StaffApiService);
  private errorMapper = inject(ApiErrorMapper);

  childId: string | null = null;

  isLoading = true;
  isSavingSection: string | null = null;
  errorMessage: string | null = null;

  profile: RegistrationProfileResponse | null = null;
  checklist: RegistrationOfficeUseChecklistResponse | null = null;

  collectionPassword = '';
  collectionPasswordMessage: string | null = null;
  collectionPasswordError: string | null = null;

  gdprName = '';
  gdprDate = '';
  gdprMessage: string | null = null;
  gdprError: string | null = null;

  sectionMessages: Record<string, string> = {};
  sectionErrors: Record<string, string> = {};

  ngOnInit(): void {
    this.childId = this.route.snapshot.paramMap.get('childId');
    if (!this.childId) {
      this.errorMessage = 'Child ID is required.';
      this.isLoading = false;
      return;
    }
    this.loadAll();
  }

  private loadAll(): void {
    this.isLoading = true;
    this.errorMessage = null;

    this.api.getRegistrationProfile(this.childId!).subscribe({
      next: (profile) => {
        this.profile = profile;
        this.prefillGdpr();
        this.loadChecklist();
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.errorMessage = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isLoading = false;
      },
    });
  }

  private loadChecklist(): void {
    this.api.getRegistrationOfficeUseChecklist(this.childId!).subscribe({
      next: (checklist) => {
        this.checklist = checklist;
        this.isLoading = false;
      },
      error: () => {
        this.checklist = null;
        this.isLoading = false;
      },
    });
  }

  get childName(): string {
    return this.profile?.child?.fullName ?? '';
  }

  get childDob(): string {
    return this.profile?.child?.dateOfBirth ?? '';
  }

  get profileCompleteness() {
    return this.profile?.completeness ?? null;
  }

  get officeCompleteness() {
    return this.checklist?.completeness ?? null;
  }

  get profileCompletionBadge(): string {
    return getCompletionBadgeClass(this.profileCompleteness?.isComplete ?? false);
  }

  get officeCompletionBadge(): string {
    return getCompletionBadgeClass(this.officeCompleteness?.isComplete ?? false);
  }

  saveSection(sectionKey: string, patch: Record<string, unknown>): void {
    if (!this.childId) return;

    this.isSavingSection = sectionKey;
    this.sectionMessages[sectionKey] = '';
    this.sectionErrors[sectionKey] = '';

    this.api.patchRegistrationProfile(this.childId, patch).subscribe({
      next: (updated) => {
        this.profile = updated;
        this.isSavingSection = null;
        this.sectionMessages[sectionKey] = 'Section saved.';
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.sectionErrors[sectionKey] = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  saveContacts(contactType: string, contacts: RegistrationContactEntry[]): void {
    this.saveSection(contactType, { [contactType]: contacts });
  }

  saveOfficeChecklist(patch: Partial<OfficeUseChecklist>): void {
    if (!this.childId) return;

    this.isSavingSection = 'office_use_checklist';
    this.sectionMessages['office_use_checklist'] = '';
    this.sectionErrors['office_use_checklist'] = '';

    this.api.patchRegistrationOfficeUseChecklist(this.childId, patch).subscribe({
      next: (updated) => {
        this.checklist = updated;
        this.isSavingSection = null;
        this.sectionMessages['office_use_checklist'] = 'Office-use checklist saved.';
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.sectionErrors['office_use_checklist'] = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  setCollectionPassword(): void {
    if (!this.childId || !this.collectionPassword) return;

    this.collectionPasswordMessage = '';
    this.collectionPasswordError = '';
    this.isSavingSection = 'collection';

    this.api.setRegistrationCollectionPassword(this.childId, this.collectionPassword).subscribe({
      next: (updated) => {
        this.profile = updated;
        this.collectionPassword = '';
        this.collectionPasswordMessage = 'Collection password set.';
        this.isSavingSection = null;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.collectionPasswordError = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  saveGdprDeclaration(): void {
    if (!this.childId) return;

    this.gdprMessage = '';
    this.gdprError = '';
    this.isSavingSection = 'gdpr';

    const patch: Record<string, unknown> = {
      gdpr_declaration: {
        gdpr_declared_by_name: this.gdprName || null,
        gdpr_declaration_date: this.gdprDate || null,
      },
    };

    this.api.patchRegistrationProfile(this.childId, patch).subscribe({
      next: (updated) => {
        this.profile = updated;
        this.prefillGdpr();
        this.gdprMessage = 'GDPR declaration saved.';
        this.isSavingSection = null;
      },
      error: (err) => {
        const mapped = this.errorMapper.mapAndHandle(err);
        this.gdprError = formatPresentedApiError(presentApiError(mapped, 'people.child'));
        this.isSavingSection = null;
      },
    });
  }

  private prefillGdpr(): void {
    const gdpr = this.profile?.gdprDeclaration;
    this.gdprName = gdpr?.gdprDeclaredByName ?? '';
    this.gdprDate = gdpr?.gdprDeclarationDate ?? '';
  }

  protected readonly formatProfileSectionLabel = formatProfileSectionLabel;
  protected readonly formatProfileMissingFieldLabel = formatProfileMissingFieldLabel;
  protected readonly formatOfficeItemLabel = formatOfficeItemLabel;
  protected readonly formatOfficeCheckStatusLabel = formatOfficeCheckStatusLabel;
  protected readonly formatCompletionStatus = formatCompletionStatus;
}
