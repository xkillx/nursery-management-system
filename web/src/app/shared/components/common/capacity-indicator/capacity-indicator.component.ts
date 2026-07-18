import { CommonModule } from '@angular/common';
import { Component, Input } from '@angular/core';

@Component({
  selector: 'app-capacity-indicator',
  imports: [CommonModule],
  template: `
    <div class="flex items-center gap-3">
      <div
        class="h-2 flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-800"
        role="progressbar"
        [attr.aria-valuenow]="current"
        [attr.aria-valuemax]="max"
        [attr.aria-label]="ariaLabel"
      >
        <div
          class="h-full rounded-full transition-all duration-300"
          [ngClass]="barColorClass"
          [style.width.%]="percentage"
        ></div>
      </div>
      <span class="text-xs font-medium whitespace-nowrap" [ngClass]="textColorClass">
        {{ current }}/{{ max }}
      </span>
    </div>
  `,
})
export class CapacityIndicatorComponent {
  @Input() current = 0;
  @Input() max = 0;
  @Input() ariaLabel = 'Capacity';

  get percentage(): number {
    if (this.max <= 0) return 0;
    return Math.min(100, (this.current / this.max) * 100);
  }

  get statusLevel(): 'green' | 'amber' | 'red' {
    if (this.max <= 0) return 'green';
    const ratio = this.current / this.max;
    if (ratio >= 1) return 'red';
    if (ratio >= 0.8) return 'amber';
    return 'green';
  }

  get barColorClass(): string {
    const colors: Record<string, string> = {
      green: 'bg-success-500',
      amber: 'bg-warning-500',
      red: 'bg-error-500',
    };
    return colors[this.statusLevel];
  }

  get textColorClass(): string {
    const colors: Record<string, string> = {
      green: 'text-success-600 dark:text-success-500',
      amber: 'text-warning-600 dark:text-warning-500',
      red: 'text-error-600 dark:text-error-500',
    };
    return colors[this.statusLevel];
  }
}
