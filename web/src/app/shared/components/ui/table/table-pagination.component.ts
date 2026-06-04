import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';

@Component({
  selector: 'app-table-pagination',
  imports: [CommonModule],
  template: `
    <div class="mt-4 flex items-center justify-end gap-2">
      <button
        type="button"
        class="rounded border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-white/5"
        [disabled]="offset === 0 || loading"
        (click)="previous.emit()"
      >
        Previous
      </button>
      <span class="text-xs text-gray-500 dark:text-gray-400">
        Offset {{ offset }} / Limit {{ limit }}
      </span>
      <button
        type="button"
        class="rounded border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-white/5"
        [disabled]="itemCount < limit || loading"
        (click)="next.emit()"
      >
        Next
      </button>
    </div>
  `,
})
export class TablePaginationComponent {
  @Input() offset = 0;
  @Input() limit = 10;
  @Input() itemCount = 0;
  @Input() loading = false;

  @Output() previous = new EventEmitter<void>();
  @Output() next = new EventEmitter<void>();
}
