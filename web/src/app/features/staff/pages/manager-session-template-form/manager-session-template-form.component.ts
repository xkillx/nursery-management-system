import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArrowLeft,
  heroClipboardDocumentList,
  heroInformationCircle,
  heroPencilSquare,
  heroPlus,
  heroTrash,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { ROLE_ROUTES } from '../../../../core/constants/roles';
import { AuthService } from '../../../../core/services/auth.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { TextAreaComponent } from '../../../../shared/components/form/input/text-area.component';
import { SelectComponent, type Option } from '../../../../shared/components/form/select/select.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { StaffSessionTemplatesApiService } from '../../data/session-templates-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import { SessionTemplateInput } from '../../models/session-template.models';

interface DayEntry {
  sessionTypeId: string;
}

const DAY_LABELS: Record<number, string> = {
  1: 'Mon',
  2: 'Tue',
  3: 'Wed',
  4: 'Thu',
  5: 'Fri',
};

@Component({
  selector: 'app-manager-session-template-form',
  imports: [
    CommonModule,
    FormsModule,
    RouterModule,
    NgIcon,
    AlertComponent,
    FormFieldComponent,
    InputFieldComponent,
    TextAreaComponent,
    SelectComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-session-template-form.component.html',
  providers: [
    provideIcons({
      heroArrowLeft,
      heroClipboardDocumentList,
      heroInformationCircle,
      heroPencilSquare,
      heroPlus,
      heroTrash,
      heroXMark,
    }),
  ],
})
export class ManagerSessionTemplateFormComponent implements OnInit {
  private readonly templatesApi = inject(StaffSessionTemplatesApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  mode: 'create' | 'edit' = 'create';
  templateId: string | null = null;
  siteId: string | null = null;
  siteName = '';
  loading = false;
  saving = false;
  pageError: string | null = null;
  sessionTypes: StaffSessionType[] = [];

  readonly listRoute = ROLE_ROUTES.managerSessionTemplates;

  formName = '';
  formDescription = '';
  formEntries: Record<number, DayEntry[]> = { 1: [], 2: [], 3: [], 4: [], 5: [] };
  formFieldErrors: { name?: string; description?: string; entries?: string } = {};

  readonly dayOptions: Option[] = (Object.keys(DAY_LABELS) as unknown as number[])
    .sort((a, b) => a - b)
    .map((d) => ({ value: String(d), label: DAY_LABELS[d] }));

  get sessionTypeOptions(): Option[] {
    return this.sessionTypes
      .filter((t) => t.isActive)
      .map((t) => ({ value: t.id, label: `${t.name} (${t.startTime}-${t.endTime})` }));
  }

  get hasAnyEntry(): boolean {
    for (const d of [1, 2, 3, 4, 5, 6, 7]) {
      if ((this.formEntries[d] ?? []).length > 0) return true;
    }
    return false;
  }

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (membership?.branch_id) {
      this.siteId = membership.branch_id;
      this.siteName = membership.branch_name ?? 'Assigned site';
    }

    const id = this.route.snapshot.paramMap.get('templateId');
    if (id) {
      this.mode = 'edit';
      this.templateId = id;
      this.loadExisting(id);
    }

    this.loadSessionTypes();
  }

  private loadSessionTypes(): void {
    if (!this.siteId) return;
    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => {
        this.sessionTypes = types;
      },
      error: () => {
        this.sessionTypes = [];
      },
    });
  }

  private loadExisting(id: string): void {
    if (!this.siteId) {
      this.pageError = 'No site available.';
      return;
    }
    this.loading = true;
    this.templatesApi.getSessionTemplate(this.siteId, id).subscribe({
      next: (t) => {
        this.formName = t.name;
        this.formDescription = t.description ?? '';
        this.formEntries = { 1: [], 2: [], 3: [], 4: [], 5: [] };
        for (const e of t.entries) {
          const day = e.dayOfWeek;
          if (day >= 1 && day <= 5) {
            this.formEntries[day].push({ sessionTypeId: e.sessionType.id });
          }
        }
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = err?.error?.message ?? 'Failed to load template.';
      },
    });
  }

  addEntry(day: number): void {
    this.formEntries[day] = [...(this.formEntries[day] ?? []), { sessionTypeId: '' }];
  }

  removeEntry(day: number, index: number): void {
    const arr = [...(this.formEntries[day] ?? [])];
    arr.splice(index, 1);
    this.formEntries[day] = arr;
  }

  setEntryType(day: number, index: number, value: string): void {
    const arr = [...(this.formEntries[day] ?? [])];
    arr[index] = { sessionTypeId: value };
    this.formEntries[day] = arr;
  }

  copyMonToAllWeekdays(): void {
    const monday = this.formEntries[1] ?? [];
    this.formEntries = {
      1: [...monday],
      2: [...monday],
      3: [...monday],
      4: [...monday],
      5: [...monday],
    };
  }

  clearAllEntries(): void {
    this.formEntries = { 1: [], 2: [], 3: [], 4: [], 5: [] };
  }

  entriesForDay(day: number): DayEntry[] {
    return this.formEntries[day] ?? [];
  }

  buildPayload(): SessionTemplateInput {
    const entries: { dayOfWeek: number; sessionTypeId: string }[] = [];
    for (let d = 1; d <= 5; d++) {
      for (const e of this.entriesForDay(d)) {
        if (e.sessionTypeId) {
          entries.push({ dayOfWeek: d, sessionTypeId: e.sessionTypeId });
        }
      }
    }
    return {
      name: this.formName.trim(),
      description: this.formDescription.trim() ? this.formDescription.trim() : null,
      entries,
    };
  }

  onSubmit(): void {
    this.formFieldErrors = {};
    this.pageError = null;

    if (!this.siteId) {
      this.pageError = 'No site available.';
      return;
    }

    const payload = this.buildPayload();
    if (!payload.name) {
      this.formFieldErrors.name = 'Name is required.';
      return;
    }
    if (payload.entries.length === 0) {
      this.formFieldErrors.entries = 'Add at least one booked session.';
      return;
    }

    this.saving = true;
    const op =
      this.mode === 'edit' && this.templateId
        ? this.templatesApi.updateSessionTemplate(this.siteId, this.templateId, payload)
        : this.templatesApi.createSessionTemplate(this.siteId, payload);

    op.subscribe({
      next: () => {
        this.saving = false;
        this.router.navigateByUrl(this.listRoute);
      },
      error: (err) => {
        this.saving = false;
        const mapped = this.errorMapper.mapAndHandle(err);
        const fieldErrors = mapped.fieldErrors ?? {};
        this.formFieldErrors = {
          name: fieldErrors['name'],
          description: fieldErrors['description'],
          entries: fieldErrors['entries'],
        };
        this.pageError = formatPresentedApiError(presentApiError(mapped, 'sessionTemplates.create')) ?? 'Could not save template.';
      },
    });
  }
}
