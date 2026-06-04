import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, Output } from '@angular/core';
import { ModalComponent } from './modal.component';

@Component({
  selector: 'app-confirmation-dialog',
  imports: [CommonModule, ModalComponent],
  template: `
    <app-modal
      [isOpen]="isOpen"
      (close)="onCancel()"
      [ariaLabel]="title"
      className="max-w-md p-6"
      [closeOnBackdrop]="true"
      [closeOnEscape]="true"
    >
      <div>
        <h3 class="text-lg font-semibold text-gray-800 dark:text-white/90">{{ title }}</h3>
        @if (message) {
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ message }}</p>
        }
        <ng-content></ng-content>
        <div class="mt-6 flex items-center justify-end gap-3">
          <button
            type="button"
            class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-white/5"
            (click)="onCancel()"
          >
            {{ cancelText }}
          </button>
          <button
            type="button"
            [disabled]="loading || disabled"
            [ngClass]="confirmButtonClasses"
            (click)="onConfirm()"
          >
            @if (loading) {
              <svg class="mr-2 h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
            }
            {{ confirmText }}
          </button>
        </div>
      </div>
    </app-modal>
  `,
})
export class ConfirmationDialogComponent {
  @Input() isOpen = false;
  @Input() title = 'Confirm';
  @Input() message = '';
  @Input() confirmText = 'Confirm';
  @Input() cancelText = 'Cancel';
  @Input() variant: 'primary' | 'danger' = 'primary';
  @Input() loading = false;
  @Input() disabled = false;

  @Output() confirmed = new EventEmitter<void>();
  @Output() cancelled = new EventEmitter<void>();

  get confirmButtonClasses(): string {
    const base = 'inline-flex items-center rounded-lg px-4 py-2 text-sm font-medium text-white shadow-theme-xs disabled:opacity-50';
    if (this.variant === 'danger') {
      return `${base} bg-error-500 hover:bg-error-600`;
    }
    return `${base} bg-brand-500 hover:bg-brand-600`;
  }

  onConfirm() {
    if (!this.loading && !this.disabled) {
      this.confirmed.emit();
    }
  }

  onCancel() {
    if (!this.loading) {
      this.cancelled.emit();
    }
  }
}
