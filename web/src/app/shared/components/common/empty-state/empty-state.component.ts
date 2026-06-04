import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-empty-state',
  imports: [CommonModule],
  template: `
    <div
      class="text-center"
      [class.py-8]="!compact"
      [class.py-4]="compact"
    >
      @if (icon) {
        <div class="mb-3 text-gray-400 dark:text-gray-500" [innerHTML]="icon"></div>
      }
      <h3 class="text-sm font-medium text-gray-800 dark:text-white/90">{{ title }}</h3>
      @if (message) {
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ message }}</p>
      }
      <div class="mt-4">
        <ng-content></ng-content>
      </div>
    </div>
  `,
})
export class EmptyStateComponent {
  @Input() title = 'No data found';
  @Input() message = '';
  @Input() icon = '';
  @Input() compact = false;
}
