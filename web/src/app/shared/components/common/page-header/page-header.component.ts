import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-page-header',
  imports: [CommonModule],
  template: `
    <div
      class="flex flex-col gap-4 rounded-2xl border border-gray-200 bg-white p-5 md:flex-row md:items-center md:justify-between dark:border-gray-800 dark:bg-white/[0.03]"
      [class]="className"
    >
      <div class="min-w-0">
        <h1 class="text-xl font-semibold text-gray-800 dark:text-white/90">{{ title }}</h1>
        @if (description) {
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ description }}</p>
        }
      </div>
      @if (hasActions) {
        <div class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row sm:flex-wrap sm:items-center sm:justify-end">
          <ng-content select="[actions]"></ng-content>
          <ng-content></ng-content>
        </div>
      }
    </div>
  `,
})
export class PageHeaderComponent {
  @Input() title = '';
  @Input() description = '';
  @Input() className = '';
  @Input() hasActions = false;
}
