import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { ToastService, ToastVariant } from '../../../services/toast.service';

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
          <div class="-mt-0.5 flex-shrink-0" [ngClass]="iconColorClass(toast.variant)">
            <svg class="fill-current" width="24" height="24" viewBox="0 0 24 24" fill="none">
              @switch (toast.variant) {
                @case ('error') {
                  <path fill-rule="evenodd" clip-rule="evenodd" d="M20.3499 12.0004C20.3499 16.612 16.6115 20.3504 11.9999 20.3504C7.38832 20.3504 3.6499 16.612 3.6499 12.0004C3.6499 7.38881 7.38833 3.65039 11.9999 3.65039C16.6115 3.65039 20.3499 7.38881 20.3499 12.0004ZM11.9999 22.1504C17.6056 22.1504 22.1499 17.6061 22.1499 12.0004C22.1499 6.3947 17.6056 1.85039 11.9999 1.85039C6.39421 1.85039 1.8499 6.3947 1.8499 12.0004C1.8499 17.6061 6.39421 22.1504 11.9999 22.1504ZM13.0008 16.4753C13.0008 15.923 12.5531 15.4753 12.0008 15.4753L11.9998 15.4753C11.4475 15.4753 10.9998 15.923 10.9998 16.4753C10.9998 17.0276 11.4475 17.4753 11.9998 17.4753L12.0008 17.4753C12.5531 17.4753 13.0008 17.0276 13.0008 16.4753ZM11.9998 6.62898C12.414 6.62898 12.7498 6.96476 12.7498 7.37898L12.7498 13.0555C12.7498 13.4697 12.414 13.8055 11.9998 13.8055C11.5856 13.8055 11.2498 13.4697 11.2498 13.0555L11.2498 7.37898C11.2498 6.96476 11.5856 6.62898 11.9998 6.62898Z"/>
                }
                @case ('success') {
                  <path fill-rule="evenodd" clip-rule="evenodd" d="M20.3499 12.0004C20.3499 16.612 16.6115 20.3504 11.9999 20.3504C7.38832 20.3504 3.6499 16.612 3.6499 12.0004C3.6499 7.38881 7.38833 3.65039 11.9999 3.65039C16.6115 3.65039 20.3499 7.38881 20.3499 12.0004ZM11.9999 22.1504C17.6056 22.1504 22.1499 17.6061 22.1499 12.0004C22.1499 6.3947 17.6056 1.85039 11.9999 1.85039C6.39421 1.85039 1.8499 6.3947 1.8499 12.0004C1.8499 17.6061 6.39421 22.1504 11.9999 22.1504ZM16.3515 9.1257C16.742 8.73518 16.742 8.10201 16.3515 7.71149C15.9609 7.32097 15.3278 7.32097 14.9372 7.71149L10.5845 12.0642L9.05711 10.5368C8.66658 10.1463 8.03342 10.1463 7.64289 10.5368C7.25237 10.9273 7.25237 11.5605 7.64289 11.951L10.2309 14.539C10.4175 14.7256 10.6718 14.831 10.9372 14.831C11.2026 14.831 11.457 14.7256 11.6436 14.539L16.3515 9.1257Z"/>
                }
                @case ('warning') {
                  <path fill-rule="evenodd" clip-rule="evenodd" d="M20.3499 12.0004C20.3499 16.612 16.6115 20.3504 11.9999 20.3504C7.38832 20.3504 3.6499 16.612 3.6499 12.0004C3.6499 7.38881 7.38833 3.65039 11.9999 3.65039C16.6115 3.65039 20.3499 7.38881 20.3499 12.0004ZM11.9999 22.1504C17.6056 22.1504 22.1499 17.6061 22.1499 12.0004C22.1499 6.3947 17.6056 1.85039 11.9999 1.85039C6.39421 1.85039 1.8499 6.3947 1.8499 12.0004C1.8499 17.6061 6.39421 22.1504 11.9999 22.1504ZM13.0008 16.4753C13.0008 15.923 12.5531 15.4753 12.0008 15.4753L11.9998 15.4753C11.4475 15.4753 10.9998 15.923 10.9998 16.4753C10.9998 17.0276 11.4475 17.4753 11.9998 17.4753L12.0008 17.4753C12.5531 17.4753 13.0008 17.0276 13.0008 16.4753ZM11.9998 6.62898C12.414 6.62898 12.7498 6.96476 12.7498 7.37898L12.7498 13.0555C12.7498 13.4697 12.414 13.8055 11.9998 13.8055C11.5856 13.8055 11.2498 13.4697 11.2498 13.0555L11.2498 7.37898C11.2498 6.96476 11.5856 6.62898 11.9998 6.62898Z"/>
                }
                @case ('info') {
                  <path fill-rule="evenodd" clip-rule="evenodd" d="M20.3499 12.0004C20.3499 16.612 16.6115 20.3504 11.9999 20.3504C7.38832 20.3504 3.6499 16.612 3.6499 12.0004C3.6499 7.38881 7.38833 3.65039 11.9999 3.65039C16.6115 3.65039 20.3499 7.38881 20.3499 12.0004ZM11.9999 22.1504C17.6056 22.1504 22.1499 17.6061 22.1499 12.0004C22.1499 6.3947 17.6056 1.85039 11.9999 1.85039C6.39421 1.85039 1.8499 6.3947 1.8499 12.0004C1.8499 17.6061 6.39421 22.1504 11.9999 22.1504ZM11.9998 6.62898C12.414 6.62898 12.7498 6.96476 12.7498 7.37898L12.7498 13.0555C12.7498 13.4697 12.414 13.8055 11.9998 13.8055C11.5856 13.8055 11.2498 13.4697 11.2498 13.0555L11.2498 7.37898C11.2498 6.96476 11.5856 6.62898 11.9998 6.62898ZM13.0008 16.4753C13.0008 15.923 12.5531 15.4753 12.0008 15.4753L11.9998 15.4753C11.4475 15.4753 10.9998 15.923 10.9998 16.4753C10.9998 17.0276 11.4475 17.4753 11.9998 17.4753L12.0008 17.4753C12.5531 17.4753 13.0008 17.0276 13.0008 16.4753Z"/>
                }
              }
            </svg>
          </div>
          <div class="min-w-0 flex-1">
            @if (toast.title) {
              <p class="text-sm font-semibold text-gray-800 dark:text-white/90">{{ toast.title }}</p>
            }
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ toast.message }}</p>
          </div>
          <button
            type="button"
            aria-label="Dismiss"
            class="flex-shrink-0 ml-2 -mr-1 -mt-1 rounded-lg p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:text-gray-500 dark:hover:text-gray-300 dark:hover:bg-white/10 focus:outline-hidden focus:ring-2 focus:ring-gray-400/50"
            (click)="toastService.dismiss(toast.id)"
          >
            <svg fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="h-4 w-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"></path>
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

  iconColorClass(variant: ToastVariant): string {
    const map: Record<ToastVariant, string> = {
      success: 'text-success-500',
      error: 'text-error-500',
      warning: 'text-warning-500',
      info: 'text-blue-light-500',
    };
    return map[variant];
  }
}
