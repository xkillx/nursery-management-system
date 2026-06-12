import { CommonModule } from '@angular/common';
import { Component, EventEmitter, forwardRef, Input, Output } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

@Component({
  selector: 'app-text-area',
  imports: [CommonModule],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => TextAreaComponent),
      multi: true,
    },
  ],
  template: `
    <div class="relative">
      <textarea
        [id]="id"
        [name]="name"
        [placeholder]="placeholder"
        [rows]="rows"
        [value]="value"
        [disabled]="disabled"
        [attr.autocomplete]="autocomplete || null"
        [attr.aria-invalid]="error"
        [attr.aria-describedby]="describedBy || null"
        (input)="onInput($event)"
        (blur)="markTouched()"
        [ngClass]="textareaClasses"
      ></textarea>
      @if (hint) {
        <p
          class="mt-2 text-sm"
          [ngClass]="error ? 'text-error-500' : 'text-gray-500 dark:text-gray-400'">
          {{ hint }}
        </p>
      }
    </div>
  `,
  styles: ``,
})
export class TextAreaComponent implements ControlValueAccessor {

  @Input() id?: string = '';
  @Input() name?: string = '';
  @Input() placeholder = 'Enter your message';
  @Input() rows = 3;
  @Input() value = '';
  @Input() className = '';
  @Input() disabled = false;
  @Input() error = false;
  @Input() hint = '';
  @Input() describedBy?: string;
  @Input() autocomplete?: string;

  @Output() valueChange = new EventEmitter<string>();
  @Output() blurred = new EventEmitter<void>();

  private onChange: (value: string) => void = () => {};
  private onTouched: () => void = () => {};

  onInput(event: Event) {
    const val = (event.target as HTMLTextAreaElement).value;
    this.value = val;
    this.onChange(val);
    this.valueChange.emit(val);
  }

  markTouched() {
    this.onTouched();
    this.blurred.emit();
  }

  writeValue(value: string | null): void {
    this.value = value ?? '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }

  get textareaClasses(): string {
    let base = `w-full rounded-xl border px-4 py-3 text-sm shadow-theme-xs focus:outline-hidden min-h-[120px] ${this.className} `;
    if (this.disabled) {
      base += 'bg-gray-100 opacity-50 text-gray-500 border-gray-300 cursor-not-allowed dark:bg-gray-800 dark:text-gray-400 dark:border-gray-700';
    } else if (this.error) {
      base += 'bg-transparent border-error-500 focus:border-error-300 focus:ring-3 focus:ring-error-500/10 dark:border-gray-700 dark:bg-gray-900 dark:text-white/90 dark:focus:border-error-800';
    } else {
      base += 'bg-transparent text-gray-900 dark:text-gray-300 border-gray-300 focus:border-brand-300 focus:ring-3 focus:ring-brand-500/10 dark:border-gray-700 dark:bg-gray-900 dark:text-white/90 dark:focus:border-brand-800';
    }
    return base;
  }
}
