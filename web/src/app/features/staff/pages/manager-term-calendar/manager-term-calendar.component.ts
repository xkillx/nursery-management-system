import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendar,
  heroCheck,
  heroChevronDown,
  heroClock,
  heroPencilSquare,
  heroPlus,
  heroTrash,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { AlertComponent } from '../../../../shared/components/ui/alert/alert.component';
import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
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
  ],
  templateUrl: './manager-term-calendar.component.html',
  providers: [
    provideIcons({
      heroCalendar,
      heroCheck,
      heroChevronDown,
      heroClock,
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
  pageError: string | null = null;
  includeArchived = false;
  terms = signal<AcademicTerm[]>([]);

  editorMode: 'closed' | 'create' | 'edit' = 'closed';
  editingTermId: string | null = null;
  form: AcademicTermInput = { name: '', kind: 'autumn', start_date: '', end_date: '' };
  formSaving = false;
  formError: string | null = null;
  formFieldErrors: { name?: string; kind?: string; start_date?: string; end_date?: string } = {};

  readonly kindOptions = KIND_OPTIONS;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.pageError = 'No site is attached to this manager session.';
      return;
    }
    this.siteId = membership.branch_id;
    this.siteName = membership.branch_name ?? 'Assigned site';
    this.loadTerms();
  }

  loadTerms(): void {
    if (!this.siteId) return;
    this.loading = true;
    this.pageError = null;
    this.api.listTerms(this.siteId, { includeArchived: this.includeArchived }).subscribe({
      next: (terms) => {
        this.terms.set(terms);
        this.loading = false;
      },
      error: (err) => {
        this.loading = false;
        this.pageError = 'Failed to load academic terms.';
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
    if (!this.siteId) return;
    this.api.archiveTerm(this.siteId, term.id).subscribe({
      next: () => {
        this.toast.success('Term archived.');
        this.loadTerms();
      },
      error: () => {
        this.toast.error('Failed to archive term.');
      },
    });
  }
}
