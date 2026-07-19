import { CommonModule } from '@angular/common';
import { Component, forwardRef, input } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

export interface RadioCardOption {
  value: string;
  label: string;
  description?: string;
}

@Component({
  selector: 'app-radio-card-group',
  imports: [CommonModule],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => RadioCardGroupComponent),
      multi: true,
    },
  ],
  template: `
    <div class="flex flex-col sm:flex-row gap-4">
      @for (opt of options(); track opt.value) {
        <label
          class="relative flex flex-1 cursor-pointer select-none items-center gap-4 rounded-xl border p-4 transition-all shadow-theme-xs"
          [ngClass]="value === opt.value
            ? 'border-brand-500 bg-white dark:bg-gray-900/10 text-gray-900 dark:text-white'
            : 'border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900/10 text-gray-700 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-700'"
        >
          <input
            type="radio"
            [name]="name()"
            [value]="opt.value"
            [checked]="value === opt.value"
            (change)="select(opt.value)"
            class="sr-only"
          />
          <span
            class="flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-[1.25px] transition-all"
            [ngClass]="value === opt.value
              ? 'border-brand-500 bg-brand-500'
              : 'bg-transparent border-gray-400 dark:border-gray-600'"
          >
            <span
              class="h-2 w-2 rounded-full bg-white"
              [ngClass]="value === opt.value ? 'block' : 'hidden'"
            ></span>
          </span>
          <span class="flex flex-col">
            <span class="text-sm font-semibold text-gray-900 dark:text-white leading-tight">{{ opt.label }}</span>
            @if (opt.description) {
              <span class="mt-1 text-xs text-gray-500 dark:text-gray-400 leading-normal">{{ opt.description }}</span>
            }
          </span>
        </label>
      }
    </div>
  `,
})
export class RadioCardGroupComponent implements ControlValueAccessor {
  options = input.required<RadioCardOption[]>();
  name = input('radio-card');

  value = '';

  private propagateChange: (value: string) => void = () => { /* Set via registerOnChange */ };
  private propagateTouched: () => void = () => { /* Set via registerOnTouched */ };

  select(value: string): void {
    this.value = value;
    this.propagateChange(value);
    this.propagateTouched();
  }

  writeValue(value: string): void {
    this.value = value ?? '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.propagateChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.propagateTouched = fn;
  }
}
