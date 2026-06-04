import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-table-shell',
  imports: [CommonModule],
  template: `
    <div
      class="rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]"
      [class]="className"
    >
      @if (title) {
        <div class="mb-4 flex items-center justify-between">
          <div>
            <h3 class="text-sm font-semibold text-gray-800 dark:text-white/90">{{ title }}</h3>
            @if (description) {
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ description }}</p>
            }
          </div>
          <ng-content select="[shell-actions]"></ng-content>
        </div>
      }
      <div class="overflow-x-auto">
        <ng-content></ng-content>
      </div>
      <ng-content select="[shell-footer]"></ng-content>
    </div>
  `,
})
export class TableShellComponent {
  @Input() title = '';
  @Input() description = '';
  @Input() className = '';
}
