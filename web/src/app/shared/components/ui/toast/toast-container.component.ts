import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { ToastService, Toast, ToastVariant } from '../../../services/toast.service';

@Component({
  selector: 'app-toast-container',
  imports: [CommonModule],
  template: `
    <div
      class="fixed right-4 top-4 z-[100000] flex flex-col gap-2"
      aria-live="polite"
      aria-atomic="false"
    >
      @for (toast of toastService.state.toasts(); track toast.id) {
        <div
          class="flex w-80 items-start gap-3 rounded-xl border p-4 shadow-theme-md"
          [ngClass]="containerClasses(toast.variant)"
          [attr.role]="roleForVariant(toast.variant)"
        >
          <div class="flex-shrink-0" [innerHTML]="iconForVariant(toast.variant)"></div>
          <div class="min-w-0 flex-1">
            @if (toast.title) {
              <p class="text-sm font-semibold text-gray-800 dark:text-white/90">{{ toast.title }}</p>
            }
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ toast.message }}</p>
          </div>
          <button
            type="button"
            class="flex-shrink-0 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            (click)="toastService.dismiss(toast.id)"
            aria-label="Dismiss"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
              <path fillRule="evenodd" clipRule="evenodd" d="M6.04289 16.5413C5.65237 16.9318 5.65237 17.565 6.04289 17.9555C6.43342 18.346 7.06658 18.346 7.45711 17.9555L11.9987 13.4139L16.5408 17.956C16.9313 18.3466 17.5645 18.3466 17.955 17.956C18.3455 17.5655 18.3455 16.9323 17.955 16.5418L13.4129 11.9997L17.955 7.4576C18.3455 7.06707 18.3455 6.43391 17.955 6.04338C17.5645 5.65286 16.9313 5.65286 16.5408 6.04338L11.9987 10.5855L7.45711 6.0439C7.06658 5.65338 6.43342 5.65338 6.04289 6.0439C5.65237 6.43442 5.65237 7.06759 6.04289 7.45811L10.5845 11.9997L6.04289 16.5413Z" fill="currentColor"/>
            </svg>
          </button>
        </div>
      }
    </div>
  `,
})
export class ToastContainerComponent {
  readonly toastService = inject(ToastService);

  containerClasses(variant: ToastVariant): string {
    const map: Record<ToastVariant, string> = {
      success: 'border-success-500 bg-success-50 dark:border-success-500/30 dark:bg-success-500/15',
      error: 'border-error-500 bg-error-50 dark:border-error-500/30 dark:bg-error-500/15',
      warning: 'border-warning-500 bg-warning-50 dark:border-warning-500/30 dark:bg-warning-500/15',
      info: 'border-blue-light-500 bg-blue-light-50 dark:border-blue-light-500/30 dark:bg-blue-light-500/15',
    };
    return map[variant];
  }

  roleForVariant(variant: ToastVariant): string {
    return variant === 'error' || variant === 'warning' ? 'alert' : 'status';
  }

  iconForVariant(variant: ToastVariant): string {
    const colorMap: Record<ToastVariant, string> = {
      success: 'text-success-500',
      error: 'text-error-500',
      warning: 'text-warning-500',
      info: 'text-blue-light-500',
    };
    return `<svg class="${colorMap[variant]}" width="20" height="20" viewBox="0 0 24 24" fill="none">
      ${variant === 'success' ? '<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none"/><path d="M8 12l3 3 5-5" stroke="currentColor" stroke-width="2" fill="none"/>' : ''}
      ${variant === 'error' ? '<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none"/><path d="M8 8l8 8M16 8l-8 8" stroke="currentColor" stroke-width="2"/>' : ''}
      ${variant === 'warning' ? '<path d="M12 2L2 20h20L12 2z" stroke="currentColor" stroke-width="2" fill="none"/><path d="M12 9v4M12 16h.01" stroke="currentColor" stroke-width="2"/>' : ''}
      ${variant === 'info' ? '<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2" fill="none"/><path d="M12 8h.01M12 11v5" stroke="currentColor" stroke-width="2"/>' : ''}
    </svg>`;
  }
}
