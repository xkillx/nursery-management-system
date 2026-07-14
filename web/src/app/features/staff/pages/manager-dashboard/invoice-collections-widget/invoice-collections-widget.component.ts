import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { ManagerInvoicesApiService } from '../../../data/manager-invoices-api.service';
import { OverdueSummary } from '../../../models/manager-invoices.models';
import { formatGbp } from '../manager-dashboard.models';
import { ChildAvatarComponent } from '../../../../../shared/components/ui/avatar/child-avatar/child-avatar.component';
import { EmptyStateComponent } from '../../../../../shared/components/common/empty-state/empty-state.component';

type LoadState = 'loading' | 'ready' | 'error';

@Component({
  selector: 'app-invoice-collections-widget',
  standalone: true,
  imports: [CommonModule, RouterLink, ChildAvatarComponent, EmptyStateComponent],
  templateUrl: './invoice-collections-widget.component.html',
})
export class InvoiceCollectionsWidgetComponent {
  private readonly api = inject(ManagerInvoicesApiService);

  readonly state = signal<LoadState>('loading');
  readonly summary = signal<OverdueSummary | null>(null);
  readonly errorMessage = signal('');

  constructor() {
    this.load();
  }

  load(): void {
    this.state.set('loading');
    this.api.getOverdueSummary().subscribe({
      next: (data) => {
        this.summary.set(data);
        this.state.set('ready');
      },
      error: () => {
        this.errorMessage.set('Failed to load overdue invoices.');
        this.state.set('error');
      },
    });
  }

  formatGbp(minorUnits: number): string {
    return formatGbp(minorUnits);
  }

  daysOverdueClass(days: number): string {
    if (days >= 15) {
      return 'bg-error-50 text-error-700 border border-error-100 dark:bg-error-500/15 dark:text-error-400 dark:border-error-500/20';
    }
    if (days >= 1) {
      return 'bg-warning-50 text-warning-700 border border-warning-100 dark:bg-warning-500/15 dark:text-warning-400 dark:border-warning-500/20';
    }
    return 'bg-gray-50 text-gray-600 border border-gray-200 dark:bg-gray-500/10 dark:text-gray-400 dark:border-gray-500/20';
  }
}
