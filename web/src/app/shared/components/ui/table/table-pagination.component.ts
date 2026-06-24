import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';

@Component({
  selector: 'app-table-pagination',
  imports: [CommonModule],
  template: `
    <div class="flex items-center justify-between gap-4">
      <p class="text-sm text-gray-500 dark:text-gray-400">
        Showing
        <span class="font-medium text-gray-700 dark:text-gray-300">{{ startItem }}</span>
        &ndash;
        <span class="font-medium text-gray-700 dark:text-gray-300">{{ endItem }}</span>
        of
        <span class="font-medium text-gray-700 dark:text-gray-300">{{ total }}</span>
      </p>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="inline-flex min-h-9 items-center justify-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40 dark:border-gray-700 dark:bg-transparent dark:text-gray-300 dark:hover:bg-white/5"
          [disabled]="offset === 0 || loading"
          (click)="previous.emit()"
        >
          ‹ Previous
        </button>
        <button
          type="button"
          class="inline-flex min-h-9 items-center justify-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3.5 text-sm font-medium text-gray-700 transition hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40 dark:border-gray-700 dark:bg-transparent dark:text-gray-300 dark:hover:bg-white/5"
          [disabled]="endItem >= total || loading"
          (click)="next.emit()"
        >
          Next ›
        </button>
      </div>
    </div>
  `,
})
export class TablePaginationComponent {
  @Input() offset = 0;
  @Input() limit = 25;
  @Input() total = 0;
  @Input() loading = false;

  @Output() previous = new EventEmitter<void>();
  @Output() next = new EventEmitter<void>();

  get startItem(): number {
    return this.total === 0 ? 0 : Math.min(this.offset + 1, this.total);
  }

  get endItem(): number {
    return Math.min(this.offset + this.limit, this.total);
  }
}
