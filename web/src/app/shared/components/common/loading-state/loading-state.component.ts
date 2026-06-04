import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-loading-state',
  imports: [CommonModule],
  template: `
    @if (variant === 'table-row') {
      <tr>
        <td [attr.colspan]="colspan" class="py-3 text-center text-gray-500 dark:text-gray-400">
          <div class="flex items-center justify-center gap-2">
            <svg class="h-4 w-4 animate-spin text-gray-400" viewBox="0 0 24 24" fill="none">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            {{ label || 'Loading...' }}
          </div>
        </td>
      </tr>
    } @else if (variant === 'inline') {
      <span class="inline-flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
        <svg class="h-3.5 w-3.5 animate-spin text-gray-400" viewBox="0 0 24 24" fill="none">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        {{ label || 'Loading...' }}
      </span>
    } @else {
      <div class="py-8 text-center text-gray-500 dark:text-gray-400">
        <div class="flex items-center justify-center gap-2">
          <svg class="h-5 w-5 animate-spin text-gray-400" viewBox="0 0 24 24" fill="none">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          {{ label || 'Loading...' }}
        </div>
      </div>
    }
  `,
})
export class LoadingStateComponent {
  @Input() label = 'Loading...';
  @Input() rows = 1;
  @Input() variant: 'block' | 'table-row' | 'inline' = 'block';
  @Input() colspan = 1;
}
