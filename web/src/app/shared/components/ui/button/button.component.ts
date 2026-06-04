import { CommonModule } from '@angular/common';
import { Component, Input, Output, EventEmitter } from '@angular/core';
import { SafeHtmlPipe } from '../../../pipe/safe-html.pipe';

type ButtonSize = 'xs' | 'sm' | 'md' | 'lg';
type ButtonVariant =
  | 'primary'
  | 'outline'
  | 'secondary'
  | 'success'
  | 'warning'
  | 'danger'
  | 'ghost'
  | 'link';

@Component({
  selector: 'app-button',
  imports: [
    CommonModule,
    SafeHtmlPipe,
  ],
  templateUrl: './button.component.html',
  styles: ``,
})
export class ButtonComponent {

  @Input() size: ButtonSize = 'md';
  @Input() variant: ButtonVariant = 'primary';
  @Input() type: 'button' | 'submit' | 'reset' = 'button';
  @Input() disabled = false;
  @Input() loading = false;
  @Input() className = '';
  @Input() startIcon?: string;
  @Input() endIcon?: string;

  @Output() btnClick = new EventEmitter<Event>();

  get isInteractive(): boolean {
    return !this.disabled && !this.loading;
  }

  get sizeClasses(): string {
    const map: Record<ButtonSize, string> = {
      xs: 'px-2.5 py-1.5 text-xs',
      sm: 'px-4 py-3 text-sm',
      md: 'px-5 py-3.5 text-sm',
      lg: 'px-6 py-4 text-base',
    };
    return map[this.size];
  }

  get variantClasses(): string {
    const map: Record<ButtonVariant, string> = {
      primary: 'bg-brand-500 text-white shadow-theme-xs hover:bg-brand-600 disabled:bg-brand-300',
      outline: 'bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700 dark:hover:bg-white/[0.03] dark:hover:text-gray-300',
      secondary: 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700',
      success: 'bg-success-500 text-white shadow-theme-xs hover:bg-success-600 disabled:bg-success-300',
      warning: 'bg-warning-500 text-white shadow-theme-xs hover:bg-warning-600 disabled:bg-warning-300',
      danger: 'bg-error-500 text-white shadow-theme-xs hover:bg-error-600 disabled:bg-error-300',
      ghost: 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-white/[0.05]',
      link: 'text-brand-500 underline hover:text-brand-600 dark:text-brand-400',
    };
    return map[this.variant];
  }

  get disabledClasses(): string {
    if (this.loading) return 'cursor-wait opacity-70';
    if (this.disabled) return 'cursor-not-allowed opacity-50';
    return '';
  }

  onClick(event: Event) {
    if (this.isInteractive) {
      this.btnClick.emit(event);
    }
  }
}
