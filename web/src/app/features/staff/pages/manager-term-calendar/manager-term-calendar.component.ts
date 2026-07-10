import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendar,
  heroCheck,
  heroChevronDown,
  heroClock,
  heroInformationCircle,
  heroPencilSquare,
  heroPlus,
  heroTrash,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { BadgeComponent } from '../../../../shared/components/ui/badge/badge.component';
import { SelectComponent } from '../../../../shared/components/form/select/select.component';
import { DatePickerComponent } from '../../../../shared/components/form/date-picker/date-picker.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { AcademicTermsApiService } from '../../data/academic-terms-api.service';
import { AcademicTerm, AcademicTermKind, AcademicTermInput } from '../../models/academic-term.models';

const KIND_OPTIONS: { value: AcademicTermKind; label: string }[] = [
  { value: 'autumn', label: 'Autumn' },
  { value: 'spring', label: 'Spring' },
  { value: 'summer', label: 'Summer' },
];

@Component({
  selector: 'app-manager-term-calendar',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    AlertComponent,
    BadgeComponent,
    SelectComponent,
    DatePickerComponent,
    ConfirmationDialogComponent,
  ],
  templateUrl: './manager-term-calendar.component.html',
  providers: [
    provideIcons({
      heroCalendar,
      heroCheck,
      heroChevronDown,
      heroClock,
      heroInformationCircle,
      heroPencilSquare,
      heroPlus,
      heroTrash,
      heroXMark,
    }),
  ],
})
export class ManagerTermCalendarComponent implements OnInit {
  private readonly api = inject(AcademicTermsApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  siteName = '';
  loading = false;
  includeArchived = false;
  terms = signal<AcademicTerm[]>([]);

  totalTermsCount = computed(() => this.terms().length);
  activeTermsCount = computed(() => this.terms().filter(t => t.is_active).length);
  archivedTermsCount = computed(() => this.terms().filter(t => !t.is_active).length);

  editorMode: 'closed' | 'create' | 'edit' = 'closed';
  editingTermId: string | null = null;
  form: AcademicTermInput = { name: '', kind: 'autumn', start_date: '', end_date: '' };
  formSaving = false;
  formError: string | null = null;
  formFieldErrors: { name?: string; kind?: string; start_date?: string; end_date?: string } = {};

  isConfirmArchiveOpen = false;
  termToArchive: AcademicTerm | null = null;
  archiveSaving = false;

  readonly kindOptions = KIND_OPTIONS;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.toast.error('No site is attached to this manager session.');
      return;
    }
    this.siteId = membership.branch_id;
    this.siteName = membership.branch_name ?? 'Assigned site';
    this.loadTerms();
  }

  loadTerms(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.api.listTerms(this.siteId, { includeArchived: this.includeArchived }).subscribe({
      next: (terms) => {
        this.terms.set(terms);
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.toast.error('Failed to load academic terms.');
      },
    });
  }

  toggleIncludeArchived(): void {
    this.includeArchived = !this.includeArchived;
    this.loadTerms();
  }

  openCreate(): void {
    this.editorMode = 'create';
    this.editingTermId = null;
    this.form = { name: '', kind: 'autumn', start_date: '', end_date: '' };
    this.formError = null;
    this.formFieldErrors = {};
  }

  openEdit(term: AcademicTerm): void {
    this.editorMode = 'edit';
    this.editingTermId = term.id;
    this.form = { name: term.name, kind: term.kind, start_date: term.start_date, end_date: term.end_date };
    this.formError = null;
    this.formFieldErrors = {};
  }

  closeEditor(): void {
    this.editorMode = 'closed';
    this.editingTermId = null;
    this.formError = null;
    this.formFieldErrors = {};
  }

  save(): void {
    if (!this.siteId) return;
    this.formSaving = true;
    this.formError = null;
    this.formFieldErrors = {};

    const req$ = this.editorMode === 'edit' && this.editingTermId
      ? this.api.updateTerm(this.siteId, this.editingTermId, this.form)
      : this.api.createTerm(this.siteId, this.form);

    req$.subscribe({
      next: () => {
        this.formSaving = false;
        this.toast.success(this.editorMode === 'edit' ? 'Term updated.' : 'Term created.');
        this.closeEditor();
        this.loadTerms();
      },
      error: (err) => {
        this.formSaving = false;
        const body = err?.error;
        if (body?.code === 'validation_error' && body?.fields) {
          const fields = body.fields as Record<string, string>;
          this.formFieldErrors = {
            name: fields['name'],
            kind: fields['kind'],
            start_date: fields['start_date'],
            end_date: fields['end_date'],
          };
          this.formError = 'Please correct the highlighted fields.';
        } else {
          this.formError = body?.message ?? 'Failed to save term.';
        }
      },
    });
  }

  archive(term: AcademicTerm): void {
    this.termToArchive = term;
    this.isConfirmArchiveOpen = true;
  }

  confirmArchive(): void {
    if (!this.siteId || !this.termToArchive) return;
    this.archiveSaving = true;
    this.api.archiveTerm(this.siteId, this.termToArchive.id).subscribe({
      next: () => {
        this.archiveSaving = false;
        this.isConfirmArchiveOpen = false;
        this.termToArchive = null;
        this.toast.success('Term archived.');
        this.loadTerms();
      },
      error: () => {
        this.archiveSaving = false;
        this.isConfirmArchiveOpen = false;
        this.termToArchive = null;
        this.toast.error('Failed to archive term.');
      },
    });
  }

  cancelArchive(): void {
    this.isConfirmArchiveOpen = false;
    this.termToArchive = null;
  }
}
