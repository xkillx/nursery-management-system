import { CommonModule } from '@angular/common';
import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroArchiveBox,
  heroArrowPath,
  heroClipboardDocumentList,
  heroPlus,
  heroTrash,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { ApiErrorMapper } from '../../../../core/errors/api-error.mapper';
import { presentApiError, formatPresentedApiError } from '../../../../core/errors/api-error-presenter';
import { AuthService } from '../../../../core/services/auth.service';
import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { ButtonComponent } from '../../../../shared/components/ui/button/button.component';
import { FormFieldComponent } from '../../../../shared/components/form/form-field/form-field.component';
import { InputFieldComponent } from '../../../../shared/components/form/input/input-field.component';
import { TextAreaComponent } from '../../../../shared/components/form/input/text-area.component';
import { SelectComponent, type Option } from '../../../../shared/components/form/select/select.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import {
  StaffSessionTemplatesApiService,
} from '../../data/session-templates-api.service';
import { StaffSessionTypesApiService, StaffSessionType } from '../../data/session-types-api.service';
import {
  SessionTemplate,
  SessionTemplateInput,
  SessionTemplateListItem,
} from '../../models/session-template.models';

type EditorMode = 'closed' | 'create' | 'edit';

type DayEntry = {
  sessionTypeId: string;
};

const DAY_LABELS: Record<number, string> = {
  1: 'Mon',
  2: 'Tue',
  3: 'Wed',
  4: 'Thu',
  5: 'Fri',
};

@Component({
  selector: 'app-manager-session-templates',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    AlertComponent,
    ButtonComponent,
    FormFieldComponent,
    InputFieldComponent,
    TextAreaComponent,
    SelectComponent,
    LoadingStateComponent,
  ],
  templateUrl: './manager-session-templates.component.html',
  providers: [
    provideIcons({
      heroArchiveBox,
      heroArrowPath,
      heroClipboardDocumentList,
      heroPlus,
      heroTrash,
      heroXMark,
    }),
  ],
})
export class ManagerSessionTemplatesComponent implements OnInit {
  private readonly templatesApi = inject(StaffSessionTemplatesApiService);
  private readonly sessionTypesApi = inject(StaffSessionTypesApiService);
  private readonly auth = inject(AuthService);
  private readonly errorMapper = inject(ApiErrorMapper);

  siteId: string | null = null;
  siteName = '';
  loading = false;
  pageError: string | null = null;
  mutatingId: string | null = null;
  includeArchived = false;
  templates: SessionTemplateListItem[] = [];
  sessionTypes: StaffSessionType[] = [];

  editorMode: EditorMode = 'closed';
  editingTemplateId: string | null = null;
  formName = '';
  formDescription = '';
  formEntries: Record<number, DayEntry[]> = { 1: [], 2: [], 3: [], 4: [], 5: [] };
  formSaving = false;
  formError: string | null = null;
  formFieldErrors: { name?: string; description?: string; entries?: string } = {};


  readonly dayOptions: Option[] = (Object.keys(DAY_LABELS) as unknown as number[])
    .sort((a, b) => a - b)
    .map((d) => ({ value: String(d), label: DAY_LABELS[d] }));

  get sessionTypeOptions(): Option[] {
    return this.sessionTypes
      .filter((t) => t.isActive)
      .map((t) => ({ value: t.id, label: `${t.name} (${t.startTime}-${t.endTime})` }));
  }

  get visibleRows(): SessionTemplateListItem[] {
    return this.templates;
  }

  get isEditorOpen(): boolean {
    return this.editorMode !== 'closed';
  }

  get editorTitle(): string {
    return this.editorMode === 'edit' ? 'Edit session template' : 'New session template';
  }

  get editorSubmitLabel(): string {
    return this.editorMode === 'edit' ? 'Save changes' : 'Create template';
  }

  get hasAnyEntry(): boolean {
    for (const d of [1, 2, 3, 4, 5, 6, 7]) {
      if ((this.formEntries[d] ?? []).length > 0) return true;
    }
    return false;
  }

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.pageError = 'No site is attached to this manager session.';
      return;
    }
    this.siteId = membership.branch_id;
    this.siteName = membership.branch_name ?? 'Assigned site';
    this.loadAll();
  }

  private loadAll(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.pageError = null;

    this.templatesApi.listSessionTemplates(this.siteId, { includeArchived: this.includeArchived }).subscribe({
      next: (templates) => {
        this.templates = templates;
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = this.formatError(err, 'Failed to load session templates.');
      },
    });

    this.sessionTypesApi.listSessionTypes(this.siteId, { includeArchived: false }).subscribe({
      next: (types) => {
        this.sessionTypes = types;
      },
      error: (err) => {
        // Non-fatal: the form will show a helpful error when the user opens it.
        this.sessionTypes = [];
        this.pageError = this.pageError ?? this.formatError(err, 'Failed to load session types.');
      },
    });
  }

  reload(): void {
    this.loadAll();
  }

  toggleIncludeArchived(): void {
    this.includeArchived = !this.includeArchived;
    this.loadAll();
  }

  openCreate(): void {
    this.editorMode = 'create';
    this.editingTemplateId = null;
    this.formName = '';
    this.formDescription = '';
    this.formEntries = { 1: [], 2: [], 3: [], 4: [], 5: [], 6: [], 7: [] };
    this.formError = null;
    this.formFieldErrors = {};
  }

  openEdit(t: SessionTemplateListItem): void {
    if (!this.siteId) return;
    this.loading = true;
    this.templatesApi.getSessionTemplate(this.siteId, t.id).subscribe({
      next: (full) => {
        this.loading = false;
        this.populateEditor(full);
      },
      error: (err) => {
        this.loading = false;
        this.pageError = this.formatError(err, 'Failed to load template.');
      },
    });
  }

  private populateEditor(t: SessionTemplate): void {
    this.editorMode = 'edit';
    this.editingTemplateId = t.id;
    this.formName = t.name;
    this.formDescription = t.description ?? '';
    this.formEntries = { 1: [], 2: [], 3: [], 4: [], 5: [], 6: [], 7: [] };
    for (const e of t.entries) {
      const day = e.dayOfWeek;
      if (day >= 1 && day <= 7) {
        this.formEntries[day].push({ sessionTypeId: e.sessionType.id });
      }
    }
    this.formError = null;
    this.formFieldErrors = {};
  }

  closeEditor(): void {
    this.editorMode = 'closed';
    this.editingTemplateId = null;
    this.formError = null;
    this.formFieldErrors = {};
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

  submit(): void {
    this.formError = null;
    this.formFieldErrors = {};

    if (!this.siteId) return;
    if (!this.siteId) return;
    if (!this.siteId) return;
    const payload = this.buildPayload();
    if (!payload.name) {
      this.formFieldErrors.name = 'Name is required.';
      return;
    }
    if (payload.entries.length === 0) {
      this.formFieldErrors.entries = 'Add at least one booked session.';
      return;
    }

    this.formSaving = true;
    const op =
      this.editorMode === 'edit' && this.editingTemplateId
        ? this.templatesApi.updateSessionTemplate(this.siteId, this.editingTemplateId, payload)
        : this.templatesApi.createSessionTemplate(this.siteId, payload);

    op.subscribe({
      next: () => {
        this.formSaving = false;
        this.closeEditor();
        this.reload();
      },
      error: (err) => {
        this.formSaving = false;
        const mapped = this.errorMapper.mapAndHandle(err);
        const fieldErrors = mapped.fieldErrors ?? {};
        this.formFieldErrors = {
          name: fieldErrors['name'],
          description: fieldErrors['description'],
          entries: fieldErrors['entries'],
        };
        this.formError = formatPresentedApiError(presentApiError(mapped, 'sessionTemplates.create')) ?? 'Could not save template.';
      },
    });
  }

  archive(t: SessionTemplateListItem): void {
    if (!this.siteId || !t.isActive) return;
    if (!confirm(`Archive "${t.name}"? It can no longer be used when creating new booking patterns.`)) return;
    this.mutatingId = t.id;
    this.templatesApi.archiveSessionTemplate(this.siteId, t.id).subscribe({
      next: () => {
        this.mutatingId = null;
        this.reload();
      },
      error: (err) => {
        this.mutatingId = null;
        this.pageError = this.formatError(err, 'Failed to archive template.');
      },
    });
  }

  reactivate(t: SessionTemplateListItem): void {
    if (!this.siteId || t.isActive) return;
    this.mutatingId = t.id;
    this.templatesApi.reactivateSessionTemplate(this.siteId, t.id).subscribe({
      next: () => {
        this.mutatingId = null;
        this.reload();
      },
      error: (err) => {
        this.mutatingId = null;
        this.pageError = this.formatError(err, 'Failed to reactivate template.');
      },
    });
  }

  formatEntriesForList(t: SessionTemplateListItem): string {
    if (!t.entries) return '';
    return `${(t as unknown as { entryCount?: number }).entryCount ?? ''}`;
  }

  formatError(err: unknown, fallback: string): string {
    const mapped = this.errorMapper.mapAndHandle(err);
    return formatPresentedApiError(presentApiError(mapped, 'sessionTemplates.list')) ?? fallback;
  }
}
