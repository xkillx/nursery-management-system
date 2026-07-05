import { CommonModule } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { NgIcon, provideIcons } from '@ng-icons/core';
import {
  heroCalendarDays,
  heroPlus,
  heroTrash,
  heroXMark,
} from '@ng-icons/heroicons/outline';

import { EmptyStateComponent } from '../../../../shared/components/common/empty-state/empty-state.component';
import { LoadingStateComponent } from '../../../../shared/components/common/loading-state/loading-state.component';
import { ConfirmationDialogComponent } from '../../../../shared/components/ui/modal/confirmation-dialog.component';
import { AuthService } from '../../../../core/services/auth.service';
import { ToastService } from '../../../../shared/services/toast.service';
import { ClosureDaysApiService } from '../../data/closure-days-api.service';
import { ClosureDay } from '../../models/closure-day.models';

@Component({
  selector: 'app-manager-closure-days',
  imports: [
    CommonModule,
    FormsModule,
    NgIcon,
    LoadingStateComponent,
    EmptyStateComponent,
    ConfirmationDialogComponent,
  ],
  templateUrl: './manager-closure-days.component.html',
  providers: [
    provideIcons({
      heroCalendarDays,
      heroPlus,
      heroTrash,
      heroXMark,
    }),
  ],
})
export class ManagerClosureDaysComponent implements OnInit {
  private readonly api = inject(ClosureDaysApiService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);

  siteId: string | null = null;
  loading = false;
  closureDays = signal<ClosureDay[]>([]);

  formDate = '';
  formReason = '';
  formSaving = false;
  formError: string | null = null;

  isConfirmDeleteOpen = false;
  dayToDelete: ClosureDay | null = null;
  deleteSaving = false;

  ngOnInit(): void {
    const membership = this.auth.activeMembership();
    if (!membership?.branch_id) {
      this.toast.error('No site is attached to this manager session.');
      return;
    }
    this.siteId = membership.branch_id;
    this.loadClosureDays();
  }

  loadClosureDays(): void {
    if (!this.siteId) return;
    this.loading = true;
    const year = new Date().getFullYear();
    const from = `${year}-01-01`;
    const to = `${year}-12-31`;
    this.api.list(this.siteId, from, to).subscribe({
      next: (days) => {
        this.closureDays.set(days);
        this.loading = false;
      },
      error: () => {
        this.loading = false;
        this.toast.error('Failed to load closure days.');
      },
    });
  }

  addClosureDay(form: NgForm): void {
    if (!this.siteId || !this.formDate) return;
    this.formSaving = true;
    this.formError = null;
    this.api.create(this.siteId, this.formDate, this.formReason || undefined).subscribe({
      next: () => {
        this.formSaving = false;
        this.toast.success('Closure day added.');
        this.formDate = '';
        this.formReason = '';
        form.resetForm();
        this.loadClosureDays();
      },
      error: (err) => {
        this.formSaving = false;
        const body = err?.error;
        this.formError = body?.message ?? 'Failed to add closure day.';
      },
    });
  }

  confirmDelete(day: ClosureDay): void {
    this.dayToDelete = day;
    this.isConfirmDeleteOpen = true;
  }

  doDelete(): void {
    if (!this.siteId || !this.dayToDelete) return;
    this.deleteSaving = true;
    this.api.delete(this.siteId, this.dayToDelete.id).subscribe({
      next: () => {
        this.deleteSaving = false;
        this.isConfirmDeleteOpen = false;
        this.dayToDelete = null;
        this.toast.success('Closure day removed.');
        this.loadClosureDays();
      },
      error: () => {
        this.deleteSaving = false;
        this.isConfirmDeleteOpen = false;
        this.dayToDelete = null;
        this.toast.error('Failed to remove closure day.');
      },
    });
  }

  cancelDelete(): void {
    this.isConfirmDeleteOpen = false;
    this.dayToDelete = null;
  }
}
